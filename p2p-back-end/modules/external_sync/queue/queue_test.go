package queue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"p2p-back-end/logs"
)

func TestMain(m *testing.M) {
	logs.Loginit()
	os.Exit(m.Run())
}

func setupRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(func() { mr.Close() })
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb, mr
}

func newJob(jt, year string, months ...string) *Job {
	return &Job{
		JobType:     jt,
		Year:        year,
		Months:      months,
		TriggeredBy: "TEST",
	}
}

// ─────────────────────────── PriorityFor ───────────────────────────

func TestPriorityFor(t *testing.T) {
	cases := []struct {
		jobType  string
		expected int
	}{
		{"DW_SYNC", 1},
		{"TIER2_FULL", 2},
		{"ACTUAL_FACT", 3},
		{"MANUAL", 3},
		{"TIER1_FAST", 5},
		{"UNKNOWN", 3}, // default
		{"", 3},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, PriorityFor(c.jobType), "PriorityFor(%q)", c.jobType)
	}
}

func TestMakeScore_PriorityDominates(t *testing.T) {
	// A high-priority job enqueued LATER must still have a lower score than
	// a low-priority job enqueued EARLIER.
	earlyTier1 := makeScore(5, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	lateDW := makeScore(1, time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC))
	assert.Less(t, lateDW, earlyTier1, "DW (P1) should run before old Tier 1 (P5)")
}

func TestMakeScore_FIFOWithinPriority(t *testing.T) {
	first := makeScore(3, time.Now())
	time.Sleep(2 * time.Millisecond)
	second := makeScore(3, time.Now())
	assert.Less(t, first, second, "earlier enqueue should have lower score")
}

// ─────────────────────────── DedupKey ───────────────────────────

func TestDedupKey_StableAcrossMonthOrder(t *testing.T) {
	a := newJob("ACTUAL_FACT", "2026", "JAN", "FEB", "MAR")
	b := newJob("ACTUAL_FACT", "2026", "MAR", "JAN", "FEB")
	assert.Equal(t, a.DedupKey(), b.DedupKey(), "month order should not affect dedup key")
}

func TestDedupKey_DifferentForDifferentJobs(t *testing.T) {
	a := newJob("ACTUAL_FACT", "2026", "APR")
	b := newJob("ACTUAL_FACT", "2026", "MAY")
	c := newJob("DW_SYNC", "2026")
	d := newJob("ACTUAL_FACT", "2025", "APR")

	keys := []string{a.DedupKey(), b.DedupKey(), c.DedupKey(), d.DedupKey()}
	seen := map[string]bool{}
	for _, k := range keys {
		assert.False(t, seen[k], "dedup keys should be unique across distinct jobs")
		seen[k] = true
	}
}

// ─────────────────────────── Enqueue + dedup ───────────────────────────

func TestEnqueue_AssignsIDAndTimestamp(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)

	j := newJob("ACTUAL_FACT", "2026", "APR")
	enq, err := q.Enqueue(context.Background(), j)
	require.NoError(t, err)
	assert.True(t, enq)
	assert.NotEqual(t, uuid.Nil, j.ID)
	assert.False(t, j.EnqueuedAt.IsZero())
	assert.Equal(t, JobStatusPending, j.Status)
}

func TestEnqueue_DedupsIdenticalJob(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	first, err := q.Enqueue(ctx, newJob("ACTUAL_FACT", "2026", "APR"))
	require.NoError(t, err)
	assert.True(t, first, "first enqueue should succeed")

	second, err := q.Enqueue(ctx, newJob("ACTUAL_FACT", "2026", "APR"))
	require.NoError(t, err)
	assert.False(t, second, "duplicate enqueue should be rejected")

	pending, err := q.GetPending(ctx)
	require.NoError(t, err)
	assert.Len(t, pending, 1, "only one job should be pending")
}

func TestEnqueue_DifferentMonthsAreNotDuplicates(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	a, err := q.Enqueue(ctx, newJob("ACTUAL_FACT", "2026", "APR"))
	require.NoError(t, err)
	assert.True(t, a)

	b, err := q.Enqueue(ctx, newJob("ACTUAL_FACT", "2026", "MAY"))
	require.NoError(t, err)
	assert.True(t, b)

	pending, err := q.GetPending(ctx)
	require.NoError(t, err)
	assert.Len(t, pending, 2)
}

// ─────────────────────────── Priority ordering ───────────────────────────

func TestGetPending_OrdersByPriority(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	// Enqueue in REVERSE priority order to prove that ordering is by score, not insertion order.
	tier1Job := newJob("TIER1_FAST", "2026", "APR")
	tier2Job := newJob("TIER2_FULL", "2026")
	manualJob := newJob("MANUAL", "2026", "APR")
	dwJob := newJob("DW_SYNC", "2026")

	for _, j := range []*Job{tier1Job, manualJob, tier2Job, dwJob} {
		_, err := q.Enqueue(ctx, j)
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond) // ensure distinct enqueue times
	}

	pending, err := q.GetPending(ctx)
	require.NoError(t, err)
	require.Len(t, pending, 4)

	assert.Equal(t, "DW_SYNC", pending[0].JobType, "P1 first")
	assert.Equal(t, "TIER2_FULL", pending[1].JobType, "P2 second")
	assert.Equal(t, "MANUAL", pending[2].JobType, "P3 third")
	assert.Equal(t, "TIER1_FAST", pending[3].JobType, "P5 last")
}

func TestPopNext_ReturnsHighestPriorityFirst(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	_, _ = q.Enqueue(ctx, newJob("TIER1_FAST", "2026", "APR"))
	time.Sleep(2 * time.Millisecond)
	_, _ = q.Enqueue(ctx, newJob("DW_SYNC", "2026"))
	time.Sleep(2 * time.Millisecond)
	_, _ = q.Enqueue(ctx, newJob("TIER2_FULL", "2026"))

	first, err := q.PopNext(ctx)
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, "DW_SYNC", first.JobType)

	second, err := q.PopNext(ctx)
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.Equal(t, "TIER2_FULL", second.JobType)

	third, err := q.PopNext(ctx)
	require.NoError(t, err)
	require.NotNil(t, third)
	assert.Equal(t, "TIER1_FAST", third.JobType)
}

func TestPopNext_FIFOWithinSamePriority(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	a := newJob("ACTUAL_FACT", "2026", "JAN")
	_, err := q.Enqueue(ctx, a)
	require.NoError(t, err)
	time.Sleep(3 * time.Millisecond)
	b := newJob("ACTUAL_FACT", "2026", "FEB")
	_, err = q.Enqueue(ctx, b)
	require.NoError(t, err)

	first, err := q.PopNext(ctx)
	require.NoError(t, err)
	assert.Equal(t, a.ID, first.ID, "earlier enqueue runs first")
}

// ─────────────────────────── Cancel + Promote ───────────────────────────

func TestCancel_RemovesFromQueueAndDedup(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	j := newJob("ACTUAL_FACT", "2026", "APR")
	_, _ = q.Enqueue(ctx, j)

	require.NoError(t, q.Cancel(ctx, j.ID))

	pending, _ := q.GetPending(ctx)
	assert.Empty(t, pending)

	// After cancel, identical job can be re-enqueued (dedup released)
	again, err := q.Enqueue(ctx, newJob("ACTUAL_FACT", "2026", "APR"))
	require.NoError(t, err)
	assert.True(t, again, "dedup key should be released after cancel")
}

func TestCancel_NonExistentJob(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	err := q.Cancel(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestPromote_MovesJobToFront(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	dwJob := newJob("DW_SYNC", "2026")
	tier2Job := newJob("TIER2_FULL", "2026")
	manualJob := newJob("MANUAL", "2026", "APR")

	_, _ = q.Enqueue(ctx, dwJob)
	time.Sleep(2 * time.Millisecond)
	_, _ = q.Enqueue(ctx, tier2Job)
	time.Sleep(2 * time.Millisecond)
	_, _ = q.Enqueue(ctx, manualJob)

	require.NoError(t, q.Promote(ctx, manualJob.ID))

	pending, err := q.GetPending(ctx)
	require.NoError(t, err)
	require.Len(t, pending, 3)
	assert.Equal(t, manualJob.ID, pending[0].ID, "promoted job should be at the front")
}

func TestPromote_NonExistentJob(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	err := q.Promote(context.Background(), uuid.New())
	assert.Error(t, err)
}

// ─────────────────────────── IsBusy ───────────────────────────

func TestIsBusy_FalseWhenIdle(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	busy, err := q.IsBusy(context.Background())
	require.NoError(t, err)
	assert.False(t, busy)
}

func TestIsBusy_TrueWhenPendingExists(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	_, _ = q.Enqueue(ctx, newJob("ACTUAL_FACT", "2026", "APR"))

	busy, err := q.IsBusy(ctx)
	require.NoError(t, err)
	assert.True(t, busy)
}

func TestIsBusy_TrueWhenJobIsRunning(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	j := newJob("ACTUAL_FACT", "2026", "APR")
	_, _ = q.Enqueue(ctx, j)
	popped, _ := q.PopNext(ctx)
	require.NotNil(t, popped)
	require.NoError(t, q.SetCurrent(ctx, popped))

	busy, err := q.IsBusy(ctx)
	require.NoError(t, err)
	assert.True(t, busy, "current set should mark queue as busy")
}

// ─────────────────────────── SetCurrent + ClearCurrent + Recover ───────────────────────────

func TestSetCurrent_UpdatesStatusAndStartedAt(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	j := newJob("ACTUAL_FACT", "2026", "APR")
	_, _ = q.Enqueue(ctx, j)
	popped, _ := q.PopNext(ctx)
	require.NoError(t, q.SetCurrent(ctx, popped))

	cur, err := q.GetCurrent(ctx)
	require.NoError(t, err)
	require.NotNil(t, cur)
	assert.Equal(t, JobStatusRunning, cur.Status)
	assert.NotNil(t, cur.StartedAt)
}

func TestClearCurrent_RemovesAllTraces(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	j := newJob("ACTUAL_FACT", "2026", "APR")
	_, _ = q.Enqueue(ctx, j)
	popped, _ := q.PopNext(ctx)
	_ = q.SetCurrent(ctx, popped)

	require.NoError(t, q.ClearCurrent(ctx, popped))

	cur, err := q.GetCurrent(ctx)
	require.NoError(t, err)
	assert.Nil(t, cur)

	// Same business key can be enqueued again — dedup released
	again, err := q.Enqueue(ctx, newJob("ACTUAL_FACT", "2026", "APR"))
	require.NoError(t, err)
	assert.True(t, again)
}

func TestRecover_RequeuesOrphanedCurrent(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	// Simulate worker crash: a job is "current" but never cleared.
	j := newJob("ACTUAL_FACT", "2026", "APR")
	_, _ = q.Enqueue(ctx, j)
	popped, _ := q.PopNext(ctx)
	_ = q.SetCurrent(ctx, popped)

	// At this point: queue is empty, current is set.
	pendingBefore, _ := q.GetPending(ctx)
	assert.Empty(t, pendingBefore)

	recovered, err := q.Recover(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, recovered)

	pendingAfter, _ := q.GetPending(ctx)
	require.Len(t, pendingAfter, 1, "orphaned job should be re-queued")
	assert.Equal(t, popped.ID, pendingAfter[0].ID)

	cur, _ := q.GetCurrent(ctx)
	assert.Nil(t, cur, "current should be cleared after recover")
}

func TestRecover_NoOpWhenNothingRunning(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)

	recovered, err := q.Recover(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, recovered)
}

// ─────────────────────────── End-to-end queue lifecycle ───────────────────────────

func TestEndToEnd_PriorityCancelPromote(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	ctx := context.Background()

	// Enqueue 4 jobs out of order.
	tier1 := newJob("TIER1_FAST", "2026", "APR")
	manual := newJob("MANUAL", "2026", "MAY")
	dw := newJob("DW_SYNC", "2026")
	tier2 := newJob("TIER2_FULL", "2026")

	_, _ = q.Enqueue(ctx, tier1)
	time.Sleep(2 * time.Millisecond)
	_, _ = q.Enqueue(ctx, manual)
	time.Sleep(2 * time.Millisecond)
	_, _ = q.Enqueue(ctx, dw)
	time.Sleep(2 * time.Millisecond)
	_, _ = q.Enqueue(ctx, tier2)

	// Cancel Tier 1.
	require.NoError(t, q.Cancel(ctx, tier1.ID))
	pending, _ := q.GetPending(ctx)
	require.Len(t, pending, 3)
	for _, p := range pending {
		assert.NotEqual(t, tier1.ID, p.ID)
	}

	// Promote MANUAL to the front (override priority).
	require.NoError(t, q.Promote(ctx, manual.ID))
	pending, _ = q.GetPending(ctx)
	require.Len(t, pending, 3)
	assert.Equal(t, manual.ID, pending[0].ID, "promoted job runs next")
	// DW + TIER2 still ordered by their priorities (DW < TIER2)
	assert.Equal(t, dw.ID, pending[1].ID)
	assert.Equal(t, tier2.ID, pending[2].ID)
}

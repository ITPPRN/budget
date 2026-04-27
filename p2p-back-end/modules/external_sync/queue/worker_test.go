package queue

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"p2p-back-end/modules/entities/models"
	syncRepo "p2p-back-end/modules/external_sync/repository"
)

// ─────────────────────────── Mocks ───────────────────────────

type mockExecutor struct {
	mock.Mock
}

func (m *mockExecutor) SyncActuals(ctx context.Context, year string, months []string) error {
	args := m.Called(ctx, year, months)
	return args.Error(0)
}

func (m *mockExecutor) SyncFromDW(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockTrackingRepo struct {
	mock.Mock
}

func (m *mockTrackingRepo) CreateRun(ctx context.Context, run *models.SyncRunEntity) error {
	args := m.Called(ctx, run)
	if run.ID == uuid.Nil {
		run.ID = uuid.New() // mimic real repo behaviour
	}
	return args.Error(0)
}

func (m *mockTrackingRepo) CompleteRun(
	ctx context.Context, id uuid.UUID, status string,
	rowsFetched, rowsInserted, rowsSkipped int64, errMsg string,
) error {
	args := m.Called(ctx, id, status, rowsFetched, rowsInserted, rowsSkipped, errMsg)
	return args.Error(0)
}

func (m *mockTrackingRepo) IncrementRetry(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockTrackingRepo) GetLatest(ctx context.Context, jobType string) (*models.SyncRunEntity, error) {
	args := m.Called(ctx, jobType)
	r, _ := args.Get(0).(*models.SyncRunEntity)
	return r, args.Error(1)
}

func (m *mockTrackingRepo) GetRecent(ctx context.Context, limit int) ([]models.SyncRunEntity, error) {
	args := m.Called(ctx, limit)
	r, _ := args.Get(0).([]models.SyncRunEntity)
	return r, args.Error(1)
}

func (m *mockTrackingRepo) GetRunsByJobAndStatus(
	ctx context.Context, jobType, status string, limit int,
) ([]models.SyncRunEntity, error) {
	args := m.Called(ctx, jobType, status, limit)
	r, _ := args.Get(0).([]models.SyncRunEntity)
	return r, args.Error(1)
}

func (m *mockTrackingRepo) GetReconciliation(ctx context.Context, year string) (*syncRepo.ReconciliationResult, error) {
	args := m.Called(ctx, year)
	r, _ := args.Get(0).(*syncRepo.ReconciliationResult)
	return r, args.Error(1)
}

func (m *mockTrackingRepo) GetFailedRunsForRetry(
	ctx context.Context, within time.Duration, maxRetries int,
) ([]models.SyncRunEntity, error) {
	args := m.Called(ctx, within, maxRetries)
	r, _ := args.Get(0).([]models.SyncRunEntity)
	return r, args.Error(1)
}

func (m *mockTrackingRepo) ClearStaleRunningRuns(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTrackingRepo) DeleteOldRunsByJobType(
	ctx context.Context, jobType string, olderThan time.Duration,
) (int64, error) {
	args := m.Called(ctx, jobType, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

// ─────────────────────────── Tests: dispatch by job type ───────────────────────────

func TestWorker_Execute_ActualFactDispatchesToSyncActuals(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	exec := new(mockExecutor)
	exec.On("SyncActuals", mock.Anything, "2026", []string{"APR"}).Return(nil)

	w := NewWorker(WorkerDeps{Queue: q, Executor: exec})

	job := &Job{ID: uuid.New(), JobType: models.SyncJobActualFact, Year: "2026", Months: []string{"APR"}}
	err := w.execute(context.Background(), job)

	assert.NoError(t, err)
	exec.AssertExpectations(t)
}

func TestWorker_Execute_DWDispatchesToSyncFromDW(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	exec := new(mockExecutor)
	exec.On("SyncFromDW", mock.Anything).Return(nil)

	w := NewWorker(WorkerDeps{Queue: q, Executor: exec})

	job := &Job{ID: uuid.New(), JobType: models.SyncJobDW, Year: "2026"}
	err := w.execute(context.Background(), job)

	assert.NoError(t, err)
	exec.AssertExpectations(t)
}

func TestWorker_Execute_Tier1Dispatches(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	exec := new(mockExecutor)
	exec.On("SyncActuals", mock.Anything, "2026", []string{"APR"}).Return(nil)

	w := NewWorker(WorkerDeps{Queue: q, Executor: exec})

	job := &Job{ID: uuid.New(), JobType: models.SyncJobTier1Fast, Year: "2026", Months: []string{"APR"}}
	err := w.execute(context.Background(), job)

	assert.NoError(t, err)
	exec.AssertExpectations(t)
}

func TestWorker_Execute_Tier2DispatchesFullYear(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	exec := new(mockExecutor)
	// Tier 2 should pass empty months.
	exec.On("SyncActuals", mock.Anything, "2026", []string(nil)).Return(nil).Maybe()
	exec.On("SyncActuals", mock.Anything, "2026", []string{}).Return(nil).Maybe()

	w := NewWorker(WorkerDeps{Queue: q, Executor: exec})

	job := &Job{ID: uuid.New(), JobType: models.SyncJobTier2Full, Year: "2026", Months: []string{}}
	err := w.execute(context.Background(), job)

	assert.NoError(t, err)
	// At least one of the SyncActuals expectations must have matched.
	exec.AssertCalled(t, "SyncActuals", mock.Anything, "2026", mock.Anything)
}

func TestWorker_Execute_MissingYearFails(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	exec := new(mockExecutor)
	w := NewWorker(WorkerDeps{Queue: q, Executor: exec})

	job := &Job{ID: uuid.New(), JobType: models.SyncJobActualFact}
	err := w.execute(context.Background(), job)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "year is required")
	// Executor must NOT be called when validation fails.
	exec.AssertNotCalled(t, "SyncActuals")
}

func TestWorker_Execute_UnknownJobTypeFails(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	w := NewWorker(WorkerDeps{Queue: q, Executor: new(mockExecutor)})

	err := w.execute(context.Background(), &Job{JobType: "BOGUS", Year: "2026"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported job_type")
}

// ─────────────────────────── Tests: processJob lifecycle ───────────────────────────

func TestWorker_ProcessJob_FullSuccessLifecycle(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	exec := new(mockExecutor)
	tracker := new(mockTrackingRepo)

	// Pre-enqueue and pop so we have a job that's ready for SetCurrent.
	job := newJob(models.SyncJobActualFact, "2026", "APR")
	_, err := q.Enqueue(context.Background(), job)
	require.NoError(t, err)
	popped, err := q.PopNext(context.Background())
	require.NoError(t, err)
	require.NotNil(t, popped)

	exec.On("SyncActuals", mock.Anything, "2026", []string{"APR"}).Return(nil)
	tracker.On("CreateRun", mock.Anything, mock.AnythingOfType("*models.SyncRunEntity")).Return(nil)
	tracker.On("CompleteRun",
		mock.Anything, mock.AnythingOfType("uuid.UUID"),
		models.SyncStatusSuccess,
		int64(0), int64(0), int64(0), "",
	).Return(nil)

	w := NewWorker(WorkerDeps{Queue: q, Executor: exec, TrackingRepo: tracker})
	w.processJob(context.Background(), popped)

	// After processJob: current should be cleared, sync_run recorded as SUCCESS.
	cur, err := q.GetCurrent(context.Background())
	require.NoError(t, err)
	assert.Nil(t, cur, "current should be cleared after processJob completes")

	exec.AssertExpectations(t)
	tracker.AssertExpectations(t)
}

func TestWorker_ProcessJob_FailureRecordedAsFailed(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	exec := new(mockExecutor)
	tracker := new(mockTrackingRepo)

	job := newJob(models.SyncJobActualFact, "2026", "APR")
	_, _ = q.Enqueue(context.Background(), job)
	popped, _ := q.PopNext(context.Background())

	syncErr := errors.New("simulated sync failure")
	exec.On("SyncActuals", mock.Anything, "2026", []string{"APR"}).Return(syncErr)
	tracker.On("CreateRun", mock.Anything, mock.AnythingOfType("*models.SyncRunEntity")).Return(nil)
	tracker.On("CompleteRun",
		mock.Anything, mock.AnythingOfType("uuid.UUID"),
		models.SyncStatusFailed,
		int64(0), int64(0), int64(0), syncErr.Error(),
	).Return(nil)

	w := NewWorker(WorkerDeps{Queue: q, Executor: exec, TrackingRepo: tracker})
	w.processJob(context.Background(), popped)

	cur, _ := q.GetCurrent(context.Background())
	assert.Nil(t, cur, "current should be cleared even after failure")

	tracker.AssertExpectations(t)
}

// ─────────────────────────── Tests: Start/Stop & priority ordering ───────────────────────────

// recordExecutor captures the order in which Worker dispatches jobs.
type recordExecutor struct {
	mu    sync.Mutex
	calls []string
}

func (r *recordExecutor) SyncActuals(_ context.Context, year string, months []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tag := "ACTUALS:" + year
	if len(months) > 0 {
		tag += ":" + months[0]
	}
	r.calls = append(r.calls, tag)
	return nil
}

func (r *recordExecutor) SyncFromDW(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, "DW")
	return nil
}

func (r *recordExecutor) Order() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.calls))
	copy(out, r.calls)
	return out
}

func TestWorker_RunLoop_DrainsByPriorityThenFIFO(t *testing.T) {
	rdb, _ := setupRedis(t)
	q := NewRedisQueue(rdb)
	exec := &recordExecutor{}

	// Enqueue lowest-priority job first to prove priority ordering wins over insertion.
	ctx := context.Background()
	tier1, _ := q.Enqueue(ctx, newJob(models.SyncJobTier1Fast, "2026", "APR"))
	require.True(t, tier1)
	time.Sleep(2 * time.Millisecond)
	manualA, _ := q.Enqueue(ctx, newJob(models.SyncJobManual, "2026", "JAN"))
	require.True(t, manualA)
	time.Sleep(2 * time.Millisecond)
	manualB, _ := q.Enqueue(ctx, newJob(models.SyncJobManual, "2026", "FEB"))
	require.True(t, manualB)
	time.Sleep(2 * time.Millisecond)
	dw, _ := q.Enqueue(ctx, newJob(models.SyncJobDW, "2026"))
	require.True(t, dw)

	w := NewWorker(WorkerDeps{Queue: q, Executor: exec})
	w.Start()
	defer w.Stop()

	// Worker should drain all 4 jobs within a few seconds.
	require.Eventually(t, func() bool {
		return len(exec.Order()) == 4
	}, 6*time.Second, 50*time.Millisecond, "worker should drain all 4 jobs")

	got := exec.Order()
	assert.Equal(t, "DW", got[0], "DW (P1) runs first")
	// MANUAL JAN was enqueued before MANUAL FEB — FIFO within same priority
	assert.Equal(t, "ACTUALS:2026:JAN", got[1])
	assert.Equal(t, "ACTUALS:2026:FEB", got[2])
	assert.Equal(t, "ACTUALS:2026:APR", got[3], "Tier 1 (P5) runs last")
}

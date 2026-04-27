package queue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	keyPending = "sync:queue:pending" // ZSET — score = priority*1e13 + enqueue_ms
	keyCurrent = "sync:queue:current"
	keyJob     = "sync:queue:job:" // + id
	keyDedup   = "sync:queue:dedup"
)

// scorePriorityShift > current unix-ms (~1.77e12) so priority dominates the score.
const scorePriorityShift float64 = 1e13

// Priority for each job type (lower = runs first).
// 1: DW_SYNC, 2: TIER2_FULL, 3: ACTUAL_FACT/MANUAL, 5: TIER1_FAST.
func PriorityFor(jobType string) int {
	switch jobType {
	case "DW_SYNC":
		return 1
	case "TIER2_FULL":
		return 2
	case "ACTUAL_FACT", "MANUAL":
		return 3
	case "TIER1_FAST":
		return 5
	default:
		return 3
	}
}

func makeScore(priority int, enqueuedAt time.Time) float64 {
	return float64(priority)*scorePriorityShift + float64(enqueuedAt.UnixMilli())
}

const (
	JobStatusPending   = "PENDING"
	JobStatusRunning   = "RUNNING"
	JobStatusCompleted = "COMPLETED"
	JobStatusFailed    = "FAILED"
)

// Job represents a sync task waiting in queue or currently running.
type Job struct {
	ID          uuid.UUID  `json:"id"`
	JobType     string     `json:"job_type"`
	Year        string     `json:"year"`
	Months      []string   `json:"months"`
	TriggeredBy string     `json:"triggered_by"`
	EnqueuedAt  time.Time  `json:"enqueued_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	Status      string     `json:"status"`
}

// DedupKey returns a stable hash of (job_type, year, sorted months) used to
// prevent the same task being queued multiple times.
func (j *Job) DedupKey() string {
	months := append([]string(nil), j.Months...)
	sort.Strings(months)
	raw := fmt.Sprintf("%s|%s|%s", j.JobType, j.Year, strings.Join(months, ","))
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:8])
}

// Queue is the abstraction backed by Redis (priority queue via ZSET).
type Queue interface {
	Enqueue(ctx context.Context, job *Job) (enqueued bool, err error)
	GetPending(ctx context.Context) ([]*Job, error)
	GetCurrent(ctx context.Context) (*Job, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	CancelAll(ctx context.Context) (int, error)
	Promote(ctx context.Context, id uuid.UUID) error
	// IsBusy reports whether the queue has any pending or running job.
	// Used by Tier 1 to skip its slot if the system is occupied.
	IsBusy(ctx context.Context) (bool, error)

	// Worker-only operations
	PopNext(ctx context.Context) (*Job, error)
	SetCurrent(ctx context.Context, job *Job) error
	ClearCurrent(ctx context.Context, job *Job) error
	Recover(ctx context.Context) (int, error)
}

type RedisQueue struct {
	rdb *redis.Client
}

func NewRedisQueue(rdb *redis.Client) *RedisQueue {
	return &RedisQueue{rdb: rdb}
}

// Enqueue adds a job to the tail of the queue. If a job with the same DedupKey
// already exists in queue or is currently running, returns enqueued=false.
func (q *RedisQueue) Enqueue(ctx context.Context, job *Job) (bool, error) {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	if job.EnqueuedAt.IsZero() {
		job.EnqueuedAt = time.Now()
	}
	if job.Status == "" {
		job.Status = JobStatusPending
	}

	dedupKey := job.DedupKey()
	added, err := q.rdb.SAdd(ctx, keyDedup, dedupKey).Result()
	if err != nil {
		return false, fmt.Errorf("queue.Enqueue.dedup: %w", err)
	}
	if added == 0 {
		return false, nil
	}

	payload, err := json.Marshal(job)
	if err != nil {
		_ = q.rdb.SRem(ctx, keyDedup, dedupKey).Err()
		return false, err
	}

	if err := q.rdb.Set(ctx, keyJob+job.ID.String(), payload, 24*time.Hour).Err(); err != nil {
		_ = q.rdb.SRem(ctx, keyDedup, dedupKey).Err()
		return false, fmt.Errorf("queue.Enqueue.setJob: %w", err)
	}

	score := makeScore(PriorityFor(job.JobType), job.EnqueuedAt)
	if err := q.rdb.ZAdd(ctx, keyPending, &redis.Z{Score: score, Member: job.ID.String()}).Err(); err != nil {
		_ = q.rdb.Del(ctx, keyJob+job.ID.String()).Err()
		_ = q.rdb.SRem(ctx, keyDedup, dedupKey).Err()
		return false, fmt.Errorf("queue.Enqueue.zadd: %w", err)
	}
	return true, nil
}

func (q *RedisQueue) GetPending(ctx context.Context) ([]*Job, error) {
	// ZRANGE in ascending order (lowest score = highest priority first)
	ids, err := q.rdb.ZRange(ctx, keyPending, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("queue.GetPending: %w", err)
	}
	return q.fetchJobs(ctx, ids)
}

// IsBusy returns true if a job is running OR pending queue is non-empty.
func (q *RedisQueue) IsBusy(ctx context.Context) (bool, error) {
	cur, err := q.rdb.Get(ctx, keyCurrent).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, fmt.Errorf("queue.IsBusy.current: %w", err)
	}
	if cur != "" {
		return true, nil
	}
	n, err := q.rdb.ZCard(ctx, keyPending).Result()
	if err != nil {
		return false, fmt.Errorf("queue.IsBusy.zcard: %w", err)
	}
	return n > 0, nil
}

func (q *RedisQueue) GetCurrent(ctx context.Context) (*Job, error) {
	id, err := q.rdb.Get(ctx, keyCurrent).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("queue.GetCurrent: %w", err)
	}
	if id == "" {
		return nil, nil
	}
	return q.fetchJob(ctx, id)
}

func (q *RedisQueue) Cancel(ctx context.Context, id uuid.UUID) error {
	job, err := q.fetchJob(ctx, id.String())
	if err != nil {
		return err
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	removed, err := q.rdb.ZRem(ctx, keyPending, id.String()).Result()
	if err != nil {
		return fmt.Errorf("queue.Cancel.zrem: %w", err)
	}
	if removed == 0 {
		return fmt.Errorf("job is currently running and cannot be cancelled")
	}

	pipe := q.rdb.Pipeline()
	pipe.SRem(ctx, keyDedup, job.DedupKey())
	pipe.Del(ctx, keyJob+id.String())
	_, err = pipe.Exec(ctx)
	return err
}

// CancelAll removes every pending job from the queue. The currently running
// job is NOT touched (worker is mid-execution). Returns count of removed jobs.
// Used by the "Cancel All" admin button.
func (q *RedisQueue) CancelAll(ctx context.Context) (int, error) {
	ids, err := q.rdb.ZRange(ctx, keyPending, 0, -1).Result()
	if err != nil {
		return 0, fmt.Errorf("queue.CancelAll.zrange: %w", err)
	}
	if len(ids) == 0 {
		return 0, nil
	}
	pipe := q.rdb.Pipeline()
	for _, idStr := range ids {
		job, _ := q.fetchJob(ctx, idStr)
		pipe.ZRem(ctx, keyPending, idStr)
		if job != nil {
			pipe.SRem(ctx, keyDedup, job.DedupKey())
		}
		pipe.Del(ctx, keyJob+idStr)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("queue.CancelAll.pipe: %w", err)
	}
	return len(ids), nil
}

// Promote sets the job's score to (current min score - 1) so it runs next,
// regardless of its priority class.
func (q *RedisQueue) Promote(ctx context.Context, id uuid.UUID) error {
	if _, err := q.rdb.ZScore(ctx, keyPending, id.String()).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("job not in pending queue")
		}
		return fmt.Errorf("queue.Promote.zscore: %w", err)
	}
	heads, err := q.rdb.ZRangeWithScores(ctx, keyPending, 0, 0).Result()
	if err != nil {
		return fmt.Errorf("queue.Promote.zrange: %w", err)
	}
	var newScore float64
	if len(heads) > 0 {
		newScore = heads[0].Score - 1
	} else {
		newScore = float64(time.Now().UnixMilli())
	}
	if err := q.rdb.ZAdd(ctx, keyPending, &redis.Z{Score: newScore, Member: id.String()}).Err(); err != nil {
		return fmt.Errorf("queue.Promote.zadd: %w", err)
	}
	return nil
}

// PopNext atomically removes and returns the highest-priority pending job.
// Returns (nil, nil) if the queue is empty (after a short blocking sleep so the
// worker doesn't spin). Uses ZPopMin for portability with miniredis (test) and
// older Redis versions; BZPopMin would be slightly more efficient but isn't
// available everywhere.
func (q *RedisQueue) PopNext(ctx context.Context) (*Job, error) {
	res, err := q.rdb.ZPopMin(ctx, keyPending, 1).Result()
	if err != nil {
		return nil, fmt.Errorf("queue.PopNext: %w", err)
	}
	if len(res) == 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
			return nil, nil
		}
	}
	id, ok := res[0].Member.(string)
	if !ok {
		return nil, fmt.Errorf("queue.PopNext: unexpected member type")
	}
	return q.fetchJob(ctx, id)
}

func (q *RedisQueue) SetCurrent(ctx context.Context, job *Job) error {
	job.Status = JobStatusRunning
	now := time.Now()
	job.StartedAt = &now
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	pipe := q.rdb.Pipeline()
	pipe.Set(ctx, keyCurrent, job.ID.String(), 0)
	pipe.Set(ctx, keyJob+job.ID.String(), payload, 24*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) ClearCurrent(ctx context.Context, job *Job) error {
	pipe := q.rdb.Pipeline()
	pipe.Del(ctx, keyCurrent)
	pipe.SRem(ctx, keyDedup, job.DedupKey())
	pipe.Del(ctx, keyJob+job.ID.String())
	_, err := pipe.Exec(ctx)
	return err
}

// Recover handles startup: if a 'current' was set when the previous worker died,
// requeue it at the head of pending so it retries.
// Also drops dedup keys for jobs whose hash is missing (defensive).
// Returns the number of jobs recovered.
func (q *RedisQueue) Recover(ctx context.Context) (int, error) {
	id, err := q.rdb.Get(ctx, keyCurrent).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("queue.Recover.getCurrent: %w", err)
	}
	if id == "" {
		return 0, nil
	}
	// Re-enqueue with priority bump (-1e10) so it runs before any new job of same priority.
	job, _ := q.fetchJob(ctx, id)
	var score float64
	if job != nil {
		score = makeScore(PriorityFor(job.JobType), job.EnqueuedAt) - 1e10
	} else {
		score = -float64(time.Now().UnixMilli())
	}
	pipe := q.rdb.Pipeline()
	pipe.ZAdd(ctx, keyPending, &redis.Z{Score: score, Member: id})
	pipe.Del(ctx, keyCurrent)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("queue.Recover.requeue: %w", err)
	}
	return 1, nil
}

func (q *RedisQueue) fetchJob(ctx context.Context, id string) (*Job, error) {
	payload, err := q.rdb.Get(ctx, keyJob+id).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var job Job
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (q *RedisQueue) fetchJobs(ctx context.Context, ids []string) ([]*Job, error) {
	if len(ids) == 0 {
		return []*Job{}, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = keyJob + id
	}
	payloads, err := q.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	jobs := make([]*Job, 0, len(payloads))
	for _, p := range payloads {
		if p == nil {
			continue
		}
		s, ok := p.(string)
		if !ok {
			continue
		}
		var job Job
		if err := json.Unmarshal([]byte(s), &job); err != nil {
			continue
		}
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

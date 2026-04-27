package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	syncRepo "p2p-back-end/modules/external_sync/repository"
)

// JobExecutor abstracts the actual sync work so the worker is decoupled from services.
type JobExecutor interface {
	SyncActuals(ctx context.Context, year string, months []string) error
	SyncFromDW(ctx context.Context, year string, months []string) error
}

type executor struct {
	actualSrv  models.ActualService
	extSyncSrv models.ExternalSyncService
}

func NewExecutor(actualSrv models.ActualService, extSyncSrv models.ExternalSyncService) JobExecutor {
	return &executor{actualSrv: actualSrv, extSyncSrv: extSyncSrv}
}

func (e *executor) SyncActuals(ctx context.Context, year string, months []string) error {
	if e.actualSrv == nil {
		return fmt.Errorf("actual service not configured")
	}
	return e.actualSrv.SyncActuals(ctx, year, months)
}

func (e *executor) SyncFromDW(ctx context.Context, year string, months []string) error {
	if e.extSyncSrv == nil {
		return fmt.Errorf("external sync service not configured")
	}
	return e.extSyncSrv.SyncFromDW(ctx, year, months)
}

type WorkerDeps struct {
	Queue        Queue
	Executor     JobExecutor
	TrackingRepo syncRepo.SyncTrackingRepository
}

type Worker struct {
	deps   WorkerDeps
	stopCh chan struct{}
}

func NewWorker(deps WorkerDeps) *Worker {
	return &Worker{deps: deps, stopCh: make(chan struct{})}
}

func (w *Worker) Start() {
	go w.run()
}

func (w *Worker) Stop() {
	close(w.stopCh)
}

func (w *Worker) run() {
	logs.Info("⚙️ Sync Queue Worker started")

	if recovered, err := w.deps.Queue.Recover(context.Background()); err != nil {
		logs.Errorf("⚙️ Worker: Recover failed: %v", err)
	} else if recovered > 0 {
		logs.Warnf("⚙️ Worker: Recovered %d job(s) from previous run", recovered)
	}

	for {
		select {
		case <-w.stopCh:
			logs.Info("⚙️ Sync Queue Worker stopping")
			return
		default:
		}

		ctx := context.Background()
		job, err := w.deps.Queue.PopNext(ctx)
		if err != nil {
			logs.Errorf("⚙️ Worker: PopNext: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if job == nil {
			continue
		}
		w.processJob(ctx, job)
	}
}

func (w *Worker) processJob(ctx context.Context, job *Job) {
	logs.Infof("⚙️ Worker: starting %s (id=%s, year=%s, months=%v, by=%s)",
		job.JobType, job.ID, job.Year, job.Months, job.TriggeredBy)

	if err := w.deps.Queue.SetCurrent(ctx, job); err != nil {
		logs.Errorf("⚙️ Worker: SetCurrent: %v", err)
	}

	runEntity := &models.SyncRunEntity{
		JobType:     job.JobType,
		Year:        job.Year,
		TriggeredBy: job.TriggeredBy,
	}
	if len(job.Months) == 1 {
		runEntity.Month = job.Months[0]
	}
	var runID uuid.UUID
	if w.deps.TrackingRepo != nil {
		if err := w.deps.TrackingRepo.CreateRun(ctx, runEntity); err == nil {
			runID = runEntity.ID
		}
	}

	jobErr := w.execute(ctx, job)

	if w.deps.TrackingRepo != nil && runID != uuid.Nil {
		status := models.SyncStatusSuccess
		errMsg := ""
		if jobErr != nil {
			status = models.SyncStatusFailed
			errMsg = jobErr.Error()
		}
		_ = w.deps.TrackingRepo.CompleteRun(ctx, runID, status, 0, 0, 0, errMsg)
	}

	if err := w.deps.Queue.ClearCurrent(ctx, job); err != nil {
		logs.Errorf("⚙️ Worker: ClearCurrent: %v", err)
	}

	if jobErr != nil {
		logs.Errorf("⚙️ Worker: %s FAILED: %v", job.JobType, jobErr)
	} else {
		logs.Infof("⚙️ Worker: %s SUCCESS", job.JobType)
	}
}

func (w *Worker) execute(ctx context.Context, job *Job) error {
	switch job.JobType {
	case models.SyncJobTier1Fast, models.SyncJobTier2Full, models.SyncJobActualFact, models.SyncJobManual:
		if job.Year == "" {
			return fmt.Errorf("year is required for %s", job.JobType)
		}
		return w.deps.Executor.SyncActuals(ctx, job.Year, job.Months)
	case models.SyncJobDW:
		if job.Year == "" {
			return fmt.Errorf("year is required for %s", job.JobType)
		}
		return w.deps.Executor.SyncFromDW(ctx, job.Year, job.Months)
	default:
		return fmt.Errorf("unsupported job_type: %s", job.JobType)
	}
}

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	syncQueue "p2p-back-end/modules/external_sync/queue"
	_repo "p2p-back-end/modules/external_sync/repository"
	"p2p-back-end/pkg/middlewares"
)

// SyncAdminController — admin-only endpoints for sync observability + queue management
type SyncAdminController struct {
	trackingRepo _repo.SyncTrackingRepository
	queue        syncQueue.Queue
	authSrv      models.AuthService
}

func NewSyncAdminController(
	r fiber.Router,
	authSrv models.AuthService,
	trackingRepo _repo.SyncTrackingRepository,
	queue syncQueue.Queue,
) {
	c := &SyncAdminController{
		trackingRepo: trackingRepo,
		queue:        queue,
		authSrv:      authSrv,
	}

	// Observability
	r.Get("/sync/status", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.getStatus, models.RoleAdmin)))
	r.Get("/sync/history", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.getHistory, models.RoleAdmin)))
	r.Get("/sync/reconciliation", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.getReconciliation, models.RoleAdmin)))

	// Trigger + queue management
	r.Post("/sync/trigger", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.triggerSync, models.RoleAdmin)))
	r.Get("/sync/queue", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.getQueue, models.RoleAdmin)))
	r.Post("/sync/queue/cancel/:id", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.cancelQueueItem, models.RoleAdmin)))
	r.Post("/sync/queue/cancel-all", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.cancelAllQueue, models.RoleAdmin)))
	r.Post("/sync/queue/promote/:id", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.promoteQueueItem, models.RoleAdmin)))
}

func (c *SyncAdminController) getStatus(ctx *fiber.Ctx, user *models.UserInfo) error {
	result := fiber.Map{}
	jobTypes := []string{models.SyncJobDW, models.SyncJobTier1Fast, models.SyncJobTier2Full, models.SyncJobManual}
	for _, jt := range jobTypes {
		run, err := c.trackingRepo.GetLatest(ctx.UserContext(), jt)
		if err != nil {
			result[jt] = fiber.Map{"error": err.Error()}
			continue
		}
		result[jt] = run
	}
	return ctx.JSON(fiber.Map{"latest_by_type": result})
}

func (c *SyncAdminController) getHistory(ctx *fiber.Ctx, user *models.UserInfo) error {
	limit := ctx.QueryInt("limit", 50)
	if limit > 500 {
		limit = 500
	}
	jobType := ctx.Query("job_type", "")
	status := ctx.Query("status", "")

	var (
		runs []models.SyncRunEntity
		err  error
	)
	if jobType == "" && status == "" {
		runs, err = c.trackingRepo.GetRecent(ctx.UserContext(), limit)
	} else {
		runs, err = c.trackingRepo.GetRunsByJobAndStatus(ctx.UserContext(), jobType, status, limit)
	}
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"runs": runs, "count": len(runs)})
}

func (c *SyncAdminController) getReconciliation(ctx *fiber.Ctx, user *models.UserInfo) error {
	year := ctx.Query("year", "")
	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}
	rec, err := c.trackingRepo.GetReconciliation(ctx.UserContext(), year)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(rec)
}

type triggerRequest struct {
	JobType string   `json:"job_type"`
	Year    string   `json:"year"`
	Months  []string `json:"months"`
}

// triggerSync enqueues a sync job. Returns 202 + job id, or 409 if duplicate.
func (c *SyncAdminController) triggerSync(ctx *fiber.Ctx, user *models.UserInfo) error {
	var req triggerRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	job := buildJobFromRequest(&req, "ADMIN:"+user.ID)
	if job == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported job_type or missing required fields"})
	}

	enqueued, err := c.queue.Enqueue(ctx.UserContext(), job)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !enqueued {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "an identical job is already queued or running",
		})
	}

	logs.Infof("Sync trigger by %s queued: %s year=%s months=%v id=%s",
		job.TriggeredBy, job.JobType, job.Year, job.Months, job.ID)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"accepted":     true,
		"job_id":       job.ID,
		"job_type":     job.JobType,
		"year":         job.Year,
		"months":       job.Months,
		"triggered_by": job.TriggeredBy,
	})
}

// buildJobFromRequest expands defaults per job_type and returns nil for invalid input.
func buildJobFromRequest(req *triggerRequest, triggeredBy string) *syncQueue.Job {
	jt := req.JobType
	if jt == "" {
		jt = models.SyncJobActualFact
	}

	year := req.Year
	months := req.Months

	switch jt {
	case models.SyncJobTier1Fast:
		if year == "" {
			year = fmt.Sprintf("%d", time.Now().Year())
		}
		if len(months) == 0 {
			abbr := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
			months = []string{abbr[time.Now().Month()-1]}
		}
	case models.SyncJobTier2Full:
		if year == "" {
			year = fmt.Sprintf("%d", time.Now().Year())
		}
		months = []string{} // always full year
	case models.SyncJobActualFact, models.SyncJobManual:
		if year == "" {
			return nil
		}
	case models.SyncJobDW:
		// year/months ignored by service, but still record what user passed
	default:
		return nil
	}

	return &syncQueue.Job{
		JobType:     jt,
		Year:        year,
		Months:      months,
		TriggeredBy: triggeredBy,
	}
}

// getQueue returns current running job + pending list + per-type ETAs.
// ETA for each job uses the avg duration of that job_type's last 7 SUCCESS runs.
func (c *SyncAdminController) getQueue(ctx *fiber.Ctx, user *models.UserInfo) error {
	current, err := c.queue.GetCurrent(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	pending, err := c.queue.GetPending(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Build avg cache once (keyed by job_type) so we don't re-query per loop iteration.
	avgCache := map[string]int64{}
	collectType := func(jt string) {
		if jt == "" {
			return
		}
		if _, ok := avgCache[jt]; ok {
			return
		}
		avgCache[jt] = c.avgDurationMs(ctx.UserContext(), jt)
	}
	if current != nil {
		collectType(current.JobType)
	}
	for _, p := range pending {
		collectType(p.JobType)
	}

	var currentInfo fiber.Map
	if current != nil {
		avgMs := avgCache[current.JobType]
		var elapsedMs, remainingMs int64
		if current.StartedAt != nil {
			elapsedMs = time.Since(*current.StartedAt).Milliseconds()
			if avgMs > 0 {
				remainingMs = avgMs - elapsedMs
				if remainingMs < 0 {
					remainingMs = 0
				}
			}
		}
		currentInfo = fiber.Map{
			"id":           current.ID,
			"job_type":     current.JobType,
			"year":         current.Year,
			"months":       current.Months,
			"triggered_by": current.TriggeredBy,
			"enqueued_at":  current.EnqueuedAt,
			"started_at":   current.StartedAt,
			"elapsed_ms":   elapsedMs,
			"avg_total_ms": avgMs,
			"remaining_ms": remainingMs,
		}
	}

	pendingList := make([]fiber.Map, 0, len(pending))
	// running tally of "time until this pending job starts"
	cumulativeMs := int64(0)
	if currentInfo != nil {
		if r, ok := currentInfo["remaining_ms"].(int64); ok {
			cumulativeMs = r
		}
	}
	for i, p := range pending {
		myAvg := avgCache[p.JobType]
		pendingList = append(pendingList, fiber.Map{
			"id":           p.ID,
			"job_type":     p.JobType,
			"year":         p.Year,
			"months":       p.Months,
			"triggered_by": p.TriggeredBy,
			"enqueued_at":  p.EnqueuedAt,
			"position":     i + 1,
			"priority":     syncQueue.PriorityFor(p.JobType),
			"avg_total_ms": myAvg,         // expected duration of this job (from its own type)
			"eta_start_ms": cumulativeMs,  // when it will start = sum of preceding job durations
		})
		cumulativeMs += myAvg // next pending starts after this one finishes
	}

	return ctx.JSON(fiber.Map{
		"current": currentInfo,
		"pending": pendingList,
		"count":   len(pending),
	})
}

// avgDurationMs returns avg duration_ms of last 7 SUCCESS runs of the SAME job_type.
// Returns 0 if no history for that type. Per-type isolation guarantees that
// estimating a Tier 1 job doesn't pull in long ACTUAL_FACT durations.
func (c *SyncAdminController) avgDurationMs(ctx context.Context, jobType string) int64 {
	if c.trackingRepo == nil || jobType == "" {
		return 0
	}
	runs, err := c.trackingRepo.GetRunsByJobAndStatus(ctx, jobType, models.SyncStatusSuccess, 7)
	if err != nil || len(runs) == 0 {
		return 0
	}
	var sum int64
	var n int64
	for _, r := range runs {
		if r.DurationMs > 0 {
			sum += r.DurationMs
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / n
}

func (c *SyncAdminController) cancelQueueItem(ctx *fiber.Ctx, user *models.UserInfo) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := c.queue.Cancel(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	logs.Infof("Sync queue: %s cancelled job %s", "ADMIN:"+user.ID, idStr)
	return ctx.JSON(fiber.Map{"cancelled": true, "id": idStr})
}

// cancelAllQueue clears every pending job in the queue AND marks any FAILED
// runs that the retry cron would otherwise pick up as CANCELED. Together this
// guarantees nothing the admin just cleared comes back automatically. The
// currently RUNNING job is not interrupted (worker is mid-execution).
func (c *SyncAdminController) cancelAllQueue(ctx *fiber.Ctx, user *models.UserInfo) error {
	cancelled, err := c.queue.CancelAll(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var disabledRetries int64
	if c.trackingRepo != nil {
		// Match retry cron's window/limits so we exactly cover what it would re-pick.
		disabledRetries, err = c.trackingRepo.MarkRetryableFailedAsCanceled(ctx.UserContext(), 24*time.Hour, 3)
		if err != nil {
			// Log but continue — queue is cleared which is the primary intent.
			logs.Errorf("cancel-all: failed to mark retryable runs as CANCELED: %v", err)
		}
	}

	logs.Infof("Sync queue: %s cancelled all (queue=%d, disabled_retries=%d)",
		"ADMIN:"+user.ID, cancelled, disabledRetries)
	return ctx.JSON(fiber.Map{
		"cancelled_queue":  cancelled,
		"disabled_retries": disabledRetries,
	})
}

func (c *SyncAdminController) promoteQueueItem(ctx *fiber.Ctx, user *models.UserInfo) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := c.queue.Promote(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	logs.Infof("Sync queue: %s promoted job %s", "ADMIN:"+user.ID, idStr)
	return ctx.JSON(fiber.Map{"promoted": true, "id": idStr})
}

package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	_repo "p2p-back-end/modules/external_sync/repository"
	"p2p-back-end/pkg/middlewares"
)

// SyncAdminController — admin-only endpoints for sync observability + manual trigger
type SyncAdminController struct {
	trackingRepo _repo.SyncTrackingRepository
	extSyncSrv   models.ExternalSyncService
	actualSrv    models.ActualService
	authSrv      models.AuthService
	syncMutex    *sync.Mutex
}

func NewSyncAdminController(
	r fiber.Router,
	authSrv models.AuthService,
	trackingRepo _repo.SyncTrackingRepository,
	extSyncSrv models.ExternalSyncService,
	actualSrv models.ActualService,
	syncMutex *sync.Mutex,
) {
	c := &SyncAdminController{
		trackingRepo: trackingRepo,
		extSyncSrv:   extSyncSrv,
		actualSrv:    actualSrv,
		authSrv:      authSrv,
		syncMutex:    syncMutex,
	}

	// GET /admin/sync/status — latest run + recent history
	r.Get("/sync/status", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.getStatus, models.RoleAdmin)))
	// GET /admin/sync/history — paginated history with optional filters
	r.Get("/sync/history", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.getHistory, models.RoleAdmin)))
	// GET /admin/sync/reconciliation?year=2026 — row count health check
	r.Get("/sync/reconciliation", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.getReconciliation, models.RoleAdmin)))
	// POST /admin/sync/trigger — manual trigger (body: {job_type, year, month})
	r.Post("/sync/trigger", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(c.triggerSync, models.RoleAdmin)))
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
	JobType string `json:"job_type"` // ACTUAL_FACT | DW_SYNC
	Year    string `json:"year"`     // required for ACTUAL_FACT
	Months  []string `json:"months"` // optional; empty = full year
}

func (c *SyncAdminController) triggerSync(ctx *fiber.Ctx, user *models.UserInfo) error {
	var req triggerRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	triggeredBy := "ADMIN:" + user.ID

	// Try to acquire sync mutex — refuse if another sync is already running
	// (prevents concurrent syncs that cause duplicate transaction inserts)
	if c.syncMutex != nil && !c.syncMutex.TryLock() {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "another sync is already running — please wait until it finishes",
		})
	}

	// Run in background so the HTTP request doesn't time out for large syncs
	go func() {
		if c.syncMutex != nil {
			defer c.syncMutex.Unlock()
		}
		bgCtx := context.Background()

		run := &models.SyncRunEntity{
			JobType:     req.JobType,
			Year:        req.Year,
			TriggeredBy: triggeredBy,
		}
		if len(req.Months) == 1 {
			run.Month = req.Months[0]
		}
		var runID = run.ID
		if c.trackingRepo != nil {
			if err := c.trackingRepo.CreateRun(bgCtx, run); err == nil {
				runID = run.ID
			}
		}

		var err error
		switch req.JobType {
		case models.SyncJobManual, models.SyncJobActualFact, "":
			if req.Year == "" {
				err = fmt.Errorf("year is required")
			} else {
				err = c.actualSrv.SyncActuals(bgCtx, req.Year, req.Months)
			}
		case models.SyncJobTier1Fast:
			year := req.Year
			months := req.Months
			if year == "" {
				year = fmt.Sprintf("%d", time.Now().Year())
			}
			if len(months) == 0 {
				monthMap := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
				months = []string{monthMap[time.Now().Month()-1]}
			}
			err = c.actualSrv.SyncActuals(bgCtx, year, months)
		case models.SyncJobTier2Full:
			year := req.Year
			if year == "" {
				year = fmt.Sprintf("%d", time.Now().Year())
			}
			err = c.actualSrv.SyncActuals(bgCtx, year, []string{})
		case models.SyncJobDW:
			if c.extSyncSrv == nil {
				err = fmt.Errorf("external sync service not configured")
			} else {
				err = c.extSyncSrv.SyncFromDW(bgCtx)
			}
		default:
			err = fmt.Errorf("unsupported job_type: %s", req.JobType)
		}

		if c.trackingRepo != nil {
			status := models.SyncStatusSuccess
			errMsg := ""
			if err != nil {
				status = models.SyncStatusFailed
				errMsg = err.Error()
			}
			_ = c.trackingRepo.CompleteRun(bgCtx, runID, status, 0, 0, 0, errMsg)
		}
		if err != nil {
			logs.Errorf("Manual sync trigger by %s failed: %v", triggeredBy, err)
		} else {
			logs.Infof("Manual sync trigger by %s completed: %s", triggeredBy, req.JobType)
		}
	}()

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"accepted":     true,
		"job_type":     req.JobType,
		"year":         req.Year,
		"months":       req.Months,
		"triggered_by": triggeredBy,
	})
}

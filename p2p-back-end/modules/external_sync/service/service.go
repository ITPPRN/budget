package service

import (
	"context"
	"fmt"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	_extSyncRepo "p2p-back-end/modules/external_sync/repository"
	"time"

	"github.com/google/uuid"
)

type externalSyncService struct {
	repo         models.ExternalSyncRepository
	actualSrv    models.ActualService
	trackingRepo _extSyncRepo.SyncTrackingRepository
}

func NewExternalSyncService(
	repo models.ExternalSyncRepository,
	actualSrv models.ActualService,
	trackingRepo _extSyncRepo.SyncTrackingRepository,
) models.ExternalSyncService {
	return &externalSyncService{
		repo:         repo,
		actualSrv:    actualSrv,
		trackingRepo: trackingRepo,
	}
}

func (s *externalSyncService) SyncFromDW(ctx context.Context) error {
	logs.Info("🔄 DW Sync: Starting data synchronization from 2025 to present...")

	// Create tracking run (if tracking is configured)
	var runID uuid.UUID
	if s.trackingRepo != nil {
		run := &models.SyncRunEntity{
			JobType:     models.SyncJobDW,
			TriggeredBy: "CRON",
		}
		if err := s.trackingRepo.CreateRun(ctx, run); err == nil {
			runID = run.ID
		}
	}

	// Stats for tracking
	var totalRowsFetched, totalRowsSkipped int64
	var anyError bool
	var firstErrMsg string

	// Increase timeout to 6 hours for initial full sync of large data
	ctx, cancel := context.WithTimeout(ctx, 360*time.Minute)
	defer cancel()

	// Wrap in deferred function to ALWAYS record completion status
	defer func() {
		if s.trackingRepo != nil && runID != uuid.Nil {
			status := models.SyncStatusSuccess
			if anyError {
				status = models.SyncStatusPartial
			}
			// Use fresh background context in case parent ctx was cancelled
			_ = s.trackingRepo.CompleteRun(
				context.Background(), runID, status,
				totalRowsFetched, totalRowsFetched, totalRowsSkipped, firstErrMsg,
			)
		}
	}()

	now := time.Now()
	currentYear := now.Year()
	startYear := currentYear - 1
	currentMonth := int(now.Month())

	batchSize := 2000

	// We iterate by Year and Month to make the query much faster and see progress clearly
	for year := startYear; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}

		for month := 1; month <= endMonth; month++ {
			logs.Infof("🔄 DW Sync: Processing Year %d Month %d...", year, month)

			// 1. Sync HMW for this month — DELETE first (idempotent), then fetch + insert in batches
			if err := s.repo.DeleteHMWByYearMonth(ctx, year, month); err != nil {
				logs.Errorf("DW Sync: Failed to delete HMW before re-insert %d-%02d: %v", year, month, err)
				anyError = true
				if firstErrMsg == "" {
					firstErrMsg = fmt.Sprintf("HMW delete %d-%02d: %v", year, month, err)
				}
				continue
			}
			err := s.repo.FetchHMWInBatches(ctx, year, month, batchSize, func(data []models.AchHmwGleEntity) error {
				totalRowsFetched += int64(len(data))
				return s.repo.UpsertHMWLocal(ctx, data)
			})
			if err != nil {
				logs.Errorf("DW Sync: Failed HMW for %d-%02d: %v", year, month, err)
				anyError = true
				if firstErrMsg == "" {
					firstErrMsg = fmt.Sprintf("HMW fetch %d-%02d: %v", year, month, err)
				}
			}

			// 2. Sync CLIK for this month — same pattern
			if err := s.repo.DeleteCLIKByYearMonth(ctx, year, month); err != nil {
				logs.Errorf("DW Sync: Failed to delete CLIK before re-insert %d-%02d: %v", year, month, err)
				anyError = true
				if firstErrMsg == "" {
					firstErrMsg = fmt.Sprintf("CLIK delete %d-%02d: %v", year, month, err)
				}
				continue
			}
			err = s.repo.FetchCLIKInBatches(ctx, year, month, batchSize, func(data []models.ClikGleEntity) error {
				totalRowsFetched += int64(len(data))
				return s.repo.UpsertCLIKLocal(ctx, data)
			})
			if err != nil {
				logs.Errorf("DW Sync: Failed CLIK for %d-%02d: %v", year, month, err)
				anyError = true
				if firstErrMsg == "" {
					firstErrMsg = fmt.Sprintf("CLIK fetch %d-%02d: %v", year, month, err)
				}
			}
		}

		// 3. Trigger Internal Actual Sync for this year after raw data update
		yearStr := fmt.Sprintf("%d", year)
		if err := s.actualSrv.SyncActuals(ctx, yearStr, []string{}); err != nil {
			logs.Errorf("DW Sync: Internal Fact Sync failed for year %s: %v", yearStr, err)
			anyError = true
			if firstErrMsg == "" {
				firstErrMsg = fmt.Sprintf("actual fact sync %s: %v", yearStr, err)
			}
		}

		// 4. Refresh Data Inventory Metadata for Admin UI
		if err := s.actualSrv.RefreshDataInventory(ctx); err != nil {
			logs.Errorf("DW Sync: Inventory Refresh failed: %v", err)
			anyError = true
			if firstErrMsg == "" {
				firstErrMsg = fmt.Sprintf("inventory refresh: %v", err)
			}
		}
	}

	logs.Info("✅ DW Sync: Data synchronization completed successfully.")
	return nil
}

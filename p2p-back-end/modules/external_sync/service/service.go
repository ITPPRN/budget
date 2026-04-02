package service

import (
	"context"
	"fmt"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"time"
)

type externalSyncService struct {
	repo      models.ExternalSyncRepository
	actualSrv models.ActualService
}

func NewExternalSyncService(repo models.ExternalSyncRepository, actualSrv models.ActualService) models.ExternalSyncService {
	return &externalSyncService{
		repo:      repo,
		actualSrv: actualSrv,
	}
}

func (s *externalSyncService) SyncFromDW(ctx context.Context) error {
	logs.Info("🔄 DW Sync: Starting data synchronization from 2025 to present...")

	// Increase timeout to 6 hours for initial full sync of large data
	ctx, cancel := context.WithTimeout(ctx, 360*time.Minute)
	defer cancel()

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

			// 1. Sync HMW for this month
			err := s.repo.FetchHMWInBatches(ctx, year, month, batchSize, func(data []models.AchHmwGleEntity) error {
				return s.repo.UpsertHMWLocal(ctx, data)
			})
			if err != nil {
				logs.Errorf("DW Sync: Failed HMW for %d-%02d: %v", year, month, err)
			}

			// 2. Sync CLIK for this month
			err = s.repo.FetchCLIKInBatches(ctx, year, month, batchSize, func(data []models.ClikGleEntity) error {
				return s.repo.UpsertCLIKLocal(ctx, data)
			})
			if err != nil {
				logs.Errorf("DW Sync: Failed CLIK for %d-%02d: %v", year, month, err)
			}
		}

		// 3. Trigger Internal Actual Sync for this year after raw data update
		yearStr := fmt.Sprintf("%d", year)
		if err := s.actualSrv.SyncActuals(ctx, yearStr, []string{}); err != nil {
			logs.Errorf("DW Sync: Internal Fact Sync failed for year %s: %v", yearStr, err)
		}

		// 4. Refresh Data Inventory Metadata for Admin UI
		if err := s.actualSrv.RefreshDataInventory(ctx); err != nil {
			logs.Errorf("DW Sync: Inventory Refresh failed: %v", err)
		}
	}

	logs.Info("✅ DW Sync: Data synchronization completed successfully.")
	return nil
}

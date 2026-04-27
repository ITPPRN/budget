package service

import (
	"context"
	"errors"
	"fmt"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	_extSyncRepo "p2p-back-end/modules/external_sync/repository"
	"strconv"
	"sync"
	"time"
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

// monthCodeToInt maps "JAN".."DEC" → 1..12.
var monthCodeToInt = map[string]int{
	"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
	"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
}

const dwPerMonthTimeout = 30 * time.Minute

// SyncFromDW pulls raw HMW + CLIK for one or more months of the given year and
// then re-projects ActualFact for those months. Designed to be invoked once per
// month (cron fans out 17 jobs to cover the rolling 17-month window) so a TCP
// reset on one month does not cascade into 6-hour zombies.
//
// Tracking is handled by the queue Worker via sync_run_entities — this function
// just returns error/nil. HMW + CLIK within one month run in parallel: different
// DW tables, different local tables, no contention; cuts wall-clock ~half.
func (s *externalSyncService) SyncFromDW(ctx context.Context, year string, months []string) error {
	yearInt, err := strconv.Atoi(year)
	if err != nil || yearInt < 1900 {
		return fmt.Errorf("SyncFromDW: invalid year %q", year)
	}

	monthInts, err := resolveMonthsForSync(months, yearInt)
	if err != nil {
		return err
	}
	if len(monthInts) == 0 {
		return nil
	}

	logs.Infof("🔄 DW Sync: year=%s months=%v starting", year, months)
	batchSize := 5000

	for _, month := range monthInts {
		// Per-month deadline: prevents a hung DW connection from blocking forever.
		// 30 min is generous for ~14M rows in a single month at our batch sizes.
		monthCtx, cancel := context.WithTimeout(ctx, dwPerMonthTimeout)
		if err := s.syncOneMonth(monthCtx, yearInt, month, batchSize); err != nil {
			cancel()
			logs.Errorf("DW Sync %d-%02d: %v", yearInt, month, err)
			return fmt.Errorf("DW sync %d-%02d: %w", yearInt, month, err)
		}
		cancel()
	}

	// Re-project actuals for ONLY the months we just pulled (not the full year).
	// SyncActuals expects 3-letter month codes — convert back from int.
	intToCode := []string{"", "JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	syncedCodes := make([]string, 0, len(monthInts))
	for _, m := range monthInts {
		syncedCodes = append(syncedCodes, intToCode[m])
	}
	if err := s.actualSrv.SyncActuals(ctx, year, syncedCodes); err != nil {
		return fmt.Errorf("actual fact sync %s/%v: %w", year, syncedCodes, err)
	}

	if err := s.actualSrv.RefreshDataInventory(ctx); err != nil {
		// Inventory refresh failure is non-fatal — log and continue.
		logs.Errorf("DW Sync: Inventory refresh failed (non-fatal): %v", err)
	}

	logs.Infof("✅ DW Sync: year=%s months=%v done", year, syncedCodes)
	return nil
}

// resolveMonthsForSync converts the queue's month code list to integers. An
// empty list defaults to the current month (single-month safety, since the cron
// fan-out always supplies an explicit month).
func resolveMonthsForSync(months []string, year int) ([]int, error) {
	if len(months) == 0 {
		return []int{int(time.Now().Month())}, nil
	}
	out := make([]int, 0, len(months))
	seen := make(map[int]struct{}, len(months))
	for _, code := range months {
		m, ok := monthCodeToInt[code]
		if !ok {
			return nil, fmt.Errorf("SyncFromDW: invalid month code %q", code)
		}
		if _, dup := seen[m]; dup {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	return out, nil
}

// syncOneMonth pulls HMW + CLIK in parallel for one (year, month) and waits for
// both. Returns a combined error if either side fails.
func (s *externalSyncService) syncOneMonth(ctx context.Context, year, month, batchSize int) error {
	logs.Infof("🔄 DW Sync: %d-%02d (HMW + CLIK in parallel)", year, month)

	var wg sync.WaitGroup
	var hmwErr, clikErr error
	var hmwRows, clikRows int64

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := s.repo.DeleteHMWByYearMonth(ctx, year, month); err != nil {
			hmwErr = fmt.Errorf("HMW delete: %w", err)
			return
		}
		if err := s.repo.FetchHMWInBatches(ctx, year, month, batchSize, func(data []models.AchHmwGleEntity) error {
			hmwRows += int64(len(data))
			return s.repo.UpsertHMWLocal(ctx, data)
		}); err != nil {
			hmwErr = fmt.Errorf("HMW fetch: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := s.repo.DeleteCLIKByYearMonth(ctx, year, month); err != nil {
			clikErr = fmt.Errorf("CLIK delete: %w", err)
			return
		}
		if err := s.repo.FetchCLIKInBatches(ctx, year, month, batchSize, func(data []models.ClikGleEntity) error {
			clikRows += int64(len(data))
			return s.repo.UpsertCLIKLocal(ctx, data)
		}); err != nil {
			clikErr = fmt.Errorf("CLIK fetch: %w", err)
		}
	}()
	wg.Wait()

	logs.Infof("📦 DW Sync %d-%02d: HMW=%d CLIK=%d rows", year, month, hmwRows, clikRows)

	if hmwErr != nil && clikErr != nil {
		return fmt.Errorf("%w; %v", hmwErr, clikErr)
	}
	return errors.Join(hmwErr, clikErr)
}

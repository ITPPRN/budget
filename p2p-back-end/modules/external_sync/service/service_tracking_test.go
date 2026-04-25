package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"p2p-back-end/modules/entities/models"
	_repo "p2p-back-end/modules/external_sync/repository"
)

// MockSyncTrackingRepository — implements SyncTrackingRepository for testing
type MockSyncTrackingRepository struct {
	mock.Mock
}

func (m *MockSyncTrackingRepository) CreateRun(ctx context.Context, run *models.SyncRunEntity) error {
	args := m.Called(ctx, run)
	// Simulate real repo: auto-assign ID if nil
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockSyncTrackingRepository) CompleteRun(
	ctx context.Context, id uuid.UUID, status string,
	rowsFetched, rowsInserted, rowsSkipped int64, errMsg string,
) error {
	args := m.Called(ctx, id, status, rowsFetched, rowsInserted, rowsSkipped, errMsg)
	return args.Error(0)
}

func (m *MockSyncTrackingRepository) IncrementRetry(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSyncTrackingRepository) GetLatest(ctx context.Context, jobType string) (*models.SyncRunEntity, error) {
	args := m.Called(ctx, jobType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SyncRunEntity), args.Error(1)
}

func (m *MockSyncTrackingRepository) GetRecent(ctx context.Context, limit int) ([]models.SyncRunEntity, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SyncRunEntity), args.Error(1)
}

func (m *MockSyncTrackingRepository) GetRunsByJobAndStatus(
	ctx context.Context, jobType, status string, limit int,
) ([]models.SyncRunEntity, error) {
	args := m.Called(ctx, jobType, status, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SyncRunEntity), args.Error(1)
}

func (m *MockSyncTrackingRepository) GetReconciliation(ctx context.Context, year string) (*_repo.ReconciliationResult, error) {
	args := m.Called(ctx, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*_repo.ReconciliationResult), args.Error(1)
}

func (m *MockSyncTrackingRepository) GetFailedRunsForRetry(ctx context.Context, within time.Duration, maxRetries int) ([]models.SyncRunEntity, error) {
	args := m.Called(ctx, within, maxRetries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SyncRunEntity), args.Error(1)
}

func (m *MockSyncTrackingRepository) ClearStaleRunningRuns(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSyncTrackingRepository) DeleteOldRunsByJobType(ctx context.Context, jobType string, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, jobType, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

// ─────────────────────────────────────────────
// Tests: SyncFromDW with tracking + delete-before-fetch
// ─────────────────────────────────────────────

// Helper: mock delete+fetch expectations for all year-months (success path)
func setupSuccessfulMonthExpectations(repo *MockExternalSyncRepository, actualSrv *MockActualService) {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())
	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			repo.On("DeleteHMWByYearMonth", mock.Anything, year, month).Return(nil)
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).Return(nil)
			repo.On("DeleteCLIKByYearMonth", mock.Anything, year, month).Return(nil)
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil)
		}
		yearStr := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}
}

func TestSyncFromDW_RecordsTrackingRun_Success(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	tracker := new(MockSyncTrackingRepository)

	svc := NewExternalSyncService(repo, actualSrv, tracker)

	setupSuccessfulMonthExpectations(repo, actualSrv)

	// Tracker expectations: CreateRun once, then CompleteRun once with SUCCESS
	tracker.On("CreateRun", mock.Anything, mock.MatchedBy(func(run *models.SyncRunEntity) bool {
		return run.JobType == models.SyncJobDW && run.TriggeredBy == "CRON"
	})).Return(nil)
	tracker.On("CompleteRun", mock.Anything, mock.AnythingOfType("uuid.UUID"),
		models.SyncStatusSuccess,
		mock.AnythingOfType("int64"), mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
		"").Return(nil)

	err := svc.SyncFromDW(context.Background())
	assert.NoError(t, err)

	tracker.AssertCalled(t, "CreateRun", mock.Anything, mock.Anything)
	tracker.AssertCalled(t, "CompleteRun", mock.Anything,
		mock.AnythingOfType("uuid.UUID"), models.SyncStatusSuccess,
		mock.AnythingOfType("int64"), mock.AnythingOfType("int64"), mock.AnythingOfType("int64"), "")
}

func TestSyncFromDW_RecordsTrackingRun_PartialOnFetchError(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	tracker := new(MockSyncTrackingRepository)

	svc := NewExternalSyncService(repo, actualSrv, tracker)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())
	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			repo.On("DeleteHMWByYearMonth", mock.Anything, year, month).Return(nil)
			// HMW fetch returns error → anyError=true, but loop continues
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).
				Return(assert.AnError)
			repo.On("DeleteCLIKByYearMonth", mock.Anything, year, month).Return(nil)
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil)
		}
		yearStr := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}

	tracker.On("CreateRun", mock.Anything, mock.Anything).Return(nil)
	// Completion should be PARTIAL when any error occurred
	tracker.On("CompleteRun", mock.Anything,
		mock.AnythingOfType("uuid.UUID"),
		models.SyncStatusPartial,
		mock.AnythingOfType("int64"), mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
		mock.AnythingOfType("string")).Return(nil)

	err := svc.SyncFromDW(context.Background())
	assert.NoError(t, err)
	tracker.AssertCalled(t, "CompleteRun", mock.Anything,
		mock.AnythingOfType("uuid.UUID"), models.SyncStatusPartial,
		mock.AnythingOfType("int64"), mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
		mock.AnythingOfType("string"))
}

// Verify delete is always called BEFORE fetch within the same month/table
func TestSyncFromDW_DeleteCalledBeforeFetch(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	tracker := new(MockSyncTrackingRepository)

	svc := NewExternalSyncService(repo, actualSrv, tracker)

	var callOrder []string
	targetYear := time.Now().Year() - 1 // first year processed
	targetMonth := 1                    // first month processed

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())
	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			y, m := year, month
			delHMW := repo.On("DeleteHMWByYearMonth", mock.Anything, y, m)
			fetchHMW := repo.On("FetchHMWInBatches", mock.Anything, y, m, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error"))
			delCLIK := repo.On("DeleteCLIKByYearMonth", mock.Anything, y, m)
			fetchCLIK := repo.On("FetchCLIKInBatches", mock.Anything, y, m, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error"))

			if y == targetYear && m == targetMonth {
				delHMW.Run(func(args mock.Arguments) { callOrder = append(callOrder, "DeleteHMW") })
				fetchHMW.Run(func(args mock.Arguments) { callOrder = append(callOrder, "FetchHMW") })
				delCLIK.Run(func(args mock.Arguments) { callOrder = append(callOrder, "DeleteCLIK") })
				fetchCLIK.Run(func(args mock.Arguments) { callOrder = append(callOrder, "FetchCLIK") })
			}
			delHMW.Return(nil)
			fetchHMW.Return(nil)
			delCLIK.Return(nil)
			fetchCLIK.Return(nil)
		}
		yearStr := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}

	tracker.On("CreateRun", mock.Anything, mock.Anything).Return(nil)
	tracker.On("CompleteRun", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := svc.SyncFromDW(context.Background())
	require.NoError(t, err)

	// Assert order for the target month: DeleteHMW → FetchHMW → DeleteCLIK → FetchCLIK
	require.GreaterOrEqual(t, len(callOrder), 4, "target month should record 4 ordered calls")
	assert.Equal(t, "DeleteHMW", callOrder[0])
	assert.Equal(t, "FetchHMW", callOrder[1])
	assert.Equal(t, "DeleteCLIK", callOrder[2])
	assert.Equal(t, "FetchCLIK", callOrder[3])
}

// If DeleteHMW fails, fetch should be skipped (continue to next month)
func TestSyncFromDW_DeleteHMWFailure_SkipsFetch(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	tracker := new(MockSyncTrackingRepository)

	svc := NewExternalSyncService(repo, actualSrv, tracker)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Fail HMW delete for one specific month; everything else succeeds
	failYear := currentYear
	failMonth := 1
	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			if year == failYear && month == failMonth {
				repo.On("DeleteHMWByYearMonth", mock.Anything, year, month).Return(assert.AnError)
				// FetchHMW should NOT be registered (service should skip it via continue)
				// CLIK side of this month should also be skipped due to `continue`
			} else {
				repo.On("DeleteHMWByYearMonth", mock.Anything, year, month).Return(nil)
				repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).Return(nil)
				repo.On("DeleteCLIKByYearMonth", mock.Anything, year, month).Return(nil)
				repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil)
			}
		}
		yearStr := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}

	tracker.On("CreateRun", mock.Anything, mock.Anything).Return(nil)
	tracker.On("CompleteRun", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := svc.SyncFromDW(context.Background())
	assert.NoError(t, err, "SyncFromDW should not return error — only log and continue")

	// Verify FetchHMW was NOT called for the failed month
	repo.AssertNotCalled(t, "FetchHMWInBatches", mock.Anything, failYear, failMonth,
		2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error"))
}

// Verify service handles nil tracker gracefully (backwards compat for tests)
func TestSyncFromDW_NilTracker_DoesNotPanic(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)

	svc := NewExternalSyncService(repo, actualSrv, nil)

	setupSuccessfulMonthExpectations(repo, actualSrv)

	err := svc.SyncFromDW(context.Background())
	assert.NoError(t, err)
}

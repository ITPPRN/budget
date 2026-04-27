package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

func TestMain(m *testing.M) {
	logs.Loginit()
	os.Exit(m.Run())
}

// --- Mocks ---

type MockExternalSyncRepository struct {
	mock.Mock
}

func (m *MockExternalSyncRepository) FetchHMWInBatches(ctx context.Context, year int, month int, batchSize int, handle func([]models.AchHmwGleEntity) error) error {
	args := m.Called(ctx, year, month, batchSize, handle)
	return args.Error(0)
}

func (m *MockExternalSyncRepository) FetchCLIKInBatches(ctx context.Context, year int, month int, batchSize int, handle func([]models.ClikGleEntity) error) error {
	args := m.Called(ctx, year, month, batchSize, handle)
	return args.Error(0)
}

func (m *MockExternalSyncRepository) UpsertHMWLocal(ctx context.Context, data []models.AchHmwGleEntity) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockExternalSyncRepository) UpsertCLIKLocal(ctx context.Context, data []models.ClikGleEntity) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockExternalSyncRepository) DeleteHMWByYearMonth(ctx context.Context, year int, month int) error {
	args := m.Called(ctx, year, month)
	return args.Error(0)
}

func (m *MockExternalSyncRepository) DeleteCLIKByYearMonth(ctx context.Context, year int, month int) error {
	args := m.Called(ctx, year, month)
	return args.Error(0)
}

type MockActualService struct {
	mock.Mock
}

func (m *MockActualService) SyncActualsDebug(ctx context.Context, targetDocNo string) error {
	args := m.Called(ctx, targetDocNo)
	return args.Error(0)
}
func (m *MockActualService) SyncActuals(ctx context.Context, year string, months []string) error {
	args := m.Called(ctx, year, months)
	return args.Error(0)
}

func (m *MockActualService) DeleteActualFacts(ctx context.Context, year string) error {
	args := m.Called(ctx, year)
	return args.Error(0)
}

func (m *MockActualService) GetRawDate(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockActualService) RefreshDataInventory(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// --- Tests ---

func TestNewExternalSyncService(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv, nil)
	assert.NotNil(t, svc)
}

func TestSyncFromDW_InvalidYear(t *testing.T) {
	svc := NewExternalSyncService(new(MockExternalSyncRepository), new(MockActualService), nil)
	err := svc.SyncFromDW(context.Background(), "abc", []string{"JAN"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid year")
}

func TestSyncFromDW_InvalidMonthCode(t *testing.T) {
	svc := NewExternalSyncService(new(MockExternalSyncRepository), new(MockActualService), nil)
	err := svc.SyncFromDW(context.Background(), "2026", []string{"NOPE"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid month code")
}

// expectMonthSuccess wires up mocks so HMW + CLIK delete/fetch all succeed for (year, month).
func expectMonthSuccess(repo *MockExternalSyncRepository, year, month int) {
	repo.On("DeleteHMWByYearMonth", mock.Anything, year, month).Return(nil)
	repo.On("FetchHMWInBatches", mock.Anything, year, month, 5000,
		mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).Return(nil)
	repo.On("DeleteCLIKByYearMonth", mock.Anything, year, month).Return(nil)
	repo.On("FetchCLIKInBatches", mock.Anything, year, month, 5000,
		mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil)
}

func TestSyncFromDW_SuccessSingleMonth(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv, nil)

	expectMonthSuccess(repo, 2026, 4)
	actualSrv.On("SyncActuals", mock.Anything, "2026", []string{"APR"}).Return(nil)
	actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncFromDW(context.Background(), "2026", []string{"APR"})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	actualSrv.AssertExpectations(t)
}

func TestSyncFromDW_MultipleMonths(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv, nil)

	expectMonthSuccess(repo, 2026, 1)
	expectMonthSuccess(repo, 2026, 2)
	expectMonthSuccess(repo, 2026, 3)
	actualSrv.On("SyncActuals", mock.Anything, "2026", []string{"JAN", "FEB", "MAR"}).Return(nil)
	actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncFromDW(context.Background(), "2026", []string{"JAN", "FEB", "MAR"})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	actualSrv.AssertExpectations(t)
}

func TestSyncFromDW_HMWFetchError_FailFast(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv, nil)

	repo.On("DeleteHMWByYearMonth", mock.Anything, 2026, 4).Return(nil)
	repo.On("FetchHMWInBatches", mock.Anything, 2026, 4, 5000,
		mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).
		Return(errors.New("connection reset"))
	// CLIK runs in parallel; allow it to succeed (or fail) — both legal.
	repo.On("DeleteCLIKByYearMonth", mock.Anything, 2026, 4).Return(nil).Maybe()
	repo.On("FetchCLIKInBatches", mock.Anything, 2026, 4, 5000,
		mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil).Maybe()

	err := svc.SyncFromDW(context.Background(), "2026", []string{"APR"})

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "HMW") || strings.Contains(err.Error(), "connection"))
	// Subsequent steps must not run after a month-level failure.
	actualSrv.AssertNotCalled(t, "SyncActuals", mock.Anything, mock.Anything, mock.Anything)
	actualSrv.AssertNotCalled(t, "RefreshDataInventory", mock.Anything)
}

func TestSyncFromDW_BothHMWAndCLIKFail(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv, nil)

	repo.On("DeleteHMWByYearMonth", mock.Anything, 2026, 4).Return(nil)
	repo.On("FetchHMWInBatches", mock.Anything, 2026, 4, 5000,
		mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).
		Return(errors.New("hmw boom"))
	repo.On("DeleteCLIKByYearMonth", mock.Anything, 2026, 4).Return(nil)
	repo.On("FetchCLIKInBatches", mock.Anything, 2026, 4, 5000,
		mock.AnythingOfType("func([]models.ClikGleEntity) error")).
		Return(errors.New("clik boom"))

	err := svc.SyncFromDW(context.Background(), "2026", []string{"APR"})

	assert.Error(t, err)
	// Both errors should surface in the joined message.
	assert.Contains(t, err.Error(), "hmw boom")
	assert.Contains(t, err.Error(), "clik boom")
}

func TestSyncFromDW_DefaultsToCurrentMonthWhenEmpty(t *testing.T) {
	// When months is empty/nil, the service should sync the current month only —
	// belt-and-braces fallback if a caller forgets to specify; cron always sends
	// an explicit month.
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv, nil)

	repo.On("DeleteHMWByYearMonth", mock.Anything, 2026, mock.AnythingOfType("int")).Return(nil)
	repo.On("FetchHMWInBatches", mock.Anything, 2026, mock.AnythingOfType("int"), 5000,
		mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).Return(nil)
	repo.On("DeleteCLIKByYearMonth", mock.Anything, 2026, mock.AnythingOfType("int")).Return(nil)
	repo.On("FetchCLIKInBatches", mock.Anything, 2026, mock.AnythingOfType("int"), 5000,
		mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil)
	actualSrv.On("SyncActuals", mock.Anything, "2026", mock.AnythingOfType("[]string")).Return(nil)
	actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncFromDW(context.Background(), "2026", nil)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSyncFromDW_RefreshInventoryNonFatal(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv, nil)

	expectMonthSuccess(repo, 2026, 4)
	actualSrv.On("SyncActuals", mock.Anything, "2026", []string{"APR"}).Return(nil)
	actualSrv.On("RefreshDataInventory", mock.Anything).Return(errors.New("inventory blew up"))

	err := svc.SyncFromDW(context.Background(), "2026", []string{"APR"})

	// Inventory refresh failure is logged but does not fail the job — the raw
	// data is already in place; the inventory widget can be refreshed next run.
	assert.NoError(t, err)
}

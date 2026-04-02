package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

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

type MockActualService struct {
	mock.Mock
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

	svc := NewExternalSyncService(repo, actualSrv)

	assert.NotNil(t, svc)
}

func TestSyncFromDW_Success(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Expect calls for each year-month from 2026 to now
	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).Return(nil)
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil)
		}
		yearStr := fmt.Sprintf("%d", year)
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}

	err := svc.SyncFromDW(context.Background())

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	actualSrv.AssertExpectations(t)
}

func TestSyncFromDW_HMWFetchError_ContinuesProcessing(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			// HMW fails for all months
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).
				Return(errors.New("connection timeout"))
			// CLIK succeeds
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).
				Return(nil)
		}
		yearStr := fmt.Sprintf("%d", year)
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}

	// SyncFromDW logs errors but does NOT return error
	err := svc.SyncFromDW(context.Background())

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSyncFromDW_CLIKFetchError_ContinuesProcessing(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).Return(nil)
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).
				Return(errors.New("db error"))
		}
		yearStr := fmt.Sprintf("%d", year)
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}

	err := svc.SyncFromDW(context.Background())

	assert.NoError(t, err)
}

func TestSyncFromDW_SyncActualsError_ContinuesProcessing(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).Return(nil)
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil)
		}
		yearStr := fmt.Sprintf("%d", year)
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(errors.New("sync failed"))
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(errors.New("refresh failed"))
	}

	err := svc.SyncFromDW(context.Background())

	assert.NoError(t, err)
}

func TestSyncFromDW_AllErrors_StillReturnsNil(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).
				Return(errors.New("hmw error"))
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).
				Return(errors.New("clik error"))
		}
		yearStr := fmt.Sprintf("%d", year)
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(errors.New("sync error"))
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(errors.New("refresh error"))
	}

	err := svc.SyncFromDW(context.Background())

	// SyncFromDW always returns nil (it only logs errors)
	assert.NoError(t, err)
}

func TestSyncFromDW_FetchHMW_CallsUpsertViaHandler(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	testData := []models.AchHmwGleEntity{
		{ID: 1, GLAccountNo: "5100", Company: "ACH"},
		{ID: 2, GLAccountNo: "5200", Company: "HMW"},
	}

	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			// Capture and invoke the handler to verify UpsertHMWLocal is called
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).
				Run(func(args mock.Arguments) {
					handler := args.Get(4).(func([]models.AchHmwGleEntity) error)
					handler(testData)
				}).Return(nil)
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).Return(nil)
		}
		yearStr := fmt.Sprintf("%d", year)
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}

	repo.On("UpsertHMWLocal", mock.Anything, testData).Return(nil)

	err := svc.SyncFromDW(context.Background())

	assert.NoError(t, err)
	repo.AssertCalled(t, "UpsertHMWLocal", mock.Anything, testData)
}

func TestSyncFromDW_FetchCLIK_CallsUpsertViaHandler(t *testing.T) {
	repo := new(MockExternalSyncRepository)
	actualSrv := new(MockActualService)
	svc := NewExternalSyncService(repo, actualSrv)

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	testData := []models.ClikGleEntity{
		{ID: 1, GLAccountNo: "6100", Company: "CLIK"},
	}

	for year := currentYear - 1; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		for month := 1; month <= endMonth; month++ {
			repo.On("FetchHMWInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.AchHmwGleEntity) error")).Return(nil)
			repo.On("FetchCLIKInBatches", mock.Anything, year, month, 2000, mock.AnythingOfType("func([]models.ClikGleEntity) error")).
				Run(func(args mock.Arguments) {
					handler := args.Get(4).(func([]models.ClikGleEntity) error)
					handler(testData)
				}).Return(nil)
		}
		yearStr := fmt.Sprintf("%d", year)
		actualSrv.On("SyncActuals", mock.Anything, yearStr, []string{}).Return(nil)
		actualSrv.On("RefreshDataInventory", mock.Anything).Return(nil)
	}

	repo.On("UpsertCLIKLocal", mock.Anything, testData).Return(nil)

	err := svc.SyncFromDW(context.Background())

	assert.NoError(t, err)
	repo.AssertCalled(t, "UpsertCLIKLocal", mock.Anything, testData)
}

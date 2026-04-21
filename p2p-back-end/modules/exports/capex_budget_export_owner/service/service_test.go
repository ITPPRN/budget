package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xuri/excelize/v2"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/exports/capex_budget_export_owner/repository"
)

// --- Mock Repository ---

type MockOwnerCapexRepository struct {
	mock.Mock
}

func (m *MockOwnerCapexRepository) GetOwnerCapexData(ctx context.Context, filter map[string]interface{}) ([]models.OwnerCapexBudgetExportDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.OwnerCapexBudgetExportDTO), args.Error(1)
}

// Compile-time check that mock satisfies the interface.
var _ repository.OwnerCapexRepository = (*MockOwnerCapexRepository)(nil)

// --- Mock OwnerService ---

type MockOwnerService struct {
	mock.Mock
}

func (m *MockOwnerService) GetDashboardSummary(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) (*models.OwnerDashboardSummaryDTO, error) {
	return nil, nil
}

func (m *MockOwnerService) GetActualTransactions(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	return nil, nil
}

func (m *MockOwnerService) GetActualDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	return nil, nil
}

func (m *MockOwnerService) GetBudgetDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	return nil, nil
}

func (m *MockOwnerService) GetFilterOptions(ctx context.Context, user *models.UserInfo) (interface{}, error) {
	return nil, nil
}

func (m *MockOwnerService) GetOrganizationStructure(ctx context.Context, user *models.UserInfo) ([]models.OrganizationDTO, error) {
	return nil, nil
}

func (m *MockOwnerService) GetOwnerFilterLists(ctx context.Context, user *models.UserInfo) (*models.OwnerFilterListsDTO, error) {
	return nil, nil
}

func (m *MockOwnerService) GetActualYears(ctx context.Context, user *models.UserInfo) ([]string, error) {
	return nil, nil
}

func (m *MockOwnerService) InjectPermissions(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) map[string]interface{} {
	args := m.Called(ctx, user, filter)
	return args.Get(0).(map[string]interface{})
}

// Compile-time check that mock satisfies the interface.
var _ models.OwnerService = (*MockOwnerService)(nil)

// --- Helpers ---

func makeUser() *models.UserInfo {
	return &models.UserInfo{
		ID:       "user-1",
		Username: "testuser",
		Name:     "Test User",
		Roles:    []string{"OWNER"},
	}
}

func makeFilter() map[string]interface{} {
	return map[string]interface{}{
		"entities": []string{"ACG"},
		"year":     "2025",
	}
}

// --- Tests ---

func TestNewService(t *testing.T) {
	repo := new(MockOwnerCapexRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)
	assert.NotNil(t, svc)
}

func TestExport_Success_WithData(t *testing.T) {
	repo := new(MockOwnerCapexRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	// InjectPermissions returns filter as-is
	ownerSrv.On("InjectPermissions", mock.Anything, user, filter).Return(filter)

	rows := []models.OwnerCapexBudgetExportDTO{
		{
			Entity:        "ACG",
			Branch:        "HQ",
			Department:    "IT",
			CapexNo:       "CX001",
			CapexName:     "Server Upgrade",
			CapexCategory: "Hardware",
			Budget:        decimal.NewFromFloat(500000),
			Actual:        decimal.NewFromFloat(200000),
			Remaining:     decimal.NewFromFloat(300000),
			Percentage:    40.0,
		},
		{
			Entity:        "ACG",
			Branch:        "HQ",
			Department:    "HR",
			CapexNo:       "CX002",
			CapexName:     "Office Renovation",
			CapexCategory: "Facility",
			Budget:        decimal.NewFromFloat(100000),
			Actual:        decimal.NewFromFloat(120000),
			Remaining:     decimal.NewFromFloat(-20000),
			Percentage:    120.0,
		},
	}

	repo.On("GetOwnerCapexData", mock.Anything, filter).Return(rows, nil)

	data, filename, err := svc.ExportOwnerCapexExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Owner_Capex_Budget_")
	assert.Contains(t, filename, ".xlsx")

	// Parse the generated Excel and verify content
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Verify headers (10 columns)
	h1, _ := f.GetCellValue(sheet, "A1")
	assert.Equal(t, "Entity", h1)
	h2, _ := f.GetCellValue(sheet, "B1")
	assert.Equal(t, "Branch", h2)
	h3, _ := f.GetCellValue(sheet, "C1")
	assert.Equal(t, "Department", h3)
	h4, _ := f.GetCellValue(sheet, "D1")
	assert.Equal(t, "CAPEX NO.", h4)
	h5, _ := f.GetCellValue(sheet, "E1")
	assert.Equal(t, "CAPEX Name", h5)
	h6, _ := f.GetCellValue(sheet, "F1")
	assert.Equal(t, "CAPEX Category", h6)
	h7, _ := f.GetCellValue(sheet, "G1")
	assert.Equal(t, "Budget", h7)
	h8, _ := f.GetCellValue(sheet, "H1")
	assert.Equal(t, "Actual", h8)
	h9, _ := f.GetCellValue(sheet, "I1")
	assert.Equal(t, "Remaining", h9)
	h10, _ := f.GetCellValue(sheet, "J1")
	assert.Equal(t, "(%)", h10)

	// Verify row 2 (first data row)
	v, _ := f.GetCellValue(sheet, "A2")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "B2")
	assert.Equal(t, "HQ", v)
	v, _ = f.GetCellValue(sheet, "C2")
	assert.Equal(t, "IT", v)
	v, _ = f.GetCellValue(sheet, "D2")
	assert.Equal(t, "CX001", v)
	v, _ = f.GetCellValue(sheet, "E2")
	assert.Equal(t, "Server Upgrade", v)
	v, _ = f.GetCellValue(sheet, "F2")
	assert.Equal(t, "Hardware", v)

	// Verify row 3 (second data row)
	v, _ = f.GetCellValue(sheet, "A3")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "B3")
	assert.Equal(t, "HQ", v)
	v, _ = f.GetCellValue(sheet, "C3")
	assert.Equal(t, "HR", v)
	v, _ = f.GetCellValue(sheet, "D3")
	assert.Equal(t, "CX002", v)

	// Verify no row 4
	v, _ = f.GetCellValue(sheet, "A4")
	assert.Equal(t, "", v)

	allRows, _ := f.GetRows(sheet)
	assert.Equal(t, 3, len(allRows)) // 1 header + 2 data

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExport_Success_EmptyData(t *testing.T) {
	repo := new(MockOwnerCapexRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, filter).Return(filter)
	repo.On("GetOwnerCapexData", mock.Anything, filter).
		Return([]models.OwnerCapexBudgetExportDTO{}, nil)

	data, filename, err := svc.ExportOwnerCapexExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Owner_Capex_Budget_")
	assert.Contains(t, filename, ".xlsx")

	// Verify the Excel has headers but no data rows
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	allRows, _ := f.GetRows(sheet)
	assert.Equal(t, 1, len(allRows))
	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExport_RepoError(t *testing.T) {
	repo := new(MockOwnerCapexRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, filter).Return(filter)
	repo.On("GetOwnerCapexData", mock.Anything, filter).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportOwnerCapexExcel(context.Background(), user, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)
	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExport_FilenameFormat(t *testing.T) {
	repo := new(MockOwnerCapexRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, filter).Return(filter)
	repo.On("GetOwnerCapexData", mock.Anything, filter).
		Return([]models.OwnerCapexBudgetExportDTO{}, nil)

	_, filename, err := svc.ExportOwnerCapexExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.Regexp(t, `^Owner_Capex_Budget_\d{14}\.xlsx$`, filename)
	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExport_SheetName(t *testing.T) {
	repo := new(MockOwnerCapexRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, filter).Return(filter)
	repo.On("GetOwnerCapexData", mock.Anything, filter).
		Return([]models.OwnerCapexBudgetExportDTO{}, nil)

	data, _, err := svc.ExportOwnerCapexExcel(context.Background(), user, filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	assert.Equal(t, "Capex Budget Status", sheet)
	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

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
	"p2p-back-end/modules/exports/budgetvsactual_export_owner/repository"
)

// Compile-time checks
var _ repository.OwnerBudgetVsActualRepository = (*MockOwnerBudgetVsActualRepository)(nil)
var _ models.OwnerService = (*MockOwnerService)(nil)

// --- Mock Repository ---

type MockOwnerBudgetVsActualRepository struct {
	mock.Mock
}

func (m *MockOwnerBudgetVsActualRepository) GetOwnerBudgetVsActual(ctx context.Context, filter map[string]interface{}) ([]models.BudgetVsActualExportDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetVsActualExportDTO), args.Error(1)
}

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

func (m *MockOwnerService) GetAdminPermittedMonths(ctx context.Context) []string {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

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

func makeSampleRows() []models.BudgetVsActualExportDTO {
	return []models.BudgetVsActualExportDTO{
		{
			Entity:     "ACG",
			Branch:     "HQ",
			Department: "IT",
			Type:       "Budget",
			Group:      "OpEx",
			Group2:     "General",
			Group3:     "Admin",
			ConsoGL:    "5100",
			GLName:     "Office Supplies",
			MonthsAmounts: map[string]interface{}{
				"JAN": decimal.NewFromFloat(1000),
				"FEB": decimal.NewFromFloat(2000),
				"MAR": decimal.NewFromFloat(1500),
			},
			YearTotal: decimal.NewFromFloat(4500),
		},
		{
			Entity:     "ACG",
			Branch:     "HQ",
			Department: "IT",
			Type:       "Actual",
			Group:      "OpEx",
			Group2:     "General",
			Group3:     "Admin",
			ConsoGL:    "5100",
			GLName:     "Office Supplies",
			MonthsAmounts: map[string]interface{}{
				"JAN": decimal.NewFromFloat(900),
				"FEB": decimal.NewFromFloat(2100),
			},
			YearTotal: decimal.NewFromFloat(3000),
		},
	}
}

// --- Tests ---

func TestNewService(t *testing.T) {
	repo := new(MockOwnerBudgetVsActualRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)
	assert.NotNil(t, svc)
}

func TestExport_Success_WithData(t *testing.T) {
	repo := new(MockOwnerBudgetVsActualRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()
	rows := makeSampleRows()

	// InjectPermissions is called after SanitizeFilter; use mock.Anything for the sanitized filter
	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(mock.Anything).
		Run(func(args mock.Arguments) {
			// return the filter as-is
		})
	// Override: make InjectPermissions return the filter argument unchanged
	ownerSrv.ExpectedCalls = nil
	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(map[string]interface{}{"entities": []string{"ACG"}, "year": "2025"})

	repo.On("GetOwnerBudgetVsActual", mock.Anything, mock.Anything).
		Return(rows, nil)

	data, filename, err := svc.ExportOwnerBudgetVsActualExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Owner_Budget_vs_Actual_")
	assert.Contains(t, filename, ".xlsx")

	// Parse Excel
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Verify headers (22 columns)
	expectedHeaders := []string{
		"Entity", "Branch", "Department", "Type",
		"Group", "Group2", "Group3", "Conso GL", "GL Name",
		"JAN", "FEB", "MAR", "APR", "MAY", "JUN",
		"JUL", "AUG", "SEP", "OCT", "NOV", "DEC",
		"YEARTOTAL",
	}
	allRows, _ := f.GetRows(sheet)
	assert.GreaterOrEqual(t, len(allRows), 3) // header + 2 data rows
	for i, expected := range expectedHeaders {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		v, _ := f.GetCellValue(sheet, colName+"1")
		assert.Equal(t, expected, v, "header column %s", colName)
	}

	// Verify first data row
	v, _ := f.GetCellValue(sheet, "A2")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "B2")
	assert.Equal(t, "HQ", v)
	v, _ = f.GetCellValue(sheet, "C2")
	assert.Equal(t, "IT", v)
	v, _ = f.GetCellValue(sheet, "D2")
	assert.Equal(t, "Budget", v)
	v, _ = f.GetCellValue(sheet, "E2")
	assert.Equal(t, "OpEx", v)
	v, _ = f.GetCellValue(sheet, "F2")
	assert.Equal(t, "General", v)
	v, _ = f.GetCellValue(sheet, "G2")
	assert.Equal(t, "Admin", v)
	v, _ = f.GetCellValue(sheet, "H2")
	assert.Equal(t, "5100", v)
	v, _ = f.GetCellValue(sheet, "I2")
	assert.Equal(t, "Office Supplies", v)

	// Verify JAN amount in row 2 (column J)
	v, _ = f.GetCellValue(sheet, "J2")
	assert.Equal(t, "1000", v)

	// Verify YEARTOTAL in row 2 (column V)
	v, _ = f.GetCellValue(sheet, "V2")
	assert.Equal(t, "4500", v)

	// Verify second data row type
	v, _ = f.GetCellValue(sheet, "D3")
	assert.Equal(t, "Actual", v)

	// Verify YEARTOTAL in row 3
	v, _ = f.GetCellValue(sheet, "V3")
	assert.Equal(t, "3000", v)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExport_Success_EmptyData(t *testing.T) {
	repo := new(MockOwnerBudgetVsActualRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(map[string]interface{}{"entities": []string{"ACG"}, "year": "2025"})

	repo.On("GetOwnerBudgetVsActual", mock.Anything, mock.Anything).
		Return([]models.BudgetVsActualExportDTO{}, nil)

	data, filename, err := svc.ExportOwnerBudgetVsActualExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".xlsx")

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
	repo := new(MockOwnerBudgetVsActualRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(map[string]interface{}{"entities": []string{"ACG"}, "year": "2025"})

	repo.On("GetOwnerBudgetVsActual", mock.Anything, mock.Anything).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportOwnerBudgetVsActualExcel(context.Background(), user, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)
	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExport_FilenameFormat(t *testing.T) {
	repo := new(MockOwnerBudgetVsActualRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(map[string]interface{}{"entities": []string{"ACG"}, "year": "2025"})

	repo.On("GetOwnerBudgetVsActual", mock.Anything, mock.Anything).
		Return([]models.BudgetVsActualExportDTO{}, nil)

	_, filename, err := svc.ExportOwnerBudgetVsActualExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.Regexp(t, `^Owner_Budget_vs_Actual_\d{14}\.xlsx$`, filename)
	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExport_SheetName(t *testing.T) {
	repo := new(MockOwnerBudgetVsActualRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(map[string]interface{}{"entities": []string{"ACG"}, "year": "2025"})

	repo.On("GetOwnerBudgetVsActual", mock.Anything, mock.Anything).
		Return([]models.BudgetVsActualExportDTO{}, nil)

	data, _, err := svc.ExportOwnerBudgetVsActualExcel(context.Background(), user, filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	assert.Equal(t, "Owner Budget vs Actual", sheet)
	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

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
	"p2p-back-end/modules/exports/budget_detail_export_owner/repository"
)

// --- Mock Repository ---

type MockOwnerBudgetExportRepository struct {
	mock.Mock
}

func (m *MockOwnerBudgetExportRepository) GetOwnerBudgetExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetExportDTO, error) {
	args := m.Called(ctx, user, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetExportDTO), args.Error(1)
}

// Compile-time check
var _ repository.OwnerBudgetExportRepository = (*MockOwnerBudgetExportRepository)(nil)

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

// --- Helpers ---

func makeUser() *models.UserInfo {
	return &models.UserInfo{
		ID:       "user-1",
		Username: "testuser",
		Name:     "Test User",
		Roles:    []string{"ADMIN"},
	}
}

func makeFilter() map[string]interface{} {
	return map[string]interface{}{
		"entities": []string{"ACG"},
		"year":     "2025",
	}
}

func makeBudgetRows() []models.BudgetExportDTO {
	return []models.BudgetExportDTO{
		{
			Entity:     "ACG",
			Branch:     "HQ",
			Department: "IT",
			Group:      "OpEx",
			Group2:     "General",
			Group3:     "Admin",
			ConsoGL:    "5100",
			GLName:     "Supplies",
			MonthsAmounts: map[string]interface{}{
				"JAN": decimal.NewFromFloat(100),
				"FEB": decimal.NewFromFloat(200),
			},
			YearTotal: decimal.NewFromFloat(300),
		},
		{
			Entity:     "HMW",
			Branch:     "BKK",
			Department: "HR",
			Group:      "OpEx",
			Group2:     "Personnel",
			Group3:     "Training",
			ConsoGL:    "5200",
			GLName:     "Training",
			MonthsAmounts: map[string]interface{}{
				"MAR": decimal.NewFromFloat(500),
				"APR": decimal.NewFromFloat(600),
			},
			YearTotal: decimal.NewFromFloat(1100),
		},
	}
}

// --- Tests ---

func TestNewService(t *testing.T) {
	repo := new(MockOwnerBudgetExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)
	assert.NotNil(t, svc)
}

func TestExportOwnerBudgetDetailExcel_Success_WithData(t *testing.T) {
	repo := new(MockOwnerBudgetExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()
	rows := makeBudgetRows()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerBudgetExportDetails", mock.Anything, user, mock.Anything).
		Return(rows, nil)

	data, filename, err := svc.ExportOwnerBudgetDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Owner_Budget_Detail_")
	assert.Contains(t, filename, ".xlsx")

	// Parse Excel
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Verify all 21 headers
	h1, _ := f.GetCellValue(sheet, "A1")
	assert.Equal(t, "Entity", h1)
	h2, _ := f.GetCellValue(sheet, "B1")
	assert.Equal(t, "Branch", h2)
	h3, _ := f.GetCellValue(sheet, "C1")
	assert.Equal(t, "Department", h3)
	h4, _ := f.GetCellValue(sheet, "D1")
	assert.Equal(t, "GROUP1", h4)
	h5, _ := f.GetCellValue(sheet, "E1")
	assert.Equal(t, "GROUP2", h5)
	h6, _ := f.GetCellValue(sheet, "F1")
	assert.Equal(t, "GROUP3", h6)
	h7, _ := f.GetCellValue(sheet, "G1")
	assert.Equal(t, "GL Code", h7)
	h8, _ := f.GetCellValue(sheet, "H1")
	assert.Equal(t, "GL Name", h8)
	h9, _ := f.GetCellValue(sheet, "I1")
	assert.Equal(t, "JAN", h9)
	h10, _ := f.GetCellValue(sheet, "J1")
	assert.Equal(t, "FEB", h10)
	h11, _ := f.GetCellValue(sheet, "K1")
	assert.Equal(t, "MAR", h11)
	h12, _ := f.GetCellValue(sheet, "L1")
	assert.Equal(t, "APR", h12)
	h17, _ := f.GetCellValue(sheet, "Q1")
	assert.Equal(t, "SEP", h17)
	h21, _ := f.GetCellValue(sheet, "U1")
	assert.Equal(t, "YEARTOTAL", h21)

	// Verify row 2 data
	v, _ := f.GetCellValue(sheet, "A2")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "B2")
	assert.Equal(t, "HQ", v)
	v, _ = f.GetCellValue(sheet, "C2")
	assert.Equal(t, "IT", v)
	v, _ = f.GetCellValue(sheet, "D2")
	assert.Equal(t, "OpEx", v)
	v, _ = f.GetCellValue(sheet, "G2")
	assert.Equal(t, "5100", v)
	v, _ = f.GetCellValue(sheet, "H2")
	assert.Equal(t, "Supplies", v)

	// JAN=100, FEB=200
	v, _ = f.GetCellValue(sheet, "I2")
	assert.Equal(t, "100", v)
	v, _ = f.GetCellValue(sheet, "J2")
	assert.Equal(t, "200", v)

	// YEARTOTAL=300
	v, _ = f.GetCellValue(sheet, "U2")
	assert.Equal(t, "300", v)

	// Verify row 3
	v, _ = f.GetCellValue(sheet, "A3")
	assert.Equal(t, "HMW", v)
	v, _ = f.GetCellValue(sheet, "C3")
	assert.Equal(t, "HR", v)

	// No row 4
	v, _ = f.GetCellValue(sheet, "A4")
	assert.Equal(t, "", v)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExportOwnerBudgetDetailExcel_Success_EmptyData(t *testing.T) {
	repo := new(MockOwnerBudgetExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerBudgetExportDetails", mock.Anything, user, mock.Anything).
		Return([]models.BudgetExportDTO{}, nil)

	data, filename, err := svc.ExportOwnerBudgetDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Owner_Budget_Detail_")
	assert.Contains(t, filename, ".xlsx")

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	allRows, _ := f.GetRows(sheet)
	assert.Equal(t, 1, len(allRows)) // header only

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExportOwnerBudgetDetailExcel_RepoError(t *testing.T) {
	repo := new(MockOwnerBudgetExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerBudgetExportDetails", mock.Anything, user, mock.Anything).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportOwnerBudgetDetailExcel(context.Background(), user, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExportOwnerBudgetDetailExcel_FilenameFormat(t *testing.T) {
	repo := new(MockOwnerBudgetExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerBudgetExportDetails", mock.Anything, user, mock.Anything).
		Return([]models.BudgetExportDTO{}, nil)

	_, filename, err := svc.ExportOwnerBudgetDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.Regexp(t, `^Owner_Budget_Detail_\d{14}\.xlsx$`, filename)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExportOwnerBudgetDetailExcel_SheetName(t *testing.T) {
	repo := new(MockOwnerBudgetExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerBudgetExportDetails", mock.Anything, user, mock.Anything).
		Return([]models.BudgetExportDTO{}, nil)

	data, _, err := svc.ExportOwnerBudgetDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	assert.Equal(t, "Budget Details", sheet)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

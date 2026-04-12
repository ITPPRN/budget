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
	"p2p-back-end/modules/exports/actual_detail_export_owner/repository"
)

// --- Mock Repository ---

type MockOwnerActualExportRepository struct {
	mock.Mock
}

func (m *MockOwnerActualExportRepository) GetOwnerActualExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualExportDTO, error) {
	args := m.Called(ctx, user, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualExportDTO), args.Error(1)
}

// Compile-time check
var _ repository.OwnerActualExportRepository = (*MockOwnerActualExportRepository)(nil)

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

// --- Tests ---

func TestNewService(t *testing.T) {
	repo := new(MockOwnerActualExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)
	assert.NotNil(t, svc)
}

func TestExportOwnerActualDetailExcel_Success_WithData(t *testing.T) {
	repo := new(MockOwnerActualExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	rows := []models.ActualExportDTO{
		{
			Entity:      "ACG",
			Branch:      "HQ",
			Department:  "IT",
			Group:       "OpEx",
			Group2:      "General",
			Group3:      "Admin",
			ConsoGL:     "5100",
			GLName:      "Office Supplies",
			DocumentNo:  "INV-001",
			Amount:      decimal.NewFromFloat(15000.50),
			VendorName:  "ABC Corp",
			Description: "Monthly supplies",
			PostingDate: "2025-03-15",
		},
		{
			Entity:      "HMW",
			Branch:      "BKK",
			Department:  "HR",
			Group:       "OpEx",
			Group2:      "Personnel",
			Group3:      "Training",
			ConsoGL:     "5200",
			GLName:      "Training Expense",
			DocumentNo:  "INV-002",
			Amount:      decimal.NewFromFloat(25000.00),
			VendorName:  "Training Co.",
			Description: "Q1 Training",
			PostingDate: "2025-03-20",
		},
	}

	// InjectPermissions is called with sanitized filter; passthrough
	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerActualExportDetails", mock.Anything, user, mock.Anything).
		Return(rows, nil)

	data, filename, err := svc.ExportOwnerActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Owner_Actual_Detail_")
	assert.Contains(t, filename, ".xlsx")

	// Parse Excel
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Verify headers
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
	assert.Equal(t, "Document No.", h9)
	h10, _ := f.GetCellValue(sheet, "J1")
	assert.Equal(t, "Amount", h10)
	h11, _ := f.GetCellValue(sheet, "K1")
	assert.Equal(t, "Vendor", h11)
	h12, _ := f.GetCellValue(sheet, "L1")
	assert.Equal(t, "Description", h12)
	h13, _ := f.GetCellValue(sheet, "M1")
	assert.Equal(t, "Date", h13)

	// Verify row 2
	v, _ := f.GetCellValue(sheet, "A2")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "B2")
	assert.Equal(t, "HQ", v)
	v, _ = f.GetCellValue(sheet, "C2")
	assert.Equal(t, "IT", v)
	v, _ = f.GetCellValue(sheet, "G2")
	assert.Equal(t, "5100", v)
	v, _ = f.GetCellValue(sheet, "I2")
	assert.Equal(t, "INV-001", v)
	v, _ = f.GetCellValue(sheet, "K2")
	assert.Equal(t, "ABC Corp", v)
	v, _ = f.GetCellValue(sheet, "M2")
	assert.Equal(t, "2025-03-15", v)

	// Verify row 3
	v, _ = f.GetCellValue(sheet, "A3")
	assert.Equal(t, "HMW", v)
	v, _ = f.GetCellValue(sheet, "C3")
	assert.Equal(t, "HR", v)
	v, _ = f.GetCellValue(sheet, "I3")
	assert.Equal(t, "INV-002", v)

	// No row 4
	v, _ = f.GetCellValue(sheet, "A4")
	assert.Equal(t, "", v)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExportOwnerActualDetailExcel_Success_EmptyData(t *testing.T) {
	repo := new(MockOwnerActualExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerActualExportDetails", mock.Anything, user, mock.Anything).
		Return([]models.ActualExportDTO{}, nil)

	data, filename, err := svc.ExportOwnerActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Owner_Actual_Detail_")
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

func TestExportOwnerActualDetailExcel_RepoError(t *testing.T) {
	repo := new(MockOwnerActualExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerActualExportDetails", mock.Anything, user, mock.Anything).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportOwnerActualDetailExcel(context.Background(), user, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExportOwnerActualDetailExcel_FilenameFormat(t *testing.T) {
	repo := new(MockOwnerActualExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerActualExportDetails", mock.Anything, user, mock.Anything).
		Return([]models.ActualExportDTO{}, nil)

	_, filename, err := svc.ExportOwnerActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.Regexp(t, `^Owner_Actual_Detail_\d{14}\.xlsx$`, filename)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

func TestExportOwnerActualDetailExcel_SheetName(t *testing.T) {
	repo := new(MockOwnerActualExportRepository)
	ownerSrv := new(MockOwnerService)
	svc := NewService(repo, ownerSrv)

	user := makeUser()
	filter := makeFilter()

	ownerSrv.On("InjectPermissions", mock.Anything, user, mock.Anything).
		Return(filter)
	repo.On("GetOwnerActualExportDetails", mock.Anything, user, mock.Anything).
		Return([]models.ActualExportDTO{}, nil)

	data, _, err := svc.ExportOwnerActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	assert.Equal(t, "Actual Details", sheet)

	repo.AssertExpectations(t)
	ownerSrv.AssertExpectations(t)
}

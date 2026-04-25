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
)

// --- Mock ---

type MockActualExportRepository struct {
	mock.Mock
}

func (m *MockActualExportRepository) GetActualExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualExportDTO, error) {
	args := m.Called(ctx, user, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualExportDTO), args.Error(1)
}

// --- Helper ---

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
	repo := new(MockActualExportRepository)
	svc := NewService(repo)
	assert.NotNil(t, svc)
}

func TestExportActualDetailExcel_Success_EmptyData(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := makeFilter()

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return([]models.ActualExportDTO{}, nil)

	data, filename, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Actual_Detail_")
	assert.Contains(t, filename, ".xlsx")

	// Verify the Excel has headers but no data rows
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	allRows, _ := f.GetRows(sheet)
	// Only 1 header row, no data rows
	assert.Equal(t, 1, len(allRows))
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_Success_WithData(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

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
			Entity:      "ACG",
			Branch:      "HQ",
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
		{
			Entity:      "HMW",
			Branch:      "BKK",
			Department:  "Sales",
			Group:       "Revenue",
			Group2:      "",
			Group3:      "",
			ConsoGL:     "4100",
			GLName:      "Sales Revenue",
			DocumentNo:  "INV-003",
			Amount:      decimal.NewFromFloat(-5000.75),
			VendorName:  "",
			Description: "Credit note",
			PostingDate: "2025-02-10",
		},
	}

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return(rows, nil)

	data, filename, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Actual_Detail_")
	assert.Contains(t, filename, ".xlsx")

	// Parse the generated Excel and verify content
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

	// Verify row 2 (first data row)
	v, _ := f.GetCellValue(sheet, "A2")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "B2")
	assert.Equal(t, "HQ", v)
	v, _ = f.GetCellValue(sheet, "C2")
	assert.Equal(t, "IT", v)
	v, _ = f.GetCellValue(sheet, "D2")
	assert.Equal(t, "OpEx", v)
	v, _ = f.GetCellValue(sheet, "E2")
	assert.Equal(t, "General", v)
	v, _ = f.GetCellValue(sheet, "F2")
	assert.Equal(t, "Admin", v)
	v, _ = f.GetCellValue(sheet, "G2")
	assert.Equal(t, "5100", v)
	v, _ = f.GetCellValue(sheet, "H2")
	assert.Equal(t, "Office Supplies", v)
	v, _ = f.GetCellValue(sheet, "I2")
	assert.Equal(t, "INV-001", v)
	v, _ = f.GetCellValue(sheet, "K2")
	assert.Equal(t, "ABC Corp", v)
	v, _ = f.GetCellValue(sheet, "L2")
	assert.Equal(t, "Monthly supplies", v)
	v, _ = f.GetCellValue(sheet, "M2")
	assert.Equal(t, "2025-03-15", v)

	// Verify row 3 (second data row)
	v, _ = f.GetCellValue(sheet, "A3")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "C3")
	assert.Equal(t, "HR", v)
	v, _ = f.GetCellValue(sheet, "I3")
	assert.Equal(t, "INV-002", v)

	// Verify row 4 (third data row - HMW)
	v, _ = f.GetCellValue(sheet, "A4")
	assert.Equal(t, "HMW", v)
	v, _ = f.GetCellValue(sheet, "C4")
	assert.Equal(t, "Sales", v)
	v, _ = f.GetCellValue(sheet, "I4")
	assert.Equal(t, "INV-003", v)

	// Verify no row 5 (only 3 data rows)
	v, _ = f.GetCellValue(sheet, "A5")
	assert.Equal(t, "", v)

	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_RepoError(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := makeFilter()

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_NilFilter(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	var filter map[string]interface{}

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return([]models.ActualExportDTO{}, nil)

	data, filename, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".xlsx")
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_NilUser(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetActualExportDetails", mock.Anything, (*models.UserInfo)(nil), filter).
		Return([]models.ActualExportDTO{}, nil)

	data, filename, err := svc.ExportActualDetailExcel(context.Background(), nil, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".xlsx")
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_LargeDataset(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := makeFilter()

	// Generate 500 rows
	var rows []models.ActualExportDTO
	for i := 0; i < 500; i++ {
		rows = append(rows, models.ActualExportDTO{
			Entity:      "ACG",
			Branch:      "HQ",
			Department:  "IT",
			Group:       "OpEx",
			ConsoGL:     "5100",
			GLName:      "Expense",
			DocumentNo:  "INV-" + string(rune(i)),
			Amount:      decimal.NewFromInt(int64(i * 100)),
			VendorName:  "Vendor",
			Description: "Desc",
			PostingDate: "2025-01-01",
		})
	}

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return(rows, nil)

	data, filename, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".xlsx")

	// Parse and verify row count
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	allRows, _ := f.GetRows(sheet)
	// 1 header + 500 data rows
	assert.Equal(t, 501, len(allRows))
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_SpecialCharacters(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := makeFilter()

	rows := []models.ActualExportDTO{
		{
			Entity:      "ACG",
			Branch:      "HQ",
			Department:  "IT & Support",
			Group:       "Op\"Ex",
			ConsoGL:     "5100",
			GLName:      "ค่าใช้จ่าย/Office",
			DocumentNo:  "INV-001<test>",
			Amount:      decimal.NewFromFloat(1234.56),
			VendorName:  "บริษัท ABC จำกัด",
			Description: "รายละเอียด & notes",
			PostingDate: "2025-06-30",
		},
	}

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return(rows, nil)

	data, _, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Parse and verify Thai/special chars preserved
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheet, "C2")
	assert.Equal(t, "IT & Support", v)
	v, _ = f.GetCellValue(sheet, "H2")
	assert.Equal(t, "ค่าใช้จ่าย/Office", v)
	v, _ = f.GetCellValue(sheet, "K2")
	assert.Equal(t, "บริษัท ABC จำกัด", v)
	v, _ = f.GetCellValue(sheet, "L2")
	assert.Equal(t, "รายละเอียด & notes", v)
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_ZeroAndNegativeAmounts(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := makeFilter()

	rows := []models.ActualExportDTO{
		{
			Entity:      "ACG",
			Branch:      "HQ",
			Department:  "ACC",
			ConsoGL:     "5100",
			GLName:      "Zero Amount",
			DocumentNo:  "INV-ZERO",
			Amount:      decimal.Zero,
			PostingDate: "2025-01-01",
		},
		{
			Entity:      "ACG",
			Branch:      "HQ",
			Department:  "ACC",
			ConsoGL:     "5200",
			GLName:      "Negative Amount",
			DocumentNo:  "INV-NEG",
			Amount:      decimal.NewFromFloat(-99999.99),
			PostingDate: "2025-01-02",
		},
		{
			Entity:      "ACG",
			Branch:      "HQ",
			Department:  "ACC",
			ConsoGL:     "5300",
			GLName:      "Large Amount",
			DocumentNo:  "INV-LARGE",
			Amount:      decimal.NewFromFloat(9999999.99),
			PostingDate: "2025-01-03",
		},
	}

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return(rows, nil)

	data, filename, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".xlsx")

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Row 2: zero amount
	v, _ := f.GetCellValue(sheet, "J2")
	assert.Equal(t, "0", v)

	// Row 3: negative amount
	v, _ = f.GetCellValue(sheet, "J3")
	assert.Equal(t, "-99999.99", v)

	// Row 4: large amount
	v, _ = f.GetCellValue(sheet, "J4")
	assert.Equal(t, "9999999.99", v)

	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_EmptyFieldsInData(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := makeFilter()

	rows := []models.ActualExportDTO{
		{
			Entity:      "ACG",
			Branch:      "",
			Department:  "",
			Group:       "",
			Group2:      "",
			Group3:      "",
			ConsoGL:     "5100",
			GLName:      "",
			DocumentNo:  "",
			Amount:      decimal.NewFromFloat(100),
			VendorName:  "",
			Description: "",
			PostingDate: "",
		},
	}

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return(rows, nil)

	data, filename, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".xlsx")

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Empty fields should be empty strings
	v, _ := f.GetCellValue(sheet, "B2")
	assert.Equal(t, "", v)
	v, _ = f.GetCellValue(sheet, "C2")
	assert.Equal(t, "", v)
	v, _ = f.GetCellValue(sheet, "H2")
	assert.Equal(t, "", v)

	// Entity and ConsoGL should have values
	v, _ = f.GetCellValue(sheet, "A2")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "G2")
	assert.Equal(t, "5100", v)
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_FilenameFormat(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := makeFilter()

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return([]models.ActualExportDTO{}, nil)

	_, filename, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	// Filename should be: Actual_Detail_YYYYMMDDHHMMSS.xlsx
	assert.Regexp(t, `^Actual_Detail_\d{14}\.xlsx$`, filename)
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_ExcelSheetName(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := makeFilter()

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return([]models.ActualExportDTO{}, nil)

	data, _, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	// Sheet should be named "Actual Details"
	sheet := f.GetSheetName(0)
	assert.Equal(t, "Actual Details", sheet)
	repo.AssertExpectations(t)
}

func TestExportActualDetailExcel_ComplexFilter(t *testing.T) {
	repo := new(MockActualExportRepository)
	svc := NewService(repo)

	user := makeUser()
	filter := map[string]interface{}{
		"entities":   []string{"ACG", "HMW", "CLIK"},
		"branches":   []string{"HQ", "BKK"},
		"departments": []string{"IT", "HR", "ACC"},
		"conso_gls":  []string{"5100", "5200"},
		"year":       "2025",
		"months":     []string{"JAN", "FEB", "MAR"},
		"start_date": "2025-01-01",
		"end_date":   "2025-03-31",
	}

	repo.On("GetActualExportDetails", mock.Anything, user, filter).
		Return([]models.ActualExportDTO{
			{
				Entity: "ACG", Branch: "HQ", Department: "IT",
				ConsoGL: "5100", GLName: "Supplies", DocumentNo: "D001",
				Amount: decimal.NewFromFloat(500), PostingDate: "2025-01-15",
			},
		}, nil)

	data, fname, err := svc.ExportActualDetailExcel(context.Background(), user, filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, fname, ".xlsx")
	repo.AssertExpectations(t)
}


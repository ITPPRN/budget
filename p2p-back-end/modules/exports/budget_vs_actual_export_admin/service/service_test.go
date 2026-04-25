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
	"p2p-back-end/modules/exports/budget_vs_actual_export_admin/repository"
)

// Compile-time check: MockBudgetVsActualRepository implements repository.BudgetVsActualRepository
var _ repository.BudgetVsActualRepository = (*MockBudgetVsActualRepository)(nil)

// --- Mock ---

type MockBudgetVsActualRepository struct {
	mock.Mock
}

func (m *MockBudgetVsActualRepository) GetBudgetVsActualData(ctx context.Context, filter map[string]interface{}) ([]models.BudgetVsActualExportDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetVsActualExportDTO), args.Error(1)
}

// --- Helpers ---

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
	repo := new(MockBudgetVsActualRepository)
	svc := NewService(repo)
	assert.NotNil(t, svc)
}

func TestExport_Success_WithData(t *testing.T) {
	repo := new(MockBudgetVsActualRepository)
	svc := NewService(repo)

	filter := makeFilter()
	rows := makeSampleRows()

	// Service sanitizes the filter, so match with mock.Anything
	repo.On("GetBudgetVsActualData", mock.Anything, mock.Anything).
		Return(rows, nil)

	data, filename, err := svc.ExportBudgetVsActualExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Budget_vs_Actual_Admin_")
	assert.Contains(t, filename, ".xlsx")

	// Parse Excel
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Verify headers (19 columns)
	expectedHeaders := []string{
		"Entity", "Branch", "Department", "Type",
		"GL Code", "GL Name",
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
	assert.Equal(t, "5100", v)
	v, _ = f.GetCellValue(sheet, "F2")
	assert.Equal(t, "Office Supplies", v)

	// Verify JAN amount in row 2 (column G)
	v, _ = f.GetCellValue(sheet, "G2")
	assert.Equal(t, "1000", v)

	// Verify YEARTOTAL in row 2 (column S)
	v, _ = f.GetCellValue(sheet, "S2")
	assert.Equal(t, "4500", v)

	// Verify second data row type
	v, _ = f.GetCellValue(sheet, "D3")
	assert.Equal(t, "Actual", v)

	// Verify YEARTOTAL in row 3
	v, _ = f.GetCellValue(sheet, "S3")
	assert.Equal(t, "3000", v)

	repo.AssertExpectations(t)
}

func TestExport_Success_EmptyData(t *testing.T) {
	repo := new(MockBudgetVsActualRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetBudgetVsActualData", mock.Anything, mock.Anything).
		Return([]models.BudgetVsActualExportDTO{}, nil)

	data, filename, err := svc.ExportBudgetVsActualExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".xlsx")

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	allRows, _ := f.GetRows(sheet)
	// Only 1 header row, no data rows
	assert.Equal(t, 1, len(allRows))
	repo.AssertExpectations(t)
}

func TestExport_RepoError(t *testing.T) {
	repo := new(MockBudgetVsActualRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetBudgetVsActualData", mock.Anything, mock.Anything).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportBudgetVsActualExcel(context.Background(), filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)
	repo.AssertExpectations(t)
}

func TestExport_FilenameFormat(t *testing.T) {
	repo := new(MockBudgetVsActualRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetBudgetVsActualData", mock.Anything, mock.Anything).
		Return([]models.BudgetVsActualExportDTO{}, nil)

	_, filename, err := svc.ExportBudgetVsActualExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.Regexp(t, `^Budget_vs_Actual_Admin_\d{14}\.xlsx$`, filename)
	repo.AssertExpectations(t)
}

func TestExport_SheetName(t *testing.T) {
	repo := new(MockBudgetVsActualRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetBudgetVsActualData", mock.Anything, mock.Anything).
		Return([]models.BudgetVsActualExportDTO{}, nil)

	data, _, err := svc.ExportBudgetVsActualExcel(context.Background(), filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	assert.Equal(t, "Budget vs Actual", sheet)
	repo.AssertExpectations(t)
}

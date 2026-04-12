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
	"p2p-back-end/modules/exports/department_budget_status_export_admin/repository"
)

// Compile-time check
var _ repository.DeptBudgetStatusRepository = (*MockDeptBudgetStatusRepository)(nil)

// --- Mock ---

type MockDeptBudgetStatusRepository struct {
	mock.Mock
}

func (m *MockDeptBudgetStatusRepository) GetDeptBudgetStatus(ctx context.Context, filter map[string]interface{}) ([]models.DeptBudgetStatusDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DeptBudgetStatusDTO), args.Error(1)
}

// --- Helpers ---

func makeFilter() map[string]interface{} {
	return map[string]interface{}{
		"entities": []string{"ACG"},
		"year":     "2025",
	}
}

func makeSampleRows() []models.DeptBudgetStatusDTO {
	return []models.DeptBudgetStatusDTO{
		{
			Status:     "Over Budget",
			Department: "IT",
			Budget:     decimal.NewFromFloat(100000),
			Spend:      decimal.NewFromFloat(120000),
			Remaining:  decimal.NewFromFloat(-20000),
			Percentage: 120.0,
		},
		{
			Status:     "Normal",
			Department: "HR",
			Budget:     decimal.NewFromFloat(80000),
			Spend:      decimal.NewFromFloat(50000),
			Remaining:  decimal.NewFromFloat(30000),
			Percentage: 62.5,
		},
	}
}

// --- Tests ---

func TestNewService(t *testing.T) {
	repo := new(MockDeptBudgetStatusRepository)
	svc := NewService(repo)
	assert.NotNil(t, svc)
}

func TestExport_Success_WithData(t *testing.T) {
	repo := new(MockDeptBudgetStatusRepository)
	svc := NewService(repo)

	filter := makeFilter()
	rows := makeSampleRows()

	repo.On("GetDeptBudgetStatus", mock.Anything, mock.Anything).
		Return(rows, nil)

	data, filename, err := svc.ExportDeptBudgetStatusExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Dept_Budget_Status_")
	assert.Contains(t, filename, ".xlsx")

	// Parse Excel
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Verify headers (6 columns)
	expectedHeaders := []string{"Status", "Department", "Budget", "Spend", "Remaining", "(%)"}
	allRows, _ := f.GetRows(sheet)
	assert.GreaterOrEqual(t, len(allRows), 3) // header + 2 data rows
	for i, expected := range expectedHeaders {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		v, _ := f.GetCellValue(sheet, colName+"1")
		assert.Equal(t, expected, v, "header column %s", colName)
	}

	// Verify first data row
	v, _ := f.GetCellValue(sheet, "A2")
	assert.Equal(t, "Over Budget", v)
	v, _ = f.GetCellValue(sheet, "B2")
	assert.Equal(t, "IT", v)
	v, _ = f.GetCellValue(sheet, "C2")
	assert.Equal(t, "100000", v)
	v, _ = f.GetCellValue(sheet, "D2")
	assert.Equal(t, "120000", v)
	v, _ = f.GetCellValue(sheet, "E2")
	assert.Equal(t, "-20000", v)
	v, _ = f.GetCellValue(sheet, "F2")
	assert.Equal(t, "120", v)

	// Verify second data row
	v, _ = f.GetCellValue(sheet, "A3")
	assert.Equal(t, "Normal", v)
	v, _ = f.GetCellValue(sheet, "B3")
	assert.Equal(t, "HR", v)

	repo.AssertExpectations(t)
}

func TestExport_Success_EmptyData(t *testing.T) {
	repo := new(MockDeptBudgetStatusRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetDeptBudgetStatus", mock.Anything, mock.Anything).
		Return([]models.DeptBudgetStatusDTO{}, nil)

	data, filename, err := svc.ExportDeptBudgetStatusExcel(context.Background(), filter)

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
}

func TestExport_RepoError(t *testing.T) {
	repo := new(MockDeptBudgetStatusRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetDeptBudgetStatus", mock.Anything, mock.Anything).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportDeptBudgetStatusExcel(context.Background(), filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)
	repo.AssertExpectations(t)
}

func TestExport_FilenameFormat(t *testing.T) {
	repo := new(MockDeptBudgetStatusRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetDeptBudgetStatus", mock.Anything, mock.Anything).
		Return([]models.DeptBudgetStatusDTO{}, nil)

	_, filename, err := svc.ExportDeptBudgetStatusExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.Regexp(t, `^Dept_Budget_Status_\d{14}\.xlsx$`, filename)
	repo.AssertExpectations(t)
}

func TestExport_SheetName(t *testing.T) {
	repo := new(MockDeptBudgetStatusRepository)
	svc := NewService(repo)

	filter := makeFilter()

	repo.On("GetDeptBudgetStatus", mock.Anything, mock.Anything).
		Return([]models.DeptBudgetStatusDTO{}, nil)

	data, _, err := svc.ExportDeptBudgetStatusExcel(context.Background(), filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	assert.Equal(t, "Department Status", sheet)
	repo.AssertExpectations(t)
}

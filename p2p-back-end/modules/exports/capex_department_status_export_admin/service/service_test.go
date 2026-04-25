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
	"p2p-back-end/modules/exports/capex_department_status_export_admin/repository"
)

// --- Mock ---

type MockCapexDeptStatusRepository struct {
	mock.Mock
}

func (m *MockCapexDeptStatusRepository) GetCapexDeptStatus(ctx context.Context, filter map[string]interface{}) ([]models.CapexDeptStatusDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CapexDeptStatusDTO), args.Error(1)
}

// --- Helpers ---

func makeFilter() map[string]interface{} {
	return map[string]interface{}{
		"entities": []string{"ACG"},
		"year":     "2025",
	}
}

// Compile-time check that mock satisfies the interface.
var _ repository.CapexDeptStatusRepository = (*MockCapexDeptStatusRepository)(nil)

// --- Tests ---

func TestNewService(t *testing.T) {
	repo := new(MockCapexDeptStatusRepository)
	svc := NewService(repo)
	assert.NotNil(t, svc)
}

func TestExport_Success_WithData(t *testing.T) {
	repo := new(MockCapexDeptStatusRepository)
	svc := NewService(repo)
	filter := makeFilter()

	rows := []models.CapexDeptStatusDTO{
		{
			Status:      "Normal",
			Department:  "IT",
			CapexBudget: decimal.NewFromFloat(500000),
			Spend:       decimal.NewFromFloat(200000),
			Remaining:   decimal.NewFromFloat(300000),
			Percentage:  40.0,
		},
		{
			Status:      "Over Budget",
			Department:  "HR",
			CapexBudget: decimal.NewFromFloat(100000),
			Spend:       decimal.NewFromFloat(120000),
			Remaining:   decimal.NewFromFloat(-20000),
			Percentage:  120.0,
		},
	}

	repo.On("GetCapexDeptStatus", mock.Anything, filter).Return(rows, nil)

	data, filename, err := svc.ExportCapexDeptStatusExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Capex_Dept_Status_")
	assert.Contains(t, filename, ".xlsx")

	// Parse the generated Excel and verify content
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Verify headers
	h1, _ := f.GetCellValue(sheet, "A1")
	assert.Equal(t, "Status", h1)
	h2, _ := f.GetCellValue(sheet, "B1")
	assert.Equal(t, "Department", h2)
	h3, _ := f.GetCellValue(sheet, "C1")
	assert.Equal(t, "Capex_BG", h3)
	h4, _ := f.GetCellValue(sheet, "D1")
	assert.Equal(t, "Spend", h4)
	h5, _ := f.GetCellValue(sheet, "E1")
	assert.Equal(t, "Remaining", h5)
	h6, _ := f.GetCellValue(sheet, "F1")
	assert.Equal(t, "(%)", h6)

	// Verify row 2 (first data row)
	v, _ := f.GetCellValue(sheet, "A2")
	assert.Equal(t, "Normal", v)
	v, _ = f.GetCellValue(sheet, "B2")
	assert.Equal(t, "IT", v)

	// Verify row 3 (second data row)
	v, _ = f.GetCellValue(sheet, "A3")
	assert.Equal(t, "Over Budget", v)
	v, _ = f.GetCellValue(sheet, "B3")
	assert.Equal(t, "HR", v)

	// Verify no row 4
	v, _ = f.GetCellValue(sheet, "A4")
	assert.Equal(t, "", v)

	allRows, _ := f.GetRows(sheet)
	assert.Equal(t, 3, len(allRows)) // 1 header + 2 data

	repo.AssertExpectations(t)
}

func TestExport_Success_EmptyData(t *testing.T) {
	repo := new(MockCapexDeptStatusRepository)
	svc := NewService(repo)
	filter := makeFilter()

	repo.On("GetCapexDeptStatus", mock.Anything, filter).
		Return([]models.CapexDeptStatusDTO{}, nil)

	data, filename, err := svc.ExportCapexDeptStatusExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Capex_Dept_Status_")
	assert.Contains(t, filename, ".xlsx")

	// Verify the Excel has headers but no data rows
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	allRows, _ := f.GetRows(sheet)
	assert.Equal(t, 1, len(allRows))
	repo.AssertExpectations(t)
}

func TestExport_RepoError(t *testing.T) {
	repo := new(MockCapexDeptStatusRepository)
	svc := NewService(repo)
	filter := makeFilter()

	repo.On("GetCapexDeptStatus", mock.Anything, filter).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportCapexDeptStatusExcel(context.Background(), filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)
	repo.AssertExpectations(t)
}

func TestExport_FilenameFormat(t *testing.T) {
	repo := new(MockCapexDeptStatusRepository)
	svc := NewService(repo)
	filter := makeFilter()

	repo.On("GetCapexDeptStatus", mock.Anything, filter).
		Return([]models.CapexDeptStatusDTO{}, nil)

	_, filename, err := svc.ExportCapexDeptStatusExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.Regexp(t, `^Capex_Dept_Status_\d{14}\.xlsx$`, filename)
	repo.AssertExpectations(t)
}

func TestExport_SheetName(t *testing.T) {
	repo := new(MockCapexDeptStatusRepository)
	svc := NewService(repo)
	filter := makeFilter()

	repo.On("GetCapexDeptStatus", mock.Anything, filter).
		Return([]models.CapexDeptStatusDTO{}, nil)

	data, _, err := svc.ExportCapexDeptStatusExcel(context.Background(), filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	assert.Equal(t, "Capex Dept Status", sheet)
	repo.AssertExpectations(t)
}

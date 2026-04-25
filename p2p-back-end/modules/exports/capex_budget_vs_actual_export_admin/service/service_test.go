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
	"p2p-back-end/modules/exports/capex_budget_vs_actual_export_admin/repository"
)

// --- Mock ---

type MockCapexVsActualRepository struct {
	mock.Mock
}

func (m *MockCapexVsActualRepository) GetCapexVsActualData(ctx context.Context, filter map[string]interface{}) ([]models.CapexVsActualExportDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CapexVsActualExportDTO), args.Error(1)
}

// --- Helpers ---

func makeFilter() map[string]interface{} {
	return map[string]interface{}{
		"entities": []string{"ACG"},
		"year":     "2025",
	}
}

// Compile-time check that mock satisfies the interface.
var _ repository.CapexVsActualRepository = (*MockCapexVsActualRepository)(nil)

// --- Tests ---

func TestNewService(t *testing.T) {
	repo := new(MockCapexVsActualRepository)
	svc := NewService(repo)
	assert.NotNil(t, svc)
}

func TestExport_Success_WithData(t *testing.T) {
	repo := new(MockCapexVsActualRepository)
	svc := NewService(repo)
	filter := makeFilter()

	rows := []models.CapexVsActualExportDTO{
		{
			Entity:        "ACG",
			Branch:        "HQ",
			Department:    "IT",
			CapexNo:       "CX001",
			CapexName:     "Server",
			CapexCategory: "Hardware",
			Type:          "Budget",
			MonthsAmounts: map[string]interface{}{
				"JAN": decimal.NewFromFloat(50000),
			},
			YearTotal: decimal.NewFromFloat(50000),
		},
		{
			Entity:        "ACG",
			Branch:        "HQ",
			Department:    "IT",
			CapexNo:       "CX001",
			CapexName:     "Server",
			CapexCategory: "Hardware",
			Type:          "Actual",
			MonthsAmounts: map[string]interface{}{
				"JAN": decimal.NewFromFloat(30000),
				"FEB": decimal.NewFromFloat(10000),
			},
			YearTotal: decimal.NewFromFloat(40000),
		},
	}

	repo.On("GetCapexVsActualData", mock.Anything, filter).Return(rows, nil)

	data, filename, err := svc.ExportCapexVsActualExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Capex_Budget_vs_Actual_Admin_")
	assert.Contains(t, filename, ".xlsx")

	// Parse the generated Excel and verify content
	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// Verify headers (20 columns)
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
	assert.Equal(t, "Type", h7)
	h8, _ := f.GetCellValue(sheet, "H1")
	assert.Equal(t, "JAN", h8)
	h20, _ := f.GetCellValue(sheet, "T1")
	assert.Equal(t, "YEARTOTAL", h20)

	// Verify row 2 (Budget row)
	v, _ := f.GetCellValue(sheet, "A2")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "B2")
	assert.Equal(t, "HQ", v)
	v, _ = f.GetCellValue(sheet, "C2")
	assert.Equal(t, "IT", v)
	v, _ = f.GetCellValue(sheet, "D2")
	assert.Equal(t, "CX001", v)
	v, _ = f.GetCellValue(sheet, "E2")
	assert.Equal(t, "Server", v)
	v, _ = f.GetCellValue(sheet, "F2")
	assert.Equal(t, "Hardware", v)
	v, _ = f.GetCellValue(sheet, "G2")
	assert.Equal(t, "Budget", v)

	// Verify row 3 (Actual row)
	v, _ = f.GetCellValue(sheet, "A3")
	assert.Equal(t, "ACG", v)
	v, _ = f.GetCellValue(sheet, "G3")
	assert.Equal(t, "Actual", v)

	// Verify no row 4
	v, _ = f.GetCellValue(sheet, "A4")
	assert.Equal(t, "", v)

	allRows, _ := f.GetRows(sheet)
	assert.Equal(t, 3, len(allRows)) // 1 header + 2 data

	repo.AssertExpectations(t)
}

func TestExport_Success_EmptyData(t *testing.T) {
	repo := new(MockCapexVsActualRepository)
	svc := NewService(repo)
	filter := makeFilter()

	repo.On("GetCapexVsActualData", mock.Anything, filter).
		Return([]models.CapexVsActualExportDTO{}, nil)

	data, filename, err := svc.ExportCapexVsActualExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, "Capex_Budget_vs_Actual_Admin_")
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
	repo := new(MockCapexVsActualRepository)
	svc := NewService(repo)
	filter := makeFilter()

	repo.On("GetCapexVsActualData", mock.Anything, filter).
		Return(nil, errors.New("database connection failed"))

	data, filename, err := svc.ExportCapexVsActualExcel(context.Background(), filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, data)
	assert.Empty(t, filename)
	repo.AssertExpectations(t)
}

func TestExport_FilenameFormat(t *testing.T) {
	repo := new(MockCapexVsActualRepository)
	svc := NewService(repo)
	filter := makeFilter()

	repo.On("GetCapexVsActualData", mock.Anything, filter).
		Return([]models.CapexVsActualExportDTO{}, nil)

	_, filename, err := svc.ExportCapexVsActualExcel(context.Background(), filter)

	assert.NoError(t, err)
	assert.Regexp(t, `^Capex_Budget_vs_Actual_Admin_\d{14}\.xlsx$`, filename)
	repo.AssertExpectations(t)
}

func TestExport_SheetName(t *testing.T) {
	repo := new(MockCapexVsActualRepository)
	svc := NewService(repo)
	filter := makeFilter()

	repo.On("GetCapexVsActualData", mock.Anything, filter).
		Return([]models.CapexVsActualExportDTO{}, nil)

	data, _, err := svc.ExportCapexVsActualExcel(context.Background(), filter)

	assert.NoError(t, err)

	f, parseErr := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, parseErr)
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	assert.Equal(t, "Capex Budget vs Actual", sheet)
	repo.AssertExpectations(t)
}

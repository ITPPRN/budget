package utils

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ExcelHelper provides utility methods for common excelize tasks
type ExcelHelper struct {
	File  *excelize.File
	Sheet string
}

// NewExcelHelper creates a new helper with an initialized file and sheet
func NewExcelHelper(sheetName string) *ExcelHelper {
	f := excelize.NewFile()
	// Rename default "Sheet1"
	f.SetSheetName("Sheet1", sheetName)
	return &ExcelHelper{
		File:  f,
		Sheet: sheetName,
	}
}

// SetHeaders sets common header styles and values
func (h *ExcelHelper) SetHeaders(headers []string) error {
	style, err := h.File.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		return err
	}

	for i, name := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s1", colName)
		h.File.SetCellValue(h.Sheet, cell, name)
		h.File.SetCellStyle(h.Sheet, cell, cell, style)
	}

	return nil
}

// AutoWidth adjusts column widths based on content estimate (Basic implementation)
func (h *ExcelHelper) AutoWidth(colCount int) {
	for i := 1; i <= colCount; i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		h.File.SetColWidth(h.Sheet, colName, colName, 15) // Default reasonable width
	}
}

// WriteToBuffer returns the file as a byte slice
func (h *ExcelHelper) WriteToBuffer() ([]byte, error) {
	buffer, err := h.File.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

package service

import (
	"context"
	"fmt"
	"p2p-back-end/modules/exports/budget_vs_actual_export_admin/repository"
	"p2p-back-end/pkg/utils"

	"github.com/shopspring/decimal"
)

type BudgetVsActualService interface {
	ExportBudgetVsActualExcel(ctx context.Context, filter map[string]interface{}) ([]byte, string, error)
}

type service struct {
	repo repository.BudgetVsActualRepository
}

func NewService(repo repository.BudgetVsActualRepository) BudgetVsActualService {
	return &service{repo: repo}
}

func (s *service) ExportBudgetVsActualExcel(ctx context.Context, filter map[string]interface{}) ([]byte, string, error) {
	// 1. Sanitize Filters
	sanitizedFilter := make(map[string]interface{})
	for k, v := range filter {
		nk, nv := utils.SanitizeFilter(k, v)
		sanitizedFilter[nk] = nv
	}

	// 2. Fetch Data
	data, err := s.repo.GetBudgetVsActualData(ctx, sanitizedFilter)
	if err != nil {
		return nil, "", err
	}

	// 3. Create Excel
	helper := utils.NewExcelHelper("Budget vs Actual")

	// Define Headers
	headers := []string{
		"Entity", "Branch", "Department", "Type",
		"GL Code", "GL Name",
		"JAN", "FEB", "MAR", "APR", "MAY", "JUN",
		"JUL", "AUG", "SEP", "OCT", "NOV", "DEC",
		"YEARTOTAL",
	}

	if err := helper.SetHeaders(headers); err != nil {
		return nil, "", err
	}

	// 4. Fill Data
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	rowIdx := 2
	for _, row := range data {
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("A%d", rowIdx), row.Entity)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("B%d", rowIdx), row.Branch)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("C%d", rowIdx), row.Department)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("D%d", rowIdx), row.Type)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("E%d", rowIdx), row.ConsoGL)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("F%d", rowIdx), row.GLName)

		// Monthly Amounts
		for i, m := range months {
			colName, _ := excelizeColumnName(7 + i)
			val := decimal.Zero
			if amt, ok := row.MonthsAmounts[m].(decimal.Decimal); ok {
				val = amt
			}
			helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("%s%d", colName, rowIdx), val.InexactFloat64())
		}

		// Year Total
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("S%d", rowIdx), row.YearTotal.InexactFloat64())
		rowIdx++
	}

	helper.AutoWidth(len(headers))

	buffer, err := helper.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Budget_vs_Actual_Admin_%s.xlsx", utils.GetTimestamp())
	return buffer, filename, nil
}

func excelizeColumnName(col int) (string, error) {
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+(col%26))) + name
		col /= 26
	}
	return name, nil
}

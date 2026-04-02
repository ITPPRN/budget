package service

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/exports/budget_detail_export_owner/repository"
	"p2p-back-end/pkg/utils"
)

type OwnerBudgetExportService interface {
	ExportOwnerBudgetDetailExcel(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]byte, string, error)
}

type service struct {
	repo     repository.OwnerBudgetExportRepository
	ownerSrv models.OwnerService
}

func NewService(repo repository.OwnerBudgetExportRepository, ownerSrv models.OwnerService) OwnerBudgetExportService {
	return &service{repo: repo, ownerSrv: ownerSrv}
}

func (s *service) ExportOwnerBudgetDetailExcel(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]byte, string, error) {
	// 1. Sanitize Filters
	sanitizedFilter := make(map[string]interface{})
	for k, v := range filter {
		nk, nv := utils.SanitizeFilter(k, v)
		sanitizedFilter[nk] = nv
	}

	// 🛠️ Enforce RBAC (Same as Dashboard)
	sanitizedFilter = s.ownerSrv.InjectPermissions(ctx, user, sanitizedFilter)

	data, err := s.repo.GetOwnerBudgetExportDetails(ctx, user, sanitizedFilter)
	if err != nil {
		return nil, "", err
	}

	helper := utils.NewExcelHelper("Budget Details")

	headers := []string{
		"Entity", "Branch", "Department",
		"GROUP1", "GROUP2", "GROUP3",
		"GL Code", "GL Name",
		"JAN", "FEB", "MAR", "APR", "MAY", "JUN",
		"JUL", "AUG", "SEP", "OCT", "NOV", "DEC",
		"YEARTOTAL",
	}

	if err := helper.SetHeaders(headers); err != nil {
		return nil, "", err
	}

	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	rowIdx := 2
	for _, row := range data {
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("A%d", rowIdx), row.Entity)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("B%d", rowIdx), row.Branch)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("C%d", rowIdx), row.Department)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("D%d", rowIdx), row.Group)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("E%d", rowIdx), row.Group2)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("F%d", rowIdx), row.Group3)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("G%d", rowIdx), row.ConsoGL)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("H%d", rowIdx), row.GLName)

		for i, m := range months {
			colName, _ := excelizeColumnName(9 + i)
			val := decimal.Zero
			if amt, ok := row.MonthsAmounts[m].(decimal.Decimal); ok {
				val = amt
			}
			_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("%s%d", colName, rowIdx), val.InexactFloat64())
		}

		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("U%d", rowIdx), row.YearTotal.InexactFloat64())
		rowIdx++
	}

	helper.AutoWidth(len(headers))

	buffer, err := helper.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Owner_Budget_Detail_%s.xlsx", utils.GetTimestamp())
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

package service

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/exports/capex_budget_export_owner/repository"
	"p2p-back-end/pkg/utils"
)

type OwnerCapexService interface {
	ExportOwnerCapexExcel(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]byte, string, error)
}

type service struct {
	repo     repository.OwnerCapexRepository
	ownerSrv models.OwnerService
}

func NewService(repo repository.OwnerCapexRepository, ownerSrv models.OwnerService) OwnerCapexService {
	return &service{repo: repo, ownerSrv: ownerSrv}
}

func (s *service) ExportOwnerCapexExcel(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]byte, string, error) {
	// 1. Sanitize Filters
	sanitizedFilter := make(map[string]interface{})
	for k, v := range filter {
		nk, nv := utils.SanitizeFilter(k, v)
		sanitizedFilter[nk] = nv
	}

	// 🛠️ Enforce RBAC
	sanitizedFilter = s.ownerSrv.InjectPermissions(ctx, user, sanitizedFilter)

	// 2. Fetch Data
	data, err := s.repo.GetOwnerCapexData(ctx, sanitizedFilter)
	if err != nil {
		return nil, "", err
	}

	// 3. Create Excel
	helper := utils.NewExcelHelper("Capex Budget Status")

	// Define Headers (per-transaction layout: Type + 12 months + YearTotal)
	headers := []string{
		"Entity", "Branch", "Department", "CAPEX NO.", "CAPEX Name", "CAPEX Category", "Type",
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
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("A%d", rowIdx), row.Entity)
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("B%d", rowIdx), row.Branch)
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("C%d", rowIdx), row.Department)
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("D%d", rowIdx), row.CapexNo)
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("E%d", rowIdx), row.CapexName)
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("F%d", rowIdx), row.CapexCategory)
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("G%d", rowIdx), row.Type)

		for i, m := range months {
			colName, _ := excelizeColumnName(8 + i)
			val := decimal.Zero
			if amt, ok := row.MonthsAmounts[m].(decimal.Decimal); ok {
				val = amt
			}
			_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("%s%d", colName, rowIdx), val.InexactFloat64())
		}

		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("T%d", rowIdx), row.YearTotal.InexactFloat64())
		rowIdx++
	}

	helper.AutoWidth(len(headers))

	buffer, err := helper.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Owner_Capex_Budget_%s.xlsx", utils.GetTimestamp())
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

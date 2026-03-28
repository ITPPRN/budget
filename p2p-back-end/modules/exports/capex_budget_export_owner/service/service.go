package service

import (
	"context"
	"fmt"
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

	// Define Headers
	headers := []string{"Entity", "Department", "CAPEX NO.", "CAPEX Name", "CAPEX Category", "Budget", "Actual", "Remaining", "(%)"}
	if err := helper.SetHeaders(headers); err != nil {
		return nil, "", err
	}

	// 4. Fill Data
	rowIdx := 2
	for _, row := range data {
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("A%d", rowIdx), row.Entity)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("B%d", rowIdx), row.Department)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("C%d", rowIdx), row.CapexNo)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("D%d", rowIdx), row.CapexName)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("E%d", rowIdx), row.CapexCategory)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("F%d", rowIdx), row.Budget.InexactFloat64())
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("G%d", rowIdx), row.Actual.InexactFloat64())
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("H%d", rowIdx), row.Remaining.InexactFloat64())
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("I%d", rowIdx), row.Percentage)
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

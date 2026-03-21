package service

import (
	"context"
	"fmt"
	"p2p-back-end/modules/exports/capex_department_status_export_admin/repository"
	"p2p-back-end/pkg/utils"
)

type CapexDeptStatusService interface {
	ExportCapexDeptStatusExcel(ctx context.Context, filter map[string]interface{}) ([]byte, string, error)
}

type service struct {
	repo repository.CapexDeptStatusRepository
}

func NewService(repo repository.CapexDeptStatusRepository) CapexDeptStatusService {
	return &service{repo: repo}
}

func (s *service) ExportCapexDeptStatusExcel(ctx context.Context, filter map[string]interface{}) ([]byte, string, error) {
	// 1. Sanitize Filters
	sanitizedFilter := make(map[string]interface{})
	for k, v := range filter {
		nk, nv := utils.SanitizeFilter(k, v)
		sanitizedFilter[nk] = nv
	}

	// 2. Fetch Data
	data, err := s.repo.GetCapexDeptStatus(ctx, sanitizedFilter)
	if err != nil {
		return nil, "", err
	}

	// 3. Create Excel
	helper := utils.NewExcelHelper("Capex Dept Status")

	// Define Headers
	headers := []string{"Status", "Department", "Capex_BG", "Spend", "Remaining", "(%)"}
	if err := helper.SetHeaders(headers); err != nil {
		return nil, "", err
	}

	// 4. Fill Data
	rowIdx := 2
	for _, row := range data {
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("A%d", rowIdx), row.Status)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("B%d", rowIdx), row.Department)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("C%d", rowIdx), row.CapexBudget.InexactFloat64())
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("D%d", rowIdx), row.Spend.InexactFloat64())
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("E%d", rowIdx), row.Remaining.InexactFloat64())
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("F%d", rowIdx), row.Percentage)
		rowIdx++
	}

	helper.AutoWidth(len(headers))

	buffer, err := helper.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Capex_Dept_Status_%s.xlsx", utils.GetTimestamp())
	return buffer, filename, nil
}

package service

import (
	"context"
	"fmt"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/exports/actual_detail_export_owner/repository"
	"p2p-back-end/pkg/utils"
)

type OwnerActualExportService interface {
	ExportOwnerActualDetailExcel(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]byte, string, error)
}

type service struct {
	repo     repository.OwnerActualExportRepository
	ownerSrv models.OwnerService
}

func NewService(repo repository.OwnerActualExportRepository, ownerSrv models.OwnerService) OwnerActualExportService {
	return &service{repo: repo, ownerSrv: ownerSrv}
}

func (s *service) ExportOwnerActualDetailExcel(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]byte, string, error) {
	// 1. Sanitize Filters
	sanitizedFilter := make(map[string]interface{})
	for k, v := range filter {
		nk, nv := utils.SanitizeFilter(k, v)
		sanitizedFilter[nk] = nv
	}

	// 🛠️ Enforce RBAC (Same as Dashboard)
	sanitizedFilter = s.ownerSrv.InjectPermissions(ctx, user, sanitizedFilter)

	data, err := s.repo.GetOwnerActualExportDetails(ctx, user, sanitizedFilter)
	if err != nil {
		return nil, "", err
	}

	helper := utils.NewExcelHelper("Actual Details")

	headers := []string{
		"Entity", "Branch", "Department",
		"GROUP1", "GROUP2", "GROUP3",
		"GL Code", "GL Name",
		"Document No.", "Amount", "Vendor", "Description", "Date",
	}

	if err := helper.SetHeaders(headers); err != nil {
		return nil, "", err
	}

	rowIdx := 2
	for _, row := range data {
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("A%d", rowIdx), row.Entity)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("B%d", rowIdx), row.Branch)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("C%d", rowIdx), row.Department)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("D%d", rowIdx), row.Group)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("E%d", rowIdx), row.Group2)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("F%d", rowIdx), row.Group3)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("G%d", rowIdx), row.ConsoGL)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("H%d", rowIdx), row.GLName)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("I%d", rowIdx), row.DocumentNo)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("J%d", rowIdx), row.Amount.InexactFloat64())
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("K%d", rowIdx), row.VendorName)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("L%d", rowIdx), row.Description)
		helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("M%d", rowIdx), row.PostingDate)
		rowIdx++
	}

	helper.AutoWidth(len(headers))

	buffer, err := helper.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Owner_Actual_Detail_%s.xlsx", utils.GetTimestamp())
	return buffer, filename, nil
}

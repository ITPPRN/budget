package service

import (
	"context"
	"fmt"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/exports/actual_detail_export/repository"
	"p2p-back-end/pkg/utils"
)

type ActualExportService interface {
	ExportActualDetailExcel(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]byte, string, error)
}

type service struct {
	repo repository.ActualExportRepository
}

func NewService(repo repository.ActualExportRepository) ActualExportService {
	return &service{repo: repo}
}

func (s *service) ExportActualDetailExcel(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]byte, string, error) {
	// 1. Fetch Data
	data, err := s.repo.GetActualExportDetails(ctx, user, filter)
	if err != nil {
		return nil, "", err
	}

	// 3. Create Excel
	helper := utils.NewExcelHelper("Actual Details")

	// Define Headers
	headers := []string{
		"Entity", "Branch", "Department",
		"GROUP1", "GROUP2", "GROUP3",
		"GL Code", "GL Name",
		"Document No.", "Amount", "Vendor", "Description", "Date", "Status",
	}

	if err := helper.SetHeaders(headers); err != nil {
		return nil, "", err
	}

	// 4. Fill Data
	rowIdx := 2
	for _, row := range data {
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("A%d", rowIdx), row.Entity)
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("B%d", rowIdx), row.Branch)
		_ = helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("C%d", rowIdx), row.Department)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("D%d", rowIdx), row.Group)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("E%d", rowIdx), row.Group2)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("F%d", rowIdx), row.Group3)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("G%d", rowIdx), row.ConsoGL)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("H%d", rowIdx), row.GLName)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("I%d", rowIdx), row.DocumentNo)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("J%d", rowIdx), row.Amount.InexactFloat64())
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("K%d", rowIdx), row.VendorName)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("L%d", rowIdx), row.Description)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("M%d", rowIdx), row.PostingDate)
		_ =helper.File.SetCellValue(helper.Sheet, fmt.Sprintf("N%d", rowIdx), row.Status)
		rowIdx++
	}

	helper.AutoWidth(len(headers))

	buffer, err := helper.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Actual_Detail_%s.xlsx", utils.GetTimestamp())
	return buffer, filename, nil
}

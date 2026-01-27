package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"

	"p2p-back-end/modules/entities/models"
)

type budgetService struct {
	repo models.BudgetRepository
}

func NewBudgetService(repo models.BudgetRepository) models.BudgetService {
	return &budgetService{repo: repo}
}

// ---------------------------------------------------------------------
// 1. Import Budget (PL)
// ---------------------------------------------------------------------

// ---------------------------------------------------------------------
// 1. Import Budget (PL) - Upload ONLY
// ---------------------------------------------------------------------
func (s *budgetService) ImportBudget(fileHeader *multipart.FileHeader, userID string, versionName string) error {
	// 1. Create File Record First to get ID
	fileNameToSave := fileHeader.Filename
	if versionName != "" {
		fileNameToSave = versionName
	}
	fileEntity := &models.FileBudgetEntity{
		ID: uuid.New(), FileName: fileNameToSave, UploadAt: time.Now(), UserID: userID,
	}

	if err := s.repo.CreateFileBudget(fileEntity); err != nil {
		return err
	}

	// 2. Save File to Disk
	if err := saveFileToDisk(fileHeader, "budget", fileEntity.ID.String()); err != nil {
		// Clean up DB if save fails
		s.repo.DeleteFileBudget(fileEntity.ID.String())
		return err
	}

	return nil
}

// ---------------------------------------------------------------------
// 2. Import Capex Budget
// ---------------------------------------------------------------------
func (s *budgetService) ImportCapexBudget(fileHeader *multipart.FileHeader, userID string, versionName string) error {
	fileNameToSave := fileHeader.Filename
	if versionName != "" {
		fileNameToSave = versionName
	}
	fileEntity := &models.FileCapexBudgetEntity{
		ID: uuid.New(), FileName: fileNameToSave, UploadAt: time.Now(), UserID: userID,
	}

	if err := s.repo.CreateFileCapexBudget(fileEntity); err != nil {
		return err
	}

	if err := saveFileToDisk(fileHeader, "capex_budget", fileEntity.ID.String()); err != nil {
		s.repo.DeleteFileCapexBudget(fileEntity.ID.String())
		return err
	}

	return nil
}

// ---------------------------------------------------------------------
// 3. Import Capex Actual
// ---------------------------------------------------------------------
func (s *budgetService) ImportCapexActual(fileHeader *multipart.FileHeader, userID string, versionName string) error {
	fileNameToSave := fileHeader.Filename
	if versionName != "" {
		fileNameToSave = versionName
	}
	fileEntity := &models.FileCapexActualEntity{
		ID: uuid.New(), FileName: fileNameToSave, UploadAt: time.Now(), UserID: userID,
	}

	if err := s.repo.CreateFileCapexActual(fileEntity); err != nil {
		return err
	}

	if err := saveFileToDisk(fileHeader, "capex_actual", fileEntity.ID.String()); err != nil {
		s.repo.DeleteFileCapexActual(fileEntity.ID.String())
		return err
	}

	return nil
}

// Sync Budget - Process & Replace Data
func (s *budgetService) SyncBudget(fileID string) error {
	// Check if file exists on disk
	filePath := fmt.Sprintf("./uploads/budget/%s.xlsx", fileID)
	xlsx, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("file not found on server: %v", err)
	}
	defer xlsx.Close()

	// Debug: Print all sheets
	sheets := xlsx.GetSheetList()
	fmt.Printf("[Debug] FileID: %s, Sheets found: %v\n", fileID, sheets)

	sheetName, err := findTargetSheet(xlsx, isBudgetHeader)
	if err != nil {
		return fmt.Errorf("invalid Budget file format: missing required columns")
	}
	fmt.Printf("[Debug] Selected Sheet: %s\n", sheetName)

	// 2. Transaction: Delete All & Insert New
	return s.repo.WithTrx(func(trxRepo models.BudgetRepository) error {
		// Delete All Existing Data
		if err := trxRepo.DeleteAllBudgetFacts(); err != nil {
			return err
		}

		// Process & Insert
		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processBudgetFact(xlsx, sheetName, parsedUUID)
		if err != nil {
			return err
		}
		if len(headers) > 0 {
			return trxRepo.CreateBudgetFacts(headers)
		}
		return nil
	})
}

// Sync Capex Budget
func (s *budgetService) SyncCapexBudget(fileID string) error {
	filePath := fmt.Sprintf("./uploads/capex_budget/%s.xlsx", fileID)
	xlsx, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("file not found on server: %v", err)
	}
	defer xlsx.Close()

	sheetName, err := findTargetSheet(xlsx, isCapexBudgetHeader)
	if err != nil {
		return fmt.Errorf("invalid Capex Budget file format: missing required columns")
	}

	return s.repo.WithTrx(func(trxRepo models.BudgetRepository) error {
		if err := trxRepo.DeleteAllCapexBudgetFacts(); err != nil {
			return err
		}

		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processCapexBudgetFact(xlsx, sheetName, parsedUUID)
		if err != nil {
			return err
		}
		if len(headers) > 0 {
			return trxRepo.CreateCapexBudgetFacts(headers)
		}
		return nil
	})
}

// Sync Capex Actual
func (s *budgetService) SyncCapexActual(fileID string) error {
	filePath := fmt.Sprintf("./uploads/capex_actual/%s.xlsx", fileID)
	xlsx, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("file not found on server: %v", err)
	}
	defer xlsx.Close()

	sheetName, err := findTargetSheet(xlsx, isCapexBudgetHeader)
	if err != nil {
		return fmt.Errorf("invalid Capex Actual file format: missing required columns")
	}

	return s.repo.WithTrx(func(trxRepo models.BudgetRepository) error {
		if err := trxRepo.DeleteAllCapexActualFacts(); err != nil {
			return err
		}

		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processCapexActualFact(xlsx, sheetName, parsedUUID)
		if err != nil {
			return err
		}
		if len(headers) > 0 {
			return trxRepo.CreateCapexActualFacts(headers)
		}
		return nil
	})
}

// ---------------------------------------------------------------------
// Processing Logic (Calculates YearTotal)
// ---------------------------------------------------------------------

// Helper for safe column access
func getColSafe(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

func (s *budgetService) processBudgetFact(f *excelize.File, sheetName string, fileID uuid.UUID) ([]models.BudgetFactEntity, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var headers []models.BudgetFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	fmt.Printf("[Debug] Budget Import - Sheet: %s, Total Rows: %d\n", sheetName, len(rows))
	if len(rows) > 0 {
		fmt.Printf("[Debug] Header Row: %v\n", rows[0])
	}
	if len(rows) > 1 {
		fmt.Printf("[Debug] First Data Row: %v\n", rows[1])
	}

	for i, row := range rows {
		if i == 0 {
			continue
		} // Header

		// Minimal check: Entity(0) to Dept(6)
		if len(row) < 7 {
			fmt.Printf("[Debug] Skipping Row %d: Not enough columns (%d)\n", i, len(row))
			continue
		}

		entity := getColSafe(row, 0)
		branch := getColSafe(row, 1)
		entityGL := getColSafe(row, 2)
		consoGL := getColSafe(row, 3)
		group := getColSafe(row, 4)
		glName := getColSafe(row, 5)
		dept := getColSafe(row, 6)

		headerID := uuid.New()
		header := models.BudgetFactEntity{
			ID:           headerID,
			FileBudgetID: fileID,
			Entity:       entity, Branch: branch, Group: group, EntityGL: entityGL, ConsoGL: consoGL, GLName: glName, Department: dept,
			YearTotal:     decimal.Zero,
			BudgetAmounts: []models.BudgetAmountEntity{},
		}

		// Fixed Loop for 12 months (Cols 7-18)
		for mIdx := 0; mIdx < 12; mIdx++ {
			colIdx := 7 + mIdx
			valStr := getColSafe(row, colIdx)
			amount := parseDecimal(valStr)

			header.BudgetAmounts = append(header.BudgetAmounts, models.BudgetAmountEntity{
				ID: uuid.New(), BudgetFactID: headerID, Month: months[mIdx], Amount: amount,
			})
			header.YearTotal = header.YearTotal.Add(amount)
		}
		headers = append(headers, header)
	}
	fmt.Printf("[Debug] Budget Import - Total Saved: %d\n", len(headers))
	return headers, nil
}

func (s *budgetService) processCapexBudgetFact(f *excelize.File, sheetName string, fileID uuid.UUID) ([]models.CapexBudgetFactEntity, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var headers []models.CapexBudgetFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	fmt.Printf("[Debug] Capex Plan Import - Sheet: %s, Total Rows: %d\n", sheetName, len(rows))

	for i, row := range rows {
		if i == 0 {
			continue
		}
		// Minimal check: Entity(0) to Category(4)
		if len(row) < 5 {
			return nil, nil
		}

		entity := getColSafe(row, 0)
		dept := getColSafe(row, 1)
		cNo := getColSafe(row, 2)
		cName := getColSafe(row, 3)
		cCat := getColSafe(row, 4)

		headerID := uuid.New()
		header := models.CapexBudgetFactEntity{
			ID:                headerID,
			FileCapexBudgetID: fileID,
			Entity:            entity, Department: dept, CapexNo: cNo, CapexName: cName, CapexCategory: cCat,
			YearTotal:          decimal.Zero,
			CapexBudgetAmounts: []models.CapexBudgetAmountEntity{},
		}

		// Fixed Loop for 12 months (Cols 5-16)
		for mIdx := 0; mIdx < 12; mIdx++ {
			colIdx := 5 + mIdx
			valStr := getColSafe(row, colIdx)
			amount := parseDecimal(valStr)

			header.CapexBudgetAmounts = append(header.CapexBudgetAmounts, models.CapexBudgetAmountEntity{
				ID: uuid.New(), CapexBudgetFactID: headerID, Month: months[mIdx], Amount: amount,
			})
			header.YearTotal = header.YearTotal.Add(amount)
		}
		headers = append(headers, header)
	}
	fmt.Printf("[Debug] Capex Plan - Total Saved: %d\n", len(headers))
	return headers, nil
}

func (s *budgetService) processCapexActualFact(f *excelize.File, sheetName string, fileID uuid.UUID) ([]models.CapexActualFactEntity, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var headers []models.CapexActualFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	fmt.Printf("[Debug] Capex Actual Import - Sheet: %s, Total Rows: %d\n", sheetName, len(rows))

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 5 {
			continue
		}

		entity := getColSafe(row, 0)
		dept := getColSafe(row, 1)
		cNo := getColSafe(row, 2)
		cName := getColSafe(row, 3)
		cCat := getColSafe(row, 4)

		headerID := uuid.New()
		header := models.CapexActualFactEntity{
			ID:                headerID,
			FileCapexActualID: fileID,
			Entity:            entity, Department: dept, CapexNo: cNo, CapexName: cName, CapexCategory: cCat,
			YearTotal:          decimal.Zero,
			CapexActualAmounts: []models.CapexActualAmountEntity{},
		}

		for mIdx := 0; mIdx < 12; mIdx++ {
			colIdx := 5 + mIdx
			valStr := getColSafe(row, colIdx)
			amount := parseDecimal(valStr)

			header.CapexActualAmounts = append(header.CapexActualAmounts, models.CapexActualAmountEntity{
				ID: uuid.New(), CapexActualFactID: headerID, Month: months[mIdx], Amount: amount,
			})
			header.YearTotal = header.YearTotal.Add(amount)
		}
		headers = append(headers, header)
	}
	fmt.Printf("[Debug] Capex Actual - Total Saved: %d\n", len(headers))
	return headers, nil
}

func parseDecimal(s string) decimal.Decimal {
	if s == "" {
		return decimal.Zero
	}
	// Robust Parsing: Remove commas and spaces
	cleanS := strings.ReplaceAll(s, ",", "")
	cleanS = strings.TrimSpace(cleanS)

	d, err := decimal.NewFromString(cleanS)
	if err != nil {
		// Log warning only for non-empty distinct strings to avoid spam
		if len(cleanS) > 0 {
			fmt.Printf("[Parse Warning] Invalid decimal: '%s' -> 0\n", s)
		}
		return decimal.Zero
	}
	return d
}

// ---------------------------------------------------------------------
// List Files Methods
// ---------------------------------------------------------------------

func (s *budgetService) ListBudgetFiles() ([]models.FileBudgetEntity, error) {
	return s.repo.ListFileBudgets()
}

func (s *budgetService) ListCapexBudgetFiles() ([]models.FileCapexBudgetEntity, error) {
	return s.repo.ListFileCapexBudgets()
}

func (s *budgetService) ListCapexActualFiles() ([]models.FileCapexActualEntity, error) {
	return s.repo.ListFileCapexActuals()
}

// ---------------------------------------------------------------------
// Management Methods (Delete / Rename)
// ---------------------------------------------------------------------

func (s *budgetService) DeleteBudgetFile(id string) error {
	return s.repo.DeleteFileBudget(id)
}
func (s *budgetService) DeleteCapexBudgetFile(id string) error {
	return s.repo.DeleteFileCapexBudget(id)
}
func (s *budgetService) DeleteCapexActualFile(id string) error {
	return s.repo.DeleteFileCapexActual(id)
}

func (s *budgetService) RenameBudgetFile(id string, newName string) error {
	return s.repo.UpdateFileBudget(id, newName)
}
func (s *budgetService) RenameCapexBudgetFile(id string, newName string) error {
	return s.repo.UpdateFileCapexBudget(id, newName)
}
func (s *budgetService) RenameCapexActualFile(id string, newName string) error {
	return s.repo.UpdateFileCapexActual(id, newName)
}

// ---------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------

func saveFileToDisk(fileHeader *multipart.FileHeader, subDir string, id string) error {
	// 1. Prepare Directory
	uploadDir := fmt.Sprintf("./uploads/%s", subDir)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return err
	}

	// 2. Open Source File
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 3. Create Dest File
	dstPath := filepath.Join(uploadDir, fmt.Sprintf("%s.xlsx", id))
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// 4. Copy
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func openExcelFile(fileHeader *multipart.FileHeader) (*excelize.File, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	return excelize.OpenReader(src)
}

func findTargetSheet(f *excelize.File, validator func([]string) bool) (string, error) {
	for _, name := range f.GetSheetList() {
		rows, err := f.GetRows(name, excelize.Options{RawCellValue: true})
		if err != nil || len(rows) == 0 {
			continue
		}
		for i := 0; i < 3 && i < len(rows); i++ {
			if validator(rows[i]) {
				return name, nil
			}
		}
	}
	return "", fmt.Errorf("not found")
}

func isBudgetHeader(row []string) bool {
	if len(row) > 1 && containsIgnoreCase(row[1], "Branch") {
		return true
	}
	return false
}

func isCapexBudgetHeader(row []string) bool {
	if len(row) > 2 && containsIgnoreCase(row[2], "CAPEX No") {
		return true
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

package service

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"

	"p2p-back-end/modules/entities/models"
)

type capexService struct {
	repo models.CapexRepository
}

func NewCapexService(repo models.CapexRepository) models.CapexService {
	return &capexService{repo: repo}
}

// ---------------------------------------------------------------------
// Helpers (Duplicated from Budget Service to keep independent)
// ---------------------------------------------------------------------

func extractYear(filename string) string {
	re := regexp.MustCompile(`\d{4}`)
	match := re.FindString(filename)
	if match != "" {
		// Convert to int to check range if needed, but string is fine
		return match
	}
	// Fallback: Current Year?
	return fmt.Sprintf("%d", time.Now().Year())
}

func parseDecimal(s string) decimal.Decimal {
	if s == "" {
		return decimal.Zero
	}
	cleanS := strings.ReplaceAll(s, ",", "")
	cleanS = strings.TrimSpace(cleanS)
	if cleanS == "-" {
		return decimal.Zero
	}

	d, err := decimal.NewFromString(cleanS)
	if err != nil {
		return decimal.Zero
	}
	return d
}

func getColSafe(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

// ---------------------------------------------------------------------
// Import & Sync
// ---------------------------------------------------------------------

func (s *capexService) ImportCapexBudget(fileHeader *multipart.FileHeader, userID string, versionName string) error {
	rows, err := parseExcelToJSON(fileHeader)
	if err != nil {
		return err
	}
	jsonData, _ := json.Marshal(rows)

	fileNameToSave := fileHeader.Filename
	if versionName != "" {
		fileNameToSave = versionName
	}
	year := extractYear(fileNameToSave)

	fileEntity := &models.FileCapexBudgetEntity{
		ID: uuid.New(), FileName: fileNameToSave, UploadAt: time.Now(), UserID: userID,
		Year: year,
		Data: datatypes.JSON(jsonData),
	}

	return s.repo.CreateFileCapexBudget(fileEntity)
}

func (s *capexService) ImportCapexActual(fileHeader *multipart.FileHeader, userID string, versionName string) error {
	rows, err := parseExcelToJSON(fileHeader)
	if err != nil {
		return err
	}
	jsonData, _ := json.Marshal(rows)

	fileNameToSave := fileHeader.Filename
	if versionName != "" {
		fileNameToSave = versionName
	}
	year := extractYear(fileNameToSave)

	fileEntity := &models.FileCapexActualEntity{
		ID: uuid.New(), FileName: fileNameToSave, UploadAt: time.Now(), UserID: userID,
		Year: year,
		Data: datatypes.JSON(jsonData),
	}

	return s.repo.CreateFileCapexActual(fileEntity)
}

// Redoing Sync to be strict

func (s *capexService) processCapexBudgetFact(rows [][]string, fileID uuid.UUID, year string) ([]models.CapexBudgetFactEntity, error) {
	var headers []models.CapexBudgetFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	if len(rows) < 2 {
		return []models.CapexBudgetFactEntity{}, nil
	}

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
		header := models.CapexBudgetFactEntity{
			ID:                headerID,
			FileCapexBudgetID: fileID,
			Entity:            entity, Department: dept, CapexNo: cNo, CapexName: cName, CapexCategory: cCat,
			Year:               year,
			YearTotal:          decimal.Zero,
			CapexBudgetAmounts: []models.CapexBudgetAmountEntity{},
		}

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
	return headers, nil
}

func (s *capexService) processCapexActualFact(rows [][]string, fileID uuid.UUID, year string) ([]models.CapexActualFactEntity, error) {
	var headers []models.CapexActualFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	if len(rows) < 2 {
		return []models.CapexActualFactEntity{}, nil
	}

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
			Year:               year,
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
	return headers, nil
}

// Re-implement Sync with correct processor calls
func (s *capexService) SyncCapexBudget(fileID string) error {
	fileEntity, err := s.repo.GetFileCapexBudget(fileID)
	if err != nil {
		return fmt.Errorf("file record not found: %v", err)
	}
	var rows [][]string
	if err := json.Unmarshal(fileEntity.Data, &rows); err != nil {
		return fmt.Errorf("failed to parse stored json data: %v", err)
	}

	return s.repo.WithTrx(func(trxRepo models.CapexRepository) error {
		if err := trxRepo.DeleteAllCapexBudgetFacts(); err != nil {
			return err
		}
		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processCapexBudgetFact(rows, parsedUUID, fileEntity.Year)
		if err != nil {
			return err
		}

		if len(headers) > 0 {
			return trxRepo.CreateCapexBudgetFacts(headers)
		}
		return nil
	})
}

func (s *capexService) SyncCapexActual(fileID string) error {
	fileEntity, err := s.repo.GetFileCapexActual(fileID)
	if err != nil {
		return fmt.Errorf("file record not found: %v", err)
	}
	var rows [][]string
	if err := json.Unmarshal(fileEntity.Data, &rows); err != nil {
		return fmt.Errorf("failed to parse stored json data: %v", err)
	}

	return s.repo.WithTrx(func(trxRepo models.CapexRepository) error {
		if err := trxRepo.DeleteAllCapexActualFacts(); err != nil {
			return err
		}
		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processCapexActualFact(rows, parsedUUID, fileEntity.Year)
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
// List Files
// ---------------------------------------------------------------------

func (s *capexService) ListCapexBudgetFiles() ([]models.FileCapexBudgetEntity, error) {
	return s.repo.ListFileCapexBudgets()
}

func (s *capexService) ListCapexActualFiles() ([]models.FileCapexActualEntity, error) {
	return s.repo.ListFileCapexActuals()
}

// ---------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------

func (s *capexService) DeleteCapexBudgetFile(id string) error {
	return s.repo.DeleteFileCapexBudget(id)
}

func (s *capexService) DeleteCapexActualFile(id string) error {
	return s.repo.DeleteFileCapexActual(id)
}

// ---------------------------------------------------------------------
// Rename
// ---------------------------------------------------------------------

func (s *capexService) RenameCapexBudgetFile(id string, newName string) error {
	return s.repo.UpdateFileCapexBudget(id, newName)
}

func (s *capexService) RenameCapexActualFile(id string, newName string) error {
	return s.repo.UpdateFileCapexActual(id, newName)
}

// ---------------------------------------------------------------------
// Dashboard
// ---------------------------------------------------------------------

func (s *capexService) GetCapexDashboardSummary(filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	return s.repo.GetCapexDashboardAggregates(filter)
}

// ---------------------------------------------------------------------
// Excel Parsing Helper
// ---------------------------------------------------------------------
func parseExcelToJSON(fileHeader *multipart.FileHeader) ([][]string, error) {
	f, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	excelFile, err := excelize.OpenReader(f)
	if err != nil {
		return nil, err
	}

	// Assuming first sheet
	sheetName := excelFile.GetSheetName(0)
	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// Placeholder since strict types used above

package service

import (
	"context"
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
	"p2p-back-end/pkg/excel"
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

var expectedCapexHeaders = []string{"Entity", "Department", "CAPEX No.", "CAPEX Name", "CAPEX Category", "JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC", "YEARTOTAL"}

func (s *capexService) ImportCapexBudget(ctx context.Context, fileHeader *multipart.FileHeader, userID string, versionName string) error {
	rows, err := excel.ParseExcelToJSONStrict(fileHeader, expectedCapexHeaders)
	if err != nil {
		return fmt.Errorf("capexSrv.ImportCapexBudget.Parse: %w", err)
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

	return s.repo.CreateFileCapexBudget(ctx, fileEntity)
}

func (s *capexService) ImportCapexActual(ctx context.Context, fileHeader *multipart.FileHeader, userID string, versionName string) error {
	rows, err := excel.ParseExcelToJSONStrict(fileHeader, expectedCapexHeaders)
	if err != nil {
		return fmt.Errorf("capexSrv.ImportCapexActual.Parse: %w", err)
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

	return s.repo.CreateFileCapexActual(ctx, fileEntity)
}

// Redoing Sync to be strict

func (s *capexService) processCapexBudgetFact(rows [][]string, fileID uuid.UUID, year string) ([]models.CapexBudgetFactEntity, error) {
	var headers []models.CapexBudgetFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	if len(rows) < 2 {
		return []models.CapexBudgetFactEntity{}, nil
	}

	headerRow := rows[0]
	colMap := make(map[string]int)
	for i, h := range headerRow {
		cleanHeader := strings.TrimSpace(strings.ToUpper(h))
		colMap[cleanHeader] = i
	}

	idxEntity := colMap["ENTITY"]
	idxDept := colMap["DEPARTMENT"]
	idxCNo := colMap["CAPEX NO."]
	idxCName := colMap["CAPEX NAME"]
	idxCCat := colMap["CAPEX CATEGORY"]

	monthIdxs := make([]int, 12)
	for m := 0; m < 12; m++ {
		monthIdxs[m] = colMap[months[m]]
	}

	for i, row := range rows {
		if i == 0 {
			continue // Skip Header
		}

		entity := getColSafe(row, idxEntity)
		dept := getColSafe(row, idxDept)
		cNo := getColSafe(row, idxCNo)
		cName := getColSafe(row, idxCName)
		cCat := getColSafe(row, idxCCat)

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
			colIdx := monthIdxs[mIdx]
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

	headerRow := rows[0]
	colMap := make(map[string]int)
	for i, h := range headerRow {
		cleanHeader := strings.TrimSpace(strings.ToUpper(h))
		colMap[cleanHeader] = i
	}

	idxEntity := colMap["ENTITY"]
	idxDept := colMap["DEPARTMENT"]
	idxCNo := colMap["CAPEX NO."]
	idxCName := colMap["CAPEX NAME"]
	idxCCat := colMap["CAPEX CATEGORY"]

	monthIdxs := make([]int, 12)
	for m := 0; m < 12; m++ {
		monthIdxs[m] = colMap[months[m]]
	}

	for i, row := range rows {
		if i == 0 {
			continue // Skip header
		}

		entity := getColSafe(row, idxEntity)
		dept := getColSafe(row, idxDept)
		cNo := getColSafe(row, idxCNo)
		cName := getColSafe(row, idxCName)
		cCat := getColSafe(row, idxCCat)

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
			colIdx := monthIdxs[mIdx]
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
func (s *capexService) SyncCapexBudget(ctx context.Context, fileID string) error {
	fileEntity, err := s.repo.GetFileCapexBudget(ctx, fileID)
	if err != nil {
		return fmt.Errorf("capexSrv.SyncCapexBudget.FetchFile: %w", err)
	}
	var rows [][]string
	if err := json.Unmarshal(fileEntity.Data, &rows); err != nil {
		return fmt.Errorf("capexSrv.SyncCapexBudget.Unmarshal: %w", err)
	}

	return s.repo.WithTrx(func(trxRepo models.CapexRepository) error {
		if err := trxRepo.DeleteAllCapexBudgetFacts(ctx); err != nil {
			return fmt.Errorf("transaction.DeleteOldFacts: %w", err)
		}
		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processCapexBudgetFact(rows, parsedUUID, fileEntity.Year)
		if err != nil {
			return fmt.Errorf("transaction.ProcessFacts: %w", err)
		}

		if len(headers) > 0 {
			if err := trxRepo.CreateCapexBudgetFacts(ctx, headers); err != nil {
				return fmt.Errorf("transaction.CreateFacts: %w", err)
			}
		}
		return nil
	})
}

func (s *capexService) SyncCapexActual(ctx context.Context, fileID string) error {
	fileEntity, err := s.repo.GetFileCapexActual(ctx, fileID)
	if err != nil {
		return fmt.Errorf("capexSrv.SyncCapexActual.FetchFile: %w", err)
	}
	var rows [][]string
	if err := json.Unmarshal(fileEntity.Data, &rows); err != nil {
		return fmt.Errorf("capexSrv.SyncCapexActual.Unmarshal: %w", err)
	}

	return s.repo.WithTrx(func(trxRepo models.CapexRepository) error {
		if err := trxRepo.DeleteAllCapexActualFacts(ctx); err != nil {
			return fmt.Errorf("transaction.DeleteOldFacts: %w", err)
		}
		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processCapexActualFact(rows, parsedUUID, fileEntity.Year)
		if err != nil {
			return fmt.Errorf("transaction.ProcessFacts: %w", err)
		}

		if len(headers) > 0 {
			if err := trxRepo.CreateCapexActualFacts(ctx, headers); err != nil {
				return fmt.Errorf("transaction.CreateFacts: %w", err)
			}
		}
		return nil
	})
}

// ---------------------------------------------------------------------
// List Files
// ---------------------------------------------------------------------

func (s *capexService) ListCapexBudgetFiles(ctx context.Context) ([]models.FileCapexBudgetEntity, error) {
	return s.repo.ListFileCapexBudgets(ctx)
}

func (s *capexService) ListCapexActualFiles(ctx context.Context) ([]models.FileCapexActualEntity, error) {
	return s.repo.ListFileCapexActuals(ctx)
}

// ---------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------

func (s *capexService) DeleteCapexBudgetFile(ctx context.Context, id string) error {
	return s.repo.DeleteFileCapexBudget(ctx, id)
}

func (s *capexService) DeleteCapexActualFile(ctx context.Context, id string) error {
	return s.repo.DeleteFileCapexActual(ctx, id)
}

// ---------------------------------------------------------------------
// Rename
// ---------------------------------------------------------------------

func (s *capexService) RenameCapexBudgetFile(ctx context.Context, id string, newName string) error {
	return s.repo.UpdateFileCapexBudget(ctx, id, newName)
}

func (s *capexService) RenameCapexActualFile(ctx context.Context, id string, newName string) error {
	return s.repo.UpdateFileCapexActual(ctx, id, newName)
}

// ---------------------------------------------------------------------
// Dashboard
// ---------------------------------------------------------------------

func (s *capexService) GetCapexDashboardSummary(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	return s.repo.GetCapexDashboardAggregates(ctx, filter)
}

// ---------------------------------------------------------------------
// Clear Data (Sync Empty)
// ---------------------------------------------------------------------

func (s *capexService) ClearCapexBudget(ctx context.Context) error {
	return s.repo.WithTrx(func(trxRepo models.CapexRepository) error {
		return trxRepo.DeleteAllCapexBudgetFacts(ctx)
	})
}

func (s *capexService) ClearCapexActual(ctx context.Context) error {
	return s.repo.WithTrx(func(trxRepo models.CapexRepository) error {
		return trxRepo.DeleteAllCapexActualFacts(ctx)
	})
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

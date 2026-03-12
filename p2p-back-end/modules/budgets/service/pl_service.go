package service

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/excel"
)

type plBudgetService struct {
	repo   models.PLBudgetRepository
	depSrv models.DepartmentService
}

func NewPLBudgetService(repo models.PLBudgetRepository, depSrv models.DepartmentService) models.PLBudgetService {
	return &plBudgetService{repo: repo, depSrv: depSrv}
}

func (s *plBudgetService) ImportBudget(fileHeader *multipart.FileHeader, userID string, versionName string) error {
	expectedHeaders := []string{"Entity", "Branch", "Entity GL", "Conso GL", "GROUP1", "GL Name", "Department", "JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC", "YEARTOTAL"}
	
	rows, err := excel.ParseExcelToJSONStrict(fileHeader, expectedHeaders)
	if err != nil {
		return err // This error will now clearly state missing columns if validation fails
	}
	jsonData, _ := json.Marshal(rows)

	fileNameToSave := fileHeader.Filename
	if versionName != "" {
		fileNameToSave = versionName
	}
	fileEntity := &models.FileBudgetEntity{
		ID: uuid.New(), FileName: fileNameToSave, UploadAt: time.Now(), UserID: userID,
		Data: datatypes.JSON(jsonData),
	}

	return s.repo.CreateFileBudget(fileEntity)
}

func (s *plBudgetService) SyncBudget(fileID string) error {
	fmt.Printf("[DEBUG] SyncBudget: Starting for FileID %s\n", fileID)
	fileEntity, err := s.repo.GetFileBudget(fileID)
	if err != nil {
		fmt.Printf("[DEBUG] SyncBudget: GetFileBudget failed: %v\n", err)
		return fmt.Errorf("file record not found: %v", err)
	}
	fmt.Printf("[DEBUG] SyncBudget: Found File. Data Size: %d bytes\n", len(fileEntity.Data))

	var rows [][]string
	if err := json.Unmarshal(fileEntity.Data, &rows); err != nil {
		fmt.Printf("[DEBUG] SyncBudget: Unmarshal failed: %v\n", err)
		return fmt.Errorf("failed to parse stored json data: %v", err)
	}
	fmt.Printf("[DEBUG] SyncBudget: Unmarshaled %d rows\n", len(rows))

	// 2. Transaction: Delete All & Insert New
	err = s.repo.WithTrx(func(trxRepo models.PLBudgetRepository) error {
		// Delete All Existing Data
		fmt.Println("[DEBUG] SyncBudget: Deleting old facts...")
		if err := trxRepo.DeleteAllBudgetFacts(); err != nil {
			fmt.Printf("[DEBUG] SyncBudget: DeleteAllBudgetFacts failed: %v\n", err)
			return err
		}

		// Process & Insert
		parsedUUID, _ := uuid.Parse(fileID)
		year := extractYear(fileEntity.FileName)
		if year == "" {
			year = "2026" // Fallback to current fiscal year
		}
		fmt.Printf("[DEBUG] SyncBudget: Processing facts for Year %s...\n", year)
		headers, err := s.processBudgetFact(rows, parsedUUID, year)
		if err != nil {
			fmt.Printf("[DEBUG] SyncBudget: processBudgetFact failed: %v\n", err)
			return err
		}
		fmt.Printf("[DEBUG] SyncBudget: To Insert %d headers\n", len(headers))

		if len(headers) > 0 {
			fmt.Println("[DEBUG] SyncBudget: Creating facts in DB...")
			if err := trxRepo.CreateBudgetFacts(headers); err != nil {
				fmt.Printf("[DEBUG] SyncBudget: CreateBudgetFacts failed: %v\n", err)
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("[DEBUG] SyncBudget: Transaction failed: %v\n", err)
		return err
	}
	fmt.Println("[DEBUG] SyncBudget: Success")
	return nil
}

func (s *plBudgetService) ClearBudget() error {
	return s.repo.WithTrx(func(trxRepo models.PLBudgetRepository) error {
		return trxRepo.DeleteAllBudgetFacts()
	})
}

func (s *plBudgetService) ListBudgetFiles() ([]models.FileBudgetEntity, error) {
	return s.repo.ListFileBudgets()
}

func (s *plBudgetService) DeleteBudgetFile(id string) error {
	return s.repo.DeleteFileBudget(id)
}

func (s *plBudgetService) RenameBudgetFile(id string, newName string) error {
	return s.repo.UpdateFileBudget(id, newName)
}

func (s *plBudgetService) processBudgetFact(rows [][]string, fileID uuid.UUID, year string) ([]models.BudgetFactEntity, error) {
	if len(rows) < 2 {
		// User requested to allow empty files (just header) to clear data
		fmt.Println("[DEBUG] ProcessBudget: File has only header or empty. Returning empty list (Clear Data).")
		return []models.BudgetFactEntity{}, nil
	}

	headerRow := rows[0]
	fmt.Printf("[Debug] Header Row: %v\n", headerRow)

	// Since we strictly validated against expectedHeaders before insert, we know exactly what columns exist
	// We'll dynamically find their precise index in case the user shuffled the order (e.g., put JAN before Entity)
	
	colMap := make(map[string]int)
	for i, h := range headerRow {
		cleanHeader := strings.TrimSpace(strings.ToUpper(h))
		colMap[cleanHeader] = i
	}

	// Lookup precisely by the known exact names
	idxEntity := colMap["ENTITY"]
	idxBranch := colMap["BRANCH"]
	idxEntityGL := colMap["ENTITY GL"]
	idxConsoGL := colMap["CONSO GL"]
	idxGroup := colMap["GROUP1"]
	idxGLName := colMap["GL NAME"]
	idxDept := colMap["DEPARTMENT"]

	var headers []models.BudgetFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	
	// Create map for month indexes
	monthIdxs := make([]int, 12)
	for m := 0; m < 12; m++ {
		monthIdxs[m] = colMap[months[m]]
	}

	// --- Normalization Helpers (Copied from SyncActuals to ensure consistency) ---
	entityNameMap := map[string]string{
		"HONDA MALIWAN":    "HMW",
		"AUTOCORP HOLDING": "ACG",
		"CLIK":             "CLIK",
		// Add commonly used variations in Budget Files if known, or rely on normalize
	}
	normalize := func(s string) string {
		return strings.TrimSpace(strings.ToUpper(s))
	}
	mapToCode := func(rawVal string, m map[string]string) string {
		norm := normalize(rawVal)
		if code, ok := m[norm]; ok {
			return code
		}
		// Fallback: Check if rawVal itself is a valid code
		return norm
	}

	for i, row := range rows {
		if i == 0 {
			continue // Skip Header
		}

		rawEntity := getColSafe(row, idxEntity)
		branch := getColSafe(row, idxBranch)
		entityGL := getColSafe(row, idxEntityGL)
		consoGL := getColSafe(row, idxConsoGL)
		group := getColSafe(row, idxGroup)
		glName := getColSafe(row, idxGLName)
		rawDept := getColSafe(row, idxDept)

		// Normalize Entity & Department for Lookup
		entity := mapToCode(rawEntity, entityNameMap)
		deptLookup := normalize(rawDept)

		// Apply Department Mapping
		dept := rawDept // Default to original
		originalDept := rawDept

		if master, err := s.depSrv.GetMasterDepartment(deptLookup, entity); err == nil && master != nil {
			dept = master.Code
		} else {
			dept = strings.TrimSpace(rawDept)
		}

		// Validate crucial data (optional)
		if group == "" && dept == "" && entityGL == "" {
			continue // Skip empty line
		}

		headerID := uuid.New()
		header := models.BudgetFactEntity{
			ID:           headerID,
			FileBudgetID: fileID,
			Entity:       entity, Branch: branch, Group: group, EntityGL: entityGL, ConsoGL: consoGL, GLName: glName,
			Department: dept, NavCode: originalDept, // Save both
			Year:          year,
			YearTotal:     decimal.Zero,
			BudgetAmounts: []models.BudgetAmountEntity{},
		}

		for mIdx := 0; mIdx < 12; mIdx++ {
			colIdx := monthIdxs[mIdx]
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

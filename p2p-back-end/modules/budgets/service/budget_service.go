package service

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"

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
	rows, err := parseExcelToJSON(fileHeader, isBudgetHeader)
	if err != nil {
		return err
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

// ---------------------------------------------------------------------
// 2. Import Capex Budget
// ---------------------------------------------------------------------
func (s *budgetService) ImportCapexBudget(fileHeader *multipart.FileHeader, userID string, versionName string) error {
	rows, err := parseExcelToJSON(fileHeader, isCapexBudgetHeader)
	if err != nil {
		return err
	}
	jsonData, _ := json.Marshal(rows)

	fileNameToSave := fileHeader.Filename
	if versionName != "" {
		fileNameToSave = versionName
	}
	fileEntity := &models.FileCapexBudgetEntity{
		ID: uuid.New(), FileName: fileNameToSave, UploadAt: time.Now(), UserID: userID,
		Data: datatypes.JSON(jsonData),
	}

	return s.repo.CreateFileCapexBudget(fileEntity)
}

// ---------------------------------------------------------------------
// 3. Import Capex Actual
// ---------------------------------------------------------------------
func (s *budgetService) ImportCapexActual(fileHeader *multipart.FileHeader, userID string, versionName string) error {
	rows, err := parseExcelToJSON(fileHeader, isCapexBudgetHeader)
	if err != nil {
		return err
	}
	jsonData, _ := json.Marshal(rows)

	fileNameToSave := fileHeader.Filename
	if versionName != "" {
		fileNameToSave = versionName
	}
	fileEntity := &models.FileCapexActualEntity{
		ID: uuid.New(), FileName: fileNameToSave, UploadAt: time.Now(), UserID: userID,
		Data: datatypes.JSON(jsonData),
	}

	return s.repo.CreateFileCapexActual(fileEntity)
}

// Sync Budget - Process & Replace Data
func (s *budgetService) SyncBudget(fileID string) error {
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
	err = s.repo.WithTrx(func(trxRepo models.BudgetRepository) error {
		// Delete All Existing Data
		fmt.Println("[DEBUG] SyncBudget: Deleting old facts...")
		if err := trxRepo.DeleteAllBudgetFacts(); err != nil {
			fmt.Printf("[DEBUG] SyncBudget: DeleteAllBudgetFacts failed: %v\n", err)
			return err
		}

		// Process & Insert
		parsedUUID, _ := uuid.Parse(fileID)
		fmt.Println("[DEBUG] SyncBudget: Processing facts...")
		headers, err := s.processBudgetFact(rows, parsedUUID)
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

// Sync Capex Budget
func (s *budgetService) SyncCapexBudget(fileID string) error {
	fileEntity, err := s.repo.GetFileCapexBudget(fileID)
	if err != nil {
		return fmt.Errorf("file record not found: %v", err)
	}

	var rows [][]string
	if err := json.Unmarshal(fileEntity.Data, &rows); err != nil {
		return fmt.Errorf("failed to parse stored json data: %v", err)
	}

	return s.repo.WithTrx(func(trxRepo models.BudgetRepository) error {
		if err := trxRepo.DeleteAllCapexBudgetFacts(); err != nil {
			return err
		}

		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processCapexBudgetFact(rows, parsedUUID)
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
	fileEntity, err := s.repo.GetFileCapexActual(fileID)
	if err != nil {
		return fmt.Errorf("file record not found: %v", err)
	}

	var rows [][]string
	if err := json.Unmarshal(fileEntity.Data, &rows); err != nil {
		return fmt.Errorf("failed to parse stored json data: %v", err)
	}

	return s.repo.WithTrx(func(trxRepo models.BudgetRepository) error {
		if err := trxRepo.DeleteAllCapexActualFacts(); err != nil {
			return err
		}

		parsedUUID, _ := uuid.Parse(fileID)
		headers, err := s.processCapexActualFact(rows, parsedUUID)
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
// 4. Sync Actuals (P2P / Operational)
// ---------------------------------------------------------------------

func (s *budgetService) SyncActuals(year string) error {
	fmt.Printf("[DEBUG] SyncActuals: Start DB Sync (Optimized) for Year %s...\n", year)

	// 2. Fetch Aggregated Data (Source)
	// We optimize by aggregating in DB (GROUP BY ...)
	hmwAgg, err := s.repo.GetAggregatedHMW(year)
	if err != nil {
		return fmt.Errorf("failed to fetch HMW aggregated data: %v", err)
	}
	fmt.Printf("[DEBUG] SyncActuals: HMW Rows Fetched for %s: %d\n", year, len(hmwAgg))

	clikAgg, err := s.repo.GetAggregatedCLIK(year)
	if err != nil {
		return fmt.Errorf("failed to fetch CLIK aggregated data: %v", err)
	}
	fmt.Printf("[DEBUG] SyncActuals: CLIK Rows Fetched for %s: %d\n", year, len(clikAgg))
	fmt.Printf("[DEBUG] SyncActuals: Fetched %d HMW rows, %d CLIK rows (Aggregated)\n", len(hmwAgg), len(clikAgg))

	// 3. Merge HMW + CLIK
	// Use Map to merge duplicates across sources (if any, though unlikely given distinct tables)
	// Key: Entity|Branch|Dept|GLCode|Month
	type AggKey struct {
		Entity, Branch, Dept, GLCode, GLName, Month string
	}
	mergedMap := make(map[AggKey]decimal.Decimal)

	// --- MAPPING LOGIC START ---
	// User Requirement: Map Source Names -> Codes for Database Storage

	// 1. Entity Map (Name -> Code)
	entityNameMap := map[string]string{
		"HONDA MALIWAN":    "HMW",
		"AUTOCORP HOLDING": "ACG",
		"CLIK":             "CLIK",
	}

	// 2. Branch Map (Name -> Code)
	branchNameMap := map[string]string{
		// HMW Branches
		"BURIRUM":      "BUR",
		"HEAD OFFICE":  "HOF",
		"KRABI":        "KBI",
		"MINI_SURIN":   "MSR",
		"MUEANG KRABI": "MKB",
		"NAKA":         "NAK",
		"NANGRONG":     "AVN",
		"PHACHA":       "PHC",
		"PHUKET":       "PRA",
		"SURIN":        "SUR",
		"VEERAWAT":     "VEE",

		// ACG Branches
		"AUTOCORP HEAD OFFICE": "HQ",

		// CLIK Branches (Normalize to Title Case or Code)
		"":           "Branch00", // Empty -> Branch00
		"BRANCH01":   "Branch01",
		"BRANCH02":   "Branch02",
		"BRANCH03":   "Branch03",
		"BRANCH04":   "Branch04",
		"BRANCH05":   "Branch05",
		"BRANCH06":   "Branch06",
		"BRANCH07":   "Branch07",
		"BRANCH08":   "Branch08",
		"BRANCH09":   "Branch09",
		"BRANCH10":   "Branch10",
		"BRANCH11":   "Branch11",
		"BRANCH12":   "Branch12",
		"BRANCH13":   "Branch13",
		"BRANCH14":   "Branch14",
		"BRANCH15":   "Branch15",
		"HEADOFFICE": "HOF", // Mapping CLIK HEADOFFICE to HOF to align with others
	}

	normalize := func(s string) string {
		return strings.TrimSpace(strings.ToUpper(s))
	}

	mapToCode := func(rawVal string, m map[string]string) string {
		norm := normalize(rawVal)
		if code, ok := m[norm]; ok {
			return code
		}
		// Fallback: If no map found, return original properly trimmed
		// Specific fix for CLIK: If it differs only by case (e.g. branch01), Map handles it if keys are UPPER.
		// If rawVal is empty string and not in map (though we added ""), return ""
		if rawVal == "" {
			if v, ok := m[""]; ok {
				return v
			}
		}
		return strings.TrimSpace(rawVal)
	}

	processAgg := func(rows []models.ActualAggregatedDTO) {
		for _, row := range rows {
			// Apply Mapping
			company := mapToCode(row.Company, entityNameMap)
			branch := mapToCode(row.Branch, branchNameMap)

			k := AggKey{
				Entity: company, Branch: branch, Dept: row.Department,
				GLCode: row.GLAccountNo, GLName: row.GLAccountName, Month: row.Month,
			}
			mergedMap[k] = mergedMap[k].Add(row.TotalAmount)
		}
	}

	processAgg(hmwAgg)
	processAgg(clikAgg)

	// 4. Transform to Entities (Header + Details)
	type HeaderKey struct {
		Entity, Branch, Dept, GLCode, GLName string
	}
	headerMap := make(map[HeaderKey][]models.ActualAmountEntity)

	for k, amt := range mergedMap {
		hk := HeaderKey{k.Entity, k.Branch, k.Dept, k.GLCode, k.GLName}
		headerMap[hk] = append(headerMap[hk], models.ActualAmountEntity{
			ID:     uuid.New(),
			Month:  k.Month,
			Amount: amt,
		})
	}

	var headers []models.ActualFactEntity
	for k, amounts := range headerMap {
		headerID := uuid.New()
		for i := range amounts {
			amounts[i].ActualFactID = headerID
		}
		total := decimal.Zero
		for _, a := range amounts {
			total = total.Add(a.Amount)
		}

		headers = append(headers, models.ActualFactEntity{
			ID:            headerID,
			Entity:        k.Entity,
			Branch:        k.Branch,
			Department:    k.Dept,
			ConsoGL:       k.GLCode,
			GLName:        k.GLName,
			YearTotal:     total,
			ActualAmounts: amounts,
			Year:          year, // ✅ Populate Year
		})
	}

	fmt.Printf("[DEBUG] SyncActuals: Merged into %d headers\n", len(headers))

	// 5. Save (Transaction)
	return s.repo.WithTrx(func(trxRepo models.BudgetRepository) error {
		// ✅ Delete ALL Existing Actuals (Clean Slate - Simple Logic)
		// This ensures what you see is what you just synced. No mixing years.
		if err := trxRepo.DeleteAllActualFacts(); err != nil {
			return err
		}
		// Batch Insert in Chunks
		chunkSize := 100 // Safe for deep structures
		if len(headers) > 0 {
			for i := 0; i < len(headers); i += chunkSize {
				end := i + chunkSize
				if end > len(headers) {
					end = len(headers)
				}
				if err := trxRepo.CreateActualFacts(headers[i:end]); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *budgetService) DeleteActualFacts(year string) error {
	if year == "" {
		return fmt.Errorf("year is required")
	}
	return s.repo.DeleteActualFactsByYear(year)
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

func (s *budgetService) processBudgetFact(rows [][]string, fileID uuid.UUID) ([]models.BudgetFactEntity, error) {
	if len(rows) < 2 {
		// User requested to allow empty files (just header) to clear data
		fmt.Println("[DEBUG] ProcessBudget: File has only header or empty. Returning empty list (Clear Data).")
		return []models.BudgetFactEntity{}, nil
	}

	headerRow := rows[0]
	fmt.Printf("[Debug] Header Row: %v\n", headerRow)

	// Dynamic Column Mapping
	colMap := make(map[string]int)
	for i, h := range headerRow {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Helper to find index by multiple possible names
	findCol := func(names ...string) int {
		for _, n := range names {
			if idx, ok := colMap[strings.ToLower(n)]; ok {
				return idx
			}
		}
		return -1
	}

	// Map required columns (Defaults to current indices if not found, to be safe?)
	// Actually, let's try to be smart. If headers found, use them. If not, fallback or error.
	// Common names based on user context
	idxEntity := findCol("entity", "company")
	idxBranch := findCol("branch", "cost center", "cost_center")
	idxEntityGL := findCol("entity gl", "entity_gl", "gl category", "gl_category")
	// Use ConsoGL field to store "GL Code" if present
	idxConsoGL := findCol("conso gl", "conso_gl", "gl code", "code", "gl_code")
	idxGroup := findCol("group", "budget group", "category")
	idxGLName := findCol("gl name", "gl_name", "description", "account name")
	idxDept := findCol("department", "dept")

	// If crucial columns missing, fallback to hardcoded (Legacy support)
	if idxGroup == -1 {
		idxGroup = 4
	}
	if idxDept == -1 {
		idxDept = 6
	}
	if idxEntityGL == -1 {
		idxEntityGL = 2
	}
	if idxGLName == -1 {
		idxGLName = 5
	}

	// Fallback for others
	if idxEntity == -1 {
		idxEntity = 0
	}
	if idxBranch == -1 {
		idxBranch = 1
	}
	if idxConsoGL == -1 {
		idxConsoGL = 3
	}

	fmt.Printf("[Debug] Mapped Columns - Group:%d, Dept:%d, EntityGL:%d, Name:%d\n", idxGroup, idxDept, idxEntityGL, idxGLName)

	var headers []models.BudgetFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	for i, row := range rows {
		if i == 0 {
			continue
		} // Skip Header

		// Safety check using max index we need
		if len(row) < 7 {
			continue // Skip malformed rows
		}

		entity := getColSafe(row, idxEntity)
		branch := getColSafe(row, idxBranch)
		entityGL := getColSafe(row, idxEntityGL)
		consoGL := getColSafe(row, idxConsoGL)
		group := getColSafe(row, idxGroup)
		glName := getColSafe(row, idxGLName)
		dept := getColSafe(row, idxDept)

		// Validate crucial data (optional)
		if group == "" && dept == "" && entityGL == "" {
			continue // Skip empty line
		}

		headerID := uuid.New()
		header := models.BudgetFactEntity{
			ID:           headerID,
			FileBudgetID: fileID,
			Entity:       entity, Branch: branch, Group: group, EntityGL: entityGL, ConsoGL: consoGL, GLName: glName, Department: dept,
			YearTotal:     decimal.Zero,
			BudgetAmounts: []models.BudgetAmountEntity{},
		}

		// Fixed Loop for 12 months (Assuming they follow Department or are at fixed offset)
		// For months, it's harder to dynamic map without precise names.
		// Let's assume they mimic the original offset (Col 7+) OR search for "JAN", "FEB"...
		// Search for JAN index
		idxJan := findCol("jan", "january")
		if idxJan == -1 {
			idxJan = 7
		} // Fallback to 7

		for mIdx := 0; mIdx < 12; mIdx++ {
			colIdx := idxJan + mIdx
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

func (s *budgetService) processCapexBudgetFact(rows [][]string, fileID uuid.UUID) ([]models.CapexBudgetFactEntity, error) {
	var headers []models.CapexBudgetFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	fmt.Printf("[Debug] Capex Plan - Total Rows: %d\n", len(rows))

	if len(rows) < 2 {
		fmt.Println("[DEBUG] ProcessCapexBudget: File has only header. Returning empty list.")
		return []models.CapexBudgetFactEntity{}, nil
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}
		// Minimal check: Entity(0) to Category(4)
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

func (s *budgetService) processCapexActualFact(rows [][]string, fileID uuid.UUID) ([]models.CapexActualFactEntity, error) {
	var headers []models.CapexActualFactEntity
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	fmt.Printf("[Debug] Capex Actual - Total Rows: %d\n", len(rows))

	if len(rows) < 2 {
		fmt.Println("[DEBUG] ProcessCapexActual: File has only header. Returning empty list.")
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
// Dashboard / Detail View
// ---------------------------------------------------------------------

func (s *budgetService) GetFilterOptions() ([]models.FilterOptionDTO, error) {
	fmt.Println("[DEBUG] Service: GetFilterOptions START")
	facts, err := s.repo.GetBudgetFilterOptions()
	if err != nil {
		fmt.Printf("[DEBUG] Service: Repo returned error: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Service: Got %d facts\n", len(facts))

	// Nested Map Structure: Group -> Dept -> EntityGL -> []Leaf
	// We use a map for the first 3 levels, and a slice for the 4th (or map if needing unique)
	// Tree: map[Group] -> map[Dept] -> map[EntityGL] -> []{Code, Name}
	type Leaf struct {
		Code string
		Name string
	}
	tree := make(map[string]map[string]map[string][]Leaf)

	for _, f := range facts {
		groupName := f.Group
		if groupName == "" {
			groupName = "(No Group)"
		}
		deptName := f.Department
		if deptName == "" {
			deptName = "(No Dept)"
		}
		entityGLName := f.EntityGL
		if entityGLName == "" {
			entityGLName = "(No Category)"
		}

		// L4 Name: "Code - Name"
		code := f.ConsoGL
		name := f.GLName
		// If code is missing, maybe rely on name?

		if tree[groupName] == nil {
			tree[groupName] = make(map[string]map[string][]Leaf)
		}
		if tree[groupName][deptName] == nil {
			tree[groupName][deptName] = make(map[string][]Leaf)
		}

		// Add Leaf if unique. Since we query distinct, it should be mostly unique,
		// but multiple rows might share same code/name if other fields differ?
		// The repo query distincts on group,dept,entity_gl,conso_gl,gl_name. So it is unique per branch.

		// Check duplicates in slice? (Inefficient but fine for small N)
		leaves := tree[groupName][deptName][entityGLName]
		exists := false
		for _, l := range leaves {
			if l.Code == code && l.Name == name {
				exists = true
				break
			}
		}
		if !exists {
			tree[groupName][deptName][entityGLName] = append(leaves, Leaf{Code: code, Name: name})
		}
	}

	// Convert Map to DTO Slice
	var rootNodes []models.FilterOptionDTO

	for grpName, deptMap := range tree {
		grpNode := models.FilterOptionDTO{
			ID:       grpName,
			Name:     grpName,
			Level:    1,
			Children: []models.FilterOptionDTO{},
		}

		for deptName, glMap := range deptMap {
			deptNode := models.FilterOptionDTO{
				ID:       deptName,
				Name:     deptName,
				Level:    2,
				Children: []models.FilterOptionDTO{},
			}

			for glName, leaves := range glMap {
				glNode := models.FilterOptionDTO{
					ID:       glName,
					Name:     glName,
					Level:    3,
					Children: []models.FilterOptionDTO{},
				}

				// Level 4 Leaves
				for _, leaf := range leaves {
					displayName := leaf.Name
					if leaf.Code != "" {
						displayName = fmt.Sprintf("%s-%s", leaf.Code, leaf.Name)
					}

					leafNode := models.FilterOptionDTO{
						ID:    leaf.Code, // Or composite? Usage depends on Frontend.
						Name:  displayName,
						Level: 4,
					}
					// If ID matches ConsoGL (Code), selecting it works for existing logic?
					// Frontend uses ID to filter. If ID is "Code", we need to make sure filter logic handles it.
					// Or we can use Name as ID if that's what we want to display.
					// Let's use Code as ID for precision if available.
					if leafNode.ID == "" {
						leafNode.ID = leafNode.Name
					}

					glNode.Children = append(glNode.Children, leafNode)
				}

				deptNode.Children = append(deptNode.Children, glNode)
			}
			grpNode.Children = append(grpNode.Children, deptNode)
		}
		rootNodes = append(rootNodes, grpNode)
	}

	return rootNodes, nil
}

func (s *budgetService) GetOrganizationStructure() ([]models.OrganizationDTO, error) {
	facts, err := s.repo.GetOrganizationStructure()
	if err != nil {
		return nil, err
	}

	// Return Raw Codes (HMW, BUR) as stored in DB.
	// User requested "Abbreviations" in filter.

	// Map Entity -> Map Branch -> []Departments
	structure := make(map[string]map[string][]string)
	for _, f := range facts {
		if f.Entity == "" {
			continue
		}

		// Use Raw Entity Code
		entityName := f.Entity
		if structure[entityName] == nil {
			structure[entityName] = make(map[string][]string)
		}

		// Use Raw Branch Code
		if f.Branch != "" {
			branchName := f.Branch
			if _, exists := structure[entityName][branchName]; !exists {
				structure[entityName][branchName] = []string{}
			}

			// Add Department if not exists
			if f.Department != "" {
				deptName := f.Department
				found := false
				for _, d := range structure[entityName][branchName] {
					if d == deptName {
						found = true
						break
					}
				}
				if !found {
					structure[entityName][branchName] = append(structure[entityName][branchName], deptName)
				}
			}
		}
	}

	var result []models.OrganizationDTO
	for entity, branchesMap := range structure {
		var branchDTOs []models.BranchDTO
		for branch, depts := range branchesMap {
			// Sort Departments
			// sort.Strings(depts) // Optional
			branchDTOs = append(branchDTOs, models.BranchDTO{
				Name:        branch,
				Departments: depts,
			})
		}
		// Sort Branches (Optional but good for UI)
		// ...

		result = append(result, models.OrganizationDTO{
			Entity:   entity,
			Branches: branchDTOs,
		})
	}
	return result, nil
}

// Helper to extract Code from "Code - Name" format
func extractCode(s string) string {
	if strings.Contains(s, " - ") {
		parts := strings.SplitN(s, " - ", 2)
		return strings.TrimSpace(parts[0])
	}
	return s
}

// Helper to sanitize filter map (Single or Slice)
func sanitizeFilter(filter map[string]interface{}) {

	// Normalize keys: Ensure entities/branches exist provided entity/branch exist
	if v, ok := filter["entity"]; ok {
		if _, exists := filter["entities"]; !exists {
			filter["entities"] = v
		}
	}
	if v, ok := filter["branch"]; ok {
		if _, exists := filter["branches"]; !exists {
			filter["branches"] = v
		}
	}
	// Normalize keys: Department
	if v, ok := filter["department"]; ok {
		if _, exists := filter["departments"]; !exists {
			filter["departments"] = v
		}
	}

	targetKeys := []string{"entities", "branches", "departments"}
	for _, key := range targetKeys {
		val, ok := filter[key]
		if !ok {
			continue
		}

		var finalSlice []string

		// Case 1: Single String
		if s, ok := val.(string); ok && s != "" {
			finalSlice = append(finalSlice, extractCode(s))
		} else if ss, ok := val.([]string); ok {
			// Case 2: Slice of Strings
			for _, v := range ss {
				finalSlice = append(finalSlice, extractCode(v))
			}
		} else if ifaceSlice, ok := val.([]interface{}); ok {
			// Case 3: Slice of Interface
			for _, v := range ifaceSlice {
				if s, ok := v.(string); ok {
					finalSlice = append(finalSlice, extractCode(s))
				}
			}
		}

		// Update Filter with correct type []string
		if len(finalSlice) > 0 {
			filter[key] = finalSlice
		} else {
			// If empty, remove to avoid confusion or empty IN clause issues
			delete(filter, key)
		}
	}
}

func (s *budgetService) GetBudgetDetails(filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	sanitizeFilter(filter)
	return s.repo.GetBudgetDetails(filter)
}

func (s *budgetService) GetActualDetails(filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	sanitizeFilter(filter)
	return s.repo.GetActualDetails(filter)
}

func (s *budgetService) GetDashboardSummary(filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	fmt.Printf("[DEBUG] GetDashboardSummary Filter (Before): %+v\n", filter)
	sanitizeFilter(filter)
	fmt.Printf("[DEBUG] GetDashboardSummary Filter (After): %+v\n", filter)
	return s.repo.GetDashboardAggregates(filter)
}

func (s *budgetService) GetActualTransactions(filter map[string]interface{}) ([]models.ActualTransactionDTO, error) {
	sanitizeFilter(filter)
	return s.repo.GetActualTransactions(filter)
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

func parseExcelToJSON(fileHeader *multipart.FileHeader, validator func([]string) bool) ([][]string, error) {
	fmt.Printf("[DEBUG] Parsing Excel File: %s\n", fileHeader.Filename)
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		fmt.Printf("[DEBUG] OpenReader failed: %v\n", err)
		return nil, err
	}
	defer f.Close()

	sheetName, err := findTargetSheet(f, validator)
	if err != nil {
		fmt.Printf("[DEBUG] Target Sheet Not Found: %v\n", err)
		return nil, fmt.Errorf("invalid file format: missing required columns")
	}
	fmt.Printf("[DEBUG] Found Target Sheet: %s\n", sheetName)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[DEBUG] Read %d rows from sheet\n", len(rows))
	return rows, nil
}

func findTargetSheet(f *excelize.File, validator func([]string) bool) (string, error) {
	for _, name := range f.GetSheetList() {
		rows, err := f.GetRows(name, excelize.Options{RawCellValue: true})
		if err != nil || len(rows) == 0 {
			continue
		}
		// Log check
		// fmt.Printf("[DEBUG] Checking Sheet: %s (Rows: %d)\n", name, len(rows))

		for i := 0; i < 3 && i < len(rows); i++ {
			if validator(rows[i]) {
				fmt.Printf("[DEBUG] Valid Header Found in Sheet: %s at row %d\n", name, i)
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

func (s *budgetService) GetRawDate() (string, error) {
	return s.repo.GetRawDate()
}

package repository

import (
	"fmt"
	"p2p-back-end/modules/entities/models"
	"sort"

	"gorm.io/gorm"
)

type budgetRepositoryDB struct {
	db *gorm.DB
}

func NewBudgetRepositoryDB(db *gorm.DB) models.BudgetRepository {
	return &budgetRepositoryDB{db: db}
}

// Transaction helper
func (r *budgetRepositoryDB) WithTrx(trxHandle func(repo models.BudgetRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewBudgetRepositoryDB(tx)
		return trxHandle(repo)
	})
}

// ---------------------------------------------------------
// File Create Methods
// ---------------------------------------------------------

func (r *budgetRepositoryDB) CreateFileBudget(file *models.FileBudgetEntity) error {
	return r.db.Create(file).Error
}

func (r *budgetRepositoryDB) CreateFileCapexBudget(file *models.FileCapexBudgetEntity) error {
	return r.db.Create(file).Error
}

func (r *budgetRepositoryDB) CreateFileCapexActual(file *models.FileCapexActualEntity) error {
	return r.db.Create(file).Error
}

// ---------------------------------------------------------
// Fact Create Methods (Batch Insert + Association)
// ---------------------------------------------------------

// 1. Budget (PL)
func (r *budgetRepositoryDB) CreateBudgetFacts(headers []models.BudgetFactEntity) error {
	// GORM CreateInBatches ไม่บันทึก Association (Amounts) โดยอัตโนมัติ
	// เราต้องแยกบันทึก Header และ Amount เองเพื่อประสิทธิภาพ 100%

	// 1.1 Insert Headers
	if err := r.db.Omit("BudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	// 1.2 Collect All Amounts
	var allAmounts []models.BudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.BudgetAmounts...)
	}

	// 1.3 Insert Amounts
	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// 2. Capex Budget
func (r *budgetRepositoryDB) CreateCapexBudgetFacts(headers []models.CapexBudgetFactEntity) error {
	if err := r.db.Omit("CapexBudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	var allAmounts []models.CapexBudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.CapexBudgetAmounts...)
	}

	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// 3. Capex Actual
func (r *budgetRepositoryDB) CreateCapexActualFacts(headers []models.CapexActualFactEntity) error {
	if err := r.db.Omit("CapexActualAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	var allAmounts []models.CapexActualAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.CapexActualAmounts...)
	}

	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// ---------------------------------------------------------
// File List Methods
// ---------------------------------------------------------

func (r *budgetRepositoryDB) ListFileBudgets() ([]models.FileBudgetEntity, error) {
	var files []models.FileBudgetEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *budgetRepositoryDB) ListFileCapexBudgets() ([]models.FileCapexBudgetEntity, error) {
	var files []models.FileCapexBudgetEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *budgetRepositoryDB) ListFileCapexActuals() ([]models.FileCapexActualEntity, error) {
	var files []models.FileCapexActualEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *budgetRepositoryDB) GetFileBudget(id string) (*models.FileBudgetEntity, error) {
	var file models.FileBudgetEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *budgetRepositoryDB) GetFileCapexBudget(id string) (*models.FileCapexBudgetEntity, error) {
	var file models.FileCapexBudgetEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *budgetRepositoryDB) GetFileCapexActual(id string) (*models.FileCapexActualEntity, error) {
	var file models.FileCapexActualEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// ---------------------------------------------------------------------
// Dashboard / Detail View
// ---------------------------------------------------------------------

func (r *budgetRepositoryDB) GetBudgetFilterOptions() ([]models.BudgetFactEntity, error) {
	fmt.Println("[DEBUG] Repo: GetBudgetFilterOptions START")
	var results []models.BudgetFactEntity
	// Select distinct combinations for hierarchy building
	err := r.db.Model(&models.BudgetFactEntity{}).
		Distinct("\"group\"", "department", "entity_gl", "conso_gl", "gl_name").
		Order("\"group\", department, entity_gl, conso_gl").
		Find(&results).Error
	fmt.Printf("[DEBUG] Repo: GetBudgetFilterOptions END - Count: %d, Err: %v\n", len(results), err)
	return results, err
}

func (r *budgetRepositoryDB) GetOrganizationStructure() ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	// Union Distinct Entities and Branches from both Budget and Actual tables
	// GORM doesn't support UNION natively in a clean way for struct scanning without raw SQL usually,
	// but we can use Raw SQL for performance and clarity here.

	query := `
        SELECT DISTINCT entity, branch FROM budget_fact_entities WHERE entity != ''
        UNION
        SELECT DISTINCT entity, branch FROM actual_fact_entities WHERE entity != ''
        ORDER BY entity, branch
    `

	err := r.db.Raw(query).Scan(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetBudgetDetails(filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	query := r.db.Model(&models.BudgetFactEntity{}).Preload("BudgetAmounts")

	// Dynamic Filtering Helper
	applyFilter := func(key string, dbCol string) {
		if val, ok := filter[key]; ok {
			var strs []string
			if s, ok := val.([]string); ok {
				strs = s
			} else if s, ok := val.([]interface{}); ok {
				for _, item := range s {
					strs = append(strs, fmt.Sprintf("%v", item))
				}
			}
			if len(strs) > 0 {
				query = query.Where(fmt.Sprintf("%s IN ?", dbCol), strs)
			}
		}
	}

	applyFilter("groups", "\"group\"")
	applyFilter("departments", "department")
	applyFilter("entity_gls", "entity_gl")
	applyFilter("conso_gls", "conso_gl")
	applyFilter("entities", "entity")
	applyFilter("branches", "branch")

	err := query.Order("\"group\", department, entity_gl, conso_gl, gl_name").Find(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetActualDetails(filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	var results []models.ActualFactEntity
	query := r.db.Model(&models.ActualFactEntity{}).Preload("ActualAmounts")

	// Dynamic Filtering Helper
	applyFilter := func(key string, dbCol string) {
		if val, ok := filter[key]; ok {
			var strs []string
			if s, ok := val.([]string); ok {
				strs = s
			} else if s, ok := val.([]interface{}); ok {
				for _, item := range s {
					strs = append(strs, fmt.Sprintf("%v", item))
				}
			}
			if len(strs) > 0 {
				query = query.Where(fmt.Sprintf("%s IN ?", dbCol), strs)
			}
		}
	}

	// Dimensions for Actuals: Entity, Branch, Department, ConsoGL (Code), GLName
	// Note: Actuals might not have "Group" or "EntityGL" if not mapped yet.
	applyFilter("departments", "department")
	applyFilter("conso_gls", "conso_gl")
	applyFilter("entities", "entity")
	applyFilter("branches", "branch")

	// Order by logic
	err := query.Order("department, conso_gl, gl_name").Find(&results).Error
	return results, err
}

// ---------------------------------------------------------
// File Delete Methods
// ---------------------------------------------------------

func (r *budgetRepositoryDB) DeleteFileBudget(id string) error {
	return r.db.Delete(&models.FileBudgetEntity{}, "id = ?", id).Error
}

func (r *budgetRepositoryDB) DeleteFileCapexBudget(id string) error {
	return r.db.Delete(&models.FileCapexBudgetEntity{}, "id = ?", id).Error
}

func (r *budgetRepositoryDB) DeleteFileCapexActual(id string) error {
	return r.db.Delete(&models.FileCapexActualEntity{}, "id = ?", id).Error
}

// ---------------------------------------------------------------------
// 4. Delete All Facts (For Sync)
// ---------------------------------------------------------------------
// ---------------------------------------------------------------------
// 4. Delete All Facts (For Sync)
// ---------------------------------------------------------------------
func (r *budgetRepositoryDB) DeleteAllBudgetFacts() error {
	// 1. Delete Amounts (Children)
	if err := r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.BudgetAmountEntity{}).Error; err != nil {
		return err
	}
	// 2. Delete Headers (Parents)
	return r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.BudgetFactEntity{}).Error
}

func (r *budgetRepositoryDB) DeleteAllCapexBudgetFacts() error {
	// 1. Delete Amounts
	if err := r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.CapexBudgetAmountEntity{}).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.CapexBudgetFactEntity{}).Error
}

func (r *budgetRepositoryDB) DeleteAllCapexActualFacts() error {
	// 1. Delete Amounts
	if err := r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.CapexActualAmountEntity{}).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.CapexActualFactEntity{}).Error
}

// ---------------------------------------------------------------------
// 5. Update Files (Rename) - Implementation
// ---------------------------------------------------------------------

func (r *budgetRepositoryDB) UpdateFileBudget(id string, filename string) error {
	return r.db.Model(&models.FileBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

func (r *budgetRepositoryDB) UpdateFileCapexBudget(id string, filename string) error {
	return r.db.Model(&models.FileCapexBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

func (r *budgetRepositoryDB) UpdateFileCapexActual(id string, filename string) error {
	return r.db.Model(&models.FileCapexActualEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

// ---------------------------------------------------------------------
// 6. Actuals (Operational) - Sync Implementation
// ---------------------------------------------------------------------

func (r *budgetRepositoryDB) CreateActualFacts(headers []models.ActualFactEntity) error {
	// 1. Insert Headers
	if err := r.db.Omit("ActualAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}
	// 2. Collect Amounts
	var allAmounts []models.ActualAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.ActualAmounts...)
	}
	// 3. Insert Amounts
	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

func (r *budgetRepositoryDB) DeleteAllActualFacts() error {
	// 1. Delete Amounts
	if err := r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ActualAmountEntity{}).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ActualFactEntity{}).Error
}

func (r *budgetRepositoryDB) DeleteActualFactsByYear(year string) error {
	// 1. Delete Amounts linked to Headers of that Year
	// (Requires join or subquery? Usually GORM handles Cascade delete if configured, but let's be safe)
	// Actually, Amounts don't have Year. We must find Headers first.
	// Subquery delete: DELETE FROM actual_amount_entities WHERE actual_fact_id IN (SELECT id FROM actual_fact_entities WHERE year = ?)
	if err := r.db.Exec(`
		DELETE FROM actual_amount_entities 
		WHERE actual_fact_id IN (SELECT id FROM actual_fact_entities WHERE year = ?)
	`, year).Error; err != nil {
		return err
	}

	// 2. Delete Headers
	return r.db.Where("year = ?", year).Delete(&models.ActualFactEntity{}).Error
}

func (r *budgetRepositoryDB) GetAggregatedHMW(year string) ([]models.ActualAggregatedDTO, error) {
	var results []models.ActualAggregatedDTO
	// Optimization: Group by Database
	// Postgres TO_CHAR(date, 'MON') returns 'JAN', 'FEB'... (uppercase)
	// Posting_Date is likely VARCHAR in DB, so cast to DATE first
	err := r.db.Table("achhmw_gle_api").
		Select(`
			company, 
			branch, 
			"Global_Dimension_1_Code" as department, 
			"G_L_Account_No" as gl_account_no, 
			"G_L_Account_Name" as gl_account_name, 
			UPPER(TO_CHAR("Posting_Date"::DATE, 'MON')) as month, 
			SUM("Credit_Amount") as total_amount
		`).
		Where("LEFT(\"Posting_Date\", 4) = ?", year). // Fix: Use String manipulation to avoid Date Cast crashes
		Group(`company, branch, "Global_Dimension_1_Code", "G_L_Account_No", "G_L_Account_Name", UPPER(TO_CHAR("Posting_Date"::DATE, 'MON'))`).
		Scan(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetAllAchHmwGle() ([]models.AchHmwGleEntity, error) {
	var results []models.AchHmwGleEntity
	err := r.db.Find(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetAggregatedCLIK(year string) ([]models.ActualAggregatedDTO, error) {
	var results []models.ActualAggregatedDTO
	// CLIK uses Global_Dimension_2_Code for Branch
	err := r.db.Table("general_ledger_entries_clik").
		Select(`
			'CLIK' as company,
			"Global_Dimension_2_Code" as branch, 
			"Global_Dimension_1_Code" as department, 
			"G_L_Account_No" as gl_account_no, 
			"G_L_Account_Name" as gl_account_name, 
			UPPER(TO_CHAR("Posting_Date"::DATE, 'MON')) as month, 
			SUM("Credit_Amount") as total_amount
		`).
		Where("LEFT(\"Posting_Date\", 4) = ?", year). // Fix: Use String manipulation
		Group(`"Global_Dimension_2_Code", "Global_Dimension_1_Code", "G_L_Account_No", "G_L_Account_Name", UPPER(TO_CHAR("Posting_Date"::DATE, 'MON'))`).
		Scan(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetAllClikGle() ([]models.ClikGleEntity, error) {
	var results []models.ClikGleEntity
	err := r.db.Find(&results).Error
	return results, err
}

// ---------------------------------------------------------------------
// 7. Dashboard Aggregation (Optimized)
// ---------------------------------------------------------------------
func (r *budgetRepositoryDB) GetActualTransactions(filter map[string]interface{}) ([]models.ActualTransactionDTO, error) {
	var results []models.ActualTransactionDTO

	// Filtering Logic
	whereClause := "1=1"
	var args []interface{}

	// Filter by GL Account No (Required for Drill Down)
	// User Requirement: Map Conso GL (Filter) -> Entity GL (Source)
	if val, ok := filter["conso_gls"]; ok {
		var consoGLs []string
		if s, ok := val.([]string); ok {
			consoGLs = s
		} else if s, ok := val.([]interface{}); ok {
			for _, item := range s {
				consoGLs = append(consoGLs, fmt.Sprintf("%v", item))
			}
		}

		if len(consoGLs) > 0 {
			// Step 1: Find corresponding Entity GLs from Mapping Table (budget_fact_entities)
			var entityGLs []string
			r.db.Model(&models.BudgetFactEntity{}).
				Where("conso_gl IN ?", consoGLs).
				Distinct("entity_gl").
				Pluck("entity_gl", &entityGLs)

			// Step 2: Filter Source Tables by Entity GLs
			if len(entityGLs) > 0 {
				whereClause += " AND \"G_L_Account_No\" IN ?"
				args = append(args, entityGLs)
			} else {
				// If no mapping found, maybe fallback to ConsoGLs themselves or return empty
				// For safety, let's include ConsoGLs too in case some match directly (though user said no)
				whereClause += " AND \"G_L_Account_No\" IN ?"
				args = append(args, consoGLs)
			}
		}
	}

	// Date Filtering
	if val, ok := filter["start_date"]; ok {
		if startDate, ok := val.(string); ok && startDate != "" {
			whereClause += " AND \"Posting_Date\"::DATE >= ?"
			args = append(args, startDate)
		}
	}
	if val, ok := filter["end_date"]; ok {
		if endDate, ok := val.(string); ok && endDate != "" {
			whereClause += " AND \"Posting_Date\"::DATE <= ?"
			args = append(args, endDate)
		}
	}

	// Helper to apply limit/order to subqueries
	applyLimit := func(db *gorm.DB) *gorm.DB {
		return db.Order("\"Posting_Date\" ASC").Limit(2000)
	}

	// ✅ FILTER LOGIC UPDATE:
	// Only show transactions for years that are currently "Synced" (Active in actual_fact_entities)
	// This matches the user expectation that "Syncing 2026" hides 2025 data.
	var activeYears []string
	r.db.Model(&models.ActualFactEntity{}).Distinct("year").Pluck("year", &activeYears)

	// If no data is synced, return empty immediately
	if len(activeYears) == 0 {
		return []models.ActualTransactionDTO{}, nil
	}

	// Add Year Filter to WhereClause
	whereClause += " AND TO_CHAR(\"Posting_Date\"::DATE, 'YYYY') IN ?"
	args = append(args, activeYears)

	// Query HMW
	queryHMW := r.db.Table("achhmw_gle_api").
		Select(`
			'HMW' as source,
			TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date,
			"Document_No" as doc_no,
			"Description" as description, 
			"G_L_Account_No" as gl_account_no,
			"G_L_Account_Name" as gl_account_name,
			"Global_Dimension_1_Code" as department,
			"Credit_Amount" as amount
		`).
		Where(whereClause, args...).
		Scopes(applyLimit)

	// Query CLIK
	queryCLIK := r.db.Table("general_ledger_entries_clik").
		Select(`
			'CLIK' as source,
			TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date,
			"Document_No" as doc_no, 
			"Description" as description,
			"G_L_Account_No" as gl_account_no,
			"G_L_Account_Name" as gl_account_name,
			"Global_Dimension_1_Code" as department,
			"Credit_Amount" as amount
		`).
		Where(whereClause, args...).
		Scopes(applyLimit)

	// Union
	var hmwRows []models.ActualTransactionDTO
	if err := queryHMW.Scan(&hmwRows).Error; err != nil {
		return nil, err
	}
	var clikRows []models.ActualTransactionDTO
	if err := queryCLIK.Scan(&clikRows).Error; err != nil {
		return nil, err
	}

	results = append(results, hmwRows...)
	results = append(results, clikRows...)

	// Sort by Date Desc (Since we appended two sorted lists, we strictly need to sort again or just accept pseudo-sort)
	// For 4000 rows, returning as-is is fine, frontend can sort or we sort here if critical.
	// Let's rely on map order being "good enough" for drill-down or let frontend DataGrid sort.
	return results, nil
}

func (r *budgetRepositoryDB) GetDashboardAggregates(filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	summary := &models.DashboardSummaryDTO{
		DepartmentData: []models.DepartmentStatDTO{},
		ChartData:      []models.MonthlyStatDTO{},
	}

	// Dynamic Filter Helper (Returns GORM Scope)
	applyFilter := func(tx *gorm.DB, tableName string) *gorm.DB {
		if val, ok := filter["entities"]; ok {
			if strs, ok := val.([]string); ok && len(strs) > 0 {
				tx = tx.Where(tableName+".entity IN ?", strs)
			}
		}
		if val, ok := filter["branches"]; ok {
			if strs, ok := val.([]string); ok && len(strs) > 0 {
				tx = tx.Where(tableName+".branch IN ?", strs)
			}
		}
		// Added: Departments Filter
		if val, ok := filter["departments"]; ok {
			if strs, ok := val.([]string); ok && len(strs) > 0 {
				tx = tx.Where(tableName+".department IN ?", strs)
			}
		}
		return tx
	}

	// 1. Department Aggregation
	// Budget
	type DeptResult struct {
		Department string
		Total      float64
	}
	var budgetDept []DeptResult
	tx1 := r.db.Table("budget_fact_entities").Select("department, SUM(year_total) as total")
	tx1 = applyFilter(tx1, "budget_fact_entities")
	if err := tx1.Group("department").Scan(&budgetDept).Error; err != nil {
		return nil, err
	}

	// Actual
	var actualDept []DeptResult
	tx2 := r.db.Table("actual_fact_entities").Select("department, SUM(year_total) as total")
	tx2 = applyFilter(tx2, "actual_fact_entities")

	if err := tx2.Group("department").Scan(&actualDept).Error; err != nil {
		// Log error but maybe continue? No, return error
		return nil, err
	}

	// Merge Department Data
	deptMap := make(map[string]*models.DepartmentStatDTO)
	for _, b := range budgetDept {
		deptMap[b.Department] = &models.DepartmentStatDTO{Department: b.Department, Budget: b.Total}
		summary.TotalBudget += b.Total
	}
	for _, a := range actualDept {
		if _, ok := deptMap[a.Department]; !ok {
			deptMap[a.Department] = &models.DepartmentStatDTO{Department: a.Department}
		}
		deptMap[a.Department].Actual += a.Total
		summary.TotalActual += a.Total
	}

	// Flatten Map
	var allDepts []models.DepartmentStatDTO
	for _, v := range deptMap {
		allDepts = append(allDepts, *v)
	}

	// Sort Logic
	sortBy := "actual" // Default
	sortOrder := "desc"
	if val, ok := filter["sort_by"]; ok {
		if s, ok := val.(string); ok && s != "" {
			sortBy = s
		}
	}
	if val, ok := filter["sort_order"]; ok {
		if s, ok := val.(string); ok && s != "" {
			sortOrder = s
		}
	}

	// Calculate Global Status Counts (Before Pagination)
	var overBudgetCount, nearLimitCount int
	for _, d := range allDepts {
		budget := d.Budget
		actual := d.Actual
		remaining := budget - actual

		// Over Budget: (Budget=0 & Actual>0) OR (Remaining < 0)
		if (budget == 0 && actual > 0) || remaining < 0 {
			overBudgetCount++
		} else if budget > 0 {
			// Near Limit: Remaining < 20%
			ratio := remaining / budget
			if ratio < 0.2 {
				nearLimitCount++
			}
		}
	}
	summary.OverBudgetCount = overBudgetCount
	summary.NearLimitCount = nearLimitCount

	sort.Slice(allDepts, func(i, j int) bool {
		var valI, valJ float64
		switch sortBy {
		case "budget":
			valI, valJ = allDepts[i].Budget, allDepts[j].Budget
		case "actual": // Spend
			valI, valJ = allDepts[i].Actual, allDepts[j].Actual
		case "remaining":
			valI = allDepts[i].Budget - allDepts[i].Actual
			valJ = allDepts[j].Budget - allDepts[j].Actual
		default:
			valI, valJ = allDepts[i].Actual, allDepts[j].Actual
		}

		if sortOrder == "asc" {
			return valI < valJ
		}
		return valI > valJ
	})

	// Pagination Logic
	page := 1
	limit := 10
	if val, ok := filter["page"]; ok {
		if p, ok := val.(float64); ok { // JSON unmarshal makes numbers float64
			page = int(p)
		} else if p, ok := val.(int); ok {
			page = p
		}
	}
	if val, ok := filter["limit"]; ok {
		if l, ok := val.(float64); ok {
			limit = int(l)
		} else if l, ok := val.(int); ok {
			limit = l
		}
	}

	summary.TotalCount = int64(len(allDepts))
	summary.Page = page
	summary.Limit = limit

	start := (page - 1) * limit
	end := start + limit

	if start > len(allDepts) {
		summary.DepartmentData = []models.DepartmentStatDTO{}
	} else {
		if end > len(allDepts) {
			end = len(allDepts)
		}
		summary.DepartmentData = allDepts[start:end]
	}

	// 2. Monthly Aggregation for Chart
	// This is trickier because we need to join with Amount tables.
	// Budget Amounts
	type MonthResult struct {
		Month string
		Total float64
	}
	var budgetMonth []MonthResult
	// Join Header to filter -> Sum Amount
	tx3 := r.db.Table("budget_amount_entities").
		Select("budget_amount_entities.month, SUM(budget_amount_entities.amount) as total").
		Joins("JOIN budget_fact_entities ON budget_amount_entities.budget_fact_id = budget_fact_entities.id")
	tx3 = applyFilter(tx3, "budget_fact_entities")
	if err := tx3.Group("budget_amount_entities.month").Scan(&budgetMonth).Error; err != nil {
		return nil, err
	}

	// Actual Amounts
	var actualMonth []MonthResult
	tx4 := r.db.Table("actual_amount_entities").
		Select("actual_amount_entities.month, SUM(actual_amount_entities.amount) as total").
		Joins("JOIN actual_fact_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id")
	tx4 = applyFilter(tx4, "actual_fact_entities")
	if err := tx4.Group("actual_amount_entities.month").Scan(&actualMonth).Error; err != nil {
		return nil, err
	}

	// Merge Chart Data
	monthMap := make(map[string]*models.MonthlyStatDTO)
	// Initialize 12 months? or just map what we have
	// Let's rely on what we have, frontend usually handles ordering or we fix it.
	// We'll normalize keys to JAN,FEB...
	for _, m := range budgetMonth {
		// m.Month might be "January", "JAN", etc. Assumed stored as JAN,FEB from Import.
		monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: m.Total}
	}
	for _, m := range actualMonth {
		if _, ok := monthMap[m.Month]; !ok {
			monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month}
		}
		monthMap[m.Month].Actual += m.Total
	}

	// Ensure logical order if possible, or just slice
	// Order: JAN, FEB, MAR...
	monthsOrder := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	for _, mon := range monthsOrder {
		if val, ok := monthMap[mon]; ok {
			summary.ChartData = append(summary.ChartData, *val)
		} else {
			summary.ChartData = append(summary.ChartData, models.MonthlyStatDTO{Month: mon, Budget: 0, Actual: 0})
		}
	}

	return summary, nil
}

// Debugging
func (r *budgetRepositoryDB) GetRawDate() (string, error) {
	var rawDate string
	// Try HMW first
	if err := r.db.Table("achhmw_gle_api").Select("\"Posting_Date\"").Limit(1).Scan(&rawDate).Error; err != nil {
		return "", err
	}
	if rawDate != "" {
		return fmt.Sprintf("HMW Date: %s", rawDate), nil
	}

	// Try CLIK if HMW empty
	if err := r.db.Table("acclik_gle_api").Select("\"Posting_Date\"").Limit(1).Scan(&rawDate).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("CLIK Date: %s", rawDate), nil
}

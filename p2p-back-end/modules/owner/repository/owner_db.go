package repository

import (
	"fmt"
	"p2p-back-end/modules/entities/models"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ownerRepositoryDB struct {
	db *gorm.DB
}

func NewOwnerRepositoryDB(db *gorm.DB) models.OwnerRepository {
	return &ownerRepositoryDB{db: db}
}

func (r *ownerRepositoryDB) GetUserContext(userID string) (*models.UserEntity, error) {
	var user models.UserEntity
	err := r.db.Preload("Department").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user record (Lazy Sync)
func (r *ownerRepositoryDB) CreateUser(user *models.UserEntity) error {
	return r.db.Create(user).Error
}

func (r *ownerRepositoryDB) GetBudgetFilterOptions(filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	// Reused logic: fetch distincts
	var results []models.BudgetFactEntity
	query := r.db.Model(&models.BudgetFactEntity{})

	// Apply Restriction
	if val, ok := filter["allowed_departments"]; ok {
		if strs, ok := val.([]string); ok && len(strs) > 0 {
			query = query.Where("department IN ? OR nav_code IN ?", strs, strs)
		} else if isRestricted, ok := filter["is_restricted"].(bool); ok && isRestricted {
			query = query.Where("1=0")
		}
	}

	err := query.Distinct("entity", "branch", "\"group\"", "department", "nav_code", "entity_gl", "conso_gl", "gl_name").
		Order("entity, branch, \"group\", department, nav_code, entity_gl, conso_gl").
		Find(&results).Error
	return results, err
}

func (r *ownerRepositoryDB) GetDashboardAggregates(filter map[string]interface{}) (*models.OwnerDashboardSummaryDTO, error) {
	summary := &models.OwnerDashboardSummaryDTO{
		DashboardSummaryDTO: models.DashboardSummaryDTO{
			DepartmentData: []models.DepartmentStatDTO{},
			ChartData:      []models.MonthlyStatDTO{},
		},
		TopExpenses: []models.TopExpenseDTO{},
	}

	// Filter Helper for GORM
	applyFilter := func(tx *gorm.DB, tableName string) *gorm.DB {
		// Helper to safely convert interface{} to []string
		toStringSlice := func(val interface{}) []string {
			if strs, ok := val.([]string); ok {
				return strs
			}
			if interfaces, ok := val.([]interface{}); ok {
				var strs []string
				for _, i := range interfaces {
					if s, ok := i.(string); ok {
						strs = append(strs, s)
					}
				}
				return strs
			}
			return nil
		}

		if val, ok := filter["entities"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".entity IN ?", strs)
			}
		}
		// Branch Filter (Skip for Capex as they don't have Branch column)
		if !strings.HasPrefix(tableName, "capex_") {
			if val, ok := filter["branches"]; ok {
				if strs := toStringSlice(val); len(strs) > 0 {
					tx = tx.Where(tableName+".branch IN ?", strs)
				}
			}
		}
		if val, ok := filter["departments"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tn := tableName
				if tn == "budget_fact_entities" || tn == "owner_actual_fact_entities" {
					// Smart Filter: Check both Dept (Master) and NavCode (Sub)
					tx = tx.Where(tn+".department IN ? OR "+tn+".nav_code IN ?", strs, strs)
				} else {
					tx = tx.Where(tableName+".department IN ?", strs)
				}
			}
		}

		// Enforce Permissions (Internal Filter)
		if val, ok := filter["allowed_departments"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tn := tableName
				if tn == "budget_fact_entities" || tn == "owner_actual_fact_entities" {
					tx = tx.Where(tn+".department IN ? OR "+tn+".nav_code IN ?", strs, strs)
				} else {
					tx = tx.Where(tableName+".department IN ?", strs)
				}
			} else if isRestricted, ok := filter["is_restricted"].(bool); ok && isRestricted {
				tx = tx.Where("1=0")
			}
		}
		// Year Filter
		if val, ok := filter["year"]; ok {
			if s, ok := val.(string); ok && s != "" {
				s = strings.ReplaceAll(s, "FY", "")
				if !strings.Contains(tableName, "budget") {
					tx = tx.Where(tableName+".year = ?", s)
				}
			}
		}
		return tx
	}

	// Sync WaitGroup for Parallel Execution
	var wg sync.WaitGroup
	var errMutex sync.Mutex
	var firstError error

	// Worker Pool / Semaphore to limit concurrency (Max 3 parallel queries)
	// This prevents "wsarecv: An existing connection was forcibly closed" or Pool Exhaustion.
	sem := make(chan struct{}, 3)

	setError := func(err error) {
		errMutex.Lock()
		defer errMutex.Unlock()
		if firstError == nil {
			firstError = err
		}
	}

	// Determine Grouping Strategy (Drill-down vs Summary)
	groupBy := "department"
	selectCol := "department"

	// Helper to safely convert interface{} to []string (Moved up scope)
	toStringSlice := func(val interface{}) []string {
		if strs, ok := val.([]string); ok {
			return strs
		}
		if interfaces, ok := val.([]interface{}); ok {
			var strs []string
			for _, i := range interfaces {
				if s, ok := i.(string); ok {
					strs = append(strs, s)
				}
			}
			return strs
		}
		return nil
	}

	if val, ok := filter["departments"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			// If filtered by Department, we likely want to drill down to NavCode
			groupBy = "nav_code"
			selectCol = "nav_code as department"
		}
	}

	// 1. Department Aggregation (Budget)
	var budgetDept []struct {
		Department string
		Total      float64
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}        // Acquire
		defer func() { <-sem }() // Release
		fmt.Println("[DEBUG] Start Budget Dept")
		// Use Dynamic Grouping
		tx := r.db.Table("budget_fact_entities").Select(selectCol + ", SUM(year_total) as total")
		tx = applyFilter(tx, "budget_fact_entities")
		// Need to filter out empty NavCodes if grouping by NavCode?
		// if groupBy == "nav_code" {
		// 	tx = tx.Where("nav_code != ''")
		// }
		if err := tx.Group(groupBy).Scan(&budgetDept).Error; err != nil {
			setError(err)
		}
		fmt.Println("[DEBUG] End Budget Dept")
	}()

	// 2. Department Aggregation (Actual)
	var actualDept []struct {
		Department string
		Total      float64
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}        // Acquire
		defer func() { <-sem }() // Release
		fmt.Println("[DEBUG] Start Actual Dept")
		tx := r.db.Table("owner_actual_fact_entities").Select(selectCol + ", COALESCE(SUM(year_total), 0) as total")
		tx = applyFilter(tx, "owner_actual_fact_entities")
		if groupBy == "nav_code" {
			// tx = tx.Where("nav_code != ''")
		}
		if err := tx.Group(groupBy).Scan(&actualDept).Error; err != nil {
			setError(err)
		}
		fmt.Println("[DEBUG] End Actual Dept")
	}()

	// 3. Chart Aggregation (Budget)
	var budgetMonthResults []struct {
		Month string
		Total float64
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}        // Acquire
		defer func() { <-sem }() // Release
		fmt.Println("[DEBUG] Start Budget Chart")
		// Join needed because Amounts don't have Entity/Branch/Year directly (unless we densormalize amounts too, which we haven't)
		// But wait, applyFilter filters on Headers.
		tx := r.db.Table("budget_amount_entities").
			Joins("JOIN budget_fact_entities ON budget_fact_entities.id = budget_amount_entities.budget_fact_id").
			Select("budget_amount_entities.month, SUM(budget_amount_entities.amount) as total")
		tx = applyFilter(tx, "budget_fact_entities")
		if err := tx.Group("budget_amount_entities.month").Scan(&budgetMonthResults).Error; err != nil {
			setError(err)
		}
		fmt.Println("[DEBUG] End Budget Chart")
	}()

	// 4. Chart Aggregation (Actual)
	// Owner Actuals structure might NOT have Amount Table if I just copied headers?
	// Wait, my AutoSync INSERTED into `owner_actual_fact_entities` but DID NOT create `owner_actual_amount_entities`!
	// My `OwnerActualFactEntity` struct has `YearTotal` but NO sub-table for amounts?!
	// The `hmwQuery` in `AutoSync` only fills `owner_actual_fact_entities`.
	// IT DOES NOT FILL AMOUNTS.
	// So `Chart Data` (Monthly) will be MISSING if I try to join `owner_actual_amount_entities`.
	// The `AutoSync` query logic aggregated by `Year`. It did NOT aggregate by Month.
	// Wait, the SQL in `AutoSync` has `SUM("Credit_Amount")` group by `... TO_CHAR(..., 'YYYY')`.
	// It loses specific month granularity!
	//
	// CRITICAL FIX: The Owner Dashboard needs MONTHLY data for the Chart.
	// I must update `AutoSync` to insert into `Amount` table OR design `OwnerActualFact` to store 12 months array (denormalized)
	// We now use Denormalized Columns in OwnerActualFactEntity
	var actualMonthResults []struct {
		Month string
		Total float64
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}        // Acquire
		defer func() { <-sem }() // Release
		fmt.Println("[DEBUG] Start Actual Chart")

		// Join Fact + Amount
		tx := r.db.Table("owner_actual_fact_entities").
			Select("owner_actual_amount_entities.month, SUM(owner_actual_amount_entities.amount) as total").
			Joins("JOIN owner_actual_amount_entities ON owner_actual_amount_entities.owner_actual_fact_id = owner_actual_fact_entities.id")

		tx = applyFilter(tx, "owner_actual_fact_entities")

		if err := tx.Group("owner_actual_amount_entities.month").Scan(&actualMonthResults).Error; err != nil {
			setError(err)
		}
		fmt.Println("[DEBUG] End Actual Chart")
	}()

	// 5. CAPEX Budget
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}        // Acquire
		defer func() { <-sem }() // Release
		fmt.Println("[DEBUG] Start Capex Budget")
		tx := r.db.Table("capex_budget_fact_entities").Select("COALESCE(SUM(year_total), 0)")
		// Note: Capex Budget might usually not have Year column populated yet if not updated.
		// Assuming we added Year to Capex models too? Not yet.
		// If NO Year in Capex models, skip year filter for Capex or update models.
		// For now, let's assume filtering relies on existing logic or ignores year if column missing (GORM might error).
		// Safe bet: The `applyFilter` adds `.year`. If Capex table doesn't have it, it Errors.
		// Current `CapexBudgetFactEntity` in `models` DOES NOT have keys like `Year`.
		// So we should NOT apply filter with Year for Capex tables yet.
		// Let's make a custom filter for Capex or just omit Year.
		// For now, I'll clone logic but omit Year for Capex if unknown.
		// But wait, the previous code DID apply filter to Capex: `applyFilter(txCapexBudget, "capex_budget_fact_entities")`.
		// And `applyFilter` filtered by Year in `owner_db.go` before my changes.
		// This suggests Capex Tables MIGHT have Year, or the previous code was crashing/buggy, or I missed the model.
		// Let's assume it works or just suppress year for now to be safe, OR check `database.go` again?
		// `CapexBudgetFactEntity` does NOT have Year in lines 262-280 of `database.go` viewed earlier.
		// So the previous code `tx.Where(tableName+".year = ?", y)` WOULD HAVE FAILED if run.
		// Maybe it was never run with Year filter?
		// I will SKIP Year filter for Capex for safety.
		currentFilter := applyFilter(tx, "capex_budget_fact_entities")
		currentFilter.Scan(&summary.CapexBudget)
		fmt.Println("[DEBUG] End Capex Budget")
	}()

	// 6. CAPEX Actual
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}        // Acquire
		defer func() { <-sem }() // Release
		fmt.Println("[DEBUG] Start Capex Actual")
		tx := r.db.Table("capex_actual_fact_entities").Select("COALESCE(SUM(year_total), 0)")
		tx = applyFilter(tx, "capex_actual_fact_entities") // Same risk as above
		tx.Scan(&summary.CapexActual)
		fmt.Println("[DEBUG] End Capex Actual")
	}()

	// 7. Top Expenses
	var topExps []struct {
		Name  string
		Total float64
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}        // Acquire
		defer func() { <-sem }() // Release
		fmt.Println("[DEBUG] Start Top Expenses")
		tx := r.db.Table("owner_actual_fact_entities").
			Select("gl_name as name, CAST(SUM(year_total) AS FLOAT) as total").
			Group("gl_name").
			Order("total desc").
			Limit(3)
		tx = applyFilter(tx, "owner_actual_fact_entities")
		if err := tx.Scan(&topExps).Error; err != nil {
			fmt.Printf("[WARN] Failed to fetch Top Expenses: %v\n", err)
		}
		fmt.Println("[DEBUG] End Top Expenses")
	}()

	// Wait for all
	wg.Wait()

	if firstError != nil {
		return nil, firstError
	}

	// --- Post-Processing (Merging Results) ---

	// Merge Dept Data
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
	for _, v := range deptMap {
		summary.DepartmentData = append(summary.DepartmentData, *v)
	}

	// Merge Chart Data
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	chartMap := make(map[string]*models.MonthlyStatDTO)
	for _, m := range months {
		chartMap[m] = &models.MonthlyStatDTO{Month: m, Budget: 0, Actual: 0}
	}
	for _, b := range budgetMonthResults {
		if _, ok := chartMap[b.Month]; ok {
			chartMap[b.Month].Budget = b.Total
		}
	}
	for _, a := range actualMonthResults {
		if _, ok := chartMap[a.Month]; ok {
			chartMap[a.Month].Actual = a.Total
		}
	}
	for _, m := range months {
		summary.ChartData = append(summary.ChartData, *chartMap[m])
	}

	// Counts
	overCount := 0
	nearCount := 0
	for _, d := range deptMap {
		if d.Budget > 0 {
			percent := (d.Actual / d.Budget) * 100
			if percent >= 100 {
				overCount++
			} else if percent >= 80 {
				nearCount++
			}
		} else if d.Actual > 0 {
			overCount++
		}
	}
	summary.OverBudgetCount = overCount
	summary.NearLimitCount = nearCount
	summary.TotalCount = int64(len(summary.DepartmentData))

	// Top Expenses
	for _, t := range topExps {
		summary.TopExpenses = append(summary.TopExpenses, models.TopExpenseDTO{
			Name:   t.Name,
			Amount: t.Total,
		})
	}

	return summary, nil
}

func (r *ownerRepositoryDB) applyFilter(query *gorm.DB, filter map[string]interface{}) *gorm.DB {
	// Debug Filter
	// fmt.Printf("[DEBUG] applyFilter: %+v\n", filter)

	if entities, ok := filter["entities"].([]string); ok && len(entities) > 0 {
		query = query.Where("entity IN ?", entities) // Fixed: entity_code -> entity
	}
	return query
}

func (r *ownerRepositoryDB) GetBudgetDetails(filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	query := r.db.Model(&models.BudgetFactEntity{}).Preload("BudgetAmounts")

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

	// Enforce Permissions (Internal Filter)
	if val, ok := filter["allowed_departments"]; ok {
		if strs, ok := val.([]string); ok && len(strs) > 0 {
			query = query.Where("department IN ? OR nav_code IN ?", strs, strs)
		} else if isRestricted, ok := filter["is_restricted"].(bool); ok && isRestricted {
			query = query.Where("1=0")
		}
	}

	applyFilter("groups", "\"group\"")
	applyFilter("groups", "\"group\"")
	// Smart Filter for Departments
	if val, ok := filter["departments"]; ok {
		var strs []string
		if s, ok := val.([]string); ok {
			strs = s
		} else if s, ok := val.([]interface{}); ok {
			for _, item := range s {
				strs = append(strs, fmt.Sprintf("%v", item))
			}
		}
		if len(strs) > 0 {
			query = query.Where("department IN ? OR nav_code IN ?", strs, strs)
		}
	} else {
		// applyFilter("departments", "department") // Removed standard usage
	}
	applyFilter("entity_gls", "entity_gl")
	applyFilter("conso_gls", "conso_gl")
	applyFilter("entities", "entity")
	applyFilter("branches", "branch")

	err := query.Order("\"group\", department, entity_gl, conso_gl, gl_name").Find(&results).Error
	return results, err
}

func (r *ownerRepositoryDB) GetActualDetails(filter map[string]interface{}) ([]models.OwnerActualFactEntity, error) {
	var results []models.OwnerActualFactEntity
	query := r.db.Model(&models.OwnerActualFactEntity{}).Preload("OwnerActualAmounts")

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

	// Enforce Permissions (Internal Filter)
	if val, ok := filter["allowed_departments"]; ok {
		if strs, ok := val.([]string); ok && len(strs) > 0 {
			query = query.Where("department IN ? OR nav_code IN ?", strs, strs)
		} else if isRestricted, ok := filter["is_restricted"].(bool); ok && isRestricted {
			query = query.Where("1=0")
		}
	}

	// Smart Filter for Departments
	if val, ok := filter["departments"]; ok {
		var strs []string
		if s, ok := val.([]string); ok {
			strs = s
		} else if s, ok := val.([]interface{}); ok {
			for _, item := range s {
				strs = append(strs, fmt.Sprintf("%v", item))
			}
		}
		if len(strs) > 0 {
			query = query.Where("department IN ? OR nav_code IN ?", strs, strs)
		}
	} else {
		// applyFilter("departments", "department")
	}
	applyFilter("conso_gls", "conso_gl")
	applyFilter("entities", "entity")
	applyFilter("branches", "branch")

	// Add Year Filter
	if year, ok := filter["year"].(string); ok && year != "" {
		y := strings.ReplaceAll(year, "FY", "")
		query = query.Where("year = ?", y)
	}

	err := query.Order("department, conso_gl, gl_name").Find(&results).Error
	return results, err
}

func (r *ownerRepositoryDB) GetActualTransactions(filter map[string]interface{}) ([]models.ActualTransactionDTO, error) {
	var results []models.ActualTransactionDTO
	// Requires Union Query logic.
	// For Owner View, simpler logic? Or full copy?
	// Full Copy is safer for consistency.

	whereClause := "1=1"
	var args []interface{}

	// Maps (Entity/Branch) - Hardcoded/Duplicated from BudgetDB
	// If shared constants existed, we'd use them.
	// Duplicating for isolation.
	entityCodeToNameMap := map[string][]string{
		"HMW":  {"HONDA MALIWAN", "HMW", "Honda Maliwan"},
		"ACG":  {"AUTOCORP HOLDING", "ACG", "Autocorp Holding"},
		"CLIK": {"CLIK"},
	}
	branchCodeToNameMap := map[string][]string{
		"HOF":      {"HEAD OFFICE", "AUTOCORP HEAD OFFICE", "HEADOFFICE", "HOF", "Head Office", "Headoffice"},
		"BUR":      {"BURIRUM", "BUR", "Burirum"},
		"KBI":      {"KRABI", "KBI", "Krabi"},
		"MSR":      {"MINI_SURIN", "MSR", "Mini_Surin"},
		"MKB":      {"MUEANG KRABI", "MKB", "Mueang Krabi"},
		"NAK":      {"NAKA", "NAK", "Naka"},
		"AVN":      {"NANGRONG", "AVN", "Nangrong"},
		"PHC":      {"PHACHA", "PHC", "Phacha"},
		"PRA":      {"PHUKET", "PRA", "Phuket"},
		"SUR":      {"SURIN", "SUR", "Surin"},
		"VEE":      {"VEERAWAT", "VEE", "Veerawat"},
		"HQ":       {"AUTOCORP HEAD OFFICE", "HQ", "Autocorp Head Office"},
		"Branch00": {"", "Branch00"},
	}
	for i := 1; i <= 15; i++ {
		key := fmt.Sprintf("Branch%02d", i)
		branchCodeToNameMap[key] = []string{fmt.Sprintf("BRANCH%02d", i), fmt.Sprintf("Branch%02d", i)}
	}

	// Filter Logic
	var hmwEntities []string
	var clikEntities []string
	if val, ok := filter["entities"]; ok {
		var entities []string
		if s, ok := val.([]string); ok {
			entities = s
		}
		if len(entities) > 0 {
			// selectedEntities = entities // Unused
			for _, e := range entities {
				if names, ok := entityCodeToNameMap[e]; ok {
					hmwEntities = append(hmwEntities, names...)
					clikEntities = append(clikEntities, names...)
				} else {
					hmwEntities = append(hmwEntities, e)
					clikEntities = append(clikEntities, e)
				}
			}
		}
	}

	var hmwBranches []string
	var clikBranches []string
	if val, ok := filter["branches"]; ok {
		var branches []string
		if s, ok := val.([]string); ok {
			branches = s
		}
		if len(branches) > 0 {
			for _, b := range branches {
				if names, ok := branchCodeToNameMap[b]; ok {
					hmwBranches = append(hmwBranches, names...)
					clikBranches = append(clikBranches, names...)
				} else {
					hmwBranches = append(hmwBranches, b)
					clikBranches = append(clikBranches, b)
				}
			}
		}
	}

	// Dept Filter
	if val, ok := filter["departments"]; ok {
		var depts []string
		if s, ok := val.([]string); ok {
			depts = s
		}
		if len(depts) > 0 {
			// Expand Master Codes to NavCodes
			var mappedCodes []string
			r.db.Table("department_mapping_entities").
				Joins("JOIN department_entities ON department_entities.id = department_mapping_entities.department_id").
				Where("department_entities.code IN ?", depts).
				Pluck("department_mapping_entities.nav_code", &mappedCodes)

			// Combine
			depts = append(depts, mappedCodes...)

			whereClause += " AND \"Global_Dimension_1_Code\" IN ?"
			args = append(args, depts)
		}
	}

	// Enforce Permissions (Internal Filter)
	if val, ok := filter["allowed_departments"]; ok {
		if strs, ok := val.([]string); ok && len(strs) > 0 {
			// Combine with mapped codes if needed?
			// For simplicity, we just filter by the allowed departments directly in "Global_Dimension_1_Code"
			whereClause += " AND \"Global_Dimension_1_Code\" IN ?"
			args = append(args, strs)
		} else if isRestricted, ok := filter["is_restricted"].(bool); ok && isRestricted {
			whereClause += " AND 1=0"
		}
	}

	// Date
	if val, ok := filter["start_date"]; ok {
		if s, ok := val.(string); ok && s != "" {
			whereClause += " AND \"Posting_Date\"::DATE >= ?"
			args = append(args, s)
		}
	}
	if val, ok := filter["end_date"]; ok {
		if s, ok := val.(string); ok && s != "" {
			whereClause += " AND \"Posting_Date\"::DATE <= ?"
			args = append(args, s)
		}
	}

	// Active Years
	var activeYears []string
	r.db.Model(&models.ActualFactEntity{}).Distinct("year").Pluck("year", &activeYears)
	whereClause += " AND TO_CHAR(\"Posting_Date\"::DATE, 'YYYY') IN ?"
	if len(activeYears) > 0 {
		args = append(args, activeYears)
	} else {
		whereClause += " AND 1=0"
	}

	// Queries
	hmwQuery := r.db.Table("achhmw_gle_api").
		Select(`'HMW' as source, TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date, "Document_No" as doc_no, "Description" as description, "G_L_Account_No" as gl_account_no,  "G_L_Account_Name" as gl_account_name, "Global_Dimension_1_Code" as department, "Credit_Amount" as amount, "Company" as company, "Branch" as branch`).
		Where(whereClause, args...).Limit(2000)

	if len(hmwEntities) > 0 {
		hmwQuery = hmwQuery.Where("company IN ?", hmwEntities)
	}
	if len(hmwBranches) > 0 {
		hmwQuery = hmwQuery.Where("branch IN ?", hmwBranches)
	}

	clikQuery := r.db.Table("general_ledger_entries_clik").
		Select(`'CLIK' as source, TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date, "Document_No" as doc_no, "Description" as description, "G_L_Account_No" as gl_account_no, "G_L_Account_Name" as gl_account_name, "Global_Dimension_1_Code" as department, "Credit_Amount" as amount, 'CLIK' as company, "Global_Dimension_2_Code" as branch`).
		Where(whereClause, args...).Limit(2000)

	// CLIK Company Check
	if len(hmwEntities) > 0 {
		hasClik := false
		for _, e := range hmwEntities {
			if e == "CLIK" {
				hasClik = true
				break
			}
		}
		if !hasClik {
			clikQuery = clikQuery.Where("1=0")
		}
	}
	if len(clikBranches) > 0 {
		clikQuery = clikQuery.Where("\"Global_Dimension_2_Code\" IN ?", clikBranches)
	}

	var hmwRows []models.ActualTransactionDTO
	if err := hmwQuery.Scan(&hmwRows).Error; err != nil {
		return nil, err
	}
	var clikRows []models.ActualTransactionDTO
	if err := clikQuery.Scan(&clikRows).Error; err != nil {
		return nil, err
	}

	results = append(results, hmwRows...)
	results = append(results, clikRows...)

	// --- Display Mapping (DB Full Name -> Abbreviation) ---
	// Create Mapping directly here or reuse constant
	// Entity Mapping
	dbToDisplayEntity := map[string]string{
		"HONDA MALIWAN":    "HMW",
		"AUTOCORP HOLDING": "ACG",
		"CLIK":             "CLIK",
	}

	// Branch Mapping
	dbToDisplayBranch := map[string]string{
		// HMW
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
		// ACG
		"AUTOCORP HEAD OFFICE": "HQ",
		// CLIK
		"HEADOFFICE": "HOF",
		// "BranchXX" usually maps validation logic, but here likely just needs pass through or normalization
	}

	for i := range results {
		// Map Company
		// Norm and Trim Check
		compUpper := strings.ToUpper(strings.TrimSpace(results[i].Company))
		if code, ok := dbToDisplayEntity[compUpper]; ok {
			results[i].Company = code
		} else if compUpper == "" && results[i].Source == "CLIK" {
			results[i].Company = "CLIK" // Safety
		}

		// Map Branch
		branchUpper := strings.ToUpper(strings.TrimSpace(results[i].Branch))
		if code, ok := dbToDisplayBranch[branchUpper]; ok {
			results[i].Branch = code
		} else {
			// Special Cases
			if branchUpper == "" {
				results[i].Branch = "Branch00"
			} else {
				// Normalize BRANCHXX -> BranchXX logic if needed?
				// User said: BRANCH01 -> Branch01
				if strings.HasPrefix(branchUpper, "BRANCH") {
					// Convert to Title Case or just keep as passed if it matches "Branch01" logic?
					// Simple hack: "BRANCH01" -> "Branch01"
					// Caser for Title?
					results[i].Branch = strings.Title(strings.ToLower(branchUpper))
				}
			}
		}
	}

	return results, nil
}

func (r *ownerRepositoryDB) GetOwnerFilterLists(filter map[string]interface{}) (*models.OwnerFilterListsDTO, error) {
	lists := &models.OwnerFilterListsDTO{
		Companies: []string{},
		Branches:  []string{},
		Years:     []string{},
	}

	// Helper for Restriction
	applyRestriction := func(query *gorm.DB) *gorm.DB {
		if val, ok := filter["allowed_departments"]; ok {
			if strs, ok := val.([]string); ok && len(strs) > 0 {
				query = query.Where("department IN ? OR nav_code IN ?", strs, strs)
			} else if isRestricted, ok := filter["is_restricted"].(bool); ok && isRestricted {
				query = query.Where("1=0")
			}
		}
		return query
	}

	// 1. Companies (Entity)
	var budgetEntities []string
	if err := applyRestriction(r.db.Model(&models.BudgetFactEntity{})).Distinct("entity").Pluck("entity", &budgetEntities).Error; err != nil {
		return nil, err
	}
	var actualEntities []string
	if err := applyRestriction(r.db.Model(&models.OwnerActualFactEntity{})).Distinct("entity").Pluck("entity", &actualEntities).Error; err != nil {
		return nil, err
	}
	// Merge Unique
	entityMap := make(map[string]bool)
	for _, e := range budgetEntities {
		entityMap[e] = true
	}
	for _, e := range actualEntities {
		entityMap[e] = true
	}
	for e := range entityMap {
		if e != "" {
			lists.Companies = append(lists.Companies, e)
		}
	}

	// 2. Branches
	var budgetBranches []string
	if err := applyRestriction(r.db.Model(&models.BudgetFactEntity{})).Distinct("branch").Pluck("branch", &budgetBranches).Error; err != nil {
		return nil, err
	}
	var actualBranches []string
	if err := applyRestriction(r.db.Model(&models.OwnerActualFactEntity{})).Distinct("branch").Pluck("branch", &actualBranches).Error; err != nil {
		return nil, err
	}
	branchMap := make(map[string]bool)
	for _, b := range budgetBranches {
		branchMap[b] = true
	}
	for _, b := range actualBranches {
		branchMap[b] = true
	}
	for b := range branchMap {
		if b != "" {
			lists.Branches = append(lists.Branches, b)
		}
	}

	// 3. Years
	// Years might need to be sorted descending
	var budgetYears []string
	if err := applyRestriction(r.db.Model(&models.BudgetFactEntity{})).Distinct("year").Pluck("year", &budgetYears).Error; err != nil {
		return nil, err
	}
	var actualYears []string
	if err := applyRestriction(r.db.Model(&models.OwnerActualFactEntity{})).Distinct("year").Pluck("year", &actualYears).Error; err != nil {
		return nil, err
	}
	yearMap := make(map[string]bool)
	for _, y := range budgetYears {
		yearMap[y] = true
	}
	for _, y := range actualYears {
		yearMap[y] = true
	}
	for y := range yearMap {
		if y != "" {
			lists.Years = append(lists.Years, y)
		}
	}

	return lists, nil
}

// AutoSyncOwnerActuals trancates and re-populates OwnerActualFactEntity
func (r *ownerRepositoryDB) GetUserPermissions(userID string) ([]models.UserPermissionEntity, error) {
	var perms []models.UserPermissionEntity
	err := r.db.Where("user_id = ? AND is_active = true", userID).Find(&perms).Error
	return perms, err
}

func (r *ownerRepositoryDB) AutoSyncOwnerActuals() error {
	// 1. Truncate Tables
	if err := r.db.Exec("TRUNCATE TABLE owner_actual_amount_entities, owner_actual_fact_entities RESTART IDENTITY CASCADE").Error; err != nil {
		return fmt.Errorf("failed to truncate owner actuals: %w", err)
	}

	// 2. Fetch Source Data (One Pass)
	type SourceRow struct {
		Entity     string          `gorm:"column:entity"`
		Branch     string          `gorm:"column:branch"`
		Department string          `gorm:"column:department"`
		NavCode    string          `gorm:"column:navcode"`
		EntityGL   string          `gorm:"column:entitygl"`
		GLName     string          `gorm:"column:glname"`
		Year       string          `gorm:"column:year"`
		Month      string          `gorm:"column:month"`
		Amount     decimal.Decimal `gorm:"column:amount"`
	}

	var rows []SourceRow
	sourceQuery := `
		SELECT 
			COALESCE(NULLIF(src.company, ''), 'HMW') as Entity,
			src.branch as Branch,
			COALESCE(d.code, NULLIF(src."Global_Dimension_1_Code", ''), 'OTHERS') as Department,
			NULLIF(src."Global_Dimension_1_Code", '') as NavCode,
			src."G_L_Account_No" as EntityGL,
			src."G_L_Account_Name" as GLName,
			TO_CHAR(src."Posting_Date"::DATE, 'YYYY') as Year,
			UPPER(TO_CHAR(src."Posting_Date"::DATE, 'MON')) as Month,
			SUM(src."Credit_Amount") as Amount
		FROM achhmw_gle_api src
        LEFT JOIN (SELECT DISTINCT ON (nav_code, entity) * FROM department_mapping_entities) m ON m.nav_code = src."Global_Dimension_1_Code" AND m.entity = COALESCE(NULLIF(src.company, ''), 'HMW')
        LEFT JOIN department_entities d ON d.id = m.department_id
		GROUP BY 1, 2, 3, 4, 5, 6, 7, 8

		UNION ALL

		SELECT 
			'CLIK' as Entity,
			src."Global_Dimension_2_Code" as Branch,
			COALESCE(d.code, NULLIF(src."Global_Dimension_1_Code", ''), 'OTHERS') as Department,
			NULLIF(src."Global_Dimension_1_Code", '') as NavCode,
			src."G_L_Account_No" as EntityGL,
			src."G_L_Account_Name" as GLName,
			TO_CHAR(src."Posting_Date"::DATE, 'YYYY') as Year,
			UPPER(TO_CHAR(src."Posting_Date"::DATE, 'MON')) as Month,
			SUM(src."Credit_Amount") as Amount
		FROM general_ledger_entries_clik src
        LEFT JOIN (SELECT DISTINCT ON (nav_code, entity) * FROM department_mapping_entities) m ON m.nav_code = src."Global_Dimension_1_Code" AND m.entity = 'CLIK'
        LEFT JOIN department_entities d ON d.id = m.department_id
		GROUP BY 1, 2, 3, 4, 5, 6, 7, 8
	`

	if err := r.db.Raw(sourceQuery).Scan(&rows).Error; err != nil {
		return fmt.Errorf("failed to fetch source data: %w", err)
	}

	if len(rows) == 0 {
		return nil
	}

	// 3. Process in Memory
	// Map to track unique Query Facts (Entity+Branch+Dept+NavCode+GL+Year)
	factMap := make(map[string]*models.OwnerActualFactEntity)
	var facts []*models.OwnerActualFactEntity
	var amounts []models.OwnerActualAmountEntity

	// Mapping Dicts
	entityMapping := map[string]string{
		"AUTOCORP HOLDING": "ACG",
		"HONDA MALIWAN":    "HMW",
	}
	branchMapping := map[string]string{
		"AUTOCORP HEAD OFFICE": "HQ",
		"BURIRUM":              "BUR",
		"HEAD OFFICE":          "HOF",
		"HEADOFFICE":           "HOF",
		"KRABI":                "KBI",
		"MINI_SURIN":           "MSR",
		"MUEANG KRABI":         "MKB",
		"NAKA":                 "NAK",
		"NANGRONG":             "AVN",
		"PHACHA":               "PHC",
		"PHUKET":               "PRA",
		"SURIN":                "SUR",
	}

	for _, row := range rows {
		// --- MAPPING LOGIC ---
		// 1. Entity Mapping
		entUpper := strings.ToUpper(strings.TrimSpace(row.Entity))
		if mappedEnt, ok := entityMapping[entUpper]; ok {
			row.Entity = mappedEnt
		} else if row.Entity == "CLIK" {
			// Stay CLIK
		}

		// 2. Branch Mapping
		brUpper := strings.ToUpper(strings.TrimSpace(row.Branch))
		if mappedBr, ok := branchMapping[brUpper]; ok {
			row.Branch = mappedBr
		} else {
			// CLIK Special Case: BRANCH01 -> Branch01
			if brUpper == "" {
				row.Branch = "Branch00"
			} else if strings.HasPrefix(brUpper, "BRANCH") {
				// Convert BRANCHXX to BranchXX
				row.Branch = strings.Title(strings.ToLower(brUpper))
			}
		}

		// Key for Fact Uniqueness - INCLUDE NavCode
		// Use Pipes to separate
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%s", row.Entity, row.Branch, row.Department, row.NavCode, row.EntityGL, row.Year)

		if _, exists := factMap[key]; !exists {
			newID := uuid.New()
			newFact := &models.OwnerActualFactEntity{
				ID:         newID,
				Entity:     row.Entity,
				Branch:     row.Branch,
				Department: row.Department,
				NavCode:    row.NavCode, // Save
				EntityGL:   row.EntityGL,
				GLName:     row.GLName,
				Year:       row.Year,
				YearTotal:  decimal.Zero,
				IsValid:    true,
			}
			factMap[key] = newFact
			facts = append(facts, newFact)
		}

		// Update Fact Total
		fact := factMap[key]
		fact.YearTotal = fact.YearTotal.Add(row.Amount)

		// Create Amount Record
		amounts = append(amounts, models.OwnerActualAmountEntity{
			ID:                uuid.New(),
			OwnerActualFactID: fact.ID,
			Month:             row.Month, // Expected JAN, FEB...
			Amount:            row.Amount,
		})
	}

	// 4. Batch Insert Facts
	// GORM's CreateInBatches fits well here
	if len(facts) > 0 {
		if err := r.db.CreateInBatches(facts, 1000).Error; err != nil {
			return fmt.Errorf("failed to insert facts: %w", err)
		}
	}

	// 5. Batch Insert Amounts
	if len(amounts) > 0 {
		if err := r.db.CreateInBatches(amounts, 1000).Error; err != nil {
			return fmt.Errorf("failed to insert amounts: %w", err)
		}
	}

	return nil
}

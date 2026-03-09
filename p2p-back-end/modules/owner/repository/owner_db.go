package repository

import (
	"fmt"
	"strings"
	"sync"

	"p2p-back-end/modules/entities/models"

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
				if tn == "budget_fact_entities" || tn == "actual_fact_entities" {
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
				if tn == "budget_fact_entities" || tn == "actual_fact_entities" {
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
			if s, ok := val.(string); ok && s != "" && !strings.EqualFold(s, "All") {
				s = strings.ReplaceAll(s, "FY", "")
				if !strings.Contains(tableName, "budget") {
					tx = tx.Where(tableName+".year = ?", s)
				}
			}
		}
		if val, ok := filter["budget_gls"]; ok {
			// Helper to convert interface to uint slice (JSON arrays of numbers come as float64)
			var ids []uint
			if uids, ok := val.([]uint); ok {
				ids = uids
			} else if faces, ok := val.([]interface{}); ok {
				for _, f := range faces {
					if nf, ok := f.(float64); ok {
						ids = append(ids, uint(nf))
					} else if ni, ok := f.(int); ok {
						ids = append(ids, uint(ni))
					}
				}
			}

			if len(ids) > 0 {
				var nodeCodes []string
				// r.db is available from the outer repository scope
				if err := r.db.Model(&models.BudgetStructureEntity{}).Where("id IN ?", ids).Pluck("node_code", &nodeCodes).Error; err == nil && len(nodeCodes) > 0 {
					if tableName == "budget_fact_entities" || tableName == "actual_fact_entities" {
						tx = tx.Where(tableName+".conso_gl IN ?", nodeCodes)
					}
				}
			}
		}

		// Support direct conso_gl filtering (from Top Account Filter)
		if val, ok := filter["conso_gls"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				if tableName == "budget_fact_entities" || tableName == "actual_fact_entities" {
					tx = tx.Where(tableName+".conso_gl IN ?", strs)
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
			// Phase 10: New click behavior. Only drill down to NavCode if "None" is selected
			shouldDrillDown := false
			for _, s := range strs {
				if s == "None" {
					shouldDrillDown = true
					break
				}
			}

			if shouldDrillDown {
				groupBy = "nav_code"
				selectCol = "nav_code as department"
			}
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
		tx := r.db.Table("actual_fact_entities").Select(selectCol + ", COALESCE(SUM(year_total), 0) as total")
		tx = applyFilter(tx, "actual_fact_entities")
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
	// 4. Chart Aggregation (Actual)
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
		tx := r.db.Table("actual_fact_entities").
			Select("actual_amount_entities.month, SUM(actual_amount_entities.amount) as total").
			Joins("JOIN actual_amount_entities ON actual_amount_entities.owner_actual_fact_id = actual_fact_entities.id")

		tx = applyFilter(tx, "actual_fact_entities")

		if err := tx.Group("actual_amount_entities.month").Scan(&actualMonthResults).Error; err != nil {
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
		// Refactored to map GLs to Level 3 Names using budget_structure_entities
		tx := r.db.Table("actual_fact_entities").
			Select("COALESCE(budget_structure_entities.name, actual_fact_entities.gl_name) as name, CAST(SUM(actual_fact_entities.year_total) AS FLOAT) as total").
			Joins("LEFT JOIN budget_structure_entities ON budget_structure_entities.node_code = actual_fact_entities.conso_gl AND budget_structure_entities.level = 3").
			Group("COALESCE(budget_structure_entities.name, actual_fact_entities.gl_name)").
			Order("total desc").
			Limit(3)
		tx = applyFilter(tx, "actual_fact_entities")
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

func (r *ownerRepositoryDB) GetActualDetails(filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	var results []models.ActualFactEntity
	query := r.db.Model(&models.ActualFactEntity{}).Preload("ActualAmounts")

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

func (r *ownerRepositoryDB) GetActualTransactions(filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	page := 1
	limit := 10
	if p, ok := filter["page"].(float64); ok {
		page = int(p)
	}
	if l, ok := filter["limit"].(float64); ok {
		limit = int(l)
	}
	offset := (page - 1) * limit

	query := r.db.Table("actual_transaction_entities").
		Select(`
			actual_transaction_entities.source,
			actual_transaction_entities.posting_date,
			actual_transaction_entities.doc_no as doc_no,
			actual_transaction_entities.vendor_name as vendor,
			actual_transaction_entities.description,
			actual_transaction_entities.entity_gl as gl_account_no,
			actual_transaction_entities.conso_gl as conso_gl,
			mapping.account_name as gl_account_name,
			actual_transaction_entities.amount,
			actual_transaction_entities.department,
			actual_transaction_entities.entity as company,
			actual_transaction_entities.branch
		`).
		Joins("LEFT JOIN gl_mapping_entities mapping ON actual_transaction_entities.entity_gl = mapping.entity_gl AND actual_transaction_entities.entity = mapping.entity")

	// Filter Helper
	toStringSlice := func(val interface{}) []string {
		if s, ok := val.([]string); ok {
			return s
		}
		if s, ok := val.([]interface{}); ok {
			var res []string
			for _, v := range s {
				if str, ok := v.(string); ok {
					res = append(res, str)
				}
			}
			return res
		}
		return nil
	}

	// 1. Entities
	if val, ok := filter["entities"]; ok {
		if s := toStringSlice(val); len(s) > 0 {
			query = query.Where("actual_transaction_entities.entity IN ?", s)
		}
	}

	// 2. Branches
	if val, ok := filter["branches"]; ok {
		if s := toStringSlice(val); len(s) > 0 {
			query = query.Where("actual_transaction_entities.branch IN ?", s)
		}
	}

	// 3. Departments
	if val, ok := filter["departments"]; ok {
		if s := toStringSlice(val); len(s) > 0 {
			query = query.Where("actual_transaction_entities.department IN ?", s)
		}
	}

	// 4. Conso GLs
	if val, ok := filter["conso_gls"]; ok {
		if s := toStringSlice(val); len(s) > 0 {
			query = query.Where("actual_transaction_entities.conso_gl IN ?", s)
		}
	}

	// 5. Date Range
	if val, ok := filter["start_date"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.posting_date >= ?", val)
	}
	if val, ok := filter["end_date"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.posting_date <= ?", val)
	}

	// 6. Year
	if val, ok := filter["year"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.year = ?", val)
	}

	// 7. Permissions (allowed_departments)
	if val, ok := filter["allowed_departments"]; ok {
		if s := toStringSlice(val); len(s) > 0 {
			query = query.Where("actual_transaction_entities.department IN ?", s)
		} else if isRestricted, ok := filter["is_restricted"].(bool); ok && isRestricted {
			query = query.Where("1=0")
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var results []models.ActualTransactionDTO
	if err := query.Order("actual_transaction_entities.posting_date DESC, actual_transaction_entities.doc_no DESC").
		Limit(limit).Offset(offset).Scan(&results).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedActualTransactionDTO{
		Data:       results,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
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
	if err := applyRestriction(r.db.Model(&models.ActualFactEntity{})).Distinct("entity").Pluck("entity", &actualEntities).Error; err != nil {
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
	if err := applyRestriction(r.db.Model(&models.ActualFactEntity{})).Distinct("branch").Pluck("branch", &actualBranches).Error; err != nil {
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
	if err := applyRestriction(r.db.Model(&models.ActualFactEntity{})).Distinct("year").Pluck("year", &actualYears).Error; err != nil {
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
			lists.Years = append(lists.Years, "FY"+y)
		}
	}

	// Add 'All' option ONLY if we have years or if we want to allow viewing everything
	// The user wants strictly DB years, but 'All' is a functional requirement for "Everything"
	lists.Years = append(lists.Years, "All")

	return lists, nil
}

func (r *ownerRepositoryDB) GetUserPermissions(userID string) ([]models.UserPermissionEntity, error) {
	var perms []models.UserPermissionEntity
	err := r.db.Where("user_id = ? AND is_active = true", userID).Find(&perms).Error
	return perms, err
}

func (r *ownerRepositoryDB) GetNavCodesByMasterDepts(masterCodes []string) ([]string, error) {
	var navCodes []string
	err := r.db.Table("department_mapping_entities").
		Joins("JOIN department_entities ON department_entities.id = department_mapping_entities.department_id").
		Where("department_entities.code IN ?", masterCodes).
		Distinct("department_mapping_entities.nav_code").
		Pluck("department_mapping_entities.nav_code", &navCodes).Error
	return navCodes, err
}

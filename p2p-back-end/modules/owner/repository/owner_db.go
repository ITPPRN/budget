package repository

import (
	"context"
	"fmt"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ownerRepository struct {
	db *gorm.DB
}

func NewOwnerRepository(db *gorm.DB) models.OwnerRepository {
	return &ownerRepository{db: db}
}

// toStringSlice safely converts interface{} (from JSON filter map) to []string
func toStringSlice(val interface{}) []string {
	if strs, ok := val.([]string); ok {
		return strs
	}
	if interfaces, ok := val.([]interface{}); ok {
		var strs []string
		for _, item := range interfaces {
			if s, ok := item.(string); ok {
				strs = append(strs, s)
			} else if item != nil {
				strs = append(strs, fmt.Sprintf("%v", item))
			}
		}
		return strs
	}
	if s, ok := val.(string); ok && s != "" {
		return []string{s}
	}
	return nil
}

func (r *ownerRepository) GetBudgetFilterOptions(ctx context.Context) ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	if err := r.db.WithContext(ctx).Model(&models.BudgetFactEntity{}).
		Distinct("\"group\"", "department", "entity_gl", "conso_gl", "gl_name").
		Order("\"group\", department, entity_gl, conso_gl").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetBudgetFilterOptions: %w", err)
	}
	return results, nil
}

func (r *ownerRepository) GetOrganizationStructure(ctx context.Context) ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	query := `
        SELECT DISTINCT entity, branch, department FROM budget_fact_entities WHERE entity != ''
        UNION
        SELECT DISTINCT entity, branch, department FROM actual_fact_entities WHERE entity != ''
        ORDER BY entity, branch, department
    `
	if err := r.db.WithContext(ctx).Raw(query).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetOrganizationStructure: %w", err)
	}
	return results, nil
}

func (r *ownerRepository) GetDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	summary := &models.DashboardSummaryDTO{
		DepartmentData: []models.DepartmentStatDTO{},
		ChartData:      []models.MonthlyStatDTO{},
	}

	logs.Infof("[DEBUG] OwnerRepository: Starting GetDashboardAggregates (Clean Implementation) with Filter: %+v", filter)

	// Dynamic Filter Logic - Standardized
	applyCommonFilters := func(tx *gorm.DB, tableName string) *gorm.DB {
		// Internal helper for strings
		toStrings := func(val interface{}) []string {
			if s, ok := val.([]string); ok {
				return s
			}
			if i, ok := val.([]interface{}); ok {
				var res []string
				for _, it := range i {
					if s, ok := it.(string); ok {
						res = append(res, s)
					}
				}
				return res
			}
			return nil
		}
		if val, ok := filter["entities"]; ok {
			if strs := toStrings(val); len(strs) > 0 {
				tx = tx.Where(tableName+".entity IN ?", strs)
			}
		}
		if val, ok := filter["branches"]; ok {
			if strs := toStrings(val); len(strs) > 0 {
				tx = tx.Where(tableName+".branch IN ?", strs)
			}
		}
		if val, ok := filter["departments"]; ok {
			if strs := toStrings(val); len(strs) > 0 {
				hasNone := false
				var filteredStrs []string
				for _, s := range strs {
					if strings.EqualFold(strings.TrimSpace(s), "None") {
						hasNone = true
					} else if s != "" {
						filteredStrs = append(filteredStrs, s)
					}
				}

				// Use COALESCE logic to match Dashboard grouping
				condition := "(COALESCE(NULLIF(" + tableName + ".department, ''), " + tableName + ".nav_code) IN ?)"
				if hasNone {
					if len(filteredStrs) > 0 {
						tx = tx.Where("("+condition+" OR COALESCE(NULLIF("+tableName+".department, ''), "+tableName+".nav_code) = '' OR COALESCE(NULLIF("+tableName+".department, ''), "+tableName+".nav_code) IS NULL OR COALESCE(NULLIF("+tableName+".department, ''), "+tableName+".nav_code) = 'None')", filteredStrs)
					} else {
						tx = tx.Where("(COALESCE(NULLIF(" + tableName + ".department, ''), " + tableName + ".nav_code) = '' OR COALESCE(NULLIF(" + tableName + ".department, ''), " + tableName + ".nav_code) IS NULL OR COALESCE(NULLIF(" + tableName + ".department, ''), " + tableName + ".nav_code) = 'None')")
					}
				} else {
					tx = tx.Where(condition, strs)
				}
			}
		}
		glVal, hasGL := filter["conso_gls"]
		if !hasGL {
			glVal, hasGL = filter["budget_gls"]
		}
		if hasGL {
			if strs := toStrings(glVal); len(strs) > 0 {
				tx = tx.Where(tableName+".conso_gl IN ?", strs)
			}
		}
		return tx
	}

	// Determine Grouping (Same as Admin: Drill-down to nav_code if "None" is uniquely selected)
	groupBy := "COALESCE(NULLIF(department, ''), nav_code)"
	selectCol := "COALESCE(NULLIF(department, ''), nav_code) as department"

	if depts, ok := filter["departments"]; ok {
		strs := toStringSlice(depts)
		if len(strs) == 1 && strs[0] == "None" {
			groupBy = "nav_code"
			selectCol = "nav_code as department"
		}
	}

	// A. Budget Baseline (Static - uses budget_file_id)
	type DeptResult struct {
		Department string
		Total      decimal.Decimal
	}
	var budgetDept []DeptResult

	// Year for fallback
	targetYear := ""
	if val, ok := filter["year"]; ok {
		if s, ok := val.(string); ok && s != "" {
			targetYear = strings.ReplaceAll(s, "FY", "")
		}
	}

	txB := r.db.WithContext(ctx).Model(&models.BudgetFactEntity{}).
		Select(selectCol + ", SUM(ba.amount) as total").
		Joins("JOIN budget_amount_entities ba ON ba.budget_fact_id = budget_fact_entities.id AND ba.deleted_at IS NULL")

	txB = applyCommonFilters(txB, "budget_fact_entities")

	// IMPORTANT: Use budget_file_id for static baseline
	if fid, ok := filter["budget_file_id"].(string); ok && fid != "" {
		txB = txB.Where("file_budget_id = ?", fid)
	} else {
		// 🛠️ Fallback: If no ID provided, use the most recent budget file matching the targetYear
		var latestFid string
		subQuery := r.db.WithContext(ctx).Model(&models.BudgetFactEntity{})
		if targetYear != "" {
			subQuery = subQuery.Where("year = ?", targetYear)
			txB = txB.Where("budget_fact_entities.year = ?", targetYear)
		}
		if err := subQuery.Order("created_at desc").Limit(1).Pluck("file_budget_id", &latestFid).Error; err == nil && latestFid != "" {
			logs.Infof("[DEBUG] OwnerRepository: Fallback BudgetFileID Found: %s", latestFid)
			txB = txB.Where("file_budget_id = ?", latestFid)
		} else {
			logs.Warn("[WARN] OwnerRepository: No budget_file_id and no fallback found.")
			txB = txB.Where("1 = 0")
		}
	}

	if err := txB.Group(groupBy).Scan(&budgetDept).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetDashboardAggregates.Budget: %w", err)
	}

	// B. Actual Spending (Dynamic - uses year and strict months)
	var actualDept []DeptResult
	txA := r.db.WithContext(ctx).Model(&models.ActualFactEntity{}).
		Select(selectCol + ", SUM(aa.amount) as total").
		Joins("JOIN actual_amount_entities aa ON aa.actual_fact_id = actual_fact_entities.id AND aa.deleted_at IS NULL")

	// Strict Month Filter Requirement
	if mVal, ok := filter["months"]; ok {
		mstrs := toStringSlice(mVal)
		if len(mstrs) > 0 {
			txA = txA.Where("aa.month IN ?", mstrs)
		} else {
			txA = txA.Where("1 = 0")
		}
	}

	txA = applyCommonFilters(txA, "actual_fact_entities")
	if targetYear != "" {
		txA = txA.Where("actual_fact_entities.year = ?", targetYear)
	}

	if err := txA.Group(groupBy).Scan(&actualDept).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetDashboardAggregates.Actual: %w", err)
	}

	// C. Merge Department Data
	deptMap := make(map[string]*models.DepartmentStatDTO)
	summary.TotalBudget = decimal.Zero
	summary.TotalActual = decimal.Zero

	for _, b := range budgetDept {
		deptMap[b.Department] = &models.DepartmentStatDTO{Department: b.Department, Budget: b.Total, Actual: decimal.Zero}
		summary.TotalBudget = summary.TotalBudget.Add(b.Total)
	}
	for _, a := range actualDept {
		if _, ok := deptMap[a.Department]; !ok {
			deptMap[a.Department] = &models.DepartmentStatDTO{Department: a.Department, Budget: decimal.Zero, Actual: a.Total}
		} else {
			deptMap[a.Department].Actual = a.Total
		}
		summary.TotalActual = summary.TotalActual.Add(a.Total)
	}

	for _, v := range deptMap {
		summary.DepartmentData = append(summary.DepartmentData, *v)
	}

	// --- 3. Monthly Aggregation (Chart Data) ---

	type MonthResult struct {
		Month string
		Total decimal.Decimal
	}

	// Budget Monthly
	var budgetM []MonthResult
	txBM := r.db.WithContext(ctx).Table("budget_amount_entities").
		Select("budget_amount_entities.month, SUM(budget_amount_entities.amount) as total").
		Joins("JOIN budget_fact_entities ON budget_amount_entities.budget_fact_id = budget_fact_entities.id")

	txBM = applyCommonFilters(txBM, "budget_fact_entities")
	if fid, ok := filter["budget_file_id"].(string); ok && fid != "" {
		txBM = txBM.Where("budget_fact_entities.file_budget_id = ?", fid)
	} else {
		// Fallback for Chart Budget
		var latestFid string
		subQuery := r.db.WithContext(ctx).Model(&models.BudgetFactEntity{})
		if targetYear != "" {
			subQuery = subQuery.Where("year = ?", targetYear)
			txBM = txBM.Where("budget_fact_entities.year = ?", targetYear)
		}
		if err := subQuery.Order("created_at desc").Limit(1).Pluck("file_budget_id", &latestFid).Error; err == nil && latestFid != "" {
			txBM = txBM.Where("budget_fact_entities.file_budget_id = ?", latestFid)
		} else {
			txBM = txBM.Where("1 = 0")
		}
	}

	if err := txBM.Group("budget_amount_entities.month").Scan(&budgetM).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetDashboardAggregates.BudgetMonthly: %w", err)
	}

	// Actual Monthly
	var actualM []MonthResult
	txAM := r.db.WithContext(ctx).Table("actual_amount_entities").
		Select("actual_amount_entities.month, SUM(actual_amount_entities.amount) as total").
		Joins("JOIN actual_fact_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id")

	txAM = applyCommonFilters(txAM, "actual_fact_entities")
	if targetYear != "" {
		txAM = txAM.Where("actual_fact_entities.year = ?", targetYear)
	}
	// Months filter for chart: we only show what's requested. If none, show all.
	if mVal, ok := filter["months"]; ok {
		mstrs := toStringSlice(mVal)
		if len(mstrs) > 0 {
			txAM = txAM.Where("actual_amount_entities.month IN ?", mstrs)
		}
	}

	if err := txAM.Group("actual_amount_entities.month").Scan(&actualM).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetDashboardAggregates.ActualMonthly: %w", err)
	}

	// D. Merge Chart Data (JAN -> DEC order)
	monthMap := make(map[string]*models.MonthlyStatDTO)
	for _, m := range budgetM {
		monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: m.Total, Actual: decimal.Zero}
	}
	for _, m := range actualM {
		if _, ok := monthMap[m.Month]; !ok {
			monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: decimal.Zero, Actual: m.Total}
		} else {
			monthMap[m.Month].Actual = m.Total
		}
	}

	// --- 4. Top Expenses (Group 3) ---
	var topExpenses []models.TopExpenseDTO
	txTE := r.db.WithContext(ctx).Table("actual_amount_entities").
		Select("mapping.group3 as name, SUM(actual_amount_entities.amount) as amount").
		Joins("JOIN actual_fact_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id").
		Joins("JOIN (SELECT DISTINCT conso_gl, group3 FROM gl_grouping_entities) mapping ON actual_fact_entities.conso_gl = mapping.conso_gl")

	txTE = applyCommonFilters(txTE, "actual_fact_entities")

	if targetYear != "" {
		txTE = txTE.Where("actual_fact_entities.year = ?", targetYear)
	}

	if mVal, ok := filter["months"]; ok {
		mstrs := toStringSlice(mVal)
		if len(mstrs) > 0 {
			txTE = txTE.Where("actual_amount_entities.month IN ?", mstrs)
		}
	}

	if err := txTE.Group("mapping.group3").
		Order("amount DESC").
		Limit(3).
		Scan(&topExpenses).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetDashboardAggregates.TopExpenses: %w", err)
	}
	summary.TopExpenses = topExpenses

	monthsOrder := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	for _, mon := range monthsOrder {
		if val, ok := monthMap[mon]; ok {
			summary.ChartData = append(summary.ChartData, *val)
		} else {
			summary.ChartData = append(summary.ChartData, models.MonthlyStatDTO{Month: mon, Budget: decimal.Zero, Actual: decimal.Zero})
		}
	}

	// --- 5. Status Counts (Over Budget / Near Limit) ---
	var overBudgetCount, nearLimitCount int
	redLimit := 100.0
	yellowLimit := 80.0

	// Thresholds from filter or defaults
	if val, ok := filter["red_threshold"].(float64); ok && val > 0 {
		redLimit = val
	}
	if val, ok := filter["yellow_threshold"].(float64); ok && val > 0 {
		yellowLimit = val
	}

	for _, d := range summary.DepartmentData {
		budget := d.Budget
		actual := d.Actual

		if budget.IsZero() {
			if actual.GreaterThan(decimal.Zero) {
				overBudgetCount++
			}
			continue
		}

		// Calculate spend percentage
		spendPct, _ := actual.Div(budget).Mul(decimal.NewFromInt(100)).Float64()
		if spendPct >= redLimit {
			overBudgetCount++
		} else if spendPct >= yellowLimit {
			nearLimitCount++
		}
	}
	summary.OverBudgetCount = overBudgetCount
	summary.NearLimitCount = nearLimitCount

	summary.TotalCount = int64(len(summary.DepartmentData))
	summary.Page = 1
	summary.Limit = len(summary.DepartmentData)
	if summary.Limit == 0 {
		summary.Limit = 10
	}

	logs.Infof("[DEBUG] OwnerRepository: Final Result - Budget: %v, Actual: %v", summary.TotalBudget, summary.TotalActual)
	return summary, nil
}

func (r *ownerRepository) GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	query := r.db.WithContext(ctx).Table("actual_transaction_entities").
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
		Joins("LEFT JOIN gl_grouping_entities mapping ON actual_transaction_entities.entity_gl = mapping.entity_gl AND actual_transaction_entities.entity = mapping.entity")

	// 2. Filters (Standardized Alignment)
	if val, ok := filter["entities"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("actual_transaction_entities.entity IN ?", strs)
		}
	}
	if val, ok := filter["branches"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("actual_transaction_entities.branch IN ?", strs)
		}
	}
	if val, ok := filter["departments"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			hasNone := false
			var filteredStrs []string
			for _, s := range strs {
				if strings.EqualFold(strings.TrimSpace(s), "None") {
					hasNone = true
				} else if s != "" {
					filteredStrs = append(filteredStrs, s)
				}
			}

			tableName := "actual_transaction_entities"
			condition := "(COALESCE(NULLIF(" + tableName + ".department, ''), " + tableName + ".nav_code) IN ?)"
			if hasNone {
				if len(filteredStrs) > 0 {
					query = query.Where("("+condition+" OR COALESCE(NULLIF("+tableName+".department, ''), "+tableName+".nav_code) = '' OR COALESCE(NULLIF("+tableName+".department, ''), "+tableName+".nav_code) IS NULL OR COALESCE(NULLIF("+tableName+".department, ''), "+tableName+".nav_code) = 'None')", filteredStrs)
				} else {
					query = query.Where("(COALESCE(NULLIF(" + tableName + ".department, ''), " + tableName + ".nav_code) = '' OR COALESCE(NULLIF(" + tableName + ".department, ''), " + tableName + ".nav_code) IS NULL OR COALESCE(NULLIF(" + tableName + ".department, ''), " + tableName + ".nav_code) = 'None')")
				}
			} else {
				query = query.Where(condition, strs)
			}
		}
	}
	if val, ok := filter["conso_gls"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("actual_transaction_entities.conso_gl IN ?", strs)
		}
	} else if val, ok := filter["budget_gls"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("actual_transaction_entities.conso_gl IN ?", strs)
		}
	}

	if val, ok := filter["months"]; ok {
		mstrs := toStringSlice(val)
		if len(mstrs) > 0 {
			query = query.Where("UPPER(TO_CHAR(actual_transaction_entities.posting_date::DATE, 'MON')) IN ?", mstrs)
		} else {
			query = query.Where("1 = 0")
		}
	}

	if val, ok := filter["start_date"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.posting_date >= ?", val)
	}
	if val, ok := filter["end_date"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.posting_date <= ?", val)
	}
	if val, ok := filter["year"].(string); ok && val != "" && val != "All" {
		query = query.Where("actual_transaction_entities.year = ?", strings.ReplaceAll(val, "FY", ""))
	}

	// 3. Pagination & Count
	page := 1
	limit := 10
	if p, ok := filter["page"].(float64); ok {
		page = int(p)
	}
	if l, ok := filter["limit"].(float64); ok {
		limit = int(l)
	}
	offset := (page - 1) * limit

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetActualTransactions.Count: %w", err)
	}

	var results []models.ActualTransactionDTO
	if err := query.Order("actual_transaction_entities.posting_date ASC, actual_transaction_entities.doc_no ASC").
		Limit(limit).Offset(offset).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetActualTransactions.Scan: %w", err)
	}

	return &models.PaginatedActualTransactionDTO{
		Data:       results,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (r *ownerRepository) GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	var results []models.ActualFactEntity
	if err := r.db.WithContext(ctx).Model(&models.ActualFactEntity{}).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetActualDetails: %w", err)
	}
	return results, nil
}

func (r *ownerRepository) GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	tx := r.db.WithContext(ctx).Model(&models.BudgetFactEntity{})
	if fid, ok := filter["budget_file_id"].(string); ok && fid != "" {
		tx = tx.Where("file_budget_id = ?", fid)
	}
	if err := tx.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetBudgetDetails: %w", err)
	}
	return results, nil
}

func (r *ownerRepository) GetActualYears(ctx context.Context) ([]string, error) {
	var years []string
	if err := r.db.WithContext(ctx).Model(&models.ActualFactEntity{}).
		Distinct("year").
		Order("year DESC").
		Pluck("year", &years).Error; err != nil {
		return nil, fmt.Errorf("ownerRepo.GetActualYears: %w", err)
	}
	return years, nil
}

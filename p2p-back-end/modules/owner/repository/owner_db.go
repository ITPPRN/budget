package repository

import (
	"fmt"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"strings"

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

func (r *ownerRepository) GetBudgetFilterOptions() ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	err := r.db.Model(&models.BudgetFactEntity{}).
		Distinct("\"group\"", "department", "entity_gl", "conso_gl", "gl_name").
		Order("\"group\", department, entity_gl, conso_gl").
		Find(&results).Error
	return results, err
}

func (r *ownerRepository) GetOrganizationStructure() ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	query := `
        SELECT DISTINCT entity, branch, department FROM budget_fact_entities WHERE entity != ''
        UNION
        SELECT DISTINCT entity, branch, department FROM actual_fact_entities WHERE entity != ''
        ORDER BY entity, branch, department
    `
	err := r.db.Raw(query).Scan(&results).Error
	return results, err
}

func (r *ownerRepository) GetDashboardAggregates(filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	summary := &models.DashboardSummaryDTO{
		DepartmentData: []models.DepartmentStatDTO{},
		ChartData:      []models.MonthlyStatDTO{},
	}

	logs.Infof("[DEBUG] OwnerRepository: Starting GetDashboardAggregates (Clean Implementation) with Filter: %+v", filter)

	// --- 1. Helper: Unified Filtering Scope ---
	applyCommonFilters := func(tx *gorm.DB, tableName string) *gorm.DB {
		// Entity Filter
		if val, ok := filter["entities"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".entity IN ?", strs)
			}
		}
		// Branch Filter
		if val, ok := filter["branches"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".branch IN ?", strs)
			}
		}
		// Department Filter
		if val, ok := filter["departments"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".department IN ?", strs)
			}
		}
		// Account (GL) Filter - Supports conso_gls or budget_gls (Owner vs Admin naming)
		glVal, hasGL := filter["conso_gls"]
		if !hasGL {
			glVal, hasGL = filter["budget_gls"]
		}
		if hasGL {
			if strs := toStringSlice(glVal); len(strs) > 0 {
				tx = tx.Where(tableName+".conso_gl IN ?", strs)
			}
		}
		return tx
	}

	// --- 2. Department Calculation (Stats Cards & Root List) ---

	// Determine Grouping (Same as Admin: Drill-down to nav_code if "None" is uniquely selected)
	groupBy := "department"
	selectCol := "department"
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
		Total      float64
	}
	var budgetDept []DeptResult

	txB := r.db.Model(&models.BudgetFactEntity{}).Select(selectCol + ", SUM(year_total) as total")
	txB = applyCommonFilters(txB, "budget_fact_entities")

	// IMPORTANT: Use budget_file_id for static baseline
	if fid, ok := filter["budget_file_id"].(string); ok && fid != "" {
		txB = txB.Where("file_budget_id = ?", fid)
	} else {
		// 🛠️ Fallback: If no ID provided, use the most recent synced budget file in the system
		var latestFid string
		if err := r.db.Model(&models.BudgetFactEntity{}).Order("created_at desc").Limit(1).Pluck("file_budget_id", &latestFid).Error; err == nil && latestFid != "" {
			logs.Infof("[DEBUG] OwnerRepository: Fallback BudgetFileID Found: %s", latestFid)
			txB = txB.Where("file_budget_id = ?", latestFid)
		} else {
			logs.Warn("[WARN] OwnerRepository: No budget_file_id and no fallback found.")
			txB = txB.Where("1 = 0")
		}
	}

	if err := txB.Group(groupBy).Scan(&budgetDept).Error; err != nil {
		logs.Errorf("[ERROR] OwnerRepository: Budget Aggregation Failed: %v", err)
		return nil, err
	}

	// B. Actual Spending (Dynamic - uses year and strict months)
	var actualDept []DeptResult
	txA := r.db.Model(&models.ActualFactEntity{})

	// Strict Month Filter Requirement
	if mVal, ok := filter["months"]; ok {
		mstrs := toStringSlice(mVal)
		if len(mstrs) > 0 {
			txA = txA.Select(selectCol+", SUM(actual_amount_entities.amount) as total").
				Joins("JOIN actual_amount_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id").
				Where("actual_amount_entities.month IN ?", mstrs)
		} else {
			txA = txA.Where("1 = 0")
		}
	} else {
		txA = txA.Select(selectCol+", SUM(year_total) as total")
	}

	txA = applyCommonFilters(txA, "actual_fact_entities")
	if yVal, ok := filter["year"].(string); ok && yVal != "" && yVal != "All" {
		yVal = strings.ReplaceAll(yVal, "FY", "")
		txA = txA.Where("actual_fact_entities.year = ?", yVal)
	}

	if err := txA.Group(groupBy).Scan(&actualDept).Error; err != nil {
		logs.Errorf("[ERROR] OwnerRepository: Actual Aggregation Failed: %v", err)
		return nil, err
	}

	// C. Merge Department Data
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

	// --- 3. Monthly Aggregation (Chart Data) ---

	type MonthResult struct {
		Month string
		Total float64
	}

	// Budget Monthly
	var budgetM []MonthResult
	txBM := r.db.Table("budget_amount_entities").
		Select("budget_amount_entities.month, SUM(budget_amount_entities.amount) as total").
		Joins("JOIN budget_fact_entities ON budget_amount_entities.budget_fact_id = budget_fact_entities.id")

	txBM = applyCommonFilters(txBM, "budget_fact_entities")
	if fid, ok := filter["budget_file_id"].(string); ok && fid != "" {
		txBM = txBM.Where("budget_fact_entities.file_budget_id = ?", fid)
	} else {
		// Same fallback for Chart Budget
		var latestFid string
		r.db.Model(&models.BudgetFactEntity{}).Order("created_at desc").Limit(1).Pluck("file_budget_id", &latestFid)
		if latestFid != "" {
			txBM = txBM.Where("budget_fact_entities.file_budget_id = ?", latestFid)
		} else {
			txBM = txBM.Where("1 = 0")
		}
	}

	txBM.Group("budget_amount_entities.month").Scan(&budgetM)

	// Actual Monthly
	var actualM []MonthResult
	txAM := r.db.Table("actual_amount_entities").
		Select("actual_amount_entities.month, SUM(actual_amount_entities.amount) as total").
		Joins("JOIN actual_fact_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id")

	txAM = applyCommonFilters(txAM, "actual_fact_entities")
	if yVal, ok := filter["year"].(string); ok && yVal != "" && yVal != "All" {
		yVal = strings.ReplaceAll(yVal, "FY", "")
		txAM = txAM.Where("actual_fact_entities.year = ?", yVal)
	}
	// Months filter for chart: we only show what's requested. If none, show all.
	if mVal, ok := filter["months"]; ok {
		mstrs := toStringSlice(mVal)
		if len(mstrs) > 0 {
			txAM = txAM.Where("actual_amount_entities.month IN ?", mstrs)
		}
	}

	txAM.Group("actual_amount_entities.month").Scan(&actualM)

	// D. Merge Chart Data (JAN -> DEC order)
	monthMap := make(map[string]*models.MonthlyStatDTO)
	for _, m := range budgetM {
		monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: m.Total}
	}
	for _, m := range actualM {
		if _, ok := monthMap[m.Month]; !ok {
			monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month}
		}
		monthMap[m.Month].Actual += m.Total
	}

	// --- 4. Top Expenses (Group 3) ---
	var topExpenses []models.TopExpenseDTO
	txTE := r.db.Table("actual_amount_entities").
		Select("budget_structure_entities.group3 as name, SUM(actual_amount_entities.amount) as amount").
		Joins("JOIN actual_fact_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id").
		Joins("JOIN budget_structure_entities ON actual_fact_entities.conso_gl = budget_structure_entities.conso_gl")

	txTE = applyCommonFilters(txTE, "actual_fact_entities")

	if yVal, ok := filter["year"].(string); ok && yVal != "" && yVal != "All" {
		yVal = strings.ReplaceAll(yVal, "FY", "")
		txTE = txTE.Where("actual_fact_entities.year = ?", yVal)
	}

	if mVal, ok := filter["months"]; ok {
		mstrs := toStringSlice(mVal)
		if len(mstrs) > 0 {
			txTE = txTE.Where("actual_amount_entities.month IN ?", mstrs)
		}
	}

	if err := txTE.Group("budget_structure_entities.group3").
		Order("amount DESC").
		Limit(3).
		Scan(&topExpenses).Error; err != nil {
		logs.Errorf("[ERROR] OwnerRepository: Top Expenses Calculation Failed: %v", err)
	}
	summary.TopExpenses = topExpenses

	monthsOrder := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	for _, mon := range monthsOrder {
		if val, ok := monthMap[mon]; ok {
			summary.ChartData = append(summary.ChartData, *val)
		} else {
			summary.ChartData = append(summary.ChartData, models.MonthlyStatDTO{Month: mon})
		}
	}

	summary.TotalCount = int64(len(summary.DepartmentData))
	summary.Page = 1
	summary.Limit = len(summary.DepartmentData)
	if summary.Limit == 0 {
		summary.Limit = 10
	}

	logs.Infof("[DEBUG] OwnerRepository: Final Result - Budget: %.2f, Actual: %.2f", summary.TotalBudget, summary.TotalActual)
	return summary, nil
}

func (r *ownerRepository) GetActualTransactions(filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	// Standard Transaction logic, typically similar to admin but can be specialized
	var txs []models.ActualTransactionEntity
	var total int64

	tx := r.db.Model(&models.ActualTransactionEntity{})

	// Apply Filter Logic (Subset of Dashboard logic but tailored for Owner)
	if val, ok := filter["entities"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			tx = tx.Where("entity IN ?", strs)
		}
	}
	if val, ok := filter["branches"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			tx = tx.Where("branch IN ?", strs)
		}
	}
	if val, ok := filter["departments"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			tx = tx.Where("department IN ?", strs)
		}
	}
	if yVal, ok := filter["year"].(string); ok && yVal != "" {
		tx = tx.Where("year = ?", strings.ReplaceAll(yVal, "FY", ""))
	}

	tx.Count(&total)

	// Pagination
	page := 1
	limit := 50
	if p, ok := filter["page"].(int); ok {
		page = p
	}
	if l, ok := filter["limit"].(int); ok {
		limit = l
	}
	tx.Offset((page - 1) * limit).Limit(limit).Find(&txs)

	// Map to DTO
	var dtos []models.ActualTransactionDTO
	for _, t := range txs {
		dtos = append(dtos, models.ActualTransactionDTO{
			Source:        t.Source,
			PostingDate:   t.PostingDate,
			DocNo:         t.DocNo,
			Vendor:        t.VendorName,
			Description:   t.Description,
			GLAccountNo:   t.ConsoGL,
			GLAccountName: "-",
			Amount:        t.Amount,
			Department:    t.Department,
			Company:       t.Entity,
			Branch:        t.Branch,
		})
	}

	return &models.PaginatedActualTransactionDTO{
		Data:       dtos,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (r *ownerRepository) GetActualDetails(filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	var results []models.ActualFactEntity
	tx := r.db.Model(&models.ActualFactEntity{})
	// Basic filtering...
	tx.Find(&results)
	return results, nil
}

func (r *ownerRepository) GetBudgetDetails(filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	tx := r.db.Model(&models.BudgetFactEntity{})
	// IMPORTANT: respects budget_file_id
	if fid, ok := filter["budget_file_id"].(string); ok && fid != "" {
		tx = tx.Where("file_budget_id = ?", fid)
	}
	tx.Find(&results)
	return results, nil
}

func (r *ownerRepository) GetActualYears() ([]string, error) {
	var years []string
	err := r.db.Model(&models.ActualFactEntity{}).
		Distinct("year").
		Order("year DESC").
		Pluck("year", &years).Error
	return years, err
}

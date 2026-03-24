package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) models.DashboardRepository {
	return &dashboardRepository{db: db}
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

func (r *dashboardRepository) GetBudgetFilterOptions(ctx context.Context) ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	if err := r.db.WithContext(ctx).Model(&models.BudgetFactEntity{}).
		Distinct("\"group\"", "department", "entity_gl", "conso_gl", "gl_name").
		Order("\"group\", department, entity_gl, conso_gl").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetBudgetFilterOptions: %w", err)
	}
	return results, nil
}

func (r *dashboardRepository) GetOrganizationStructure(ctx context.Context) ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	// รวม Entity และ Branch ที่ไม่ซ้ำกันจากทั้งตาราง Budget และ Actual
	// GORM ไม่รองรับ UNION ในการ Scan struct โดยตรงได้ง่ายๆ
	// เราจึงใช้ Raw SQL เพื่อความชัดเจนและประสิทธิภาพ

	query := `
        SELECT DISTINCT entity, branch, department FROM budget_fact_entities WHERE entity != ''
        UNION
        SELECT DISTINCT entity, branch, department FROM actual_fact_entities WHERE entity != ''
        ORDER BY entity, branch, department
    `

	if err := r.db.WithContext(ctx).Raw(query).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetOrganizationStructure: %w", err)
	}
	return results, nil
}

func (r *dashboardRepository) GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]models.BudgetDetailDTO, error) {
	var results []models.BudgetDetailDTO
	query := r.db.WithContext(ctx).Model(&models.BudgetFactEntity{})

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
				query = query.Where(fmt.Sprintf("budget_fact_entities.%s IN ?", dbCol), strs)
			}
		}
	}

	applyFilter("groups", "\"group\"")
	applyFilter("departments", "department")
	applyFilter("entity_gls", "entity_gl")
	applyFilter("conso_gls", "conso_gl")
	applyFilter("entities", "entity")
	applyFilter("branches", "branch")

	type flatResult struct {
		ConsoGL string
		GLName  string
		Month   string
		Amount  float64
	}

	if val, ok := filter["months"]; ok {
		mstrs := toStringSlice(val)
		if len(mstrs) > 0 {
			query = query.Joins("JOIN budget_amount_entities ba2 ON ba2.budget_fact_id = budget_fact_entities.id").
				Where("ba2.month IN ?", mstrs)
		} else {
			// Strict Month Filter: If months list is empty, return no data
			query = query.Where("1 = 0")
		}
	}

	if val, ok := filter["budget_file_id"]; ok && val != "" {
		query = query.Where("budget_fact_entities.file_budget_id = ?", val)
	}

	var flatData []flatResult
	err := query.
		Joins("JOIN budget_amount_entities ba ON ba.budget_fact_id = budget_fact_entities.id AND ba.deleted_at IS NULL").
		Select("budget_fact_entities.conso_gl, MAX(budget_fact_entities.gl_name) as gl_name, ba.month, SUM(ba.amount) as amount").
		Group("budget_fact_entities.conso_gl, ba.month").
		Scan(&flatData).Error

	if err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetBudgetDetails: %w", err)
	}

	// Map flat results to nested DTO
	budgetMap := make(map[string]*models.BudgetDetailDTO)
	for _, row := range flatData {
		dto, exists := budgetMap[row.ConsoGL]
		if !exists {
			dto = &models.BudgetDetailDTO{
				ConsoGL:       row.ConsoGL,
				GLName:        row.GLName,
				YearTotal:     0,
				BudgetAmounts: []models.BudgetAmountDTO{},
			}
			budgetMap[row.ConsoGL] = dto
		}
		dto.BudgetAmounts = append(dto.BudgetAmounts, models.BudgetAmountDTO{
			Month:  row.Month,
			Amount: row.Amount,
		})
		dto.YearTotal += row.Amount
	}

	// Convert map to slice
	for _, dto := range budgetMap {
		results = append(results, *dto)
	}

	return results, nil
}

func (r *dashboardRepository) GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	var results []models.ActualFactEntity
	query := r.db.WithContext(ctx).Model(&models.ActualFactEntity{}).Preload("ActualAmounts")

	// Dynamic Filtering Helper
	applyFilter := func(key string, dbCol string) {
		if val, ok := filter[key]; ok {
			strs := toStringSlice(val)
			if len(strs) > 0 {
				query = query.Where(fmt.Sprintf("%s IN ?", dbCol), strs)
			}
		}
	}

	// Unified GL Filter (Supports both naming conventions)
	if val, ok := filter["budget_gls"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("conso_gl IN ?", strs)
		}
	} else if val, ok := filter["conso_gls"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("conso_gl IN ?", strs)
		}
	}

	applyFilter("departments", "department")
	applyFilter("entities", "entity")
	applyFilter("branches", "branch")

	if val, ok := filter["year"].(string); ok && val != "" {
		query = query.Where("year = ?", val)
	}

	if val, ok := filter["months"]; ok {
		mstrs := toStringSlice(val)
		if len(mstrs) > 0 {
			// Filter the fact records that have at least one of these months
			query = query.Joins("JOIN actual_amount_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id").
				Where("actual_amount_entities.month IN ?", mstrs).
				Group("actual_fact_entities.id")
		} else {
			// Strict Month Filter
			query = query.Where("1 = 0")
		}
	}

	// เรียงลำดับข้อมูล
	if err := query.Order("department, conso_gl, gl_name").Find(&results).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetActualDetails: %w", err)
	}
	return results, nil
}

func (r *dashboardRepository) GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	query := r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
		Select(`
			actual_transaction_entities.source,
			actual_transaction_entities.posting_date,
			actual_transaction_entities.doc_no as doc_no,
			actual_transaction_entities.vendor_name as vendor,
			actual_transaction_entities.description,
			COALESCE(
				NULLIF(mapping.conso_gl, ''),
				NULLIF((SELECT conso_gl FROM gl_mapping_entities WHERE entity_gl = actual_transaction_entities.entity_gl AND conso_gl != '' LIMIT 1), ''),
				actual_transaction_entities.entity_gl
			) as gl_account_no,
			actual_transaction_entities.conso_gl as conso_gl,
			COALESCE(
				NULLIF(mapping.account_name, ''),
				NULLIF((SELECT account_name FROM gl_mapping_entities WHERE entity_gl = actual_transaction_entities.entity_gl AND account_name != '' LIMIT 1), ''),
				NULLIF(actual_transaction_entities.gl_account_name, ''),
				'Unmapped GL'
			) as gl_account_name,
			actual_transaction_entities.amount,
			actual_transaction_entities.department,
			actual_transaction_entities.entity as company,
			actual_transaction_entities.branch
		`).
		Joins("LEFT JOIN gl_mapping_entities mapping ON actual_transaction_entities.entity_gl = mapping.entity_gl AND actual_transaction_entities.entity = mapping.entity")

	// 2. Filters (Only apply if not empty)
	if val, ok := filter["entities"]; ok {
		if s, ok := val.([]string); ok && len(s) > 0 {
			query = query.Where("actual_transaction_entities.entity IN ?", s)
		} else if s, ok := val.([]interface{}); ok && len(s) > 0 {
			query = query.Where("actual_transaction_entities.entity IN ?", s)
		}
	}
	if val, ok := filter["branches"]; ok {
		if s, ok := val.([]string); ok && len(s) > 0 {
			query = query.Where("actual_transaction_entities.branch IN ?", s)
		} else if s, ok := val.([]interface{}); ok && len(s) > 0 {
			query = query.Where("actual_transaction_entities.branch IN ?", s)
		}
	}
	if val, ok := filter["departments"]; ok {
		if s, ok := val.([]string); ok && len(s) > 0 {
			query = query.Where("actual_transaction_entities.department IN ?", s)
		} else if s, ok := val.([]interface{}); ok && len(s) > 0 {
			query = query.Where("actual_transaction_entities.department IN ?", s)
		}
	}
	if val, ok := filter["conso_gls"]; ok {
		if s, ok := val.([]string); ok && len(s) > 0 {
			query = query.Where("actual_transaction_entities.conso_gl IN ?", s)
		} else if s, ok := val.([]interface{}); ok && len(s) > 0 {
			query = query.Where("actual_transaction_entities.conso_gl IN ?", s)
		}
	}
	if val, ok := filter["months"]; ok {
		mstrs := toStringSlice(val)
		if len(mstrs) > 0 {
			query = query.Where("UPPER(TO_CHAR(actual_transaction_entities.posting_date::DATE, 'MON')) IN ?", mstrs)
		} else {
			// Strict Month Filter
			query = query.Where("1 = 0")
		}
	}
	if val, ok := filter["start_date"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.posting_date >= ?", val)
	}
	if val, ok := filter["end_date"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.posting_date <= ?", val)
	}
	if val, ok := filter["year"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.year = ?", val)
	}

	// 3. Pagination
	page := 0
	limit := 10
	if p, ok := filter["page"].(float64); ok {
		page = int(p)
	}
	if l, ok := filter["limit"].(float64); ok {
		limit = int(l)
	}
	offset := page * limit

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetActualTransactions.Count: %w", err)
	}
	fmt.Printf("[DEBUG] GetActualTransactions Total Found: %d\n", total)

	var results []models.ActualTransactionDTO
	if err := query.Order("actual_transaction_entities.posting_date ASC, actual_transaction_entities.doc_no ASC").
		Limit(limit).Offset(offset).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetActualTransactions.Scan: %w", err)
	}

	if len(results) > 0 {
		fmt.Printf("[DEBUG] GetActualTransactions Sample: %+v\n", results[0])
	}

	return &models.PaginatedActualTransactionDTO{
		Data:       results,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (r *dashboardRepository) GetDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	summary := &models.DashboardSummaryDTO{
		DepartmentData: []models.DepartmentStatDTO{},
		ChartData:      []models.MonthlyStatDTO{},
	}

	logs.Infof("[DEBUG] Repository: GetDashboardAggregates with Filters: %+v", filter)

	// ตัวช่วยกรองข้อมูลแบบไดนามิก (คืนค่า GORM Scope)
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
		if val, ok := filter["branches"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".branch IN ?", strs)
			}
		}
		// Added: Departments Filter
		if val, ok := filter["departments"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".department IN ?", strs)
			}
		}
		// Added: NavCodes Filter (For Sub-Department Drill-down)
		if val, ok := filter["nav_codes"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".nav_code IN ?", strs)
			}
		}
		// Added: Budget GLs Filter (Supports both naming conventions)
		glVal, hasGL := filter["budget_gls"]
		if !hasGL {
			glVal, hasGL = filter["conso_gls"]
		}

		if hasGL {
			if strs := toStringSlice(glVal); len(strs) > 0 {
				tx = tx.Where(tableName+".conso_gl IN ?", strs)
			}
		}

		// Added: Year Filter (Only apply to Actual tables, Budget uses file_id)
		if tableName == "actual_fact_entities" || tableName == "actual_amount_entities" || tableName == "actual_transaction_entities" {
			if val, ok := filter["year"]; ok {
				if s, ok := val.(string); ok && s != "" {
					s = strings.ReplaceAll(s, "FY", "")
					tx = tx.Where(tableName+".year = ?", s)
				}
			}
		}

		// Added: ID-based version filters
		if tableName == "budget_fact_entities" {
			if val, ok := filter["budget_file_id"]; ok && val != "" {
				tx = tx.Where(tableName+".file_budget_id = ?", val)
			}
		}
		// Assuming similar file ID fields for capex if needed later, but focusing on Budget for now.
		if tableName == "capex_fact_entities" {
			if val, ok := filter["capex_file_id"]; ok && val != "" {
				tx = tx.Where(tableName+".file_capex_id = ?", val)
			}
		}

		// Added: Restricted flag (Extra safety)
		if val, ok := filter["is_restricted"].(bool); ok && val {
			if _, hasDept := filter["departments"]; !hasDept {
				// If restricted but no departments given, and not admin (handled in service),
				// we should probably enforce a non-match.
				// But service should handle this.
			}
		}
		return tx
	}

	// Determine Grouping Strategy (Drill-down vs Summary)
	groupBy := "department"
	selectCol := "department"

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

	// 1. รวมยอดตาม Department (Department Aggregation)
	// Budget
	type DeptResult struct {
		Department string
		Total      float64
	}
	var budgetDept []DeptResult
	tx1 := r.db.Model(&models.BudgetFactEntity{}).Select(selectCol + ", SUM(year_total) as total")
	tx1 = applyFilter(tx1, "budget_fact_entities")
	if err := tx1.WithContext(ctx).Group(groupBy).Scan(&budgetDept).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetDashboardAggregates.Budget: %w", err)
	}
	fmt.Printf("[DEBUG] Aggregates Budget Count: %d, SQL: %s\n", len(budgetDept), tx1.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx }))

	// Actual
	var actualDept []DeptResult
	tx2 := r.db.Model(&models.ActualFactEntity{}).Select(selectCol + ", SUM(actual_fact_entities.year_total) as total")

	// If months filter is provided, we must sum from the amount table instead of using year_total
	if val, ok := filter["months"]; ok {
		mstrs := toStringSlice(val)
		if len(mstrs) > 0 {
			tx2 = tx2.Select(selectCol+", SUM(actual_amount_entities.amount) as total").
				Joins("JOIN actual_amount_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id").
				Where("actual_amount_entities.month IN ?", mstrs)
		} else {
			// No months selected: set total to 0 explicitly to avoid SQL errors
			tx2 = tx2.Select(selectCol + ", 0 as total").Where("1 = 0")
		}
	} else {
		// No month filter: Use the pre-aggregated year_total (Default behavior)
		// No changes needed, tx2 already has the default Select
	}

	tx2 = applyFilter(tx2, "actual_fact_entities")

	if err := tx2.WithContext(ctx).Group(groupBy).Scan(&actualDept).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetDashboardAggregates.Actual: %w", err)
	}
	fmt.Printf("[DEBUG] Aggregates Actual Count: %d, SQL: %s\n", len(actualDept), tx2.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx }))

	// รวมข้อมูล (Merge Department Data)
	deptMap := make(map[string]*models.DepartmentStatDTO)
	for _, b := range budgetDept {
		fmt.Printf("[DEBUG] Row Budget: Dept=%q, Total=%f\n", b.Department, b.Total)
		deptMap[b.Department] = &models.DepartmentStatDTO{Department: b.Department, Budget: b.Total}
		summary.TotalBudget += b.Total
	}
	for _, a := range actualDept {
		fmt.Printf("[DEBUG] Row Actual: Dept=%q, Total=%f\n", a.Department, a.Total)
		if _, ok := deptMap[a.Department]; !ok {
			deptMap[a.Department] = &models.DepartmentStatDTO{Department: a.Department}
		}
		deptMap[a.Department].Actual += a.Total
		summary.TotalActual += a.Total
	}
	fmt.Printf("[DEBUG] Aggregates Final - TotalBudget: %f, TotalActual: %f, DeptCount: %d\n", summary.TotalBudget, summary.TotalActual, len(deptMap))

	// แปลง Map เป็น Slice (Flatten Map)
	var allDepts []models.DepartmentStatDTO
	for _, v := range deptMap {
		allDepts = append(allDepts, *v)
	}

	// ตรรกะการเรียงลำดับ (Sort Logic)
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

	// คำนวณสถานะภาพรวม (ก่อนแบ่งหน้า)
	var overBudgetCount, nearLimitCount int
	for _, d := range allDepts {
		budget := d.Budget
		actual := d.Actual
		remaining := budget - actual

		// เกินงบ: (Budget=0 & Actual>0) หรือ (คงเหลือ < 0)
		if (budget == 0 && actual > 0) || remaining < 0 {
			overBudgetCount++
		} else if budget > 0 {
			// ใกล้เต็ม: คงเหลือ < 20%
			ratio := remaining / budget
			if ratio <= 0.2 {
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
		case "remaining_pct":
			if allDepts[i].Budget != 0 {
				valI = (allDepts[i].Budget - allDepts[i].Actual) / allDepts[i].Budget
			} else {
				valI = -999999 // Treat no budget as lowest priority or handled specifically
			}
			if allDepts[j].Budget != 0 {
				valJ = (allDepts[j].Budget - allDepts[j].Actual) / allDepts[j].Budget
			} else {
				valJ = -999999
			}
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
	if err := tx3.WithContext(ctx).Group("budget_amount_entities.month").Scan(&budgetMonth).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetDashboardAggregates.ChartBudget: %w", err)
	}

	// Actual Amounts
	var actualMonth []MonthResult
	tx4 := r.db.Table("actual_amount_entities").
		Select("actual_amount_entities.month, SUM(actual_amount_entities.amount) as total").
		Joins("JOIN actual_fact_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id")

	if val, ok := filter["months"]; ok {
		mstrs := toStringSlice(val)
		if len(mstrs) > 0 {
			tx4 = tx4.Where("actual_amount_entities.month IN ?", mstrs)
		} else {
			// Strict Month Filter
			tx4 = tx4.Where("1 = 0")
		}
	} else {
		// Strict Month Filter
		tx4 = tx4.Where("1 = 0")
	}

	tx4 = applyFilter(tx4, "actual_fact_entities")
	if err := tx4.WithContext(ctx).Group("actual_amount_entities.month").Scan(&actualMonth).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetDashboardAggregates.ChartActual: %w", err)
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

func (r *dashboardRepository) GetActualYears(ctx context.Context) ([]string, error) {
	var years []string
	if err := r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
		Distinct().
		Order("year DESC").
		Pluck("year", &years).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetActualYears: %w", err)
	}
	return years, nil
}

func (r *dashboardRepository) GetAvailableMonths(ctx context.Context, year string) ([]string, error) {
	var months []string
	// Query from the fast Data Inventory metadata table
	if err := r.db.WithContext(ctx).Model(&models.DataInventoryEntity{}).
		Where("year = ?", year).
		Order(`CASE 
			WHEN month = 'JAN' THEN 1 WHEN month = 'FEB' THEN 2 WHEN month = 'MAR' THEN 3
			WHEN month = 'APR' THEN 4 WHEN month = 'MAY' THEN 5 WHEN month = 'JUN' THEN 6
			WHEN month = 'JUL' THEN 7 WHEN month = 'AUG' THEN 8 WHEN month = 'SEP' THEN 9
			WHEN month = 'OCT' THEN 10 WHEN month = 'NOV' THEN 11 WHEN month = 'DEC' THEN 12
		END`).
		Pluck("month", &months).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetAvailableMonths: %w", err)
	}
	return months, nil
}

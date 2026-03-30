package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/shopspring/decimal"
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
		Amount  decimal.Decimal
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
		Where("budget_fact_entities.conso_gl != '' AND budget_fact_entities.conso_gl IS NOT NULL").
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
				YearTotal:     decimal.Zero,
				BudgetAmounts: []models.BudgetAmountDTO{},
			}
			budgetMap[row.ConsoGL] = dto
		}
		dto.BudgetAmounts = append(dto.BudgetAmounts, models.BudgetAmountDTO{
			Month:  row.Month,
			Amount: row.Amount,
		})
		dto.YearTotal = dto.YearTotal.Add(row.Amount)
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
				Where("actual_fact_entities.conso_gl != '' AND actual_fact_entities.conso_gl IS NOT NULL").
				Group("actual_fact_entities.id")
		} else {
			// Strict Month Filter
			query = query.Where("1 = 0")
		}
	} else {
		// Even without month filter, hide unmapped GLs
		query = query.Where("conso_gl != '' AND conso_gl IS NOT NULL")
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
			actual_transaction_entities.id,
			actual_transaction_entities.source,
			actual_transaction_entities.posting_date,
			actual_transaction_entities.doc_no as doc_no,
			actual_transaction_entities.vendor_name as vendor,
			actual_transaction_entities.description,
			actual_transaction_entities.conso_gl as conso_gl,
			actual_transaction_entities.gl_account_name as gl_account_name,
			actual_transaction_entities.amount,
			actual_transaction_entities.department,
			actual_transaction_entities.entity as company,
			actual_transaction_entities.branch,
			CASE 
				WHEN alrie.id IS NOT NULL THEN 'Request Change'
				WHEN al.status = 'CONFIRMED' THEN 'Approved'
				ELSE 'None'
			END as audit_status
		`).
		Joins("LEFT JOIN audit_log_rejected_item_entities alrie ON alrie.transaction_id = actual_transaction_entities.id").
		Joins("LEFT JOIN audit_log_entities al ON al.entity = actual_transaction_entities.entity AND al.branch = actual_transaction_entities.branch AND al.department = actual_transaction_entities.department AND al.year = actual_transaction_entities.year AND al.month = UPPER(TO_CHAR(actual_transaction_entities.posting_date::DATE, 'MON')) AND al.status = 'CONFIRMED'").
		Joins("LEFT JOIN gl_grouping_entities mapping ON actual_transaction_entities.entity_gl = mapping.entity_gl AND actual_transaction_entities.entity = mapping.entity")

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
		// Added: Departments Filter (Handles "None" for unmapped data)
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

				if hasNone {
					if len(filteredStrs) > 0 {
						tx = tx.Where("("+tableName+".department IN ? OR "+tableName+".department = '' OR "+tableName+".department IS NULL OR "+tableName+".department = 'None')", filteredStrs)
					} else {
						tx = tx.Where("("+tableName+".department = '' OR "+tableName+".department IS NULL OR "+tableName+".department = 'None')")
					}
				} else {
					tx = tx.Where(tableName+".department IN ?", strs)
				}
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

		// Added: Year Filter
		targetYear := ""
		if val, ok := filter["year"]; ok {
			if s, ok := val.(string); ok && s != "" {
				targetYear = strings.ReplaceAll(s, "FY", "")
				// Apply to actual table always. Apply to budget only if no explicit file ID.
				if tableName != "budget_fact_entities" && tableName != "capex_budget_fact_entities" {
					tx = tx.Where(tableName+".year = ?", targetYear)
				}
			}
		}

		// Added: ID-based version filters with Fallback
		if tableName == "budget_fact_entities" {
			if fid, ok := filter["budget_file_id"]; ok && fid != "" {
				tx = tx.Where(tableName+".file_budget_id = ?", fid)
			} else if yr, ok := filter["year"]; ok && yr != "" {
				// 🛡️ Strict Fallback: Sum only THE latest PL file for this year
				var latestFid string
				r.db.Model(&models.BudgetFactEntity{}).
					Where("year = ?", yr).
					Order("created_at DESC").
					Limit(1).
					Pluck("file_budget_id", &latestFid)

				if latestFid != "" {
					tx = tx.Where(tableName+".file_budget_id = ?", latestFid)
				} else {
					// No file found for this year = No data instead of summing ALL random files
					tx = tx.Where("1 = 0")
				}
			}
		}

		if tableName == "capex_budget_fact_entities" {
			if val, ok := filter["capex_file_id"]; ok && val != "" {
				tx = tx.Where(tableName+".file_capex_budget_id = ?", val)
			} else {
				// Fallback for Capex
				var latestFid string
				subQuery := r.db.Model(&models.CapexBudgetFactEntity{})
				if val, ok := filter["year"]; ok {
					if s, ok := val.(string); ok && s != "" {
						s = strings.ReplaceAll(s, "FY", "")
						subQuery = subQuery.Where("year = ?", s)
					}
				}
				if err := subQuery.Order("created_at desc").Limit(1).Pluck("file_capex_budget_id", &latestFid).Error; err == nil && latestFid != "" {
					tx = tx.Where(tableName+".file_capex_budget_id = ?", latestFid)
				}
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
	groupBy := "COALESCE(NULLIF(department, ''), nav_code)"
	selectCol := "COALESCE(NULLIF(department, ''), nav_code) as department"

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
	// Budget - ALWAYS JOIN AMOUNT FOR PARITY
	type DeptResult struct {
		Department string
		Total      decimal.Decimal
	}
	var budgetDept []DeptResult
	tx1 := r.db.Model(&models.BudgetFactEntity{}).
		Select(selectCol+", SUM(ba.amount) as total").
		Joins("JOIN budget_amount_entities ba ON ba.budget_fact_id = budget_fact_entities.id AND ba.deleted_at IS NULL")

	tx1 = applyFilter(tx1, "budget_fact_entities")

	if err := tx1.WithContext(ctx).Group(groupBy).Scan(&budgetDept).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetDashboardAggregates.Budget: %w", err)
	}

	// Actual - ALWAYS JOIN AMOUNT FOR PARITY
	var actualDept []DeptResult
	tx2 := r.db.Model(&models.ActualFactEntity{}).
		Select(selectCol+", SUM(aa.amount) as total").
		Joins("JOIN actual_amount_entities aa ON aa.actual_fact_id = actual_fact_entities.id AND aa.deleted_at IS NULL")

	// If months filter is provided, it's already using join-summing, but we've unified it above.
	// We just apply the month filter if present.
	if val, ok := filter["months"]; ok {
		mstrs := toStringSlice(val)
		if len(mstrs) > 0 {
			tx2 = tx2.Where("aa.month IN ?", mstrs)
		} else {
			tx2 = tx2.Where("1 = 0")
		}
	}

	tx2 = applyFilter(tx2, "actual_fact_entities")

	if err := tx2.WithContext(ctx).Group(groupBy).Scan(&actualDept).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetDashboardAggregates.Actual: %w", err)
	}

	// รวมข้อมูล (Merge Department Data)
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

	redLimit := 100.0
	yellowLimit := 80.0
	if val, ok := filter["red_threshold"].(float64); ok && val > 0 {
		redLimit = val
	} else if val, ok := filter["red_threshold"].(int); ok && val > 0 {
		redLimit = float64(val)
	}
	if val, ok := filter["yellow_threshold"].(float64); ok && val > 0 {
		yellowLimit = val
	} else if val, ok := filter["yellow_threshold"].(int); ok && val > 0 {
		yellowLimit = float64(val)
	}

	for _, d := range allDepts {
		budget := d.Budget
		actual := d.Actual

		if budget.IsZero() {
			if actual.GreaterThan(decimal.Zero) {
				overBudgetCount++
			}
			continue
		}

		spendPct, _ := actual.Div(budget).Mul(decimal.NewFromInt(100)).Float64()
		if spendPct >= redLimit {
			overBudgetCount++
		} else if spendPct >= yellowLimit {
			nearLimitCount++
		}
	}
	summary.OverBudgetCount = overBudgetCount
	summary.NearLimitCount = nearLimitCount

	sort.Slice(allDepts, func(i, j int) bool {
		var valI, valJ decimal.Decimal
		switch sortBy {
		case "budget":
			valI, valJ = allDepts[i].Budget, allDepts[j].Budget
		case "actual": // Spend
			valI, valJ = allDepts[i].Actual, allDepts[j].Actual
		case "remaining":
			valI = allDepts[i].Budget.Sub(allDepts[i].Actual)
			valJ = allDepts[j].Budget.Sub(allDepts[j].Actual)
		case "remaining_pct":
			if !allDepts[i].Budget.IsZero() {
				valI = allDepts[i].Budget.Sub(allDepts[i].Actual).Div(allDepts[i].Budget)
			} else {
				valI = decimal.NewFromInt(-999999)
			}
			if !allDepts[j].Budget.IsZero() {
				valJ = allDepts[j].Budget.Sub(allDepts[j].Actual).Div(allDepts[j].Budget)
			} else {
				valJ = decimal.NewFromInt(-999999)
			}
		default:
			valI, valJ = allDepts[i].Actual, allDepts[j].Actual
		}

		if sortOrder == "asc" {
			return valI.LessThan(valJ)
		}
		return valI.GreaterThan(valJ)
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
	type MonthResult struct {
		Month string
		Total decimal.Decimal
	}
	var budgetMonth []MonthResult
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
		}
	}

	tx4 = applyFilter(tx4, "actual_fact_entities")
	if err := tx4.WithContext(ctx).Group("actual_amount_entities.month").Scan(&actualMonth).Error; err != nil {
		return nil, fmt.Errorf("dashboardRepo.GetDashboardAggregates.ChartActual: %w", err)
	}

	// Merge Chart Data
	monthMap := make(map[string]*models.MonthlyStatDTO)
	for _, m := range budgetMonth {
		monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: m.Total, Actual: decimal.Zero}
	}
	for _, m := range actualMonth {
		if _, ok := monthMap[m.Month]; !ok {
			monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: decimal.Zero, Actual: m.Total}
		} else {
			monthMap[m.Month].Actual = m.Total
		}
	}

	monthsOrder := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	for _, mon := range monthsOrder {
		if val, ok := monthMap[mon]; ok {
			summary.ChartData = append(summary.ChartData, *val)
		} else {
			summary.ChartData = append(summary.ChartData, models.MonthlyStatDTO{Month: mon, Budget: decimal.Zero, Actual: decimal.Zero})
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

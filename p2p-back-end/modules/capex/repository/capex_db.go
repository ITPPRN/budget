package repository

import (
	"context"
	"fmt"
	"p2p-back-end/modules/entities/models"
	"sort"

	"gorm.io/gorm"
)

type capexRepositoryDB struct {
	db *gorm.DB
}

func NewCapexRepositoryDB(db *gorm.DB) models.CapexRepository {
	return &capexRepositoryDB{db: db}
}

// Transaction Helper
func (r *capexRepositoryDB) WithTrx(trxHandle func(repo models.CapexRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewCapexRepositoryDB(tx)
		return trxHandle(repo)
	})
}

// ---------------------------------------------------------
// File Create Methods
// ---------------------------------------------------------

func (r *capexRepositoryDB) CreateFileCapexBudget(ctx context.Context, file *models.FileCapexBudgetEntity) error {
	if err := r.db.WithContext(ctx).Create(file).Error; err != nil {
		return fmt.Errorf("capexRepo.CreateFileCapexBudget: %w", err)
	}
	return nil
}

func (r *capexRepositoryDB) CreateFileCapexActual(ctx context.Context, file *models.FileCapexActualEntity) error {
	if err := r.db.WithContext(ctx).Create(file).Error; err != nil {
		return fmt.Errorf("capexRepo.CreateFileCapexActual: %w", err)
	}
	return nil
}

// ---------------------------------------------------------
// Fact Create Methods
// ---------------------------------------------------------

func (r *capexRepositoryDB) CreateCapexBudgetFacts(ctx context.Context, headers []models.CapexBudgetFactEntity) error {
	if err := r.db.WithContext(ctx).Omit("CapexBudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return fmt.Errorf("capexRepo.CreateCapexBudgetFacts.Headers: %w", err)
	}
	var allAmounts []models.CapexBudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.CapexBudgetAmounts...)
	}
	if len(allAmounts) > 0 {
		if err := r.db.WithContext(ctx).CreateInBatches(&allAmounts, 1000).Error; err != nil {
			return fmt.Errorf("capexRepo.CreateCapexBudgetFacts.Amounts: %w", err)
		}
	}
	return nil
}

func (r *capexRepositoryDB) CreateCapexActualFacts(ctx context.Context, headers []models.CapexActualFactEntity) error {
	if err := r.db.WithContext(ctx).Omit("CapexActualAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return fmt.Errorf("capexRepo.CreateCapexActualFacts.Headers: %w", err)
	}
	var allAmounts []models.CapexActualAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.CapexActualAmounts...)
	}
	if len(allAmounts) > 0 {
		if err := r.db.WithContext(ctx).CreateInBatches(&allAmounts, 1000).Error; err != nil {
			return fmt.Errorf("capexRepo.CreateCapexActualFacts.Amounts: %w", err)
		}
	}
	return nil
}

// ---------------------------------------------------------
// File List Methods
// ---------------------------------------------------------

func (r *capexRepositoryDB) ListFileCapexBudgets(ctx context.Context) ([]models.FileCapexBudgetEntity, error) {
	var files []models.FileCapexBudgetEntity
	if err := r.db.WithContext(ctx).Order("upload_at desc").Find(&files).Error; err != nil {
		return nil, fmt.Errorf("capexRepo.ListFileCapexBudgets: %w", err)
	}
	return files, nil
}

func (r *capexRepositoryDB) ListFileCapexActuals(ctx context.Context) ([]models.FileCapexActualEntity, error) {
	var files []models.FileCapexActualEntity
	if err := r.db.WithContext(ctx).Order("upload_at desc").Find(&files).Error; err != nil {
		return nil, fmt.Errorf("capexRepo.ListFileCapexActuals: %w", err)
	}
	return files, nil
}

// ---------------------------------------------------------
// Get Single File
// ---------------------------------------------------------

func (r *capexRepositoryDB) GetFileCapexBudget(ctx context.Context, id string) (*models.FileCapexBudgetEntity, error) {
	var file models.FileCapexBudgetEntity
	if err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("capexRepo.GetFileCapexBudget: %w", err)
	}
	return &file, nil
}

func (r *capexRepositoryDB) GetFileCapexActual(ctx context.Context, id string) (*models.FileCapexActualEntity, error) {
	var file models.FileCapexActualEntity
	if err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("capexRepo.GetFileCapexActual: %w", err)
	}
	return &file, nil
}

// ---------------------------------------------------------
// Delete Files
// ---------------------------------------------------------

func (r *capexRepositoryDB) DeleteFileCapexBudget(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.FileCapexBudgetEntity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("capexRepo.DeleteFileCapexBudget: %w", err)
	}
	return nil
}

func (r *capexRepositoryDB) DeleteFileCapexActual(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.FileCapexActualEntity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("capexRepo.DeleteFileCapexActual: %w", err)
	}
	return nil
}

func (r *capexRepositoryDB) DeleteCapexBudgetFactsByFileID(ctx context.Context, fileID string) error {
	if err := r.db.WithContext(ctx).Where("file_capex_budget_id = ?", fileID).Delete(&models.CapexBudgetFactEntity{}).Error; err != nil {
		return fmt.Errorf("capexRepo.DeleteCapexBudgetFactsByFileID: %w", err)
	}
	return nil
}

func (r *capexRepositoryDB) DeleteCapexActualFactsByFileID(ctx context.Context, fileID string) error {
	if err := r.db.WithContext(ctx).Where("file_capex_actual_id = ?", fileID).Delete(&models.CapexActualFactEntity{}).Error; err != nil {
		return fmt.Errorf("capexRepo.DeleteCapexActualFactsByFileID: %w", err)
	}
	return nil
}

// ---------------------------------------------------------
// Delete Facts
// ---------------------------------------------------------

func (r *capexRepositoryDB) DeleteAllCapexBudgetFacts(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Exec("TRUNCATE TABLE capex_budget_fact_entities, capex_budget_amount_entities RESTART IDENTITY CASCADE").Error; err != nil {
		return fmt.Errorf("capexRepo.DeleteAllCapexBudgetFacts: %w", err)
	}
	return nil
}

func (r *capexRepositoryDB) DeleteAllCapexActualFacts(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Exec("TRUNCATE TABLE capex_actual_fact_entities, capex_actual_amount_entities RESTART IDENTITY CASCADE").Error; err != nil {
		return fmt.Errorf("capexRepo.DeleteAllCapexActualFacts: %w", err)
	}
	return nil
}

// ---------------------------------------------------------
// Update Files (Rename)
// ---------------------------------------------------------

func (r *capexRepositoryDB) UpdateFileCapexBudget(ctx context.Context, id string, filename string) error {
	if err := r.db.WithContext(ctx).Model(&models.FileCapexBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error; err != nil {
		return fmt.Errorf("capexRepo.UpdateFileCapexBudget: %w", err)
	}
	return nil
}

func (r *capexRepositoryDB) UpdateFileCapexActual(ctx context.Context, id string, filename string) error {
	if err := r.db.WithContext(ctx).Model(&models.FileCapexActualEntity{}).Where("id = ?", id).Update("file_name", filename).Error; err != nil {
		return fmt.Errorf("capexRepo.UpdateFileCapexActual: %w", err)
	}
	return nil
}

// ---------------------------------------------------------
// Dashboard Aggregation
// ---------------------------------------------------------

func (r *capexRepositoryDB) GetCapexDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	summary := &models.DashboardSummaryDTO{
		DepartmentData: []models.DepartmentStatDTO{},
		ChartData:      []models.MonthlyStatDTO{},
	}

	// Filter Helper
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
		if val, ok := filter["departments"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".department IN ?", strs)
			}
		}
		if val, ok := filter["year"]; ok {
			if s, ok := val.(string); ok && s != "" && s != "All" {
				tx = tx.Where(tableName+".year = ?", s)
			}
		}

		// 🛠️ Key Mapping: Standardize between Frontend and Backend Repo
		if fid, ok := filter["capex_file_id"].(string); ok && fid != "" {
			filter["file_capex_budget_id"] = fid
		}
		if fid, ok := filter["capex_actual_file_id"].(string); ok && fid != "" {
			filter["file_capex_actual_id"] = fid
		}

		if tableName == "capex_budget_fact_entities" {
			if fid, ok := filter["file_capex_budget_id"].(string); ok && fid != "" {
				tx = tx.Where("file_capex_budget_id = ?", fid)
			} else {
				// Fallback for independent budget
				var latestFid string
				r.db.Model(&models.CapexBudgetFactEntity{}).Order("created_at desc").Select("file_capex_budget_id").Limit(1).Take(&latestFid)
				if latestFid != "" {
					tx = tx.Where("file_capex_budget_id = ?", latestFid)
				}
			}
		}
		if tableName == "capex_actual_fact_entities" {
			if fid, ok := filter["file_capex_actual_id"].(string); ok && fid != "" {
				tx = tx.Where("file_capex_actual_id = ?", fid)
			} else {
				// Fallback for actual
				var latestFid string
				r.db.Model(&models.CapexActualFactEntity{}).Order("created_at desc").Select("file_capex_actual_id").Limit(1).Take(&latestFid)
				if latestFid != "" {
					tx = tx.Where("file_capex_actual_id = ?", latestFid)
				}
			}
		}

		return tx
	}

	// 1. Department Aggregation
	type DeptResult struct {
		Department string
		Total      float64
	}

	// CAPEX Budget
	var budgetDept []DeptResult
	tx1 := r.db.Table("capex_budget_fact_entities").Select("department, SUM(year_total) as total")
	tx1 = applyFilter(tx1, "capex_budget_fact_entities")
	if err := tx1.WithContext(ctx).Group("department").Scan(&budgetDept).Error; err != nil {
		return nil, fmt.Errorf("capexRepo.GetDashboardAggregates.Budget: %w", err)
	}

	// CAPEX Actual
	var actualDept []DeptResult
	tx2 := r.db.Table("capex_actual_fact_entities").Select("department, SUM(year_total) as total")
	tx2 = applyFilter(tx2, "capex_actual_fact_entities")
	if err := tx2.WithContext(ctx).Group("department").Scan(&actualDept).Error; err != nil {
		return nil, fmt.Errorf("capexRepo.GetDashboardAggregates.Actual: %w", err)
	}

	// Merge
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

	var allDepts []models.DepartmentStatDTO
	for _, v := range deptMap {
		allDepts = append(allDepts, *v)
	}

	// Sort
	sortBy := "actual"
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

	// Status Counts
	var overBudgetCount, nearLimitCount int

	// Thresholds from filter (Default: Red >= 100%, Yellow >= 80% [20% remaining])
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

		if budget == 0 {
			if actual > 0 {
				overBudgetCount++
			}
			continue
		}

		spendPct := (actual / budget) * 100
		if spendPct >= redLimit {
			overBudgetCount++
		} else if spendPct >= yellowLimit {
			nearLimitCount++
		}
	}
	summary.OverBudgetCount = overBudgetCount
	summary.NearLimitCount = nearLimitCount

	sort.Slice(allDepts, func(i, j int) bool {
		var valI, valJ float64
		switch sortBy {
		case "budget":
			valI, valJ = allDepts[i].Budget, allDepts[j].Budget
		case "actual":
			valI, valJ = allDepts[i].Actual, allDepts[j].Actual
		case "remaining":
			valI = allDepts[i].Budget - allDepts[i].Actual
			valJ = allDepts[j].Budget - allDepts[j].Actual
		case "remaining_pct":
			if allDepts[i].Budget != 0 {
				valI = (allDepts[i].Budget - allDepts[i].Actual) / allDepts[i].Budget
			} else {
				valI = -999999
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

	// Pagination
	page := 1
	limit := 10
	if val, ok := filter["page"]; ok {
		if p, ok := val.(float64); ok {
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

	// 2. Chart Aggregation
	type MonthResult struct {
		Month string
		Total float64
	}
	var budgetMonth []MonthResult
	tx3 := r.db.Table("capex_budget_amount_entities").
		Select("capex_budget_amount_entities.month, SUM(capex_budget_amount_entities.amount) as total").
		Joins("JOIN capex_budget_fact_entities ON capex_budget_amount_entities.capex_budget_fact_id = capex_budget_fact_entities.id")
	tx3 = applyFilter(tx3, "capex_budget_fact_entities")
	if err := tx3.WithContext(ctx).Group("capex_budget_amount_entities.month").Scan(&budgetMonth).Error; err != nil {
		return nil, fmt.Errorf("capexRepo.GetDashboardAggregates.ChartBudget: %w", err)
	}

	var actualMonth []MonthResult
	tx4 := r.db.Table("capex_actual_amount_entities").
		Select("capex_actual_amount_entities.month, SUM(capex_actual_amount_entities.amount) as total").
		Joins("JOIN capex_actual_fact_entities ON capex_actual_amount_entities.capex_actual_fact_id = capex_actual_fact_entities.id")
	tx4 = applyFilter(tx4, "capex_actual_fact_entities")
	if err := tx4.WithContext(ctx).Group("capex_actual_amount_entities.month").Scan(&actualMonth).Error; err != nil {
		return nil, fmt.Errorf("capexRepo.GetDashboardAggregates.ChartActual: %w", err)
	}

	monthMap := make(map[string]*models.MonthlyStatDTO)
	for _, m := range budgetMonth {
		monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: m.Total}
	}
	for _, m := range actualMonth {
		if _, ok := monthMap[m.Month]; !ok {
			monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month}
		}
		monthMap[m.Month].Actual += m.Total
	}

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

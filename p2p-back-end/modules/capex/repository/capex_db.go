package repository

import (
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

func (r *capexRepositoryDB) CreateFileCapexBudget(file *models.FileCapexBudgetEntity) error {
	return r.db.Create(file).Error
}

func (r *capexRepositoryDB) CreateFileCapexActual(file *models.FileCapexActualEntity) error {
	return r.db.Create(file).Error
}

// ---------------------------------------------------------
// Fact Create Methods
// ---------------------------------------------------------

func (r *capexRepositoryDB) CreateCapexBudgetFacts(headers []models.CapexBudgetFactEntity) error {
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

func (r *capexRepositoryDB) CreateCapexActualFacts(headers []models.CapexActualFactEntity) error {
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

func (r *capexRepositoryDB) ListFileCapexBudgets() ([]models.FileCapexBudgetEntity, error) {
	var files []models.FileCapexBudgetEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *capexRepositoryDB) ListFileCapexActuals() ([]models.FileCapexActualEntity, error) {
	var files []models.FileCapexActualEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

// ---------------------------------------------------------
// Get Single File
// ---------------------------------------------------------

func (r *capexRepositoryDB) GetFileCapexBudget(id string) (*models.FileCapexBudgetEntity, error) {
	var file models.FileCapexBudgetEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *capexRepositoryDB) GetFileCapexActual(id string) (*models.FileCapexActualEntity, error) {
	var file models.FileCapexActualEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// ---------------------------------------------------------
// Delete Files
// ---------------------------------------------------------

func (r *capexRepositoryDB) DeleteFileCapexBudget(id string) error {
	return r.db.Delete(&models.FileCapexBudgetEntity{}, "id = ?", id).Error
}

func (r *capexRepositoryDB) DeleteFileCapexActual(id string) error {
	return r.db.Delete(&models.FileCapexActualEntity{}, "id = ?", id).Error
}

func (r *capexRepositoryDB) DeleteCapexBudgetFactsByFileID(fileID string) error {
	return r.db.Where("file_capex_budget_id = ?", fileID).Delete(&models.CapexBudgetFactEntity{}).Error
}

func (r *capexRepositoryDB) DeleteCapexActualFactsByFileID(fileID string) error {
	return r.db.Where("file_capex_actual_id = ?", fileID).Delete(&models.CapexActualFactEntity{}).Error
}

// ---------------------------------------------------------
// Delete Facts
// ---------------------------------------------------------

func (r *capexRepositoryDB) DeleteAllCapexBudgetFacts() error {
	// Cascade delete via GORM hooks or database constraints is preferred,
	// but manual cleanup ensures consistency if constraints missing.
	// For simplicity and speed, we rely on cascade or separate truncation if needed.
	// Here we just delete Parent, assuming DB handles children or we risk orphans.
	// Given previous implementation styles, DELETE FROM amounts might be needed first.
	// Let's trust GORM association handling or direct SQL if constraints exist.
	// Safest: Delete amounts then headers.
	// Efficient: Truncate?
	// Consistent with budget_db:
	// "DeleteAll" usually implies clearing EVERYTHING for a full refresh.
	return r.db.Exec("TRUNCATE TABLE capex_budget_fact_entities, capex_budget_amount_entities RESTART IDENTITY CASCADE").Error
}

func (r *capexRepositoryDB) DeleteAllCapexActualFacts() error {
	return r.db.Exec("TRUNCATE TABLE capex_actual_fact_entities, capex_actual_amount_entities RESTART IDENTITY CASCADE").Error
}

// ---------------------------------------------------------
// Update Files (Rename)
// ---------------------------------------------------------

func (r *capexRepositoryDB) UpdateFileCapexBudget(id string, filename string) error {
	return r.db.Model(&models.FileCapexBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

func (r *capexRepositoryDB) UpdateFileCapexActual(id string, filename string) error {
	return r.db.Model(&models.FileCapexActualEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

// ---------------------------------------------------------
// Dashboard Aggregation
// ---------------------------------------------------------

func (r *capexRepositoryDB) GetCapexDashboardAggregates(filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
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
			if s, ok := val.(string); ok && s != "" {
				tx = tx.Where(tableName+".year = ?", s)
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
	if err := tx1.Group("department").Scan(&budgetDept).Error; err != nil {
		return nil, err
	}

	// CAPEX Actual
	var actualDept []DeptResult
	tx2 := r.db.Table("capex_actual_fact_entities").Select("department, SUM(year_total) as total")
	tx2 = applyFilter(tx2, "capex_actual_fact_entities")
	if err := tx2.Group("department").Scan(&actualDept).Error; err != nil {
		return nil, err
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
	for _, d := range allDepts {
		budget := d.Budget
		actual := d.Actual
		remaining := budget - actual

		if (budget == 0 && actual > 0) || remaining < 0 {
			overBudgetCount++
		} else if budget > 0 {
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
	if err := tx3.Group("capex_budget_amount_entities.month").Scan(&budgetMonth).Error; err != nil {
		return nil, err
	}

	var actualMonth []MonthResult
	tx4 := r.db.Table("capex_actual_amount_entities").
		Select("capex_actual_amount_entities.month, SUM(capex_actual_amount_entities.amount) as total").
		Joins("JOIN capex_actual_fact_entities ON capex_actual_amount_entities.capex_actual_fact_id = capex_actual_fact_entities.id")
	tx4 = applyFilter(tx4, "capex_actual_fact_entities")
	if err := tx4.Group("capex_actual_amount_entities.month").Scan(&actualMonth).Error; err != nil {
		return nil, err
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

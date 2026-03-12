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

	logs.Infof("[DEBUG] OwnerRepository: Starting GetDashboardAggregates with Filter: %+v", filter)

	// Helper for type-safe year extraction
	extractYear := func() string {
		if val, ok := filter["year"]; ok {
			switch v := val.(type) {
			case string:
				return strings.ReplaceAll(v, "FY", "")
			case float64:
				return fmt.Sprintf("%.0f", v)
			case int:
				return fmt.Sprintf("%d", v)
			}
		}
		return ""
	}
	yearToUse := extractYear()

	// 1. Department Aggregation
	type DeptResult struct {
		Department string  `gorm:"column:department"`
		Total      float64 `gorm:"column:total"`
	}

	// Budget Query
	txB := r.db.Model(&models.BudgetFactEntity{}).
		Select("department, SUM(year_total) as total").
		Group("department")

	if yearToUse != "" {
		txB = txB.Where("year = ?", yearToUse)
	}
	if depts := toStringSlice(filter["departments"]); len(depts) > 0 {
		txB = txB.Where("department IN ?", depts)
	}
	if ents := toStringSlice(filter["entities"]); len(ents) > 0 {
		txB = txB.Where("entity IN ?", ents)
	}
	if brs := toStringSlice(filter["branches"]); len(brs) > 0 {
		txB = txB.Where("branch IN ?", brs)
	}
	if gls := toStringSlice(filter["conso_gls"]); len(gls) > 0 {
		txB = txB.Where("conso_gl IN ?", gls)
	} else if bgls := toStringSlice(filter["budget_gls"]); len(bgls) > 0 {
		txB = txB.Where("conso_gl IN ?", bgls)
	}

	var budgetResults []DeptResult
	if err := txB.Scan(&budgetResults).Error; err != nil {
		logs.Errorf("[ERROR] OwnerRepository: Budget Aggregation Failed: %v", err)
		return nil, err
	}
	logs.Infof("[DEBUG] OwnerRepository: Budget Aggregate found %d departments for year %s", len(budgetResults), yearToUse)

	// Actual Query
	txA := r.db.Model(&models.ActualFactEntity{}).
		Select("department, SUM(year_total) as total").
		Group("department")

	if yearToUse != "" {
		txA = txA.Where("year = ?", yearToUse)
	}
	if depts := toStringSlice(filter["departments"]); len(depts) > 0 {
		txA = txA.Where("department IN ?", depts)
	}
	if ents := toStringSlice(filter["entities"]); len(ents) > 0 {
		txA = txA.Where("entity IN ?", ents)
	}
	if brs := toStringSlice(filter["branches"]); len(brs) > 0 {
		txA = txA.Where("branch IN ?", brs)
	}
	if gls := toStringSlice(filter["conso_gls"]); len(gls) > 0 {
		txA = txA.Where("conso_gl IN ?", gls)
	}

	var actualResults []DeptResult
	if err := txA.Scan(&actualResults).Error; err != nil {
		logs.Errorf("[ERROR] OwnerRepository: Actual Aggregation Failed: %v", err)
		return nil, err
	}
	logs.Infof("[DEBUG] OwnerRepository: Actual Aggregate found %d departments for year %s", len(actualResults), yearToUse)

	// Merge Data
	deptMap := make(map[string]*models.DepartmentStatDTO)
	for _, b := range budgetResults {
		deptMap[b.Department] = &models.DepartmentStatDTO{Department: b.Department, Budget: b.Total}
		summary.TotalBudget += b.Total
	}
	for _, a := range actualResults {
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

	// 2. Chart Aggregation
	type MonthResult struct {
		Month string  `gorm:"column:month"`
		Total float64 `gorm:"column:total"`
	}

	// Budget Monthly
	txBM := r.db.Table("budget_amount_entities").
		Select("budget_amount_entities.month, SUM(budget_amount_entities.amount) as total").
		Joins("JOIN budget_fact_entities ON budget_amount_entities.budget_fact_id = budget_fact_entities.id").
		Group("budget_amount_entities.month")

	if yearToUse != "" {
		txBM = txBM.Where("budget_fact_entities.year = ?", yearToUse)
	}
	if depts := toStringSlice(filter["departments"]); len(depts) > 0 {
		txBM = txBM.Where("budget_fact_entities.department IN ?", depts)
	}
	if ents := toStringSlice(filter["entities"]); len(ents) > 0 {
		txBM = txBM.Where("budget_fact_entities.entity IN ?", ents)
	}

	var budgetMonths []MonthResult
	txBM.Scan(&budgetMonths)

	// Actual Monthly
	txAM := r.db.Table("actual_amount_entities").
		Select("actual_amount_entities.month, SUM(actual_amount_entities.amount) as total").
		Joins("JOIN actual_fact_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id").
		Group("actual_amount_entities.month")

	if yearToUse != "" {
		txAM = txAM.Where("actual_fact_entities.year = ?", yearToUse)
	}
	if depts := toStringSlice(filter["departments"]); len(depts) > 0 {
		txAM = txAM.Where("actual_fact_entities.department IN ?", depts)
	}

	var actualMonths []MonthResult
	txAM.Scan(&actualMonths)

	// Merge Charts
	monthMap := make(map[string]*models.MonthlyStatDTO)
	for _, m := range budgetMonths {
		monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: m.Total}
	}
	for _, m := range actualMonths {
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
			summary.ChartData = append(summary.ChartData, models.MonthlyStatDTO{Month: mon})
		}
	}

	// Finalize Summary
	summary.DepartmentData = allDepts
	summary.TotalCount = int64(len(allDepts))
	summary.Page = 1
	summary.Limit = len(allDepts)
	if summary.Limit == 0 {
		summary.Limit = 10
	}

	logs.Infof("[DEBUG] OwnerRepository: Aggregation Complete. TotalBudget: %.2f, TotalActual: %.2f", summary.TotalBudget, summary.TotalActual)

	return summary, nil
}

func (r *ownerRepository) GetActualTransactions(filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	return nil, nil
}

func (r *ownerRepository) GetActualDetails(filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	return nil, nil
}

func (r *ownerRepository) GetBudgetDetails(filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	return nil, nil
}

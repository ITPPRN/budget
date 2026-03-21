package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"

	"gorm.io/gorm"
)

type CapexDeptStatusRepository interface {
	GetCapexDeptStatus(ctx context.Context, filter map[string]interface{}) ([]models.CapexDeptStatusDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) CapexDeptStatusRepository {
	return &repository{db: db}
}

func (r *repository) GetCapexDeptStatus(ctx context.Context, filter map[string]interface{}) ([]models.CapexDeptStatusDTO, error) {
	type summaryRow struct {
		Department string
		Total      float64
	}

	// 1. Get Capex Budgets per Department
	var budgetRows []summaryRow
	btx := r.db.Model(&models.CapexBudgetFactEntity{}).Select("department, SUM(year_total) as total")
	btx = r.applyCommonFilters(btx, "capex_budget_fact_entities", filter)
	if err := btx.WithContext(ctx).Group("department").Scan(&budgetRows).Error; err != nil {
		return nil, err
	}

	// 2. Get Capex Actuals per Department
	var actualRows []summaryRow
	atx := r.db.Model(&models.CapexActualFactEntity{}).Select("department, SUM(year_total) as total")
	atx = r.applyCommonFilters(atx, "capex_actual_fact_entities", filter)
	if err := atx.WithContext(ctx).Group("department").Scan(&actualRows).Error; err != nil {
		return nil, err
	}

	// 3. Merge and Calculate
	deptMap := make(map[string]*models.CapexDeptStatusDTO)
	for _, b := range budgetRows {
		deptMap[b.Department] = &models.CapexDeptStatusDTO{
			Department:  b.Department,
			CapexBudget: utils.ToDecimal(b.Total),
		}
	}
	for _, a := range actualRows {
		if _, ok := deptMap[a.Department]; !ok {
			deptMap[a.Department] = &models.CapexDeptStatusDTO{Department: a.Department}
		}
		deptMap[a.Department].Spend = utils.ToDecimal(a.Total)
	}

	var results []models.CapexDeptStatusDTO
	for _, d := range deptMap {
		d.Remaining = d.CapexBudget.Sub(d.Spend)
		if !d.CapexBudget.IsZero() {
			d.Percentage = d.Spend.Div(d.CapexBudget).InexactFloat64() * 100
		} else if !d.Spend.IsZero() {
			d.Percentage = 100
		}

		if (d.CapexBudget.IsZero() && !d.Spend.IsZero()) || d.Remaining.IsNegative() {
			d.Status = "Over Budget"
		} else if !d.CapexBudget.IsZero() {
			ratio := d.Remaining.Div(d.CapexBudget).InexactFloat64()
			if ratio <= 0.2 {
				d.Status = "Near Limit"
			} else {
				d.Status = "Normal"
			}
		} else {
			d.Status = "Normal"
		}
		results = append(results, *d)
	}

	return results, nil
}

func (r *repository) applyCommonFilters(tx *gorm.DB, tableName string, filter map[string]interface{}) *gorm.DB {
	if val, ok := filter["entities"]; ok {
		if strs := utils.ToStringSlice(val); len(strs) > 0 {
			tx = tx.Where(tableName+".entity IN ?", strs)
		}
	}
	if val, ok := filter["departments"]; ok {
		if strs := utils.ToStringSlice(val); len(strs) > 0 {
			tx = tx.Where(tableName+".department IN ?", strs)
		}
	}
	if tableName != "capex_budget_fact_entities" {
		if val := utils.GetSafeString(filter, "year"); val != "" {
			tx = tx.Where(tableName+".year = ?", val)
		}
	}
	if tableName == "capex_budget_fact_entities" {
		if val := utils.GetSafeString(filter, "capex_file_id"); val != "" {
			tx = tx.Where(tableName+".file_capex_budget_id = ?", val)
		} else {
			var latestFid string
			tx.Session(&gorm.Session{}).Model(&models.CapexBudgetFactEntity{}).Order("created_at desc").Select("file_capex_budget_id").Limit(1).Take(&latestFid)
			if latestFid != "" {
				tx = tx.Where(tableName+".file_capex_budget_id = ?", latestFid)
			}
		}
	}
	if tableName == "capex_actual_fact_entities" {
		if val := utils.GetSafeString(filter, "capex_actual_file_id"); val != "" {
			tx = tx.Where(tableName+".file_capex_actual_id = ?", val)
		} else {
			var latestFid string
			tx.Session(&gorm.Session{}).Model(&models.CapexActualFactEntity{}).Order("created_at desc").Select("file_capex_actual_id").Limit(1).Take(&latestFid)
			if latestFid != "" {
				tx = tx.Where(tableName+".file_capex_actual_id = ?", latestFid)
			}
		}
	}
	return tx
}

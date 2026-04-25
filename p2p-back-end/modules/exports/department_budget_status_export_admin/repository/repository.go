package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DeptBudgetStatusRepository interface {
	GetDeptBudgetStatus(ctx context.Context, filter map[string]interface{}) ([]models.DeptBudgetStatusDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) DeptBudgetStatusRepository {
	return &repository{db: db}
}

func (r *repository) GetDeptBudgetStatus(ctx context.Context, filter map[string]interface{}) ([]models.DeptBudgetStatusDTO, error) {
	type summaryRow struct {
		Department string
		Total      decimal.Decimal
	}

	// 1. Get Budgets per Department
	var budgetRows []summaryRow
	btx := r.db.Model(&models.BudgetFactEntity{}).Select("COALESCE(NULLIF(NULLIF(department, ''), 'None'), NULLIF(nav_code, ''), 'None') as department, SUM(year_total) as total")
	btx = r.applyCommonFilters(btx, "budget_fact_entities", filter)
	if err := btx.WithContext(ctx).Group("COALESCE(NULLIF(NULLIF(department, ''), 'None'), NULLIF(nav_code, ''), 'None')").Scan(&budgetRows).Error; err != nil {
		return nil, err
	}

	// 2. Get Actuals per Department
	var actualRows []summaryRow
	atx := r.db.Model(&models.ActualFactEntity{})

	// Select base columns early to avoid GORM's default "SELECT *" which fails with GROUP BY
	colExpr := "COALESCE(NULLIF(NULLIF(actual_fact_entities.department, ''), 'None'), NULLIF(actual_fact_entities.nav_code, ''), 'None') as department"

	// Month filter logic (must sum from amount table)
	if months, ok := filter["months"]; ok {
		mstrs := utils.ToStringSlice(months)
		if len(mstrs) > 0 {
			atx = atx.Select(colExpr + ", SUM(actual_amount_entities.amount) as total").
				Joins("JOIN actual_amount_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id").
				Where("actual_amount_entities.month IN ?", mstrs)
		} else {
			// If months requested but empty, force 0 results but keep valid SQL structure
			atx = atx.Select(colExpr + ", 0 as total").Where("1 = 0")
		}
	} else {
		atx = atx.Select("COALESCE(NULLIF(NULLIF(department, ''), 'None'), NULLIF(nav_code, ''), 'None') as department, SUM(year_total) as total")
	}

	atx = r.applyCommonFilters(atx, "actual_fact_entities", filter)
	if err := atx.WithContext(ctx).Group("COALESCE(NULLIF(NULLIF(department, ''), 'None'), NULLIF(nav_code, ''), 'None')").Scan(&actualRows).Error; err != nil {
		return nil, err
	}

	deptMap := make(map[string]*models.DeptBudgetStatusDTO)
	for _, b := range budgetRows {
		deptMap[b.Department] = &models.DeptBudgetStatusDTO{
			Department: b.Department,
			Budget:     b.Total,
		}
	}
	for _, a := range actualRows {
		if _, ok := deptMap[a.Department]; !ok {
			deptMap[a.Department] = &models.DeptBudgetStatusDTO{Department: a.Department, Budget: decimal.NewFromFloat(0)}
		}
		deptMap[a.Department].Spend = a.Total
	}

	var results []models.DeptBudgetStatusDTO
	for _, d := range deptMap {
		d.Remaining = d.Budget.Sub(d.Spend)
		if !d.Budget.IsZero() {
			d.Percentage = d.Spend.Div(d.Budget).InexactFloat64() * 100
		} else if !d.Spend.IsZero() {
			d.Percentage = 100
		}

		// Status Logic (Consistent with Dashboard)
		if (d.Budget.IsZero() && !d.Spend.IsZero()) || d.Remaining.IsNegative() {
			d.Status = "Over Budget"
		} else if !d.Budget.IsZero() {
			ratio := d.Remaining.Div(d.Budget).InexactFloat64()
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
	if val, ok := filter["branches"]; ok {
		if strs := utils.ToStringSlice(val); len(strs) > 0 {
			tx = tx.Where(tableName+".branch IN ?", strs)
		}
	}
	if val, ok := filter["departments"]; ok {
		if strs := utils.ToStringSlice(val); len(strs) > 0 {
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
	if tableName != "budget_fact_entities" {
		if val := utils.GetSafeString(filter, "year"); val != "" {
			tx = tx.Where(tableName+".year = ?", val)
		}
	}

	if tableName == "budget_fact_entities" {
		if val := utils.GetSafeString(filter, "budget_file_id"); val != "" {
			tx = tx.Where(tableName+".file_budget_id = ?", val)
		}
	}
	return tx
}

package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"

	"gorm.io/gorm"
)

type OwnerCapexRepository interface {
	GetOwnerCapexData(ctx context.Context, filter map[string]interface{}) ([]models.OwnerCapexBudgetExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) OwnerCapexRepository {
	return &repository{db: db}
}

func (r *repository) GetOwnerCapexData(ctx context.Context, filter map[string]interface{}) ([]models.OwnerCapexBudgetExportDTO, error) {
	type capexRow struct {
		Entity        string
		Branch        string
		Department    string
		CapexNo       string
		CapexName     string
		CapexCategory string
		Total         float64
	}

	var budgetRows []capexRow
	btx := r.db.Table("capex_budget_fact_entities").
		Select("entity, MAX(branch) as branch, TRIM(department) as department, capex_no, MAX(capex_name) as capex_name, MAX(capex_category) as capex_category, SUM(year_total) as total").
		Group("entity, TRIM(department), capex_no")
	btx = r.applyCommonFilters(btx, "capex_budget_fact_entities", filter)
	if err := btx.Scan(&budgetRows).Error; err != nil {
		return nil, err
	}

	var actualRows []capexRow
	atx := r.db.Table("capex_actual_fact_entities").
		Select("entity, MAX(branch) as branch, TRIM(department) as department, capex_no, SUM(year_total) as total").
		Group("entity, TRIM(department), capex_no")
	atx = r.applyCommonFilters(atx, "capex_actual_fact_entities", filter)
	if err := atx.Scan(&actualRows).Error; err != nil {
		return nil, err
	}

	type key struct {
		Entity     string
		Department string
		CapexNo    string
	}
	capexMap := make(map[key]*models.OwnerCapexBudgetExportDTO)

	for _, b := range budgetRows {
		k := key{b.Entity, b.Department, b.CapexNo}
		capexMap[k] = &models.OwnerCapexBudgetExportDTO{
			Entity:        b.Entity,
			Branch:        b.Branch,
			Department:    b.Department,
			CapexNo:       b.CapexNo,
			CapexName:     b.CapexName,
			CapexCategory: b.CapexCategory,
			Budget:        utils.ToDecimal(b.Total),
		}
	}
	for _, a := range actualRows {
		k := key{a.Entity, a.Department, a.CapexNo}
		if _, ok := capexMap[k]; !ok {
			capexMap[k] = &models.OwnerCapexBudgetExportDTO{
				Entity:     a.Entity,
				Department: a.Department,
				CapexNo:    a.CapexNo,
				CapexName:  a.CapexName,
			}
		}
		capexMap[k].Actual = utils.ToDecimal(a.Total)
	}

	var results []models.OwnerCapexBudgetExportDTO
	for _, v := range capexMap {
		v.Remaining = v.Budget.Sub(v.Actual)
		if !v.Budget.IsZero() {
			v.Percentage = v.Actual.Div(v.Budget).InexactFloat64() * 100
		} else if !v.Actual.IsZero() {
			v.Percentage = 100
		}
		results = append(results, *v)
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
			tx = tx.Where("TRIM("+tableName+".department) IN ?", strs)
		}
	}
	if tableName != "capex_budget_fact_entities" && tableName != "capex_actual_fact_entities" {
		if val := utils.GetSafeString(filter, "year"); val != "" {
			tx = tx.Where(tableName+".year = ?", val)
		}
	}
	if tableName == "capex_budget_fact_entities" {
		if val := utils.GetSafeString(filter, "capex_file_id"); val != "" {
			tx = tx.Where(tableName+".file_capex_budget_id = ?", val)
		} else {
			var latestFid string
			r.db.Model(&models.CapexBudgetFactEntity{}).Order("created_at desc").Select("file_capex_budget_id").Limit(1).Take(&latestFid)
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
			r.db.Model(&models.CapexActualFactEntity{}).Order("created_at desc").Select("file_capex_actual_id").Limit(1).Take(&latestFid)
			if latestFid != "" {
				tx = tx.Where(tableName+".file_capex_actual_id = ?", latestFid)
			}
		}
	}
	return tx
}

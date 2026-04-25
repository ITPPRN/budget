package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
	"sort"

	_service "p2p-back-end/modules/budgets/service"

	"github.com/shopspring/decimal"
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
	// Inject Global Settings if missing — respects the actively selected capex files
	globalConfigs := _service.FetchGlobalSettings(r.db)
	if filter["year"] == nil || filter["year"] == "" {
		if val, ok := globalConfigs["actualYear"]; ok && val != "" {
			filter["year"] = val
		}
	}
	if filter["capex_file_id"] == nil || filter["capex_file_id"] == "" {
		if val, ok := globalConfigs["selectedCapexBg"]; ok && val != "" {
			filter["capex_file_id"] = val
		}
	}
	if filter["capex_actual_file_id"] == nil || filter["capex_actual_file_id"] == "" {
		if val, ok := globalConfigs["selectedCapexActual"]; ok && val != "" {
			filter["capex_actual_file_id"] = val
		}
	}

	type rawCapexRow struct {
		ID            string
		Entity        string
		Branch        string
		Department    string
		CapexNo       string
		CapexName     string
		CapexCategory string
		Month         string
		Amount        float64
	}

	var budgetResults []rawCapexRow
	btx := r.db.Table("capex_budget_fact_entities").
		Select(`
			capex_budget_fact_entities.id,
			capex_budget_fact_entities.entity,
			capex_budget_fact_entities.branch,
			capex_budget_fact_entities.department,
			capex_budget_fact_entities.capex_no,
			capex_budget_fact_entities.capex_name,
			capex_budget_fact_entities.capex_category,
			ba.month,
			ba.amount
		`).
		Joins("LEFT JOIN capex_budget_amount_entities ba ON ba.capex_budget_fact_id = capex_budget_fact_entities.id AND ba.deleted_at IS NULL").
		Where("COALESCE(capex_budget_fact_entities.capex_no, '') <> ''")
	btx = r.applyCommonFilters(btx, "capex_budget_fact_entities", filter)
	if err := btx.Scan(&budgetResults).Error; err != nil {
		return nil, err
	}

	var actualResults []rawCapexRow
	atx := r.db.Table("capex_actual_fact_entities").
		Select(`
			capex_actual_fact_entities.id,
			capex_actual_fact_entities.entity,
			capex_actual_fact_entities.branch,
			capex_actual_fact_entities.department,
			capex_actual_fact_entities.capex_no,
			capex_actual_fact_entities.capex_name,
			capex_actual_fact_entities.capex_category,
			aa.month,
			aa.amount
		`).
		Joins("LEFT JOIN capex_actual_amount_entities aa ON aa.capex_actual_fact_id = capex_actual_fact_entities.id AND aa.deleted_at IS NULL").
		Where("COALESCE(capex_actual_fact_entities.capex_no, '') <> ''")
	atx = r.applyCommonFilters(atx, "capex_actual_fact_entities", filter)
	if err := atx.Scan(&actualResults).Error; err != nil {
		return nil, err
	}

	// Group by fact ID so each uploaded record becomes one row — no cross-fact merging
	budgetMap := make(map[string]models.OwnerCapexBudgetExportDTO)
	actualMap := make(map[string]models.OwnerCapexBudgetExportDTO)

	process := func(data []rawCapexRow, targetMap map[string]models.OwnerCapexBudgetExportDTO, rowType string) {
		for _, res := range data {
			key := res.ID
			if row, ok := targetMap[key]; ok {
				if res.Month == "" {
					continue
				}
				amt := utils.ToDecimal(res.Amount)
				if existingAmt, ok := row.MonthsAmounts[res.Month].(decimal.Decimal); ok {
					row.MonthsAmounts[res.Month] = existingAmt.Add(amt)
				} else {
					row.MonthsAmounts[res.Month] = amt
				}
				row.YearTotal = row.YearTotal.Add(amt)
				targetMap[key] = row
			} else {
				dto := models.OwnerCapexBudgetExportDTO{
					Entity:        res.Entity,
					Branch:        res.Branch,
					Department:    res.Department,
					CapexNo:       res.CapexNo,
					CapexName:     res.CapexName,
					CapexCategory: res.CapexCategory,
					Type:          rowType,
					MonthsAmounts: map[string]interface{}{},
					YearTotal:     decimal.Zero,
				}
				if res.Month != "" {
					amt := utils.ToDecimal(res.Amount)
					dto.MonthsAmounts[res.Month] = amt
					dto.YearTotal = amt
				}
				targetMap[key] = dto
			}
		}
	}

	process(budgetResults, budgetMap, "Budget")
	process(actualResults, actualMap, "Actual")

	results := make([]models.OwnerCapexBudgetExportDTO, 0, len(budgetMap)+len(actualMap))
	for _, row := range budgetMap {
		results = append(results, row)
	}
	for _, row := range actualMap {
		results = append(results, row)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Entity != results[j].Entity {
			return results[i].Entity < results[j].Entity
		}
		if results[i].Branch != results[j].Branch {
			return results[i].Branch < results[j].Branch
		}
		if results[i].Department != results[j].Department {
			return results[i].Department < results[j].Department
		}
		if results[i].CapexNo != results[j].CapexNo {
			return results[i].CapexNo < results[j].CapexNo
		}
		return results[i].Type < results[j].Type
	})

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
			if err := r.db.Model(&models.CapexBudgetFactEntity{}).Order("created_at desc").Limit(1).Pluck("file_capex_budget_id", &latestFid).Error; err == nil && latestFid != "" {
				tx = tx.Where(tableName+".file_capex_budget_id = ?", latestFid)
			}
		}
	}
	if tableName == "capex_actual_fact_entities" {
		if val := utils.GetSafeString(filter, "capex_actual_file_id"); val != "" {
			tx = tx.Where(tableName+".file_capex_actual_id = ?", val)
		} else {
			var latestFid string
			if err := r.db.Model(&models.CapexActualFactEntity{}).Order("created_at desc").Limit(1).Pluck("file_capex_actual_id", &latestFid).Error; err == nil && latestFid != "" {
				tx = tx.Where(tableName+".file_capex_actual_id = ?", latestFid)
			}
		}
	}
	return tx
}

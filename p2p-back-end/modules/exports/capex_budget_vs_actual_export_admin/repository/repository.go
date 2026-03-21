package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
	"sort"

	"gorm.io/gorm"
)

type CapexVsActualRepository interface {
	GetCapexVsActualData(ctx context.Context, filter map[string]interface{}) ([]models.CapexVsActualExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) CapexVsActualRepository {
	return &repository{db: db}
}

func (r *repository) GetCapexVsActualData(ctx context.Context, filter map[string]interface{}) ([]models.CapexVsActualExportDTO, error) {
	type rawCapexRow struct {
		Entity        string
		Department    string
		CapexNo       string
		CapexName     string
		CapexCategory string
		Month         string
		Amount        float64
	}

	var budgetResults []rawCapexRow
	btx := r.db.Model(&models.CapexBudgetFactEntity{}).
		Select(`
			capex_budget_fact_entities.entity,
			capex_budget_fact_entities.department,
			capex_budget_fact_entities.capex_no,
			capex_budget_fact_entities.capex_name,
			capex_budget_fact_entities.capex_category,
			ba.month,
			ba.amount
		`).
		Joins("JOIN capex_budget_amount_entities ba ON ba.capex_budget_fact_id = capex_budget_fact_entities.id AND ba.deleted_at IS NULL")

	btx = r.applyCommonFilters(btx, "capex_budget_fact_entities", filter)
	if err := btx.Scan(&budgetResults).Error; err != nil {
		return nil, err
	}

	var actualResults []rawCapexRow
	atx := r.db.Model(&models.CapexActualFactEntity{}).
		Select(`
			capex_actual_fact_entities.entity,
			capex_actual_fact_entities.department,
			capex_actual_fact_entities.capex_no,
			capex_actual_fact_entities.capex_name,
			capex_actual_fact_entities.capex_category,
			aa.month,
			aa.amount
		`).
		Joins("JOIN capex_actual_amount_entities aa ON aa.capex_actual_fact_id = capex_actual_fact_entities.id AND aa.deleted_at IS NULL")

	atx = r.applyCommonFilters(atx, "capex_actual_fact_entities", filter)
	if err := atx.Scan(&actualResults).Error; err != nil {
		return nil, err
	}

	type rowKey struct {
		Entity     string
		Department string
		CapexNo    string
	}

	budgetMap := make(map[rowKey]models.CapexVsActualExportDTO)
	actualMap := make(map[rowKey]models.CapexVsActualExportDTO)

	process := func(data []rawCapexRow, targetMap map[rowKey]models.CapexVsActualExportDTO, rowType string) {
		for _, res := range data {
			key := rowKey{res.Entity, res.Department, res.CapexNo}
			if row, ok := targetMap[key]; ok {
				row.MonthsAmounts[res.Month] = utils.ToDecimal(res.Amount)
				row.YearTotal = row.YearTotal.Add(utils.ToDecimal(res.Amount))
				targetMap[key] = row
			} else {
				targetMap[key] = models.CapexVsActualExportDTO{
					Entity:        res.Entity,
					Department:    res.Department,
					CapexNo:       res.CapexNo,
					CapexName:     res.CapexName,
					CapexCategory: res.CapexCategory,
					Type:          rowType,
					MonthsAmounts: map[string]interface{}{res.Month: utils.ToDecimal(res.Amount)},
					YearTotal:     utils.ToDecimal(res.Amount),
				}
			}
		}
	}

	process(budgetResults, budgetMap, "Budget")
	process(actualResults, actualMap, "Actual")

	var finalResults []models.CapexVsActualExportDTO
	allKeysMap := make(map[rowKey]bool)
	for k := range budgetMap { allKeysMap[k] = true }
	for k := range actualMap { allKeysMap[k] = true }

	for k := range allKeysMap {
		if bRow, ok := budgetMap[k]; ok {
			finalResults = append(finalResults, bRow)
		} else {
			aRow := actualMap[k]
			finalResults = append(finalResults, models.CapexVsActualExportDTO{
				Entity: aRow.Entity, Department: aRow.Department, CapexNo: aRow.CapexNo,
				CapexName: aRow.CapexName, CapexCategory: aRow.CapexCategory, Type: "Budget",
				MonthsAmounts: make(map[string]interface{}),
			})
		}

		if aRow, ok := actualMap[k]; ok {
			finalResults = append(finalResults, aRow)
		} else {
			bRow := budgetMap[k]
			finalResults = append(finalResults, models.CapexVsActualExportDTO{
				Entity: bRow.Entity, Department: bRow.Department, CapexNo: bRow.CapexNo,
				CapexName: bRow.CapexName, CapexCategory: bRow.CapexCategory, Type: "Actual",
				MonthsAmounts: make(map[string]interface{}),
			})
		}
	}

	sort.Slice(finalResults, func(i, j int) bool {
		if finalResults[i].Entity != finalResults[j].Entity { return finalResults[i].Entity < finalResults[j].Entity }
		if finalResults[i].Department != finalResults[j].Department { return finalResults[i].Department < finalResults[j].Department }
		if finalResults[i].CapexNo != finalResults[j].CapexNo { return finalResults[i].CapexNo < finalResults[j].CapexNo }
		return finalResults[i].Type < finalResults[j].Type
	})

	return finalResults, nil
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

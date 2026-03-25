package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type BudgetVsActualRepository interface {
	GetBudgetVsActualData(ctx context.Context, filter map[string]interface{}) ([]models.BudgetVsActualExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) BudgetVsActualRepository {
	return &repository{db: db}
}

func (r *repository) GetBudgetVsActualData(ctx context.Context, filter map[string]interface{}) ([]models.BudgetVsActualExportDTO, error) {
	// 1. Fetch Budget Data
	var budgetResults []models.BudgetExportDTO
	btx := r.db.Model(&models.BudgetFactEntity{}).
		Select(`
			budget_fact_entities.entity,
			budget_fact_entities.branch,
			COALESCE(NULLIF(budget_fact_entities.department, ''), budget_fact_entities.nav_code) as department,
			bs.group1 as "group",
			bs.group2,
			bs.group3,
			budget_fact_entities.conso_gl,
			budget_fact_entities.gl_name,
			ba.month,
			ba.amount
		`).
		Joins("LEFT JOIN budget_structure_entities bs ON budget_fact_entities.conso_gl = bs.conso_gl").
		Joins("JOIN budget_amount_entities ba ON ba.budget_fact_id = budget_fact_entities.id AND ba.deleted_at IS NULL")

	btx = r.applyCommonFilters(btx, "budget_fact_entities", filter)
	if err := btx.Scan(&budgetResults).Error; err != nil {
		return nil, err
	}

	// 2. Fetch Actual Data
	var actualResults []models.BudgetExportDTO 
	atx := r.db.Model(&models.ActualFactEntity{}).
		Select(`
			actual_fact_entities.entity,
			actual_fact_entities.branch,
			COALESCE(NULLIF(actual_fact_entities.department, ''), actual_fact_entities.nav_code) as department,
			bs.group1 as "group",
			bs.group2,
			bs.group3,
			actual_fact_entities.conso_gl,
			actual_fact_entities.gl_name,
			aa.month,
			aa.amount
		`).
		Joins("LEFT JOIN budget_structure_entities bs ON actual_fact_entities.conso_gl = bs.conso_gl").
		Joins("JOIN actual_amount_entities aa ON aa.actual_fact_id = actual_fact_entities.id AND aa.deleted_at IS NULL")

	atx = r.applyCommonFilters(atx, "actual_fact_entities", filter)
	if err := atx.Scan(&actualResults).Error; err != nil {
		return nil, err
	}

	// 3. Process and Merge
	type rowKey struct {
		Entity     string
		Branch     string
		Department string
		ConsoGL    string
	}

	budgetMap := make(map[rowKey]models.BudgetVsActualExportDTO)
	actualMap := make(map[rowKey]models.BudgetVsActualExportDTO)

	process := func(data []models.BudgetExportDTO, targetMap map[rowKey]models.BudgetVsActualExportDTO, rowType string) {
		for _, res := range data {
			key := rowKey{res.Entity, res.Branch, res.Department, res.ConsoGL}
			if row, ok := targetMap[key]; ok {
				row.MonthsAmounts[res.Month] = res.Amount
				row.YearTotal = row.YearTotal.Add(res.Amount)
				targetMap[key] = row
			} else {
				targetMap[key] = models.BudgetVsActualExportDTO{
					Entity:        res.Entity,
					Branch:        res.Branch,
					Department:    res.Department,
					Type:          rowType,
					Group:         res.Group,
					Group2:        res.Group2,
					Group3:        res.Group3,
					ConsoGL:       res.ConsoGL,
					GLName:        res.GLName,
					MonthsAmounts: map[string]interface{}{res.Month: res.Amount},
					YearTotal:     res.Amount,
				}
			}
		}
	}

	process(budgetResults, budgetMap, "Budget")
	process(actualResults, actualMap, "Actual")

	var finalResults []models.BudgetVsActualExportDTO
	allKeysMap := make(map[rowKey]bool)
	for k := range budgetMap { allKeysMap[k] = true }
	for k := range actualMap { allKeysMap[k] = true }

	for k := range allKeysMap {
		if bRow, ok := budgetMap[k]; ok {
			finalResults = append(finalResults, bRow)
		} else {
			aRow := actualMap[k]
			finalResults = append(finalResults, models.BudgetVsActualExportDTO{
				Entity: aRow.Entity, Branch: aRow.Branch, Department: aRow.Department, Type: "Budget",
				Group: aRow.Group, Group2: aRow.Group2, Group3: aRow.Group3, ConsoGL: aRow.ConsoGL, GLName: aRow.GLName,
				MonthsAmounts: make(map[string]interface{}),
			})
		}

		if aRow, ok := actualMap[k]; ok {
			finalResults = append(finalResults, aRow)
		} else {
			bRow := budgetMap[k]
			finalResults = append(finalResults, models.BudgetVsActualExportDTO{
				Entity: bRow.Entity, Branch: bRow.Branch, Department: bRow.Department, Type: "Actual",
				Group: bRow.Group, Group2: bRow.Group2, Group3: bRow.Group3, ConsoGL: bRow.ConsoGL, GLName: bRow.GLName,
				MonthsAmounts: make(map[string]interface{}),
			})
		}
	}

	sort.Slice(finalResults, func(i, j int) bool {
		if finalResults[i].Entity != finalResults[j].Entity { return finalResults[i].Entity < finalResults[j].Entity }
		if finalResults[i].Branch != finalResults[j].Branch { return finalResults[i].Branch < finalResults[j].Branch }
		if finalResults[i].Department != finalResults[j].Department { return finalResults[i].Department < finalResults[j].Department }
		if finalResults[i].ConsoGL != finalResults[j].ConsoGL { return finalResults[i].ConsoGL < finalResults[j].ConsoGL }
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

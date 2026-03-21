package repository

import (
	"context"
	"fmt"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
	"sort"

	"gorm.io/gorm"
)

type BudgetExportRepository interface {
	GetBudgetExportDetails(ctx context.Context, filter map[string]interface{}) ([]models.BudgetExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) BudgetExportRepository {
	return &repository{db: db}
}

func (r *repository) GetBudgetExportDetails(ctx context.Context, filter map[string]interface{}) ([]models.BudgetExportDTO, error) {
	var results []models.BudgetExportDTO

	query := r.db.Model(&models.BudgetFactEntity{}).
		Select(`
			budget_fact_entities.entity,
			budget_fact_entities.branch,
			budget_fact_entities.department,
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

	// Apply Dynamic Filters (Normalized by Service)
	applyFilter := func(key string, dbCol string) {
		if val, ok := filter[key]; ok {
			strs := utils.ToStringSlice(val)
			if len(strs) > 0 {
				query = query.Where(fmt.Sprintf("budget_fact_entities.%s IN ?", dbCol), strs)
			}
		}
	}

	applyFilter("entities", "entity")
	applyFilter("branches", "branch")
	applyFilter("departments", "department")
	applyFilter("conso_gls", "conso_gl")



	if months, ok := filter["months"]; ok {
		mstrs := utils.ToStringSlice(months)
		if len(mstrs) > 0 {
			query = query.Where("ba.month IN ?", mstrs)
		}
	}

	if budgetFileID := utils.GetSafeString(filter, "budget_file_id"); budgetFileID != "" {
		query = query.Where("budget_fact_entities.file_budget_id = ?", budgetFileID)
	}

	err := query.Order("budget_fact_entities.entity, budget_fact_entities.branch, budget_fact_entities.department, budget_fact_entities.conso_gl").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Group results by key to calculate Year Total and ensure flat structure
	type rowKey struct {
		Entity     string
		Branch     string
		Department string
		ConsoGL    string
	}

	grouped := make(map[rowKey]models.BudgetExportDTO)
	for _, res := range results {
		key := rowKey{res.Entity, res.Branch, res.Department, res.ConsoGL}
		if row, ok := grouped[key]; ok {
			row.MonthsAmounts[res.Month] = res.Amount
			row.YearTotal = row.YearTotal.Add(res.Amount)
			grouped[key] = row
		} else {
			res.MonthsAmounts = map[string]interface{}{res.Month: res.Amount}
			res.YearTotal = res.Amount
			grouped[key] = res
		}
	}

	finalResults := make([]models.BudgetExportDTO, 0, len(grouped))
	for _, row := range grouped {
		finalResults = append(finalResults, row)
	}

	// Sort final results consistently
	sort.Slice(finalResults, func(i, j int) bool {
		if finalResults[i].Entity != finalResults[j].Entity {
			return finalResults[i].Entity < finalResults[j].Entity
		}
		if finalResults[i].Branch != finalResults[j].Branch {
			return finalResults[i].Branch < finalResults[j].Branch
		}
		if finalResults[i].Department != finalResults[j].Department {
			return finalResults[i].Department < finalResults[j].Department
		}
		return finalResults[i].ConsoGL < finalResults[j].ConsoGL
	})

	return finalResults, nil
}

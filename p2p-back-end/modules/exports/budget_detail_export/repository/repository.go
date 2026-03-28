package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
	"sort"
	"strings"

	_service "p2p-back-end/modules/budgets/service"

	"github.com/shopspring/decimal"
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
	// 🛠️ Inject Global Settings if missing
	globalConfigs := _service.FetchGlobalSettings(r.db)
	if filter["year"] == nil || filter["year"] == "" {
		if val, ok := globalConfigs["actualYear"]; ok && val != "" {
			filter["year"] = val
		}
	}
	if filter["budget_file_id"] == nil || filter["budget_file_id"] == "" {
		if val, ok := globalConfigs["selectedBudget"]; ok && val != "" {
			filter["budget_file_id"] = val
		}
	}
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
		Joins("LEFT JOIN (SELECT conso_gl, group1, group2, group3, MAX(account_name) as account_name FROM gl_grouping_entities GROUP BY conso_gl, group1, group2, group3) bs ON budget_fact_entities.conso_gl = bs.conso_gl").
		Joins("JOIN budget_amount_entities ba ON ba.budget_fact_id = budget_fact_entities.id AND ba.deleted_at IS NULL")

	// Apply Dynamic Filters (Normalized by Service)
	if val, ok := filter["entities"]; ok {
		if strs := utils.ToStringSlice(val); len(strs) > 0 {
			query = query.Where("budget_fact_entities.entity IN ?", strs)
		}
	}
	if val, ok := filter["branches"]; ok {
		if strs := utils.ToStringSlice(val); len(strs) > 0 {
			query = query.Where("budget_fact_entities.branch IN ?", strs)
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
					query = query.Where("(budget_fact_entities.department IN ? OR budget_fact_entities.department = '' OR budget_fact_entities.department IS NULL OR budget_fact_entities.department = 'None')", filteredStrs)
				} else {
					query = query.Where("(budget_fact_entities.department = '' OR budget_fact_entities.department IS NULL OR budget_fact_entities.department = 'None')")
				}
			} else {
				query = query.Where("budget_fact_entities.department IN ?", strs)
			}
		}
	}
	if val, ok := filter["conso_gls"]; ok {
		if strs := utils.ToStringSlice(val); len(strs) > 0 {
			query = query.Where("budget_fact_entities.conso_gl IN ?", strs)
		}
	}

	if months, ok := filter["months"]; ok {
		mstrs := utils.ToStringSlice(months)
		if len(mstrs) > 0 {
			query = query.Where("ba.month IN ?", mstrs)
		}
	}

	// Added: Year Filter
	targetYear := ""
	if val, ok := filter["year"]; ok {
		if s, ok := val.(string); ok && s != "" {
			targetYear = strings.ReplaceAll(s, "FY", "")
			// Apply to budget only if no explicit file ID.
		}
	}

	if val := utils.GetSafeString(filter, "budget_file_id"); val != "" {
		query = query.Where("budget_fact_entities.file_budget_id = ?", val)
	} else {
		// 🛠️ Fallback: Latest budget file ID for target year
		var latestFid string
		subQuery := r.db.Model(&models.BudgetFactEntity{})
		if targetYear != "" {
			subQuery = subQuery.Where("year = ?", targetYear)
			query = query.Where("budget_fact_entities.year = ?", targetYear)
		}
		if err := subQuery.Order("created_at desc").Limit(1).Pluck("file_budget_id", &latestFid).Error; err == nil && latestFid != "" {
			query = query.Where("budget_fact_entities.file_budget_id = ?", latestFid)
		}
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
		GLName     string
	}

	grouped := make(map[rowKey]models.BudgetExportDTO)
	for _, res := range results {
		key := rowKey{res.Entity, res.Branch, res.Department, res.ConsoGL, res.GLName}
		if row, ok := grouped[key]; ok {
			// Sum monthly amounts (fix overwrite bug)
			if existingAmt, ok := row.MonthsAmounts[res.Month].(decimal.Decimal); ok {
				row.MonthsAmounts[res.Month] = existingAmt.Add(res.Amount)
			} else {
				row.MonthsAmounts[res.Month] = res.Amount
			}
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

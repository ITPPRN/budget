package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
	"strings"

	_service "p2p-back-end/modules/budgets/service"

	"gorm.io/gorm"
)

type OwnerBudgetExportRepository interface {
	GetOwnerBudgetExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) OwnerBudgetExportRepository {
	return &repository{db: db}
}

// toStringSlice safely converts interface{} (from JSON filter map) to []string
func toStringSlice(val interface{}) []string {
	if strs, ok := val.([]string); ok {
		return strs
	}
	if interfaces, ok := val.([]interface{}); ok {
		var strs []string
		for _, item := range interfaces {
			if s, ok := item.(string); ok && s != "" {
				strs = append(strs, s)
			}
		}
		return strs
	}
	if s, ok := val.(string); ok && s != "" {
		return []string{s}
	}
	return nil
}

func (r *repository) GetOwnerBudgetExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetExportDTO, error) {
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

	query := r.db.WithContext(ctx).Model(&models.BudgetFactEntity{}).
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
		Joins(`LEFT JOIN (
			SELECT entity_gl, entity, group1, group2, group3 
			FROM gl_grouping_entities 
			GROUP BY entity_gl, entity, group1, group2, group3
		) bs ON budget_fact_entities.entity_gl = bs.entity_gl AND budget_fact_entities.entity = bs.entity`).
		Joins("LEFT JOIN budget_amount_entities ba ON ba.budget_fact_id = budget_fact_entities.id AND ba.deleted_at IS NULL")

	// 1. Entities
	if val, ok := filter["entities"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("budget_fact_entities.entity IN ?", strs)
		}
	}

	// 2. Branches
	if val, ok := filter["branches"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("budget_fact_entities.branch IN ?", strs)
		}
	}

	// 3. Departments (Robust Filtering consistent with Dashboard Bi-directional matching)
	if val, ok := filter["departments"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			if len(strs) == 1 && (strs[0] == "__RESTRICTED__" || strs[0] == "__RESTRICTED_EMPTY__" || strs[0] == "__RESTRICTED_NO_MATCH__") {
				query = query.Where("1 = 0")
			} else {
				var pConds []string
				var pVals []interface{}
				hasNone := false

				for _, s := range strs {
					if strings.EqualFold(strings.TrimSpace(s), "None") {
						hasNone = true
					} else if s != "" {
						dt := strings.ToUpper(strings.TrimSpace(s))
						// 🛡️ Match: Exactly Code, or Code as Prefix "Code - ...", or NavCode itself
						pConds = append(pConds, "UPPER(TRIM(budget_fact_entities.department)) = ? OR UPPER(TRIM(budget_fact_entities.department)) LIKE ? OR UPPER(TRIM(budget_fact_entities.department)) = ?")
						pVals = append(pVals, dt, dt+" - %", dt)
					}
				}

				if hasNone {
					if len(pConds) > 0 {
						query = query.Where("("+strings.Join(pConds, " OR ")+" OR COALESCE(TRIM(budget_fact_entities.department), '') = '' OR UPPER(TRIM(budget_fact_entities.department)) = 'NONE' OR budget_fact_entities.department IS NULL)", pVals...)
					} else {
						query = query.Where("(COALESCE(TRIM(budget_fact_entities.department), '') = '' OR UPPER(TRIM(budget_fact_entities.department)) = 'NONE' OR budget_fact_entities.department IS NULL)")
					}
				} else if len(pConds) > 0 {
					query = query.Where("("+strings.Join(pConds, " OR ")+")", pVals...)
				}
			}
		}
	}

	// 4. File Budget ID
	targetYear := ""
	if val, ok := filter["year"]; ok {
		if s, ok := val.(string); ok && s != "" {
			targetYear = strings.ReplaceAll(s, "FY", "")
		}
	}

	if fid := utils.GetSafeString(filter, "budget_file_id"); fid != "" {
		query = query.Where("budget_fact_entities.file_budget_id = ?", fid)
	} else if targetYear != "" {
		// 🛡️ Strict Fallback: Use only THE latest PL file for this year
		var latestFid string
		r.db.Model(&models.BudgetFactEntity{}).
			Where("year = ?", targetYear).
			Order("created_at DESC").
			Limit(1).
			Pluck("file_budget_id", &latestFid)

		if latestFid != "" {
			query = query.Where("budget_fact_entities.file_budget_id = ?", latestFid)
		} else {
			query = query.Where("1 = 0")
		}
		query = query.Where("budget_fact_entities.year = ?", targetYear)
	}

	// 5. Conso GLs
	if val, ok := filter["conso_gls"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("budget_fact_entities.conso_gl IN ?", strs)
		}
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	// 6. Pivot 12 Months (Same logic as admin)
	type rowKey struct {
		Entity     string
		Branch     string
		Department string
		ConsoGL    string
	}

	mergedMap := make(map[rowKey]models.BudgetExportDTO)

	for _, item := range results {
		key := rowKey{
			Entity:     item.Entity,
			Branch:     item.Branch,
			Department: item.Department,
			ConsoGL:    item.ConsoGL,
		} // Ensure 1:1 row parity with DB

		if existing, ok := mergedMap[key]; ok {
			if existing.MonthsAmounts == nil {
				existing.MonthsAmounts = make(map[string]interface{})
			}
			if item.Month != "" {
				existing.MonthsAmounts[item.Month] = item.Amount
				existing.YearTotal = existing.YearTotal.Add(item.Amount)
			}
			mergedMap[key] = existing
		} else {
			if item.MonthsAmounts == nil {
				item.MonthsAmounts = make(map[string]interface{})
			}
			if item.Month != "" {
				item.MonthsAmounts[item.Month] = item.Amount
				item.YearTotal = item.Amount
			}
			mergedMap[key] = item
		}
	}

	var finalResults []models.BudgetExportDTO
	for _, v := range mergedMap {
		finalResults = append(finalResults, v)
	}

	return finalResults, nil
}

package repository

import (
	"context"
	"fmt"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
	"sort"
	"strings"

	_service "p2p-back-end/modules/budgets/service"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BudgetExportRepository interface {
	GetBudgetExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) BudgetExportRepository {
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

func (r *repository) GetBudgetExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetExportDTO, error) {

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
			budget_fact_entities.id,
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

	// 3. Departments (Robust Filtering consistent with UI)
	if val, ok := filter["departments"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			hasNone := false
			var filteredStrs []string
			for _, s := range strs {
				if strings.EqualFold(strings.TrimSpace(s), "None") {
					hasNone = true
				} else if s != "" {
					filteredStrs = append(filteredStrs, s)
				}
			}

			tableName := "budget_fact_entities"
			condition := "(TRIM(COALESCE(NULLIF(" + tableName + ".department, ''), '')) IN ?)"
			if hasNone {
				if len(filteredStrs) > 0 {
					query = query.Where("("+condition+" OR TRIM(COALESCE(NULLIF("+tableName+".department, ''), '')) = '' OR TRIM(COALESCE(NULLIF("+tableName+".department, ''), '')) IS NULL OR TRIM(COALESCE(NULLIF("+tableName+".department, ''), '')) = 'None')", filteredStrs)
				} else {
					query = query.Where("(TRIM(COALESCE(NULLIF(" + tableName + ".department, ''), '')) = '' OR TRIM(COALESCE(NULLIF(" + tableName + ".department, ''), '')) IS NULL OR TRIM(COALESCE(NULLIF(" + tableName + ".department, ''), '')) = 'None')")
				}
			} else {
				query = query.Where(condition, strs)
			}
		}
	}

	// 4. Conso GLs
	if val, ok := filter["conso_gls"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("budget_fact_entities.conso_gl IN ?", strs)
		}
	}

	// 5. Months (Disabled for Budget Detail Export to ensure full year parity with source file)
	/*
	if months, ok := filter["months"]; ok {
		mstrs := toStringSlice(months)
		if len(mstrs) > 0 {
			query = query.Where("ba.month IN ?", mstrs)
		}
	}
	*/

	// 6. Year Filter (Strip FY)
	targetYear := ""
	if val, ok := filter["year"]; ok {
		if s, ok := val.(string); ok && s != "" {
			targetYear = strings.ReplaceAll(s, "FY", "")
		}
	}

	// 7. File Filter
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

	// Fetch matching results
	type intermediateResult struct {
		models.BudgetExportDTO
		BudgetFactID string          `gorm:"column:id"`
		Amount       decimal.Decimal `gorm:"column:amount"`
		Month        string          `gorm:"column:month"`
	}
	var rawData []intermediateResult
	if err := query.Scan(&rawData).Error; err != nil {
		return nil, fmt.Errorf("budgetExportRepo.GetBudgetExportDetails: %w", err)
	}

	// 🛠️ Pivot Data: Group by BudgetFactID
	budgetMap := make(map[string]*models.BudgetExportDTO)
	for _, raw := range rawData {
		key := raw.BudgetFactID
		if _, exists := budgetMap[key]; !exists {
			dto := &models.BudgetExportDTO{
				Entity:        raw.Entity,
				Branch:        raw.Branch,
				Department:    raw.Department,
				Group:         raw.Group,
				Group2:        raw.Group2,
				Group3:        raw.Group3,
				ConsoGL:       raw.ConsoGL,
				GLName:        raw.GLName,
				MonthsAmounts: make(map[string]interface{}),
				YearTotal:     decimal.Zero,
			}
			budgetMap[key] = dto
		}

		// Fill Monthly Amount
		if raw.Month != "" {
			budgetMap[key].MonthsAmounts[raw.Month] = raw.Amount
			budgetMap[key].YearTotal = budgetMap[key].YearTotal.Add(raw.Amount)
		}
	}

	// Convert map to slice
	results = make([]models.BudgetExportDTO, 0, len(budgetMap))
	for _, dto := range budgetMap {
		results = append(results, *dto)
	}

	// Sort results for consistency
	sort.Slice(results, func(i, j int) bool {
		if results[i].Entity != results[j].Entity { return results[i].Entity < results[j].Entity }
		if results[i].Branch != results[j].Branch { return results[i].Branch < results[j].Branch }
		if results[i].Department != results[j].Department { return results[i].Department < results[j].Department }
		return results[i].ConsoGL < results[j].ConsoGL
	})

	return results, nil
}

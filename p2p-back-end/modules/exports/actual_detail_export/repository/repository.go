package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"strings"

	"gorm.io/gorm"
)

type ActualExportRepository interface {
	GetActualExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) ActualExportRepository {
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

func (r *repository) GetActualExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualExportDTO, error) {
	var results []models.ActualExportDTO


	query := r.db.WithContext(ctx).Table("actual_transaction_entities").
		Select(`
			actual_transaction_entities.entity,
			actual_transaction_entities.branch,
			actual_transaction_entities.department,
			bs.group1 as "group",
			bs.group2,
			bs.group3,
			actual_transaction_entities.conso_gl,
			actual_transaction_entities.gl_account_name as gl_name,
			actual_transaction_entities.doc_no,
			actual_transaction_entities.amount,
			actual_transaction_entities.vendor_name,
			actual_transaction_entities.description,
			actual_transaction_entities.posting_date
		`).
		Joins(`LEFT JOIN (
			SELECT entity_gl, entity, group1, group2, group3 
			FROM gl_grouping_entities 
			GROUP BY entity_gl, entity, group1, group2, group3
		) bs ON actual_transaction_entities.entity_gl = bs.entity_gl AND actual_transaction_entities.entity = bs.entity`)

	// 1. Entities
	if val, ok := filter["entities"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("actual_transaction_entities.entity IN ?", strs)
		}
	}

	// 2. Branches
	if val, ok := filter["branches"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			query = query.Where("actual_transaction_entities.branch IN ?", strs)
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

			tableName := "actual_transaction_entities"
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
			query = query.Where("actual_transaction_entities.conso_gl IN ?", strs)
		}
	}

	// 5. Date Range
	if val, ok := filter["start_date"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.posting_date >= ?", val)
	}
	if val, ok := filter["end_date"].(string); ok && val != "" {
		query = query.Where("actual_transaction_entities.posting_date <= ?", val)
	}

	// 6. Year (Strip FY)
	if val, ok := filter["year"].(string); ok && val != "" && val != "All" {
		query = query.Where("actual_transaction_entities.year = ?", strings.ReplaceAll(val, "FY", ""))
	}

	// 7. Months
	if val, ok := filter["months"]; ok {
		mstrs := toStringSlice(val)
		if len(mstrs) > 0 {
			// Extract month from posting_date for comparison
			query = query.Where("UPPER(TO_CHAR(actual_transaction_entities.posting_date::DATE, 'MON')) IN ?", mstrs)
		}
	}

	err := query.Order("actual_transaction_entities.posting_date DESC, actual_transaction_entities.entity").
		Scan(&results).Error

	return results, err
}

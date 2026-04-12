package repository

import (
	"context"
	"p2p-back-end/modules/entities/models"
	"strings"

	"gorm.io/gorm"
)

type OwnerActualExportRepository interface {
	GetOwnerActualExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) OwnerActualExportRepository {
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

func (r *repository) GetOwnerActualExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualExportDTO, error) {
	var results []models.ActualExportDTO

	// 🛡️ Owner Permission Logic
	isAdmin := false
	for _, role := range user.Roles {
		if strings.EqualFold(strings.TrimSpace(role), "ADMIN") {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		allowedDepts := make([]string, 0)
		for _, p := range user.Permissions {
			if p.IsActive && p.DepartmentCode != "" {
				allowedDepts = append(allowedDepts, strings.TrimSpace(p.DepartmentCode))
			}
		}

		if len(allowedDepts) == 0 {
			// No permissions = No data
			filter["departments"] = []string{"__RESTRICTED_EMPTY__"}
		} else {
			// Intersect with requested departments
			if val, ok := filter["departments"]; ok {
				chosenDepts := toStringSlice(val)
				if len(chosenDepts) > 0 {
					var intersection []string
					for _, c := range chosenDepts {
						cTrim := strings.ToUpper(strings.TrimSpace(c))
						for _, a := range allowedDepts {
							aTrim := strings.ToUpper(strings.TrimSpace(a))
							// 🛡️ Case-Insensitive Bi-directional Matching: "acc" matches "ACC - Accounting"
							if cTrim == aTrim || strings.HasPrefix(cTrim, aTrim+" - ") || strings.HasPrefix(aTrim, cTrim+" - ") {
								intersection = append(intersection, c)
								break
							}
						}
					}
					if len(intersection) == 0 {
						filter["departments"] = []string{"__RESTRICTED_NO_MATCH__"}
					} else {
						filter["departments"] = intersection
					}
				} else {
					filter["departments"] = allowedDepts
				}
			} else {
				filter["departments"] = allowedDepts
			}
		}
	}

	query := r.db.WithContext(ctx).Table("actual_transaction_entities").
		Select(`
			actual_transaction_entities.entity,
			actual_transaction_entities.branch,
			actual_transaction_entities.department,
			mapping.group1 as "group",
			mapping.group2,
			mapping.group3,
			actual_transaction_entities.conso_gl as conso_gl,
			mapping.account_name as gl_name,
			actual_transaction_entities.doc_no as doc_no,
			actual_transaction_entities.amount,
			actual_transaction_entities.vendor_name,
			actual_transaction_entities.description,
			actual_transaction_entities.posting_date,
			CASE
				WHEN basket.id IS NOT NULL THEN 'In Basket'
				WHEN actual_transaction_entities.status = 'CONFIRMED' THEN 'Confirmed'
				WHEN actual_transaction_entities.status = 'COMPLETE' THEN 'Complete'
				WHEN actual_transaction_entities.status = 'REPORTED' THEN 'Reported'
				WHEN actual_transaction_entities.status = 'DRAFT' THEN 'Draft'
				ELSE 'Pending'
			END as status
		`).
		Joins("LEFT JOIN gl_grouping_entities mapping ON actual_transaction_entities.entity_gl = mapping.entity_gl AND actual_transaction_entities.entity = mapping.entity").
		Joins("LEFT JOIN audit_rejection_baskets basket ON basket.transaction_id = actual_transaction_entities.id AND basket.user_id::text = ?", user.ID)

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
						pConds = append(pConds, "UPPER(TRIM(actual_transaction_entities.department)) = ? OR UPPER(TRIM(actual_transaction_entities.department)) LIKE ? OR UPPER(TRIM(actual_transaction_entities.entity_gl)) = ?")
						pVals = append(pVals, dt, dt+" - %", dt)
					}
				}

				if hasNone {
					if len(pConds) > 0 {
						query = query.Where("("+strings.Join(pConds, " OR ")+" OR COALESCE(TRIM(actual_transaction_entities.department), '') = '' OR UPPER(TRIM(actual_transaction_entities.department)) = 'NONE' OR actual_transaction_entities.department IS NULL)", pVals...)
					} else {
						query = query.Where("(COALESCE(TRIM(actual_transaction_entities.department), '') = '' OR UPPER(TRIM(actual_transaction_entities.department)) = 'NONE' OR actual_transaction_entities.department IS NULL)")
					}
				} else if len(pConds) > 0 {
					query = query.Where("("+strings.Join(pConds, " OR ")+")", pVals...)
				}
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

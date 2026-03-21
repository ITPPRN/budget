package repository

import (
	"context"
	"fmt"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"

	"gorm.io/gorm"
)

type ActualExportRepository interface {
	GetActualExportDetails(ctx context.Context, filter map[string]interface{}) ([]models.ActualExportDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) ActualExportRepository {
	return &repository{db: db}
}

func (r *repository) GetActualExportDetails(ctx context.Context, filter map[string]interface{}) ([]models.ActualExportDTO, error) {
	var results []models.ActualExportDTO

	query := r.db.Model(&models.ActualTransactionEntity{}).
		Select(`
			actual_transaction_entities.entity,
			actual_transaction_entities.branch,
			actual_transaction_entities.department,
			bs.group1 as "group",
			bs.group2,
			bs.group3,
			actual_transaction_entities.conso_gl,
			bs.account_name as gl_name,
			actual_transaction_entities.doc_no,
			actual_transaction_entities.amount,
			actual_transaction_entities.vendor_name,
			actual_transaction_entities.description,
			actual_transaction_entities.posting_date
		`).
		Joins("LEFT JOIN budget_structure_entities bs ON actual_transaction_entities.conso_gl = bs.conso_gl")

	// Apply Dynamic Filters (Normalized by Service)
	applyFilter := func(key string, dbCol string) {
		if val, ok := filter[key]; ok {
			strs := utils.ToStringSlice(val)
			if len(strs) > 0 {
				query = query.Where(fmt.Sprintf("actual_transaction_entities.%s IN ?", dbCol), strs)
			}
		}
	}

	applyFilter("entities", "entity")
	applyFilter("branches", "branch")
	applyFilter("departments", "department")
	applyFilter("conso_gls", "conso_gl")

	if year := utils.GetSafeString(filter, "year"); year != "" {
		query = query.Where("actual_transaction_entities.year = ?", year)
	}

	err := query.Order("actual_transaction_entities.posting_date DESC, actual_transaction_entities.entity").
		Scan(&results).Error

	return results, err
}

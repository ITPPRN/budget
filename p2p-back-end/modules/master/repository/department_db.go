package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"p2p-back-end/modules/entities/models"
)

// --- Department ---

func (r *masterRepositoryDB) SyncDepartment(ctx context.Context, dept []models.Departments) ([]models.Departments, error) {
	if len(dept) == 0 {
		return nil, errors.New("masterRepo.SyncDepartment: no data Departments")
	}

	var allChangedRows []models.Departments
	batchSize := 100

	for i := 0; i < len(dept); i += batchSize {
		end := i + batchSize
		if end > len(dept) {
			end = len(dept)
		}

		currentBatch := dept[i:end]
		var changedInBatch []models.Departments

		err := r.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "central_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"name":       gorm.Expr("EXCLUDED.name"),
				"code":       gorm.Expr("EXCLUDED.code"),
				"updated_at": gorm.Expr("NOW()"),
			}),
			Where: clause.Where{
				Exprs: []clause.Expression{
					gorm.Expr(`
						departments.name IS DISTINCT FROM EXCLUDED.name OR
						departments.code IS DISTINCT FROM EXCLUDED.code
					`),
				},
			},
		}).
			Clauses(clause.Returning{}).
			Create(&currentBatch).
			Scan(&changedInBatch).
			Error
		if err != nil {
			return nil, err
		}

		allChangedRows = append(allChangedRows, changedInBatch...)
	}
	return allChangedRows, nil
}

func (r *masterRepositoryDB) GetDepartments(ctx context.Context, lastID uint, limit int) ([]models.Departments, error) {
	var departments []models.Departments
	err := r.db.WithContext(ctx).
		Where("central_id > ?", lastID).
		Order("central_id asc").
		Limit(limit).
		Find(&departments).Error
	if err != nil {
		return nil, fmt.Errorf("masterRepo.GetDepartments: %w", err)
	}
	return departments, nil
}

func (r *masterRepositoryDB) FindDeptUUID(ctx context.Context, centralID uint) (*uuid.UUID, error) {
	if centralID == 0 {
		return nil, nil
	}
	var dept models.Departments
	err := r.db.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("central_id = ?", centralID).First(&dept).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dept.ID, nil
}

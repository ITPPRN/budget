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

// --- Section ---

func (r *masterRepositoryDB) SyncSection(ctx context.Context, sec []models.Sections) ([]models.Sections, error) {
	if len(sec) == 0 {
		return nil, errors.New("masterRepo.SyncSection: no data section")
	}

	var allChangedRows []models.Sections
	batchSize := 100

	for i := 0; i < len(sec); i += batchSize {
		end := i + batchSize
		if end > len(sec) {
			end = len(sec)
		}

		currentBatch := sec[i:end]
		var changedInBatch []models.Sections

		err := r.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "central_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"name":       gorm.Expr("EXCLUDED.name"),
				"updated_at": gorm.Expr("NOW()"),
			}),
			Where: clause.Where{
				Exprs: []clause.Expression{
					gorm.Expr(`
						sections.name IS DISTINCT FROM EXCLUDED.name
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

func (r *masterRepositoryDB) GetSections(ctx context.Context, lastID uint, limit int) ([]models.Sections, error) {
	var sections []models.Sections
	err := r.db.WithContext(ctx).
		Where("central_id > ?", lastID).
		Order("central_id asc").
		Limit(limit).
		Find(&sections).Error
	if err != nil {
		return nil, fmt.Errorf("masterRepo.GetSections: %w", err)
	}
	return sections, nil
}

func (r *masterRepositoryDB) FindSectionUUID(ctx context.Context, centralID uint) (*uuid.UUID, error) {
	if centralID == 0 {
		return nil, nil
	}
	var sec models.Sections
	err := r.db.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("central_id = ?", centralID).First(&sec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sec.ID, nil
}

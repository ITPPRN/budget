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

// --- Position ---

func (r *masterRepositoryDB) SyncPosition(ctx context.Context, pos []models.Positions) ([]models.Positions, error) {
	if len(pos) == 0 {
		return nil, errors.New("masterRepo.SyncPosition: no data positions")
	}

	var allChangedRows []models.Positions
	batchSize := 100

	for i := 0; i < len(pos); i += batchSize {
		end := i + batchSize
		if end > len(pos) {
			end = len(pos)
		}

		currentBatch := pos[i:end]
		var changedInBatch []models.Positions

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
						positions.name IS DISTINCT FROM EXCLUDED.name OR
						positions.code IS DISTINCT FROM EXCLUDED.code
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

func (r *masterRepositoryDB) GetPositions(ctx context.Context, lastID uint, limit int) ([]models.Positions, error) {
	var positions []models.Positions
	err := r.db.WithContext(ctx).
		Where("central_id > ?", lastID).
		Order("central_id asc").
		Limit(limit).
		Find(&positions).Error
	if err != nil {
		return nil, fmt.Errorf("masterRepo.GetPositions: %w", err)
	}
	return positions, nil
}

func (r *masterRepositoryDB) FindPositionUUID(ctx context.Context, centralID uint) (*uuid.UUID, error) {
	if centralID == 0 {
		return nil, nil
	}
	var pos models.Positions
	err := r.db.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("central_id = ?", centralID).First(&pos).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pos.ID, nil
}

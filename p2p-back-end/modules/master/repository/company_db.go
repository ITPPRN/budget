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

// --- Company ---

func (r *masterRepositoryDB) SyncCompany(ctx context.Context, companies []models.Companies) ([]models.Companies, error) {
	if len(companies) == 0 {
		return nil, errors.New("masterRepo.SyncCompany: no data companies")
	}

	var allChangedRows []models.Companies
	batchSize := 100

	for i := 0; i < len(companies); i += batchSize {
		end := i + batchSize
		if end > len(companies) {
			end = len(companies)
		}

		currentBatch := companies[i:end]
		var changedInBatch []models.Companies

		err := r.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "central_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"name":           gorm.Expr("EXCLUDED.name"),
				"branch_name":    gorm.Expr("EXCLUDED.branch_name"),
				"branch_name_en": gorm.Expr("EXCLUDED.branch_name_en"),
				"branch_no":      gorm.Expr("EXCLUDED.branch_no"),
				"address":        gorm.Expr("EXCLUDED.address"),
				"taxid":          gorm.Expr("EXCLUDED.taxid"),
				"province":       gorm.Expr("EXCLUDED.province"),
				"updated_at":     gorm.Expr("NOW()"),
			}),
			Where: clause.Where{
				Exprs: []clause.Expression{
					gorm.Expr(`
						companies.name           IS DISTINCT FROM EXCLUDED.name OR
						companies.branch_name    IS DISTINCT FROM EXCLUDED.branch_name OR
						companies.branch_name_en IS DISTINCT FROM EXCLUDED.branch_name_en OR
						companies.branch_no 	 IS DISTINCT FROM EXCLUDED.branch_no OR
						companies.address 	     IS DISTINCT FROM EXCLUDED.address OR
						companies.taxid          IS DISTINCT FROM EXCLUDED.taxid OR
						companies.province       IS DISTINCT FROM EXCLUDED.province
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

func (r *masterRepositoryDB) GetCompanies(ctx context.Context, lastID uint, limit int) ([]models.Companies, error) {
	var companies []models.Companies
	err := r.db.WithContext(ctx).
		Where("central_id > ?", lastID).
		Order("central_id asc").
		Limit(limit).
		Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("masterRepo.GetCompanies: %w", err)
	}
	return companies, nil
}

func (r *masterRepositoryDB) FindCompanyUUID(ctx context.Context, centralID uint) (*uuid.UUID, error) {
	if centralID == 0 {
		return nil, nil
	}
	var comp models.Companies
	err := r.db.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("central_id = ?", centralID).First(&comp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &comp.ID, nil
}

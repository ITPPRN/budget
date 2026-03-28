package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

type masterDataRepository struct {
	db *gorm.DB
}

func NewMasterDataRepository(db *gorm.DB) models.MasterDataRepository {
	return &masterDataRepository{db: db}
}

func (r *masterDataRepository) WithTrx(trxHandle func(repo models.MasterDataRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewMasterDataRepository(tx)
		return trxHandle(repo)
	})
}

// --- User Config ---

func (r *masterDataRepository) GetUserConfigs(ctx context.Context, userID string) ([]models.UserConfigEntity, error) {
	var configs []models.UserConfigEntity
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("masterDataRepo.GetUserConfigs: %w", err)
	}
	return configs, nil
}

func (r *masterDataRepository) UpdateUserConfig(ctx context.Context, config *models.UserConfigEntity) error {
	var existing models.UserConfigEntity
	err := r.db.WithContext(ctx).Where("user_id = ? AND config_key = ?", config.UserID, config.ConfigKey).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
				return fmt.Errorf("masterDataRepo.UpdateUserConfig.Create: %w", err)
			}
			return nil
		}
		return fmt.Errorf("masterDataRepo.UpdateUserConfig.Find: %w", err)
	}
	existing.Value = config.Value
	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return fmt.Errorf("masterDataRepo.UpdateUserConfig.Save: %w", err)
	}
	return nil
}

// --- Unified GL Grouping ---

func (r *masterDataRepository) ListGLGroupings(ctx context.Context) ([]models.GlGroupingEntity, error) {
	var groupings []models.GlGroupingEntity
	if err := r.db.WithContext(ctx).Where("is_active = true").Find(&groupings).Error; err != nil {
		return nil, fmt.Errorf("masterDataRepo.ListGLGroupings: %w", err)
	}
	return groupings, nil
}

func (r *masterDataRepository) GetGLGroupingByID(ctx context.Context, id string) (*models.GlGroupingEntity, error) {
	var grouping models.GlGroupingEntity
	if err := r.db.WithContext(ctx).First(&grouping, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("masterDataRepo.GetGLGroupingByID: %w", err)
	}
	return &grouping, nil
}

func (r *masterDataRepository) CreateGLGrouping(ctx context.Context, g *models.GlGroupingEntity) error {
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return fmt.Errorf("masterDataRepo.CreateGLGrouping: %w", err)
	}
	return nil
}

func (r *masterDataRepository) UpdateGLGrouping(ctx context.Context, g *models.GlGroupingEntity) error {
	if err := r.db.WithContext(ctx).Save(g).Error; err != nil {
		return fmt.Errorf("masterDataRepo.UpdateGLGrouping: %w", err)
	}
	return nil
}

func (r *masterDataRepository) DeleteGLGrouping(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.GlGroupingEntity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("masterDataRepo.DeleteGLGrouping: %w", err)
	}
	return nil
}

func (r *masterDataRepository) GetGLGroupingInfo(ctx context.Context, entity string, entityGL string, target *models.GlGroupingEntity) error {
	if err := r.db.WithContext(ctx).Where("entity = ? AND entity_gl = ?", entity, entityGL).First(target).Error; err != nil {
		return fmt.Errorf("masterDataRepo.GetGLGroupingInfo: %w", err)
	}
	return nil
}

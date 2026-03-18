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

func (r *masterDataRepository) ListGLMappings(ctx context.Context) ([]models.GlMappingEntity, error) {
	var mappings []models.GlMappingEntity
	if err := r.db.WithContext(ctx).Where("is_active = true").Find(&mappings).Error; err != nil {
		return nil, fmt.Errorf("masterDataRepo.ListGLMappings: %w", err)
	}
	return mappings, nil
}

func (r *masterDataRepository) GetGLMappingByID(ctx context.Context, id string) (*models.GlMappingEntity, error) {
	var mapping models.GlMappingEntity
	if err := r.db.WithContext(ctx).First(&mapping, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("masterDataRepo.GetGLMappingByID: %w", err)
	}
	return &mapping, nil
}

func (r *masterDataRepository) CreateGLMapping(ctx context.Context, mapping *models.GlMappingEntity) error {
	if err := r.db.WithContext(ctx).Create(mapping).Error; err != nil {
		return fmt.Errorf("masterDataRepo.CreateGLMapping: %w", err)
	}
	return nil
}

func (r *masterDataRepository) UpdateGLMapping(ctx context.Context, mapping *models.GlMappingEntity) error {
	if err := r.db.WithContext(ctx).Save(mapping).Error; err != nil {
		return fmt.Errorf("masterDataRepo.UpdateGLMapping: %w", err)
	}
	return nil
}

func (r *masterDataRepository) DeleteGLMapping(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.GlMappingEntity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("masterDataRepo.DeleteGLMapping: %w", err)
	}
	return nil
}

func (r *masterDataRepository) GetGLInfo(ctx context.Context, entity string, entityGL string, target *models.GlMappingEntity) error {
	if err := r.db.WithContext(ctx).Where("entity = ? AND entity_gl = ?", entity, entityGL).First(target).Error; err != nil {
		return fmt.Errorf("masterDataRepo.GetGLInfo: %w", err)
	}
	return nil
}

func (r *masterDataRepository) CheckExactGLMapping(ctx context.Context, entity, entityGL, consoGL, accountName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.GlMappingEntity{}).
		Where("entity = ? AND entity_gl = ? AND conso_gl = ? AND account_name = ?", entity, entityGL, consoGL, accountName).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("masterDataRepo.CheckExactGLMapping: %w", err)
	}
	return count > 0, nil
}

func (r *masterDataRepository) GetBudgetStructure(ctx context.Context) ([]models.BudgetStructureEntity, error) {
	var entities []models.BudgetStructureEntity
	if err := r.db.WithContext(ctx).Order("group1 ASC, group2 ASC, group3 ASC").Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("masterDataRepo.GetBudgetStructure: %w", err)
	}
	return entities, nil
}

func (r *masterDataRepository) GetBudgetStructureByID(ctx context.Context, id uint) (*models.BudgetStructureEntity, error) {
	var entity models.BudgetStructureEntity
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, fmt.Errorf("masterDataRepo.GetBudgetStructureByID: %w", err)
	}
	return &entity, nil
}

func (r *masterDataRepository) CreateBudgetStructure(ctx context.Context, entity *models.BudgetStructureEntity) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("masterDataRepo.CreateBudgetStructure: %w", err)
	}
	return nil
}

func (r *masterDataRepository) UpdateBudgetStructure(ctx context.Context, entity *models.BudgetStructureEntity) error {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("masterDataRepo.UpdateBudgetStructure: %w", err)
	}
	return nil
}

func (r *masterDataRepository) DeleteBudgetStructure(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.BudgetStructureEntity{}, id).Error; err != nil {
		return fmt.Errorf("masterDataRepo.DeleteBudgetStructure: %w", err)
	}
	return nil
}

func (r *masterDataRepository) InsertBudgetStructures(ctx context.Context, entities []models.BudgetStructureEntity) error {
	if err := r.db.WithContext(ctx).Create(&entities).Error; err != nil {
		return fmt.Errorf("masterDataRepo.InsertBudgetStructures: %w", err)
	}
	return nil
}

func (r *masterDataRepository) DeleteAllBudgetStructures(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.BudgetStructureEntity{}).Error; err != nil {
		return fmt.Errorf("masterDataRepo.DeleteAllBudgetStructures: %w", err)
	}
	return nil
}

func (r *masterDataRepository) GetUserConfigs(ctx context.Context, userID string) ([]models.UserConfigEntity, error) {
	var configs []models.UserConfigEntity
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("masterDataRepo.GetUserConfigs: %w", err)
	}
	return configs, nil
}

func (r *masterDataRepository) UpdateUserConfig(ctx context.Context, config *models.UserConfigEntity) error {
	// Upsert based on UserID and ConfigKey
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



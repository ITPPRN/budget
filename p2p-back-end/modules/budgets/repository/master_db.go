package repository

import (
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

func (r *masterDataRepository) ListGLMappings() ([]models.GlMappingEntity, error) {
	var mappings []models.GlMappingEntity
	err := r.db.Where("is_active = true").Find(&mappings).Error
	return mappings, err
}

func (r *masterDataRepository) GetGLMappingByID(id string) (*models.GlMappingEntity, error) {
	var mapping models.GlMappingEntity
	if err := r.db.First(&mapping, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *masterDataRepository) CreateGLMapping(mapping *models.GlMappingEntity) error {
	return r.db.Create(mapping).Error
}

func (r *masterDataRepository) UpdateGLMapping(mapping *models.GlMappingEntity) error {
	return r.db.Save(mapping).Error
}

func (r *masterDataRepository) DeleteGLMapping(id string) error {
	return r.db.Delete(&models.GlMappingEntity{}, "id = ?", id).Error
}

func (r *masterDataRepository) GetGLInfo(entity string, entityGL string, target *models.GlMappingEntity) error {
	return r.db.Where("entity = ? AND entity_gl = ?", entity, entityGL).First(target).Error
}

func (r *masterDataRepository) CheckExactGLMapping(entity, entityGL, consoGL, accountName string) (bool, error) {
	var count int64
	err := r.db.Model(&models.GlMappingEntity{}).
		Where("entity = ? AND entity_gl = ? AND conso_gl = ? AND account_name = ?", entity, entityGL, consoGL, accountName).
		Count(&count).Error
	return count > 0, err
}

func (r *masterDataRepository) GetBudgetStructure() ([]models.BudgetStructureEntity, error) {
	var entities []models.BudgetStructureEntity
	if err := r.db.Order("group1 ASC, group2 ASC, group3 ASC").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *masterDataRepository) GetBudgetStructureByID(id uint) (*models.BudgetStructureEntity, error) {
	var entity models.BudgetStructureEntity
	if err := r.db.First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *masterDataRepository) CreateBudgetStructure(entity *models.BudgetStructureEntity) error {
	return r.db.Create(entity).Error
}

func (r *masterDataRepository) UpdateBudgetStructure(entity *models.BudgetStructureEntity) error {
	return r.db.Save(entity).Error
}

func (r *masterDataRepository) DeleteBudgetStructure(id uint) error {
	return r.db.Delete(&models.BudgetStructureEntity{}, id).Error
}

func (r *masterDataRepository) InsertBudgetStructures(entities []models.BudgetStructureEntity) error {
	return r.db.Create(&entities).Error
}

func (r *masterDataRepository) DeleteAllBudgetStructures() error {
	return r.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.BudgetStructureEntity{}).Error
}

func (r *masterDataRepository) GetUserConfigs(userID string) ([]models.UserConfigEntity, error) {
	var configs []models.UserConfigEntity
	if err := r.db.Where("user_id = ?", userID).Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *masterDataRepository) UpdateUserConfig(config *models.UserConfigEntity) error {
	// Upsert based on UserID and ConfigKey
	var existing models.UserConfigEntity
	err := r.db.Where("user_id = ? AND config_key = ?", config.UserID, config.ConfigKey).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(config).Error
		}
		return err
	}
	existing.Value = config.Value
	return r.db.Save(&existing).Error
}



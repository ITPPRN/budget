package repository

import (
	"p2p-back-end/modules/entities/models"

	"gorm.io/gorm"
)

type DepartmentRepositoryDB struct {
	db *gorm.DB
}

func NewDepartmentRepositoryDB(db *gorm.DB) DepartmentRepositoryDB {
	return DepartmentRepositoryDB{db: db}
}

func (r DepartmentRepositoryDB) GetDB() *gorm.DB {
	return r.db
}

// ClearMappings deletes all data from Mappings table
func (r DepartmentRepositoryDB) ClearMappings() error {
	if err := r.db.Exec("TRUNCATE TABLE department_mapping_entities CASCADE").Error; err != nil {
		if err := r.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.DepartmentMappingEntity{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// CreateDepartmentsBatch inserts multiple departments
func (r DepartmentRepositoryDB) CreateDepartmentsBatch(departments []models.DepartmentEntity) error {
	return r.db.Create(&departments).Error
}

// CreateMappingsBatch inserts multiple mappings
func (r DepartmentRepositoryDB) CreateMappingsBatch(mappings []models.DepartmentMappingEntity) error {
	return r.db.Create(&mappings).Error
}

// FindDepartmentByCode returns the master department by code
func (r DepartmentRepositoryDB) FindDepartmentByCode(code string) (*models.DepartmentEntity, error) {
	var dept models.DepartmentEntity
	err := r.db.Where("code = ?", code).Take(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dept, nil
}

// Helper to get ID map
func (r DepartmentRepositoryDB) GetDepartmentMap() (map[string]models.DepartmentEntity, error) {
	var depts []models.DepartmentEntity
	if err := r.db.Find(&depts).Error; err != nil {
		return nil, err
	}

	result := make(map[string]models.DepartmentEntity)
	for _, d := range depts {
		result[d.Code] = d
	}
	return result, nil
}

// FindMappingByNavCode finds the mapping for a given NAV code and Entity
func (r DepartmentRepositoryDB) FindMappingByNavCode(entity, navCode string) (*models.DepartmentMappingEntity, error) {
	var mapping models.DepartmentMappingEntity
	// Preload Department to get the Master Code
	// Use Take() or limit 1, and handle error gracefully to avoid log spam if frequent
	err := r.db.Preload("Department").
		Where("entity = ? AND nav_code = ?", entity, navCode).
		Take(&mapping).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, nil for not found (valid case)
		}
		return nil, err
	}
	return &mapping, nil
}

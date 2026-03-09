package repository

import (
	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

type sourceMasterRepositoryDB struct {
	db *gorm.DB
}

func NewSourceMasterRepositoryDB(db *gorm.DB) models.SourceMasterRepository {
	return &sourceMasterRepositoryDB{db: db}
}

func (r *sourceMasterRepositoryDB) GetCompanies(lastID uint, limit int) ([]models.CentralCompany, error) {
	var companies []models.CentralCompany
	err := r.db.Where("company_id > ?", lastID).
		Order("company_id ASC").
		Limit(limit).
		Find(&companies).Error
	return companies, err
}

func (r *sourceMasterRepositoryDB) GetDepartments(lastID uint, limit int) ([]models.CentralDepartment, error) {
	var departments []models.CentralDepartment
	err := r.db.Where("department_id > ?", lastID).
		Order("department_id ASC").
		Limit(limit).
		Find(&departments).Error
	return departments, err
}

func (r *sourceMasterRepositoryDB) GetSections(lasID uint, limit int) ([]models.CentralSection, error) {
	var sections []models.CentralSection
	err := r.db.Where("section_id > ?", lasID).
		Order("section_id ASC").
		Limit(limit).
		Find(&sections).Error
	return sections, err
}

func (r *sourceMasterRepositoryDB) GetPositions(lastID uint, limit int) ([]models.CentralPosition, error) {
	var positions []models.CentralPosition
	err := r.db.Where("position_id > ?", lastID).
		Order("position_id ASC").
		Limit(limit).
		Find(&positions).Error
	return positions, err
}

package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"p2p-back-end/modules/entities/models"
)

type masterRepositoryDB struct {
	db *gorm.DB
}

func NewMasterRepositoryDB(db *gorm.DB) models.MasterRepository {
	return &masterRepositoryDB{db: db}
}

// --- Company ---

func (r *masterRepositoryDB) SyncCompany(companies []models.Companies) ([]models.Companies, error) {
	if len(companies) == 0 {
		return nil, errors.New("no data companies")
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
			Columns: []clause.Column{{Name: "id"}},
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

func (r *masterRepositoryDB) GetCompanies(lastID uint, limit int) ([]models.Companies, error) {
	var companies []models.Companies
	err := r.db.
		Where("id > ?", lastID).
		Order("id asc").
		Limit(limit).
		Find(&companies).Error
	return companies, err
}

// --- Department ---

func (r *masterRepositoryDB) SyncDepartment(dept []models.Departments) ([]models.Departments, error) {
	if len(dept) == 0 {
		return nil, errors.New("no data Departments")
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
			Columns: []clause.Column{{Name: "id"}},
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

func (r *masterRepositoryDB) GetDepartments(lastID uint, limit int) ([]models.Departments, error) {
	var departments []models.Departments
	err := r.db.
		Where("id > ?", lastID).
		Order("id asc").
		Limit(limit).
		Find(&departments).Error
	return departments, err
}

// --- Section ---

func (r *masterRepositoryDB) SyncSection(sec []models.Sections) ([]models.Sections, error) {
	if len(sec) == 0 {
		return nil, errors.New("no data section")
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
			Columns: []clause.Column{{Name: "id"}},
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

func (r *masterRepositoryDB) GetSections(lastID uint, limit int) ([]models.Sections, error) {
	var sections []models.Sections
	err := r.db.
		Where("id > ?", lastID).
		Order("id asc").
		Limit(limit).
		Find(&sections).Error
	return sections, err
}

// --- Position ---

func (r *masterRepositoryDB) SyncPosition(pos []models.Positions) ([]models.Positions, error) {
	if len(pos) == 0 {
		return nil, errors.New("no data positions")
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
			Columns: []clause.Column{{Name: "id"}},
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

func (r *masterRepositoryDB) GetPositions(lastID uint, limit int) ([]models.Positions, error) {
	var positions []models.Positions
	err := r.db.
		Where("id > ?", lastID).
		Order("id asc").
		Limit(limit).
		Find(&positions).Error
	return positions, err
}

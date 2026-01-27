package repository

import (
	"p2p-back-end/modules/entities/models"

	"gorm.io/gorm"
)

type budgetRepositoryDB struct {
	db *gorm.DB
}

func NewBudgetRepositoryDB(db *gorm.DB) models.BudgetRepository {
	return &budgetRepositoryDB{db: db}
}

// Transaction helper
func (r *budgetRepositoryDB) WithTrx(trxHandle func(repo models.BudgetRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewBudgetRepositoryDB(tx)
		return trxHandle(repo)
	})
}

// ---------------------------------------------------------
// File Create Methods
// ---------------------------------------------------------

func (r *budgetRepositoryDB) CreateFileBudget(file *models.FileBudgetEntity) error {
	return r.db.Create(file).Error
}

func (r *budgetRepositoryDB) CreateFileCapexBudget(file *models.FileCapexBudgetEntity) error {
	return r.db.Create(file).Error
}

func (r *budgetRepositoryDB) CreateFileCapexActual(file *models.FileCapexActualEntity) error {
	return r.db.Create(file).Error
}

// ---------------------------------------------------------
// Fact Create Methods (Batch Insert + Association)
// ---------------------------------------------------------

// 1. Budget (PL)
func (r *budgetRepositoryDB) CreateBudgetFacts(headers []models.BudgetFactEntity) error {
	// GORM CreateInBatches ไม่บันทึก Association (Amounts) โดยอัตโนมัติ
	// เราต้องแยกบันทึก Header และ Amount เองเพื่อประสิทธิภาพ 100%

	// 1.1 Insert Headers
	if err := r.db.Omit("BudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	// 1.2 Collect All Amounts
	var allAmounts []models.BudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.BudgetAmounts...)
	}

	// 1.3 Insert Amounts
	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// 2. Capex Budget
func (r *budgetRepositoryDB) CreateCapexBudgetFacts(headers []models.CapexBudgetFactEntity) error {
	if err := r.db.Omit("CapexBudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	var allAmounts []models.CapexBudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.CapexBudgetAmounts...)
	}

	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// 3. Capex Actual
func (r *budgetRepositoryDB) CreateCapexActualFacts(headers []models.CapexActualFactEntity) error {
	if err := r.db.Omit("CapexActualAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	var allAmounts []models.CapexActualAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.CapexActualAmounts...)
	}

	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// ---------------------------------------------------------
// File List Methods
// ---------------------------------------------------------

func (r *budgetRepositoryDB) ListFileBudgets() ([]models.FileBudgetEntity, error) {
	var files []models.FileBudgetEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *budgetRepositoryDB) ListFileCapexBudgets() ([]models.FileCapexBudgetEntity, error) {
	var files []models.FileCapexBudgetEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *budgetRepositoryDB) ListFileCapexActuals() ([]models.FileCapexActualEntity, error) {
	var files []models.FileCapexActualEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

// ---------------------------------------------------------
// File Delete Methods
// ---------------------------------------------------------

func (r *budgetRepositoryDB) DeleteFileBudget(id string) error {
	return r.db.Delete(&models.FileBudgetEntity{}, "id = ?", id).Error
}

func (r *budgetRepositoryDB) DeleteFileCapexBudget(id string) error {
	return r.db.Delete(&models.FileCapexBudgetEntity{}, "id = ?", id).Error
}

func (r *budgetRepositoryDB) DeleteFileCapexActual(id string) error {
	return r.db.Delete(&models.FileCapexActualEntity{}, "id = ?", id).Error
}

// ---------------------------------------------------------------------
// 4. Delete All Facts (For Sync)
// ---------------------------------------------------------------------
// ---------------------------------------------------------------------
// 4. Delete All Facts (For Sync)
// ---------------------------------------------------------------------
func (r *budgetRepositoryDB) DeleteAllBudgetFacts() error {
	// 1. Delete Amounts (Children)
	if err := r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.BudgetAmountEntity{}).Error; err != nil {
		return err
	}
	// 2. Delete Headers (Parents)
	return r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.BudgetFactEntity{}).Error
}

func (r *budgetRepositoryDB) DeleteAllCapexBudgetFacts() error {
	// 1. Delete Amounts
	if err := r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.CapexBudgetAmountEntity{}).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.CapexBudgetFactEntity{}).Error
}

func (r *budgetRepositoryDB) DeleteAllCapexActualFacts() error {
	// 1. Delete Amounts
	if err := r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.CapexActualAmountEntity{}).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.CapexActualFactEntity{}).Error
}

// ---------------------------------------------------------------------
// 5. Update Files (Rename) - Implementation
// ---------------------------------------------------------------------

func (r *budgetRepositoryDB) UpdateFileBudget(id string, filename string) error {
	return r.db.Model(&models.FileBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

func (r *budgetRepositoryDB) UpdateFileCapexBudget(id string, filename string) error {
	return r.db.Model(&models.FileCapexBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

func (r *budgetRepositoryDB) UpdateFileCapexActual(id string, filename string) error {
	return r.db.Model(&models.FileCapexActualEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

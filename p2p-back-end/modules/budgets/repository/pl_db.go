package repository

import (
	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

type plBudgetRepository struct {
	db *gorm.DB
}

func NewPLBudgetRepository(db *gorm.DB) models.PLBudgetRepository {
	return &plBudgetRepository{db: db}
}

func (r *plBudgetRepository) WithTrx(trxHandle func(repo models.PLBudgetRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewPLBudgetRepository(tx)
		return trxHandle(repo)
	})
}

func (r *plBudgetRepository) CreateFileBudget(file *models.FileBudgetEntity) error {
	return r.db.Create(file).Error
}

func (r *plBudgetRepository) CreateBudgetFacts(headers []models.BudgetFactEntity) error {
	// GORM CreateInBatches ไม่บันทึก Association (Amounts) โดยอัตโนมัติ
	// เราต้องแยกบันทึก Header และ Amount เองเพื่อประสิทธิภาพ 100%

	// 1.1 บันทึกส่วนหัว (Headers)
	if err := r.db.Omit("BudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	// 1.2 รวบรวมข้อมูลยอดเงินทั้งหมด (Amounts)
	var allAmounts []models.BudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.BudgetAmounts...)
	}

	// 1.3 บันทึกยอดเงิน (Insert Amounts)
	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

func (r *plBudgetRepository) ListFileBudgets() ([]models.FileBudgetEntity, error) {
	var files []models.FileBudgetEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *plBudgetRepository) GetFileBudget(id string) (*models.FileBudgetEntity, error) {
	var file models.FileBudgetEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *plBudgetRepository) DeleteFileBudget(id string) error {
	return r.db.Delete(&models.FileBudgetEntity{}, "id = ?", id).Error
}

func (r *plBudgetRepository) DeleteAllBudgetFacts() error {
	// 1. ลบยอดเงิน (ลูก)
	// Use Raw SQL to guarantee execution and avoid Soft Delete confusion
	if err := r.db.Exec("DELETE FROM budget_amount_entities").Error; err != nil {
		return err
	}
	// 2. ลบส่วนหัว (แม่)
	return r.db.Exec("DELETE FROM budget_fact_entities").Error
}

func (r *plBudgetRepository) DeleteBudgetFactsByFileID(fileID string) error {
	// 1. Delete Amounts (Subquery or Join)
	if err := r.db.Exec(`
		DELETE FROM budget_amount_entities 
		WHERE budget_fact_id IN (SELECT id FROM budget_fact_entities WHERE file_budget_id = ?)
	`, fileID).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Where("file_budget_id = ?", fileID).Delete(&models.BudgetFactEntity{}).Error
}

func (r *plBudgetRepository) UpdateFileBudget(id string, filename string) error {
	return r.db.Model(&models.FileBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

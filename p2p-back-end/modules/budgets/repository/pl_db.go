package repository

import (
	"context"
	"fmt"

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

func (r *plBudgetRepository) CreateFileBudget(ctx context.Context, file *models.FileBudgetEntity) error {
	if err := r.db.WithContext(ctx).Create(file).Error; err != nil {
		return fmt.Errorf("plRepo.CreateFileBudget: %w", err)
	}
	return nil
}

func (r *plBudgetRepository) CreateBudgetFacts(ctx context.Context, headers []models.BudgetFactEntity) error {
	// GORM CreateInBatches ไม่บันทึก Association (Amounts) โดยอัตโนมัติ
	// เราต้องแยกบันทึก Header และ Amount เองเพื่อประสิทธิภาพ 100%

	// 1.1 บันทึกส่วนหัว (Headers)
	if err := r.db.WithContext(ctx).Omit("BudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return fmt.Errorf("plRepo.CreateBudgetFacts.Headers: %w", err)
	}

	// 1.2 รวบรวมข้อมูลยอดเงินทั้งหมด (Amounts)
	var allAmounts []models.BudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.BudgetAmounts...)
	}

	// 1.3 บันทึกยอดเงิน (Insert Amounts)
	if len(allAmounts) > 0 {
		if err := r.db.WithContext(ctx).CreateInBatches(&allAmounts, 1000).Error; err != nil {
			return fmt.Errorf("plRepo.CreateBudgetFacts.Amounts: %w", err)
		}
	}
	return nil
}

func (r *plBudgetRepository) ListFileBudgets(ctx context.Context) ([]models.FileBudgetEntity, error) {
	var files []models.FileBudgetEntity
	err := r.db.WithContext(ctx).Order("upload_at desc").Find(&files).Error
	if err != nil {
		return nil, fmt.Errorf("plRepo.ListFileBudgets: %w", err)
	}
	return files, nil
}

func (r *plBudgetRepository) GetFileBudget(ctx context.Context, id string) (*models.FileBudgetEntity, error) {
	var file models.FileBudgetEntity
	if err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("plRepo.GetFileBudget: %w", err)
	}
	return &file, nil
}

func (r *plBudgetRepository) DeleteFileBudget(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.FileBudgetEntity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("plRepo.DeleteFileBudget: %w", err)
	}
	return nil
}

func (r *plBudgetRepository) DeleteAllBudgetFacts(ctx context.Context) error {
	// 1. ลบยอดเงิน (ลูก)
	// Use Raw SQL to guarantee execution and avoid Soft Delete confusion
	if err := r.db.WithContext(ctx).Exec("DELETE FROM budget_amount_entities").Error; err != nil {
		return fmt.Errorf("plRepo.DeleteAllBudgetFacts.Amounts: %w", err)
	}
	// 2. ลบส่วนหัว (แม่)
	if err := r.db.WithContext(ctx).Exec("DELETE FROM budget_fact_entities").Error; err != nil {
		return fmt.Errorf("plRepo.DeleteAllBudgetFacts.Headers: %w", err)
	}
	return nil
}

func (r *plBudgetRepository) DeleteBudgetFactsByFileID(ctx context.Context, fileID string) error {
	// 1. Delete Amounts (Subquery or Join)
	if err := r.db.WithContext(ctx).Exec(`
		DELETE FROM budget_amount_entities 
		WHERE budget_fact_id IN (SELECT id FROM budget_fact_entities WHERE file_budget_id = ?)
	`, fileID).Error; err != nil {
		return fmt.Errorf("plRepo.DeleteBudgetFactsByFileID.Amounts: %w", err)
	}
	// 2. Delete Headers
	if err := r.db.WithContext(ctx).Unscoped().Where("file_budget_id = ?", fileID).Delete(&models.BudgetFactEntity{}).Error; err != nil {
		return fmt.Errorf("plRepo.DeleteBudgetFactsByFileID.Headers: %w", err)
	}
	return nil
}

func (r *plBudgetRepository) UpdateFileBudget(ctx context.Context, id string, filename string) error {
	if err := r.db.WithContext(ctx).Model(&models.FileBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error; err != nil {
		return fmt.Errorf("plRepo.UpdateFileBudget: %w", err)
	}
	return nil
}

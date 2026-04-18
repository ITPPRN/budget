package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"p2p-back-end/modules/entities/models"
)

type companyBranchCodeMappingRepository struct {
	db *gorm.DB
}

func NewCompanyBranchCodeMappingRepository(db *gorm.DB) models.CompanyBranchCodeMappingRepository {
	return &companyBranchCodeMappingRepository{db: db}
}

func (r *companyBranchCodeMappingRepository) List(ctx context.Context) ([]models.CompanyBranchCodeMappingEntity, error) {
	var rows []models.CompanyBranchCodeMappingEntity
	if err := r.db.WithContext(ctx).
		Preload("Company").
		Order("company_id ASC, branch_code ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("branchCodeMappingRepo.List: %w", err)
	}
	return rows, nil
}

// ListByCompanyID returns ALL mappings (codes) for a single company.
// Empty slice (not nil) when no mapping is configured.
func (r *companyBranchCodeMappingRepository) ListByCompanyID(ctx context.Context, companyID uuid.UUID) ([]models.CompanyBranchCodeMappingEntity, error) {
	rows := []models.CompanyBranchCodeMappingEntity{}
	if err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("branch_code ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("branchCodeMappingRepo.ListByCompanyID: %w", err)
	}
	return rows, nil
}

// Upsert is idempotent on (company_id, branch_code). Inserting an already-mapped
// pair updates only updated_at; existing rows for OTHER codes of the same
// company are untouched.
func (r *companyBranchCodeMappingRepository) Upsert(ctx context.Context, m *models.CompanyBranchCodeMappingEntity) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "company_id"}, {Name: "branch_code"}},
			DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
		}).
		Create(m).Error
	if err != nil {
		return fmt.Errorf("branchCodeMappingRepo.Upsert: %w", err)
	}
	return nil
}

func (r *companyBranchCodeMappingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.CompanyBranchCodeMappingEntity{}).Error; err != nil {
		return fmt.Errorf("branchCodeMappingRepo.Delete: %w", err)
	}
	return nil
}

func (r *companyBranchCodeMappingRepository) ListCompanies(ctx context.Context) ([]models.Companies, error) {
	var rows []models.Companies
	if err := r.db.WithContext(ctx).
		Order("name ASC, branch_no ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("branchCodeMappingRepo.ListCompanies: %w", err)
	}
	return rows, nil
}

// ListAvailableBranchCodes returns the distinct branch codes that have appeared
// in actual_transactions — the source of truth for what codes the data layer uses.
func (r *companyBranchCodeMappingRepository) ListAvailableBranchCodes(ctx context.Context) ([]string, error) {
	var codes []string
	if err := r.db.WithContext(ctx).
		Table("actual_transaction_entities").
		Distinct("branch").
		Where("branch IS NOT NULL AND branch <> ''").
		Order("branch ASC").
		Pluck("branch", &codes).Error; err != nil {
		return nil, fmt.Errorf("branchCodeMappingRepo.ListAvailableBranchCodes: %w", err)
	}
	return codes, nil
}

func (r *companyBranchCodeMappingRepository) FindCompanyByNameAndBranchNo(ctx context.Context, name, branchNo string) (*models.Companies, error) {
	var c models.Companies
	err := r.db.WithContext(ctx).
		Where("TRIM(name) = ? AND TRIM(branch_no) = ?", strings.TrimSpace(name), strings.TrimSpace(branchNo)).
		Take(&c).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("branchCodeMappingRepo.FindCompanyByNameAndBranchNo: %w", err)
	}
	return &c, nil
}

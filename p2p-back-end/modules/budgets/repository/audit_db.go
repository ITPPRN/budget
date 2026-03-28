package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) models.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) SaveAuditLog(ctx context.Context, log *models.AuditLogEntity) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("auditRepo.SaveAuditLog: %w", err)
	}
	return nil
}

func (r *auditRepository) SaveRejectedItems(ctx context.Context, items []models.AuditLogRejectedItemEntity) error {
	if len(items) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).CreateInBatches(items, 500).Error; err != nil {
		return fmt.Errorf("auditRepo.SaveRejectedItems: %w", err)
	}
	return nil
}

func (r *auditRepository) GetAuditLogs(ctx context.Context, filter map[string]interface{}) ([]models.AuditLogEntity, error) {
	var logs []models.AuditLogEntity
	query := r.db.WithContext(ctx).Order("created_at DESC")

	if dept, ok := filter["department"].(string); ok && dept != "" {
		query = query.Where("department = ?", dept)
	}
	if year, ok := filter["year"].(string); ok && year != "" {
		query = query.Where("year = ?", year)
	}
	if month, ok := filter["month"].(string); ok && month != "" {
		query = query.Where("month = ?", month)
	}
	if entity, ok := filter["entity"].(string); ok && entity != "" {
		query = query.Where("entity = ?", entity)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("auditRepo.GetAuditLogs: %w", err)
	}
	return logs, nil
}

func (r *auditRepository) GetRejectedItemsByLogID(ctx context.Context, logID string) ([]models.AuditLogRejectedItemEntity, error) {
	var items []models.AuditLogRejectedItemEntity
	if err := r.db.WithContext(ctx).Where("audit_log_id = ?", logID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("auditRepo.GetRejectedItemsByLogID: %w", err)
	}
	return items, nil
}

func (r *auditRepository) GetTransactionsByIDs(ctx context.Context, ids []uuid.UUID) ([]models.ActualTransactionEntity, error) {
	var txs []models.ActualTransactionEntity
	if len(ids) == 0 {
		return txs, nil
	}
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&txs).Error; err != nil {
		return nil, fmt.Errorf("auditRepo.GetTransactionsByIDs: %w", err)
	}
	return txs, nil
}

func (r *auditRepository) GetTransactionsByFilter(ctx context.Context, filter map[string]interface{}) ([]models.ActualTransactionEntity, error) {
	var txs []models.ActualTransactionEntity
	query := r.db.WithContext(ctx)

	if depts, ok := filter["departments"].([]string); ok && len(depts) > 0 {
		query = query.Where("department IN ?", depts)
	}
	if consoGLs, ok := filter["conso_gls"].([]string); ok && len(consoGLs) > 0 {
		query = query.Where("conso_gl IN ?", consoGLs)
	}
	if startDate, ok := filter["start_date"].(string); ok && startDate != "" {
		query = query.Where("posting_date >= ?", startDate)
	}
	if endDate, ok := filter["end_date"].(string); ok && endDate != "" {
		query = query.Where("posting_date <= ?", endDate)
	}
	if year, ok := filter["year"].(string); ok && year != "" {
		query = query.Where("year = ?", year)
	}
	if months, ok := filter["months"].([]string); ok && len(months) > 0 {
		// Only apply LIKE month if start/end dates are not provided
		if filter["start_date"] == "" && filter["end_date"] == "" {
			query = query.Where("posting_date LIKE ?", fmt.Sprintf("%s-%s-%%", filter["year"], months[0]))
		}
	}

	if search, ok := filter["search"].(string); ok && search != "" {
		pattern := "%" + search + "%"
		query = query.Where("(doc_no LIKE ? OR description LIKE ? OR conso_gl LIKE ?)", pattern, pattern, pattern)
	}

	// Handle limit
	if limit, ok := filter["limit"].(int); ok && limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&txs).Error; err != nil {
		return nil, fmt.Errorf("auditRepo.GetTransactionsByFilter: %w", err)
	}
	return txs, nil
}

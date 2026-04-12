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

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) models.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) WithTrx(trxHandle func(repo models.AuditRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewAuditRepository(tx)
		return trxHandle(repo)
	})
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

	// if search, ok := filter["search"].(string); ok && search != "" {
	// 	pattern := "%" + search + "%"
	// 	query = query.Where("(doc_no ILIKE ? OR description ILIKE ? OR conso_gl ILIKE ?)", pattern, pattern, pattern)
	// }

	searchList, hasList := filter["search_list"].([]string)
	searchStr, hasStr := filter["search"].(string)

	if hasList && len(searchList) > 0 {
		// กรณีพิมพ์หลายเลข (มีลูกน้ำคั่น) -> ค้นหาแบบตรงตัวจากใน Array
		// ใช้ doc_no IN ('00-JV2601-0035', '07-JV2601-0025')
		query = query.Where("doc_no IN ?", searchList)

	} else if hasStr && searchStr != "" {
		// กรณีพิมพ์คำเดียว -> ค้นหาแบบ ILIKE กวาดทุกคอลัมน์เหมือนเดิม
		pattern := "%" + searchStr + "%"
		query = query.Where("(doc_no ILIKE ? OR description ILIKE ? OR conso_gl ILIKE ?)", pattern, pattern, pattern)
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

func (r *auditRepository) UpdateTransactionsStatus(ctx context.Context, ids []uuid.UUID, status string) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
		Where("id IN ?", ids).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("auditRepo.UpdateTransactionsStatus: %w", err)
	}
	return nil
}

func (r *auditRepository) MarkRestAsComplete(ctx context.Context, department, year, month string, excludedIDs []uuid.UUID, targetStatus string) error {
    datePattern := fmt.Sprintf("%s-%s-%%", year, month)
    deptUpper := strings.ToUpper(strings.TrimSpace(department))

    query := r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
        Where("(UPPER(TRIM(department)) = ? OR UPPER(TRIM(department)) LIKE ?)", deptUpper, deptUpper+" - %").
        Where("year = ? AND posting_date LIKE ?", year, datePattern)

    if len(excludedIDs) > 0 {
        query = query.Where("id NOT IN ?", excludedIDs)
    }

    if err := query.Where("status IN ?", []string{models.TxStatusPending, models.TxStatusDraft}).
        Update("status", targetStatus).Error; err != nil {
        return fmt.Errorf("auditRepo.MarkRestAsComplete: %w", err)
    }
    return nil
}

// func (r *auditRepository) MarkRestAsComplete(ctx context.Context, department, year, month string, excludedIDs []uuid.UUID) error {
// 	query := r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
// 		Where("department = ? AND year = ? AND posting_date LIKE ?",
// 			department, year, fmt.Sprintf("%s-%s-%%", year, month))

// 	if len(excludedIDs) > 0 {
// 		query = query.Where("id NOT IN ?", excludedIDs)
// 	}

// 	// Only mark PENDING or DRAFT ones as COMPLETE
// 	if err := query.Where("status IN ?", []string{models.TxStatusPending, models.TxStatusDraft}).
// 		Update("status", models.TxStatusComplete).Error; err != nil {
// 		return fmt.Errorf("auditRepo.MarkRestAsComplete: %w", err)
// 	}
// 	return nil
// }

func (r *auditRepository) AddToBasket(ctx context.Context, items []models.AuditRejectBasket) error {
    if len(items) == 0 {
        return nil
    }
    const batchSize = 1000

    return r.db.WithContext(ctx).
        Clauses(clause.OnConflict{
            Columns: []clause.Column{
                {Name: "transaction_id"}, 
                {Name: "user_id"},
            },
            DoNothing: true,
        }).
        CreateInBatches(&items, batchSize).Error
}

// func (r *auditRepository) AddToBasket(ctx context.Context, items []models.AuditRejectBasket) error {
// 	if len(items) == 0 {
// 		return nil
// 	}
// 	const batchSize = 1000

// 	return r.db.WithContext(ctx).
// 		Clauses(clause.OnConflict{DoNothing: true}).
// 		CreateInBatches(&items, batchSize).Error
// }

func (r *auditRepository) GetBasketItems(ctx context.Context, userID string) ([]models.ActualTransactionEntity, error) {
	var items []models.ActualTransactionEntity

	err := r.db.WithContext(ctx).
		Table("actual_transaction_entities AS a").
		Select("a.*").
		Joins("INNER JOIN audit_rejection_baskets AS b ON a.id = b.transaction_id").
		Where("b.user_id = ?", userID).
		Find(&items).Error

	return items, err
}

func (r *auditRepository) RemoveFromBasket(ctx context.Context, userID string, transactionID string) error {
    if transactionID == "" {
        return nil
    }

    // ลบเฉพาะรายการที่ตรงกับ UserID และ TransactionID นั้นๆ
    return r.db.WithContext(ctx).
        Where("user_id = ? AND transaction_id = ?", userID, transactionID).
        Delete(&models.AuditRejectBasket{}).Error
}


func (r *auditRepository) GetBasketTransactionIDs(ctx context.Context, userID string) ([]uuid.UUID, error) {
    var ids []uuid.UUID
    err := r.db.WithContext(ctx).Model(&models.AuditRejectBasket{}).
        Where("user_id = ?", userID).
        Pluck("transaction_id", &ids).Error // Pluck จะดึงมาแค่คอลัมน์เดียวเป็น Array ให้เลย
    return ids, err
}

func (r *auditRepository) ClearBasket(ctx context.Context, userID string) error {
    return r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Delete(&models.AuditRejectBasket{}).Error
}


// func (r *auditRepository) GetActivePeriodsFromBasket(ctx context.Context, userID uuid.UUID) ([]models.YearMonth, []uuid.UUID, error) {
//     var basketIDs []uuid.UUID
//     var periods []models.YearMonth

//     // 1. ดึง ID ทั้งหมดในตะกร้า
//     err := r.db.WithContext(ctx).Model(&models.AuditRejectBasket{}).
//         Where("user_id = ?", userID).
//         Pluck("transaction_id", &basketIDs).Error
//     if err != nil || len(basketIDs) == 0 {
//         return nil, nil, err
//     }

//     // 2. หาว่าในตะกร้ามีเดือนไหนบ้าง (สกัดจาก posting_date)
//     // SQL: SELECT DISTINCT year, SUBSTRING(posting_date, 6, 2) as month ...
//     err = r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
//         Select("year, SUBSTRING(posting_date, 6, 2) as month").
//         Where("id IN ?", basketIDs).
//         Group("year, month").
//         Scan(&periods).Error

//     return periods, basketIDs, err
// }


func (r *auditRepository) ConfirmMonthTransactions(ctx context.Context, department, year, month string, excludedIDs []uuid.UUID) error {
	// สร้าง Pattern สำหรับหาเดือน เช่น "2026-04-%"
	datePattern := fmt.Sprintf("%s-%s-%%", year, month)

	query := r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
		Where("year = ? AND posting_date LIKE ?", year, datePattern)

	// ถ้าส่งชื่อแผนกมา ก็ให้เจาะจงแผนกด้วย (bi-directional match)
	if department != "" {
		deptUpper := strings.ToUpper(strings.TrimSpace(department))
		query = query.Where("(UPPER(TRIM(department)) = ? OR UPPER(TRIM(department)) LIKE ?)", deptUpper, deptUpper+" - %")
	}

	// 🌟 เว้นรายการที่อยู่ในตะกร้าไว้ ไม่ต้องยืนยัน (NOT IN)
	if len(excludedIDs) > 0 {
		query = query.Where("id NOT IN ?", excludedIDs)
	}

	// อัปเดตเฉพาะรายการที่ยัง Pending หรือ Draft ให้เป็น CONFIRMED
	if err := query.Where("status IN ?", []string{"PENDING", "DRAFT"}).
		Update("status", "CONFIRMED").Error; err != nil {
		return err
	}
	
	return nil
}

// CountPendingByDepartments นับจำนวน transaction ที่ยัง PENDING/DRAFT สำหรับ departments ในเดือน/ปีที่ระบุ
func (r *auditRepository) CountPendingByDepartments(ctx context.Context, year, month string, departments []string) (int64, error) {
	if len(departments) == 0 {
		return 0, nil
	}

	datePattern := fmt.Sprintf("%s-%s-%%", year, month)

	// สร้าง condition สำหรับ bi-directional department matching
	var deptConds []string
	var deptVals []interface{}
	for _, dept := range departments {
		deptUpper := strings.ToUpper(strings.TrimSpace(dept))
		deptConds = append(deptConds, "(UPPER(TRIM(department)) = ? OR UPPER(TRIM(department)) LIKE ?)")
		deptVals = append(deptVals, deptUpper, deptUpper+" - %")
	}

	var count int64
	query := r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
		Where("year = ? AND posting_date LIKE ?", year, datePattern).
		Where("status IN ?", []string{models.TxStatusPending, models.TxStatusDraft}).
		Where(strings.Join(deptConds, " OR "), deptVals...)

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("auditRepo.CountPendingByDepartments: %w", err)
	}
	return count, nil
}

// ตรวจสอบว่าในตะกร้ามีบิลของเดือนอื่นที่ไม่ใช่เดือนที่กำลังกด Approve หรือไม่
func (r *auditRepository) ValidateBasketScope(ctx context.Context, ids []uuid.UUID, year string, month string) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}

	var crossMonthCount int64
	datePattern := fmt.Sprintf("%s-%s-%%", year, month)

	// ค้นหาบิลในตะกร้าที่ "ไม่ใช่ปีที่เลือก" หรือ "ไม่ใช่เดือนที่เลือก"
	err := r.db.WithContext(ctx).Model(&models.ActualTransactionEntity{}).
		Where("id IN ?", ids).
		Where("year != ? OR posting_date NOT LIKE ?", year, datePattern).
		Count(&crossMonthCount).Error

	if err != nil {
		return false, err
	}

	// ถ้า count > 0 แปลว่ามีบิลแปลกปลอมข้ามเดือนหลุดมา (return false เพื่อบอกว่าไม่ผ่าน)
	if crossMonthCount > 0 {
		return false, nil
	}

	return true, nil
}
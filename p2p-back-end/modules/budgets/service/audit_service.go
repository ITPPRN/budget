package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/errs"
)

type auditService struct {
	auditRepo     models.AuditRepository
	dashboardRepo models.DashboardRepository
	userRepo      models.UserRepository
}

func NewAuditService(
	auditRepo models.AuditRepository,
	dashboardRepo models.DashboardRepository,
	userRepo models.UserRepository,
) models.AuditService {
	return &auditService{
		auditRepo:     auditRepo,
		dashboardRepo: dashboardRepo,
		userRepo:      userRepo,
	}
}

func (s *auditService) Approve(ctx context.Context, user *models.UserInfo, payload map[string]interface{}) error {
	department, _ := payload["department"].(string)
	var targets []string

	// ==========================================
	// 1. ตรวจสอบสิทธิ์และหาเป้าหมาย
	// ==========================================
	if department != "" {
		if err := s.checkOwnerPermission(ctx, user.ID, department); err != nil {
			return err
		}
		targets = append(targets, department)
	} else {
		perms, err := s.userRepo.GetUserPermissions(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch user permissions: %w", err)
		}
		for _, p := range perms {
			if p.Role == "OWNER" {
				targets = append(targets, p.DepartmentCode)
			}
		}
	}

	if len(targets) == 0 {
		return fmt.Errorf("no authorized departments found to approve")
	}

	// ==========================================
	// 2. เตรียมข้อมูลสำหรับ Query
	// ==========================================
	year, _ := payload["year"].(string)
	month, _ := payload["month"].(string)
	entity, _ := payload["entity"].(string)
	branch, _ := payload["branch"].(string)

	// ==========================================
	// 3. เริ่มกระบวนการ Database Transaction
	// ==========================================
	return s.auditRepo.WithTrx(func(trxRepo models.AuditRepository) error {
		
		// 🌟 ดึงข้อมูล Items ใน "Reject Basket" 
		rejectedTxIDs, err := trxRepo.GetBasketTransactionIDs(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch basket items: %w", err)
		}

		// 🛡️ ด่านตรวจความปลอดภัย: ป้องกันตะกร้ามีบิลข้ามเดือน (Anti-Human Error)
		if len(rejectedTxIDs) > 0 {
			isValid, err := trxRepo.ValidateBasketScope(ctx, rejectedTxIDs, year, month)
			if err != nil {
				return fmt.Errorf("failed to validate basket contents: %w", err)
			}
			if !isValid {
				return fmt.Errorf("ไม่อนุญาตให้ทำรายการ: มีรายการของเดือนอื่นค้างอยู่ในตะกร้า กรุณาเคลียร์ตะกร้าให้เรียบร้อยก่อนยืนยันเดือนนี้")
			}
		}

		// 🌟 จัดการรายการที่ถูก Reject
		var rejectedTxByDept map[string][]models.ActualTransactionEntity
		if len(rejectedTxIDs) > 0 {
			rejectedTxs, err := trxRepo.GetTransactionsByIDs(ctx, rejectedTxIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch basket transaction details: %w", err)
			}
			rejectedTxByDept = make(map[string][]models.ActualTransactionEntity, len(rejectedTxs))
			for _, tx := range rejectedTxs {
				rejectedTxByDept[tx.Department] = append(rejectedTxByDept[tx.Department], tx)
			}

			if err := trxRepo.UpdateTransactionsStatus(ctx, rejectedTxIDs, models.TxStatusReported); err != nil {
				return fmt.Errorf("failed to report basket items: %w", err)
			}
		}

		// 🌟 Auto-Complete: ยืนยันรายการที่เหลือ (ที่ไม่ได้อยู่ในตะกร้า)
		for _, dept := range targets {
			if err := trxRepo.MarkRestAsComplete(ctx, dept, year, month, rejectedTxIDs, models.TxStatusComplete); err != nil {
				return fmt.Errorf("failed to auto-complete remaining items: %w", err)
			}
		}

		// 🌟 สร้าง Audit Log บันทึกประวัติการทำรายการ
		for _, dept := range targets {
			deptRejected := rejectedTxByDept[dept]
			status := "CONFIRMED"
			if len(deptRejected) > 0 {
				status = "REPORTED"
			}
			log := &models.AuditLogEntity{
				ID:            uuid.New(),
				Entity:        entity,
				Branch:        branch,
				Department:    dept,
				Year:          year,
				Month:         month,
				Status:        status,
				RejectedCount: len(deptRejected),
				CreatedBy:     user.Name,
			}

			if err := trxRepo.SaveAuditLog(ctx, log); err != nil {
				return fmt.Errorf("failed to save approval log for dept %s: %w", dept, err)
			}

			if len(deptRejected) > 0 {
				items := make([]models.AuditLogRejectedItemEntity, 0, len(deptRejected))
				for _, tx := range deptRejected {
					items = append(items, models.AuditLogRejectedItemEntity{
						ID:            uuid.New(),
						AuditLogID:    log.ID,
						TransactionID: tx.ID,
						ConsoGL:       tx.ConsoGL,
						GLAccountName: tx.GLAccountName,
						Amount:        tx.Amount,
						Vendor:        tx.VendorName,
						DocNo:         tx.DocNo,
						Description:   tx.Description,
						PostingDate:   tx.PostingDate,
					})
				}
				if err := trxRepo.SaveRejectedItems(ctx, items); err != nil {
					return fmt.Errorf("failed to save rejected items for dept %s: %w", dept, err)
				}
			}
		}

		// 🌟 ล้างตะกร้าให้เกลี้ยง
		if err := trxRepo.ClearBasket(ctx, user.ID); err != nil {
			return fmt.Errorf("failed to clear user basket: %w", err)
		}

		return nil
	})
}

// func (s *auditService) Approve(ctx context.Context, user *models.UserInfo, payload map[string]interface{}) error {
// 	department, _ := payload["department"].(string)
// 	var targets []string

// 	if department != "" {
// 		if err := s.checkOwnerPermission(ctx, user.ID, department); err != nil {
// 			return err
// 		}
// 		targets = append(targets, department)
// 	} else {
// 		// Identify ALL authorized departments for this owner
// 		perms, err := s.userRepo.GetUserPermissions(ctx, user.ID)
// 		if err != nil {
// 			return fmt.Errorf("failed to fetch user permissions: %w", err)
// 		}
// 		for _, p := range perms {
// 			if p.Role == "OWNER" {
// 				targets = append(targets, p.DepartmentCode)
// 			}
// 		}
// 	}

// 	if len(targets) == 0 {
// 		return fmt.Errorf("no authorized departments found to approve")
// 	}

// 	year, _ := payload["year"].(string)
// 	month, _ := payload["month"].(string)
// 	entity, _ := payload["entity"].(string)
// 	branch, _ := payload["branch"].(string)
// 	rejectedIDs, _ := payload["rejected_item_ids"].([]interface{})

// 	// Wrap all DB writes in a single transaction
// 	return s.auditRepo.WithTrx(func(trxRepo models.AuditRepository) error {
// 		// Handle Composite Approval:
// 		// 1. Items in "Reject Basket" -> Status = REPORTED (Send to Admin)
// 		// 2. All other items in scope -> Status = COMPLETE (Closed/Finished)

// 		var rejectedTxIDs []uuid.UUID
// 		if len(rejectedIDs) > 0 {
// 			for _, idStr := range rejectedIDs {
// 				id, err := uuid.Parse(fmt.Sprintf("%v", idStr))
// 				if err == nil {
// 					rejectedTxIDs = append(rejectedTxIDs, id)
// 				}
// 			}

// 			// Mark items in "Reject Basket" as REPORTED
// 			if err := trxRepo.UpdateTransactionsStatus(ctx, rejectedTxIDs, models.TxStatusReported); err != nil {
// 				return fmt.Errorf("failed to report basket items: %w", err)
// 			}
// 		}

// 		// Auto-Complete: Mark UN-SELECTED (Non-Basket) items in same scope as COMPLETE
// 		for _, dept := range targets {
// 			if err := trxRepo.MarkRestAsComplete(ctx, dept, year, month, rejectedTxIDs); err != nil {
// 				return fmt.Errorf("failed to auto-complete remaining items: %w", err)
// 			}
// 		}

// 		for _, dept := range targets {
// 			log := &models.AuditLogEntity{
// 				ID:            uuid.New(),
// 				Entity:        entity,
// 				Branch:        branch,
// 				Department:    dept,
// 				Year:          year,
// 				Month:         month,
// 				Status:        "CONFIRMED",
// 				RejectedCount: len(rejectedIDs),
// 				CreatedBy:     user.Name,
// 			}

// 			if err := trxRepo.SaveAuditLog(ctx, log); err != nil {
// 				return fmt.Errorf("failed to save approval log for dept %s: %w", dept, err)
// 			}
// 		}

// 		return nil
// 	})
// }

func (s *auditService) Report(ctx context.Context, user *models.UserInfo, payload map[string]interface{}) error {
	entity, _ := payload["entity"].(string)
	branch, _ := payload["branch"].(string)
	year, _ := payload["year"].(string)
	month, _ := payload["month"].(string)
	rejectedIDs, ok := payload["rejected_item_ids"].([]interface{})
	if !ok {
		return fmt.Errorf("auditService.Report: rejected_item_ids is required")
	}

	var txIDs []uuid.UUID
	for _, idStr := range rejectedIDs {
		id, err := uuid.Parse(fmt.Sprintf("%v", idStr))
		if err == nil {
			txIDs = append(txIDs, id)
		}
	}

	// Fetch full transactions for snapshot and grouping
	txs, err := s.auditRepo.GetTransactionsByIDs(ctx, txIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions for snapshot: %w", err)
	}

	// Group transactions by Department
	deptGroups := make(map[string][]models.ActualTransactionEntity)
	for _, tx := range txs {
		deptGroups[tx.Department] = append(deptGroups[tx.Department], tx)
	}

	// For each department group, create a separate Audit Log
	for dept, items := range deptGroups {
		// Verify permission for each department being reported
		if err := s.checkOwnerPermission(ctx, user.ID, dept); err != nil {
			return fmt.Errorf("permission denied for department %s: %w", dept, err)
		}

		logID := uuid.New()
		log := &models.AuditLogEntity{
			ID:            logID,
			Entity:        entity,
			Branch:        branch,
			Department:    dept,
			Year:          year,
			Month:         month,
			Status:        "REJECTED",
			RejectedCount: len(items),
			CreatedBy:     user.Name,
		}

		if err := s.auditRepo.SaveAuditLog(ctx, log); err != nil {
			return fmt.Errorf("failed to save log for dept %s: %w", dept, err)
		}

		var rejectedItems []models.AuditLogRejectedItemEntity
		var deptTxIDs []uuid.UUID
		for _, tx := range items {
			deptTxIDs = append(deptTxIDs, tx.ID)
			rejectedItems = append(rejectedItems, models.AuditLogRejectedItemEntity{
				ID:            uuid.New(),
				AuditLogID:    logID,
				TransactionID: tx.ID,
				ConsoGL:       tx.ConsoGL,
				GLAccountName: tx.GLAccountName,
				Amount:        tx.Amount,
				Vendor:        tx.VendorName,
				DocNo:         tx.DocNo,
				Description:   tx.Description,
				PostingDate:   tx.PostingDate,
			})
		}

		if err := s.auditRepo.SaveRejectedItems(ctx, rejectedItems); err != nil {
			return fmt.Errorf("failed to save items for dept %s: %w", dept, err)
		}

		// Update only this department's transaction IDs to REPORTED
		if err := s.auditRepo.UpdateTransactionsStatus(ctx, deptTxIDs, models.TxStatusReported); err != nil {
			return fmt.Errorf("failed to mark items as reported for dept %s: %w", dept, err)
		}
	}

	return nil
}

func (s *auditService) ListLogs(ctx context.Context, filter map[string]interface{}) ([]models.AuditLogEntity, error) {
	return s.auditRepo.GetAuditLogs(ctx, filter)
}

func (s *auditService) GetRejectedItemDetails(ctx context.Context, logID string) ([]models.AuditLogRejectedItemEntity, error) {
	return s.auditRepo.GetRejectedItemsByLogID(ctx, logID)
}

func (s *auditService) GetReportableTransactions(ctx context.Context, user *models.UserInfo, payload map[string]interface{}) ([]models.ActualTransactionEntity, error) {
	department, _ := payload["department"].(string)
	var targets []string

	if department != "" {
		if err := s.checkOwnerPermission(ctx, user.ID, department); err != nil {
			return nil, err
		}
		targets = append(targets, department)
	} else {
		// Identify ALL authorized departments for this owner
		perms, err := s.userRepo.GetUserPermissions(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user permissions: %w", err)
		}
		for _, p := range perms {
			if p.Role == "OWNER" {
				targets = append(targets, p.DepartmentCode)
			}
		}
		if len(targets) == 0 {
			return nil, fmt.Errorf("no authorized departments found for your account")
		}
	}

	// Fetch all transactions (no pagination)
	consoGLs := []string{}
	if cgls, ok := payload["conso_gls"].([]interface{}); ok {
		for _, v := range cgls {
			consoGLs = append(consoGLs, fmt.Sprintf("%v", v))
		}
	}

	var searchList []string
	if sl, ok := payload["search_list"].([]interface{}); ok {
		for _, v := range sl {
			searchList = append(searchList, fmt.Sprintf("%v", v))
		}
	}

	filter := map[string]interface{}{
		"departments": targets,
		"year":        payload["year"],
		"months":      []string{fmt.Sprintf("%v", payload["month"])},
		"conso_gls":   consoGLs,
		"start_date":  payload["start_date"],
		"end_date":    payload["end_date"],
		"search":      payload["search"],
		"search_list": searchList,
		"limit":       -1,
	}

	return s.auditRepo.GetTransactionsByFilter(ctx, filter)
}

func (s *auditService) checkOwnerPermission(ctx context.Context, userID, department string) error {
	// Must have OWNER role in System
	// (Already checked by controller usually, but we double check for Dept access)

	perms, err := s.userRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to verify permissions: %w", err)
	}

	for _, p := range perms {
		if p.DepartmentCode == department && p.Role == "OWNER" {
			return nil
		}
	}
	return fmt.Errorf("permission Denied: You are not the owner of department %s", department)
}

func (s *auditService) AddToBasket(ctx context.Context, user *models.UserInfo, transactionIDs []string) error {

	if user == nil || len(transactionIDs) == 0 {
		return errors.New("invalid input: user ID and transaction IDs cannot be empty")
	}

	uid, err := uuid.Parse(user.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	var parsedIDs []uuid.UUID
	for _, txIDStr := range transactionIDs {
		if txID, err := uuid.Parse(txIDStr); err == nil {
			parsedIDs = append(parsedIDs, txID)
		}
	}

	if len(parsedIDs) == 0 {
		return errors.New("no valid transaction IDs provided")
	}

	// Validate: ต้อง status เป็น PENDING/DRAFT เท่านั้น
	// กัน user เพิ่มรายการที่ตรวจสอบไปแล้ว (COMPLETE/REPORTED) เข้าตะกร้าซ้ำ
	existing, err := s.auditRepo.GetTransactionsByIDs(ctx, parsedIDs)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}

	auditableIDs := make(map[uuid.UUID]struct{}, len(existing))
	for _, tx := range existing {
		if tx.Status == models.TxStatusPending || tx.Status == models.TxStatusDraft {
			auditableIDs[tx.ID] = struct{}{}
		}
	}

	var basketItems []models.AuditRejectBasket
	var alreadyAudited int
	for _, txID := range parsedIDs {
		if _, ok := auditableIDs[txID]; ok {
			basketItems = append(basketItems, models.AuditRejectBasket{
				TransactionID: txID,
				UserID:        uid,
			})
		} else {
			alreadyAudited++
		}
	}

	if len(basketItems) == 0 {
		return fmt.Errorf("รายการทั้งหมดถูกตรวจสอบไปแล้ว ไม่สามารถเพิ่มเข้าตะกร้าได้")
	}
	if alreadyAudited > 0 {
		return fmt.Errorf("มี %d รายการที่ถูกตรวจสอบไปแล้ว ไม่สามารถเพิ่มเข้าตะกร้าได้", alreadyAudited)
	}

	err = s.auditRepo.AddToBasket(ctx, basketItems)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}

	return nil
}

func (s *auditService) GetBasketItems(ctx context.Context, userID string) ([]models.ActualTransactionEntity, error) {
    if userID == "" {
        return nil, errors.New("user ID is required")
    }
    
    return s.auditRepo.GetBasketItems(ctx, userID)
}


func (s *auditService) CheckAuditComplete(ctx context.Context, user *models.UserInfo, year, month string) (map[string]interface{}, error) {
	// 1. ดึง departments ทั้งหมดที่ user เป็น OWNER
	perms, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user permissions: %w", err)
	}

	var ownerDepts []string
	for _, p := range perms {
		if p.Role == "OWNER" {
			ownerDepts = append(ownerDepts, p.DepartmentCode)
		}
	}

	if len(ownerDepts) == 0 {
		return map[string]interface{}{
			"is_complete":    true,
			"pending_count":  0,
			"total_count":    0,
			"reviewed_count": 0,
			"departments":    []string{},
		}, nil
	}

	// 2. นับจำนวน transaction ที่ยัง PENDING/DRAFT ในเดือน/ปีนี้
	pendingCount, err := s.auditRepo.CountPendingByDepartments(ctx, year, month, ownerDepts)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending transactions: %w", err)
	}

	// 3. นับจำนวน transaction ทั้งหมด (ทุก status) เพื่อใช้ใน progress widget
	totalCount, err := s.auditRepo.CountTotalByDepartments(ctx, year, month, ownerDepts)
	if err != nil {
		return nil, fmt.Errorf("failed to count total transactions: %w", err)
	}

	reviewedCount := totalCount - pendingCount
	if reviewedCount < 0 {
		reviewedCount = 0
	}

	return map[string]interface{}{
		"is_complete":    pendingCount == 0,
		"pending_count":  pendingCount,
		"total_count":    totalCount,
		"reviewed_count": reviewedCount,
		"departments":    ownerDepts,
	}, nil
}

func (s *auditService) RemoveFromBasket(ctx context.Context, userID string, transactionID string) error {
    if userID == "" {
        return errors.New("user ID is required")
    }
    if transactionID == "" {
        return errors.New("transaction ID to remove cannot be empty")
    }
    
    return s.auditRepo.RemoveFromBasket(ctx, userID, transactionID)
}
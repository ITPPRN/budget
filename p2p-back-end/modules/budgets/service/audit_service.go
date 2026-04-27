package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		notesByTx := map[uuid.UUID]string{}
		if len(rejectedTxIDs) > 0 {
			rejectedTxs, err := trxRepo.GetTransactionsByIDs(ctx, rejectedTxIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch basket transaction details: %w", err)
			}
			rejectedTxByDept = make(map[string][]models.ActualTransactionEntity, len(rejectedTxs))
			for _, tx := range rejectedTxs {
				rejectedTxByDept[tx.Department] = append(rejectedTxByDept[tx.Department], tx)
			}

			notesByTx, err = trxRepo.GetBasketNotes(ctx, user.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch basket notes: %w", err)
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
						Note:          notesByTx[tx.ID],
					})
				}
				if err := trxRepo.SaveRejectedItems(ctx, items); err != nil {
					return fmt.Errorf("failed to save rejected items for dept %s: %w", dept, err)
				}
			}
		}

		// 🌟 ล้างตะกร้าของ caller (OWNER) เอง
		if err := trxRepo.ClearBasket(ctx, user.ID); err != nil {
			return fmt.Errorf("failed to clear user basket: %w", err)
		}

		// 🌟 First-OWNER-wins: ลบ row เดียวกันออกจากตะกร้าของ OWNER คนอื่นด้วย
		// (delegate/branch_delegate fanned-out copies จะถูกล้างพร้อมกัน)
		if len(rejectedTxIDs) > 0 {
			if err := trxRepo.DeleteBasketRowsByTxIDs(ctx, rejectedTxIDs); err != nil {
				return fmt.Errorf("failed to clean up cross-baskets: %w", err)
			}
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

	// DELEGATE/BRANCH_DELEGATE ก็เห็นได้ตาม dept ที่ตัวเองมีสิทธิ — แค่ approve ไม่ได้
	perms, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user permissions: %w", err)
	}
	authorizedDepts := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		if p.IsActive != nil && !*p.IsActive {
			continue
		}
		switch p.Role {
		case models.RoleOwner, models.RoleDelegate, models.RoleBranchDelegate:
			authorizedDepts[p.DepartmentCode] = struct{}{}
		}
	}

	if department != "" {
		if _, ok := authorizedDepts[department]; !ok {
			return nil, fmt.Errorf("permission Denied: You do not have access to department %s", department)
		}
		targets = append(targets, department)
	} else {
		for d := range authorizedDepts {
			targets = append(targets, d)
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

func (s *auditService) AddToBasket(ctx context.Context, user *models.UserInfo, items []models.BasketAddItem) error {

	if user == nil || len(items) == 0 {
		return errors.New("invalid input: user ID and items cannot be empty")
	}

	callerUID, err := uuid.Parse(user.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	noteByTx := make(map[uuid.UUID]string, len(items))
	var parsedIDs []uuid.UUID
	for _, it := range items {
		txID, err := uuid.Parse(it.TransactionID)
		if err != nil {
			continue
		}
		parsedIDs = append(parsedIDs, txID)
		noteByTx[txID] = it.Note
	}

	if len(parsedIDs) == 0 {
		return errors.New("no valid transaction IDs provided")
	}

	// Map dept → role caller has on it. Lets us route each tx to the right basket(s).
	perms, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}
	roleByDept := make(map[string]string, len(perms))
	for _, p := range perms {
		if p.IsActive != nil && !*p.IsActive {
			continue
		}
		// OWNER beats DELEGATE/BRANCH_DELEGATE if the caller has both for the same dept
		if existing, ok := roleByDept[p.DepartmentCode]; ok {
			if existing == models.RoleOwner {
				continue
			}
		}
		roleByDept[p.DepartmentCode] = p.Role
	}

	existing, err := s.auditRepo.GetTransactionsByIDs(ctx, parsedIDs)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}

	// Index by ID + filter to PENDING/DRAFT only — กัน user เพิ่มรายการที่ตรวจสอบไปแล้ว
	txByID := make(map[uuid.UUID]models.ActualTransactionEntity, len(existing))
	for _, tx := range existing {
		if tx.Status == models.TxStatusPending || tx.Status == models.TxStatusDraft {
			txByID[tx.ID] = tx
		}
	}

	// Cache OWNER lookups per dept since multiple txs in one batch usually share a dept.
	ownersByDept := make(map[string][]string)
	getOwners := func(dept string) ([]string, error) {
		if cached, ok := ownersByDept[dept]; ok {
			return cached, nil
		}
		ids, err := s.userRepo.GetActiveOwnerIDsByDepartment(ctx, dept)
		if err != nil {
			return nil, err
		}
		ownersByDept[dept] = ids
		return ids, nil
	}

	var basketItems []models.AuditRejectBasket
	var alreadyAudited, noPermission, ownerlessDept int
	for _, txID := range parsedIDs {
		tx, ok := txByID[txID]
		if !ok {
			alreadyAudited++
			continue
		}

		role, has := roleByDept[tx.Department]
		if !has {
			noPermission++
			continue
		}

		var targetOwnerUIDs []uuid.UUID
		if role == models.RoleOwner {
			targetOwnerUIDs = []uuid.UUID{callerUID}
		} else {
			ownerIDs, err := getOwners(tx.Department)
			if err != nil {
				logs.Error(err)
				return errs.NewUnexpectedError()
			}
			if len(ownerIDs) == 0 {
				ownerlessDept++
				continue
			}
			for _, oid := range ownerIDs {
				parsed, err := uuid.Parse(oid)
				if err != nil {
					continue
				}
				targetOwnerUIDs = append(targetOwnerUIDs, parsed)
			}
		}

		for _, ownerUID := range targetOwnerUIDs {
			basketItems = append(basketItems, models.AuditRejectBasket{
				TransactionID: txID,
				UserID:        ownerUID,
				AddedBy:       callerUID,
				Note:          noteByTx[txID],
			})
		}
	}

	if len(basketItems) == 0 {
		switch {
		case noPermission > 0:
			return fmt.Errorf("ไม่มีสิทธิ์เพิ่มรายการเหล่านี้")
		case ownerlessDept > 0:
			return fmt.Errorf("ไม่พบ OWNER ของแผนก ไม่สามารถส่งเข้าตะกร้าได้")
		default:
			return fmt.Errorf("รายการทั้งหมดถูกตรวจสอบไปแล้ว ไม่สามารถเพิ่มเข้าตะกร้าได้")
		}
	}

	if err := s.auditRepo.AddToBasket(ctx, basketItems); err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}

	if alreadyAudited > 0 {
		return fmt.Errorf("เพิ่มเข้าตะกร้าแล้ว แต่มี %d รายการที่ถูกตรวจสอบไปแล้วถูกข้าม", alreadyAudited)
	}
	if noPermission > 0 {
		return fmt.Errorf("เพิ่มเข้าตะกร้าแล้ว แต่มี %d รายการที่ไม่มีสิทธิ์ถูกข้าม", noPermission)
	}
	if ownerlessDept > 0 {
		return fmt.Errorf("เพิ่มเข้าตะกร้าแล้ว แต่มี %d รายการที่หา OWNER ของแผนกไม่พบ", ownerlessDept)
	}
	return nil
}

// GetBasketItems returns:
//   - OWNER: every row in their own basket (additions by self + delegates)
//   - DELEGATE/BRANCH_DELEGATE only: rows they themselves added (across owners)
//   - mixed roles: union of the two
func (s *auditService) GetBasketItems(ctx context.Context, user *models.UserInfo) ([]models.BasketItemView, error) {
    if user == nil || user.ID == "" {
        return nil, errors.New("user is required")
    }

    isOwner, err := s.userHasOwnerPermission(ctx, user.ID)
    if err != nil {
        return nil, err
    }

    if isOwner {
        items, err := s.auditRepo.GetBasketItems(ctx, user.ID)
        if err != nil {
            return nil, err
        }
        // OWNER ที่บังเอิญเป็น DELEGATE ของ dept อื่นด้วย — รวมรายการที่ตัวเองไป add ใน OWNER คนอื่น
        if hasDelegateRole(user.Roles) {
            extra, err := s.auditRepo.GetBasketItemsAddedBy(ctx, user.ID)
            if err != nil {
                return nil, err
            }
            items = mergeBasketItems(items, extra)
        }
        return items, nil
    }

    return s.auditRepo.GetBasketItemsAddedBy(ctx, user.ID)
}

func (s *auditService) UpdateBasketNote(ctx context.Context, user *models.UserInfo, transactionID, note string) error {
    if user == nil || user.ID == "" {
        return errors.New("user is required")
    }
    if transactionID == "" {
        return errors.New("transaction ID is required")
    }

    isOwnerOnTx, err := s.callerIsOwnerOfTx(ctx, user.ID, transactionID)
    if err != nil {
        return err
    }
    if isOwnerOnTx {
        return s.auditRepo.UpdateBasketNote(ctx, user.ID, transactionID, note)
    }
    // DELEGATE/BRANCH_DELEGATE: แก้ได้เฉพาะที่ตัวเองเพิ่ม — sweep ทุก row ของ tx นี้ที่ added_by = ตัวเอง
    return s.auditRepo.UpdateBasketNoteByAddedBy(ctx, user.ID, transactionID, note)
}

// GetInBasketTxIDs returns tx IDs that are in any basket whose underlying tx
// belongs to a department the caller has access to. Frontend uses this to grey
// out items that are already in someone's basket so users do not duplicate adds.
func (s *auditService) GetInBasketTxIDs(ctx context.Context, user *models.UserInfo) ([]string, error) {
	if user == nil || user.ID == "" {
		return nil, errors.New("user is required")
	}
	perms, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	deptSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		if p.IsActive != nil && !*p.IsActive {
			continue
		}
		switch p.Role {
		case models.RoleOwner, models.RoleDelegate, models.RoleBranchDelegate:
			deptSet[p.DepartmentCode] = struct{}{}
		}
	}
	if len(deptSet) == 0 {
		return []string{}, nil
	}
	depts := make([]string, 0, len(deptSet))
	for d := range deptSet {
		depts = append(depts, d)
	}
	ids, err := s.auditRepo.GetInBasketTxIDsByDepartments(ctx, depts)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out, nil
}

// userHasOwnerPermission returns true if the user holds OWNER on any active permission.
func (s *auditService) userHasOwnerPermission(ctx context.Context, userID string) (bool, error) {
    perms, err := s.userRepo.GetUserPermissions(ctx, userID)
    if err != nil {
        return false, fmt.Errorf("failed to fetch permissions: %w", err)
    }
    for _, p := range perms {
        if p.Role == models.RoleOwner && (p.IsActive == nil || *p.IsActive) {
            return true, nil
        }
    }
    return false, nil
}

// callerIsOwnerOfTx returns true if caller is OWNER of the tx's department.
func (s *auditService) callerIsOwnerOfTx(ctx context.Context, userID, transactionID string) (bool, error) {
    txID, err := uuid.Parse(transactionID)
    if err != nil {
        return false, nil
    }
    txs, err := s.auditRepo.GetTransactionsByIDs(ctx, []uuid.UUID{txID})
    if err != nil {
        return false, err
    }
    if len(txs) == 0 {
        return false, nil
    }
    perms, err := s.userRepo.GetUserPermissions(ctx, userID)
    if err != nil {
        return false, err
    }
    for _, p := range perms {
        if p.DepartmentCode == txs[0].Department && p.Role == models.RoleOwner && (p.IsActive == nil || *p.IsActive) {
            return true, nil
        }
    }
    return false, nil
}

func hasDelegateRole(roles []string) bool {
    for _, r := range roles {
        if strings.EqualFold(r, models.RoleDelegate) || strings.EqualFold(r, models.RoleBranchDelegate) {
            return true
        }
    }
    return false
}

// mergeBasketItems unions two slices of basket items, deduplicating by tx ID.
// OWNER's own basket row wins over the added_by row when both exist.
func mergeBasketItems(primary, secondary []models.BasketItemView) []models.BasketItemView {
    seen := make(map[uuid.UUID]struct{}, len(primary)+len(secondary))
    out := make([]models.BasketItemView, 0, len(primary)+len(secondary))
    for _, it := range primary {
        seen[it.ID] = struct{}{}
        out = append(out, it)
    }
    for _, it := range secondary {
        if _, ok := seen[it.ID]; ok {
            continue
        }
        seen[it.ID] = struct{}{}
        out = append(out, it)
    }
    return out
}


func (s *auditService) CheckAuditComplete(ctx context.Context, user *models.UserInfo, year, month string) (map[string]interface{}, error) {
	// 1. ดึง departments ทั้งหมดที่ user มีสิทธิ (OWNER / DELEGATE / BRANCH_DELEGATE)
	perms, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user permissions: %w", err)
	}

	deptSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		if p.IsActive != nil && !*p.IsActive {
			continue
		}
		switch p.Role {
		case models.RoleOwner, models.RoleDelegate, models.RoleBranchDelegate:
			deptSet[p.DepartmentCode] = struct{}{}
		}
	}
	depts := make([]string, 0, len(deptSet))
	for d := range deptSet {
		depts = append(depts, d)
	}

	if len(depts) == 0 {
		return map[string]interface{}{
			"is_complete":    true,
			"pending_count":  0,
			"total_count":    0,
			"reviewed_count": 0,
			"departments":    []string{},
		}, nil
	}

	// 2. นับจำนวน transaction ที่ยัง PENDING/DRAFT ในเดือน/ปีนี้
	pendingCount, err := s.auditRepo.CountPendingByDepartments(ctx, year, month, depts)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending transactions: %w", err)
	}

	// 3. นับจำนวน transaction ทั้งหมด (ทุก status) เพื่อใช้ใน progress widget
	totalCount, err := s.auditRepo.CountTotalByDepartments(ctx, year, month, depts)
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
		"departments":    depts,
	}, nil
}

func (s *auditService) RemoveFromBasket(ctx context.Context, user *models.UserInfo, transactionID string) error {
    if user == nil || user.ID == "" {
        return errors.New("user is required")
    }
    if transactionID == "" {
        return errors.New("transaction ID to remove cannot be empty")
    }

    isOwnerOnTx, err := s.callerIsOwnerOfTx(ctx, user.ID, transactionID)
    if err != nil {
        return err
    }
    if isOwnerOnTx {
        return s.auditRepo.RemoveFromBasket(ctx, user.ID, transactionID)
    }
    // DELEGATE/BRANCH_DELEGATE: ลบเฉพาะที่ตัวเองเพิ่ม → ลบครบทุก fanned-out row ของ tx นี้
    return s.auditRepo.RemoveFromBasketByAddedBy(ctx, user.ID, transactionID)
}
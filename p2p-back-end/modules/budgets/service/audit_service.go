package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"p2p-back-end/modules/entities/models"
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

	if department != "" {
		if err := s.checkOwnerPermission(ctx, user.ID, department); err != nil {
			return err
		}
		targets = append(targets, department)
	} else {
		// Identify ALL authorized departments for this owner
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

	year, _ := payload["year"].(string)
	month, _ := payload["month"].(string)
	entity, _ := payload["entity"].(string)
	branch, _ := payload["branch"].(string)
	selectedIDs, _ := payload["selected_item_ids"].([]interface{})

	// Handle selective approval
	if len(selectedIDs) > 0 {
		var txIDs []uuid.UUID
		for _, idStr := range selectedIDs {
			id, err := uuid.Parse(fmt.Sprintf("%v", idStr))
			if err == nil {
				txIDs = append(txIDs, id)
			}
		}

		// 1. Mark selected items as REPORTED (Ready for Admin/Accountant)
		if err := s.auditRepo.UpdateTransactionsStatus(ctx, txIDs, models.TxStatusReported); err != nil {
			return fmt.Errorf("failed to report selected items: %w", err)
		}

		// 2. Mark UN-SELECTED items in same scope as COMPLETE (Won't show in dashboard anymore)
		// This achieves the "Complete" logic for everything else.
		// We first find the scope (Dept/Month/Year) which we already have in 'targets'
		for _, dept := range targets {
			if err := s.auditRepo.MarkRestAsComplete(ctx, dept, year, month, txIDs); err != nil {
				return fmt.Errorf("failed to auto-complete remaining items: %w", err)
			}
		}
	}

	for _, dept := range targets {
		log := &models.AuditLogEntity{
			ID:            uuid.New(),
			Entity:        entity,
			Branch:        branch,
			Department:    dept,
			Year:          year,
			Month:         month,
			Status:        "CONFIRMED",
			RejectedCount: 0,
			CreatedBy:     user.Name,
		}

		if err := s.auditRepo.SaveAuditLog(ctx, log); err != nil {
			return fmt.Errorf("failed to save approval log for dept %s: %w", dept, err)
		}
	}

	return nil
}

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
		for _, tx := range items {
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

		// 🛠️ Update Transaction Status to DRAFT (Basket Mode)
		if err := s.auditRepo.UpdateTransactionsStatus(ctx, txIDs, models.TxStatusDraft); err != nil {
			return fmt.Errorf("failed to mark items as draft: %w", err)
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

	filter := map[string]interface{}{
		"departments": targets,
		"year":        payload["year"],
		"months":      []string{fmt.Sprintf("%v", payload["month"])},
		"conso_gls":   consoGLs,
		"start_date":  payload["start_date"],
		"end_date":    payload["end_date"],
		"search":      payload["search"],
		"limit":       -1, 
	}
	
	return s.auditRepo.GetTransactionsByFilter(ctx, filter)
}

func (s *auditService) checkOwnerPermission(ctx context.Context, userID, department string) error {
	// Must have OWNER role in System
	// (Already checked by controller usually, but we double check for Dept access)
	
	perms, err := s.userRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return fmt.Errorf("Failed to verify permissions: %w", err)
	}

	for _, p := range perms {
		if p.DepartmentCode == department && p.Role == "OWNER" {
			return nil
		}
	}
	return fmt.Errorf("Permission Denied: You are not the owner of department %s", department)
}

package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
)

type userRepositoryDB struct {
	db *gorm.DB
}

func NewUserRepositoryDB(db *gorm.DB) models.UserRepository {
	return &userRepositoryDB{db: db}
}

func (r *userRepositoryDB) IsUserExistByID(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("user_entities").Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("userRepo.IsUserExistByID: %w", err)
	}
	return count > 0, nil
}

func applyJoins(condition *gorm.DB, optional map[string]interface{}, fieldTableMapping map[string]string) *gorm.DB {
	joined := false

	// 1. Join for Filters (Search/Sort)
	if utils.NeedsJoin(optional, fieldTableMapping, "departments") {
		condition = condition.Joins("LEFT JOIN departments ON departments.id = user_entities.department_id")
		joined = true
	}

	// 2. Join for Visibility
	if !joined && optional["visibility_allowed_depts"] != nil {
		condition = condition.Joins("LEFT JOIN departments ON departments.id = user_entities.department_id")
		joined = true
	}

	return condition
}

func (r *userRepositoryDB) GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]models.UserEntity, int, error) {
	var users []models.UserEntity
	var totalRecords64 int64

	fieldTableMapping := map[string]string{
		"username":        "user_entities.username",
		"email":           "user_entities.email",
		"department_name": "departments.name",
		"department_code": "departments.code",
	}

	logs.Infof("Repo GetAll: offset=%d, size=%d, optional=%+v", offset, size, optional)

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	condition := r.db.WithContext(ctx).Model(&models.UserEntity{})
	condition = applyJoins(condition, optional, fieldTableMapping)

	// --- Visibility Logic ---
	currentUserID, _ := optional["visibility_current_user_id"].(string)
	role, _ := optional["visibility_role"].(string)
	allowedDepts, _ := optional["visibility_allowed_depts"].([]string)

	logs.Info(fmt.Sprintf("Repo Visibility - User: %s, Role: %s, AllowedDepts: %v", currentUserID, role, allowedDepts))
	// Debug: Verify Current User ID for Sorting
	logs.Infof("GetAll: Fetching users for Management (Current User ID: %s)", currentUserID)

	if role != "ADMIN" {
		// Not Admin: Filter by (Self OR Allowed Departments with Role Restrictions)
		visibilityQuery := "user_entities.id = ?"
		args := []interface{}{currentUserID}

		notOwnerOrAdmin := `NOT EXISTS (
			SELECT 1 FROM user_permission_entities 
			WHERE user_permission_entities.user_id = user_entities.id 
			AND user_permission_entities.is_active = true 
			AND (user_permission_entities.role ILIKE '%ADMIN%' OR user_permission_entities.role ILIKE '%OWNER%')
		) AND (COALESCE(user_entities.roles::text, '') NOT ILIKE '%ADMIN%' AND COALESCE(user_entities.roles::text, '') NOT ILIKE '%OWNER%')`

		if len(allowedDepts) > 0 {
			if role == "OWNER" {
				// Owner: See Self OR (Allowed Departments AND NOT OTHER Owner/Admin)
				visibilityQuery += " OR (departments.code IN ? AND " + notOwnerOrAdmin + ")"
				args = append(args, allowedDepts)
			} else if role == "DELEGATE" {
				// Delegate: See Self OR (Allowed Departments AND NOT (Admin OR Owner))
				visibilityQuery += " OR (departments.code IN ? AND " + notOwnerOrAdmin + ")"
				args = append(args, allowedDepts)
			}
		}
		condition = condition.Where("("+visibilityQuery+")", args...)
	}

	// --- Standard Filters ---
	for field, value := range optional {
		if strings.HasPrefix(field, "visibility_") {
			continue // Skip internal visibility keys
		}
		if field == "search" && value != nil && value != "" {
			searchStr := fmt.Sprintf("%%%v%%", value)
			condition = condition.Where("(user_entities.username ILIKE ? OR user_entities.first_name ILIKE ? OR user_entities.last_name ILIKE ? OR user_entities.name_th ILIKE ? OR user_entities.name_en ILIKE ?)",
				searchStr, searchStr, searchStr, searchStr, searchStr)
			continue
		}
		if field == "status" && value != nil && value != "ALL" {
			if value == "ACTIVE" {
				condition = condition.Where("EXISTS (SELECT 1 FROM user_permission_entities WHERE user_permission_entities.user_id = user_entities.id AND user_permission_entities.is_active = true)")
			} else if value == "INACTIVE" {
				condition = condition.Where("NOT EXISTS (SELECT 1 FROM user_permission_entities WHERE user_permission_entities.user_id = user_entities.id AND user_permission_entities.is_active = true)")
			}
			continue
		}
		if field == "role" && value != nil && value != "ALL" {
			roleStr := fmt.Sprintf("%v", value)
			condition = condition.Where("(user_entities.roles::text ILIKE ? OR EXISTS (SELECT 1 FROM user_permission_entities WHERE user_permission_entities.user_id = user_entities.id AND user_permission_entities.role ILIKE ? AND user_permission_entities.is_active = true))",
				"%"+roleStr+"%", "%"+roleStr+"%")
			continue
		}
		if field == "department_code" && value != nil && value != "ALL" {
			condition = condition.Where("departments.code = ?", value)
		} else if column, ok := fieldTableMapping[field]; ok && value != nil && value != "ALL" {
			condition = utils.AddCondition(condition, value, column)
		}
	}

	if err := condition.Session(&gorm.Session{}).Count(&totalRecords64).Error; err != nil {
		return nil, 0, fmt.Errorf("error counting user entities: %w", err)
	}

	// --- Sorting: Self at Top, then Username ---
	// Use both ID and Username for robustness in "Self" detection (Case-Insensitive)
	if currentUserID != "" {
		condition = condition.Order(fmt.Sprintf("CASE WHEN LOWER(user_entities.id) = LOWER('%s') OR LOWER(user_entities.username) = LOWER('%s') THEN 0 ELSE 1 END", currentUserID, currentUserID))
	}
	condition = condition.Order("LOWER(user_entities.username) ASC")

	if err := condition.Offset(offset).Limit(size).Preload("Department").Preload("UserPermissions").Select("user_entities.*").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("error getting users: %w", err)
	}

	// Debug: check first few results
	if len(users) > 0 {
		logs.Infof("GetAll Result: First user in list is %s (ID: %s)", users[0].Username, users[0].ID)
		if users[0].ID == currentUserID {
			logs.Info("✅ SUCCESS: Logged-in user is at the top.")
		} else {
			logs.Warnf("⚠️ WARNING: Logged-in user is NOT at the top. Top ID: %s vs Current ID: %s", users[0].ID, currentUserID)
		}
	}

	logs.Info(fmt.Sprintf("Repo Result - Total: %d, ResultCount: %d", totalRecords64, len(users)))
	for _, u := range users {
		logs.Debug(fmt.Sprintf("   - Visible User: %s (ID: %s)", u.Username, u.ID))
	}

	return users, int(totalRecords64), nil
}

func (r *userRepositoryDB) CreateUser(ctx context.Context, user *models.UserEntity) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("userRepo.CreateUser: %w", err)
	}
	return nil
}

func (r *userRepositoryDB) UpdateUser(ctx context.Context, user *models.UserEntity) error {
	if err := r.db.WithContext(ctx).Model(user).Updates(user).Error; err != nil {
		return fmt.Errorf("userRepo.UpdateUser: %w", err)
	}
	return nil
}

func (r *userRepositoryDB) GetUserContext(ctx context.Context, userID string) (*models.UserEntity, error) {
	var user models.UserEntity
	err := r.db.WithContext(ctx).Preload("Department").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetUserContext: %w", err)
	}
	return &user, nil
}

func (r *userRepositoryDB) GetUserPermissions(ctx context.Context, userID string) ([]models.UserPermissionEntity, error) {
	var perms []models.UserPermissionEntity
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&perms).Error
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetUserPermissions: %w", err)
	}
	return perms, nil
}

func (r *userRepositoryDB) SetUserPermissions(ctx context.Context, userID string, permissions []models.UserPermissionEntity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete old
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserPermissionEntity{}).Error; err != nil {
			return fmt.Errorf("userRepo.SetUserPermissions (delete): %w", err)
		}
		// 2. Add new (if any)
		if len(permissions) > 0 {
			if err := tx.Create(&permissions).Error; err != nil {
				return fmt.Errorf("userRepo.SetUserPermissions (create): %w", err)
			}
		}
		return nil
	})
}
func (r *userRepositoryDB) ListDepartments(ctx context.Context) ([]models.Departments, error) {
	var results []string
	err := r.db.WithContext(ctx).
		Table("departments").
		Where("EXISTS (SELECT 1 FROM user_entities WHERE user_entities.department_id = departments.id)").
		Distinct("code").
		Order("code asc").
		Pluck("code", &results).Error

	if err != nil {
		return nil, fmt.Errorf("userRepo.ListDepartments (Pluck CodeMap): %w", err)
	}

	depts := make([]models.Departments, len(results))
	for i, code := range results {
		// Use CodeMap as both code and name for the dropdown to show Master Categories
		depts[i] = models.Departments{
			Code: code,
			Name: code,
		}
	}
	return depts, nil
}

func (r *userRepositoryDB) ListMasterDepartments(ctx context.Context) ([]models.DepartmentEntity, error) {
	var depts []models.DepartmentEntity
	err := r.db.WithContext(ctx).Order("code asc").Find(&depts).Error
	if err != nil {
		return nil, fmt.Errorf("userRepo.ListMasterDepartments: %w", err)
	}
	return depts, nil
}

func (r *userRepositoryDB) GetDepartmentByCode(ctx context.Context, code string) (*models.DepartmentEntity, error) {
	var dept models.DepartmentEntity
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&dept).Error
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetDepartmentByCode: %w", err)
	}
	return &dept, nil
}

func (r *userRepositoryDB) GetDepartmentByNavCode(ctx context.Context, navCode string) (*models.DepartmentEntity, error) {
	var mapping models.DepartmentMappingEntity
	// Find Mapping First
	err := r.db.WithContext(ctx).Preload("Department").Where("nav_code = ?", navCode).First(&mapping).Error
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetDepartmentByNavCode: %w", err)
	}
	return mapping.Department, nil
}

func (r *userRepositoryDB) FindByUsername(ctx context.Context, username string) (*models.UserEntity, error) {
	var user models.UserEntity
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("userRepo.FindByUsername: %w", err)
	}
	return &user, nil
}

func (r *userRepositoryDB) SyncUsers(ctx context.Context, users []models.UserEntity) ([]models.UserEntity, error) {
	if len(users) == 0 {
		return nil, nil
	}

	logs.Infof("[SyncUsers] Starting batch upsert for %d users...", len(users))

	// --- Final Shield: De-duplicate users within the batch to avoid SQLSTATE 21000 ---
	uniqueUsers := make(map[string]models.UserEntity)
	duplicateCount := 0
	emptyUsernameCount := 0

	for _, u := range users {
		if u.Username == "" {
			emptyUsernameCount++
			continue
		}
		if _, exists := uniqueUsers[u.Username]; exists {
			duplicateCount++
		}
		uniqueUsers[u.Username] = u
	}

	sanitizedUsers := make([]models.UserEntity, 0, len(uniqueUsers))
	for _, u := range uniqueUsers {
		sanitizedUsers = append(sanitizedUsers, u)
	}

	logs.Infof("[SyncUsers] Batch Analysis: Total=%d, Unique=%d, DuplicatesSkipped=%d, EmptyUsernameSkipped=%d",
		len(users), len(sanitizedUsers), duplicateCount, emptyUsernameCount)

	var changedRows []models.UserEntity
	// Use CentralID or Username as conflict key since source ID is ignored for PK
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "username"}}, // Or central_id
		DoUpdates: clause.Assignments(map[string]interface{}{
			"central_id":    gorm.Expr("EXCLUDED.central_id"),
			"name_th":       gorm.Expr("EXCLUDED.name_th"),
			"name_en":       gorm.Expr("EXCLUDED.name_en"),
			"company_id":    gorm.Expr("EXCLUDED.company_id"),
			"department_id": gorm.Expr("EXCLUDED.department_id"),
			"section_id":    gorm.Expr("EXCLUDED.section_id"),
			"position_id":   gorm.Expr("EXCLUDED.position_id"),
			"updated_at":    gorm.Expr("NOW()"),
		}),
		Where: clause.Where{
			Exprs: []clause.Expression{
				gorm.Expr(`
					user_entities.name_th       IS DISTINCT FROM EXCLUDED.name_th OR
					user_entities.name_en       IS DISTINCT FROM EXCLUDED.name_en OR
					user_entities.company_id    IS DISTINCT FROM EXCLUDED.company_id OR
					user_entities.department_id IS DISTINCT FROM EXCLUDED.department_id OR
					user_entities.section_id    IS DISTINCT FROM EXCLUDED.section_id OR
					user_entities.position_id   IS DISTINCT FROM EXCLUDED.position_id
				`),
			},
		},
	}).
		Clauses(clause.Returning{}).
		Create(&sanitizedUsers).
		Scan(&changedRows).
		Error

	if err != nil {
		return nil, fmt.Errorf("userRepo.SyncUsers: %w", err)
	}

	logs.Infof("[SyncUsers] Batch sync finished. Changed rows detected: %d", len(changedRows))
	return changedRows, nil
}

func (r *userRepositoryDB) GetUsers(ctx context.Context, lastID uint, limit int) ([]models.UserEntity, error) {
	var users []models.UserEntity
	err := r.db.WithContext(ctx).Preload("Company").Preload("Department").Preload("Section").Preload("Position").Where("central_id > ?", lastID).Order("central_id ASC").Limit(limit).Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetUsers: %w", err)
	}
	return users, nil
}
func (r *userRepositoryDB) UpdateUserID(ctx context.Context, oldID, newID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update User Entity ID
		if err := tx.Exec("UPDATE user_entities SET id = ? WHERE id = ?", newID, oldID).Error; err != nil {
			return err
		}
		// Update Permissions UserID
		if err := tx.Exec("UPDATE user_permission_entities SET user_id = ? WHERE user_id = ?", newID, oldID).Error; err != nil {
			return err
		}
		return nil
	})
}

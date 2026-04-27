package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
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
		// joined = true
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

	logs.Infof("[DEBUG] Repo GetAll: currentUserId=%v, role=%v, allowedDepts=%v", optional["visibility_current_user_id"], optional["visibility_role"], optional["visibility_allowed_depts"])
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
		// 🛡️ CRITICAL: Always ensure the current user can see themselves
		visibilityQuery := "user_entities.id = ? OR user_entities.username = ?"
		args := []interface{}{currentUserID, currentUserID}

		notOwnerOrAdmin := `NOT EXISTS (
			SELECT 1 FROM user_permission_entities 
			WHERE user_permission_entities.user_id = user_entities.id 
			AND user_permission_entities.is_active = true 
			AND (user_permission_entities.role ILIKE '%ADMIN%' OR user_permission_entities.role ILIKE '%OWNER%')
		) AND (COALESCE(user_entities.roles::text, '') NOT ILIKE '%ADMIN%' AND COALESCE(user_entities.roles::text, '') NOT ILIKE '%OWNER%')`

		if len(allowedDepts) > 0 {
			if role == "OWNER" || role == "DELEGATE" {
				// See Self OR (Allowed Departments [Granular OR Master Mapping] AND NOT OTHER Owner/Admin)
				visibilityQuery += " OR ((departments.code IN ? OR departments.code_map IN ?) AND " + notOwnerOrAdmin + ")"
				args = append(args, allowedDepts, allowedDepts)
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
			condition = condition.Where("(user_entities.id ILIKE ? OR user_entities.username ILIKE ? OR user_entities.first_name ILIKE ? OR user_entities.last_name ILIKE ? OR user_entities.name_th ILIKE ? OR user_entities.name_en ILIKE ?)",
				searchStr, searchStr, searchStr, searchStr, searchStr, searchStr)
			continue
		}
		// if field == "status" && value != nil && value != "ALL" {
		// 	if value == "ACTIVE" {
		// 		condition = condition.Where("(EXISTS (SELECT 1 FROM user_permission_entities WHERE user_permission_entities.user_id = user_entities.id AND user_permission_entities.is_active = true) OR COALESCE(user_entities.roles::text, '') ILIKE '%ADMIN%')")
		// 	} else if value == "INACTIVE" {
		// 		condition = condition.Where("(NOT EXISTS (SELECT 1 FROM user_permission_entities WHERE user_permission_entities.user_id = user_entities.id AND user_permission_entities.is_active = true) AND COALESCE(user_entities.roles::text, '') NOT ILIKE '%ADMIN%')")
		// 	}
		// 	continue
		// }
		if field == "status" && value != nil && value != "ALL" {
			// เปลี่ยนมาใช้ switch แทน if-else
			switch value {
			case "ACTIVE":
				condition = condition.Where("(EXISTS (SELECT 1 FROM user_permission_entities WHERE user_permission_entities.user_id = user_entities.id AND user_permission_entities.is_active = true) OR COALESCE(user_entities.roles::text, '') ILIKE '%ADMIN%')")
			case "INACTIVE":
				condition = condition.Where("(NOT EXISTS (SELECT 1 FROM user_permission_entities WHERE user_permission_entities.user_id = user_entities.id AND user_permission_entities.is_active = true) AND COALESCE(user_entities.roles::text, '') NOT ILIKE '%ADMIN%')")
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
			condition = condition.Where("departments.code_map = ?", value)
		} else if column, ok := fieldTableMapping[field]; ok && value != nil && value != "ALL" {
			condition = utils.AddCondition(condition, value, column)
		}
	}

	// --- Soft Delete Filter (NULL-Safe for transition) ---
	condition = condition.Where("(user_entities.deleted = false OR user_entities.deleted IS NULL)")

	if err := condition.Session(&gorm.Session{}).Count(&totalRecords64).Error; err != nil {
		return nil, 0, fmt.Errorf("error counting user entities: %w", err)
	}

	// --- Sorting: Self at Top, then Username ---
	if currentUserID != "" {
		logs.Infof("[SORT] Panning current user to top: %s", currentUserID)
		// 🚀 NUCLEAR OPTION: Match anything that identifies the current user
		condition = condition.Order(gorm.Expr(`
			CASE 
				WHEN user_entities.id = ? 
				OR user_entities.username = ? 
				OR LOWER(user_entities.username) = LOWER(?)
				OR (LOWER(user_entities.first_name) || ' ' || LOWER(user_entities.last_name)) = LOWER(?)
				OR user_entities.id ILIKE ?
				THEN 0 
				ELSE 1 
			END`, currentUserID, currentUserID, currentUserID, currentUserID, "%"+currentUserID+"%"))
	}
	condition = condition.Order("LOWER(user_entities.username) ASC")

	if err := condition.Offset(offset).Limit(size).Preload("Department").Preload("UserPermissions").Select("user_entities.*").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("error getting users: %w", err)
	}

	// Double check: if current user is not in the list (due to pagination), we can't easily prepend.
	// But since they are at the top of the SORT, they should be in the first page.
	if len(users) > 0 {
		logs.Infof("[DEBUG] GetAll Top User: %s (ID: %s)", users[0].Username, users[0].ID)
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
func (r *userRepositoryDB) ReactivateUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Model(&models.UserEntity{}).Where("id = ?", userID).Update("deleted", false).Error; err != nil {
		return fmt.Errorf("userRepo.ReactivateUser: %w", err)
	}
	return nil
}

func (r *userRepositoryDB) GetUserContext(ctx context.Context, userID string) (*models.UserEntity, error) {
    var user models.UserEntity
    // ใช้ LOWER() เพื่อทำให้ทั้งข้อมูลใน DB และตัวแปรที่รับมาเป็นตัวพิมพ์เล็กทั้งหมด
    err := r.db.WithContext(ctx).
        Preload("Department").
        Where("LOWER(username) = LOWER(?)", userID). 
        First(&user).Error
        
    if err != nil {
        return nil, fmt.Errorf("userRepo.GetUserContext: %w", err)
    }
    return &user, nil
}

// func (r *userRepositoryDB) GetUserContext(ctx context.Context, userID string) (*models.UserEntity, error) {
// 	var user models.UserEntity
// 	err := r.db.WithContext(ctx).Preload("Department").Where("username = ?", userID).First(&user).Error
// 	if err != nil {
// 		return nil, fmt.Errorf("userRepo.GetUserContext: %w", err)
// 	}
// 	return &user, nil
// }

func (r *userRepositoryDB) GetUserPermissions(ctx context.Context, userID string) ([]models.UserPermissionEntity, error) {
	var perms []models.UserPermissionEntity
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&perms).Error
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetUserPermissions: %w", err)
	}
	return perms, nil
}

// GetActiveOwnerIDsByDepartment returns user IDs of all active OWNERs for a dept.
// Used to fan-out a delegate's basket-add into every OWNER's basket.
func (r *userRepositoryDB) GetActiveOwnerIDsByDepartment(ctx context.Context, departmentCode string) ([]string, error) {
	if departmentCode == "" {
		return nil, nil
	}
	var ids []string
	err := r.db.WithContext(ctx).
		Table("user_permission_entities AS p").
		Joins("INNER JOIN user_entities AS u ON u.id = p.user_id").
		Where("p.department_code = ? AND p.role = ?", departmentCode, models.RoleOwner).
		Where("(p.is_active IS NULL OR p.is_active = true)").
		Where("u.is_active = true AND u.deleted = false").
		Distinct().
		Pluck("u.id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetActiveOwnerIDsByDepartment: %w", err)
	}
	return ids, nil
}

func (r *userRepositoryDB) UpdateUserPermissionsAndRoles(ctx context.Context, userID string, permissions []models.UserPermissionEntity, roles []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Update Roles
		rolesJSON, err := json.Marshal(roles)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.UserEntity{}).Where("id = ?", userID).Update("roles", datatypes.JSON(rolesJSON)).Error; err != nil {
			return err
		}

		// 2. Delete old permissions
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserPermissionEntity{}).Error; err != nil {
			return err
		}

		// 3. Add new permissions (if any)
		if len(permissions) > 0 {
			if err := tx.Create(&permissions).Error; err != nil {
				return err
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
		Where("code_map IS NOT NULL AND code_map != 'None' AND code_map != ''").
		Distinct("code_map").
		Order("code_map asc").
		Pluck("code_map", &results).Error

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
    // ใช้ LOWER(username) เพื่อเปรียบเทียบข้อมูลแบบไม่สน Case
    // และใช้ strings.ToLower(username) กับ input ที่รับเข้ามาด้วย
    err := r.db.WithContext(ctx).
        Where("LOWER(username) = LOWER(?)", username).
        First(&user).Error

    if err != nil {
        return nil, fmt.Errorf("userRepo.FindByUsername: %w", err)
    }
    return &user, nil
}

// func (r *userRepositoryDB) FindByUsername(ctx context.Context, username string) (*models.UserEntity, error) {
// 	var user models.UserEntity
// 	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
// 		return nil, fmt.Errorf("userRepo.FindByUsername: %w", err)
// 	}
// 	return &user, nil
// }

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
			"deleted":       gorm.Expr("EXCLUDED.deleted"),
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
					user_entities.position_id   IS DISTINCT FROM EXCLUDED.position_id OR
					user_entities.deleted       IS DISTINCT FROM EXCLUDED.deleted
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
	err := r.db.WithContext(ctx).Preload("Company").Preload("Department").Preload("Section").Preload("Position").
		Where("central_id > ? AND (deleted = false OR deleted IS NULL)", lastID).
		Order("central_id ASC").Limit(limit).Find(&users).Error
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

func (r *userRepositoryDB) UpdateUserRoles(ctx context.Context, userID string, roles []string) error {
	logs.Infof("[Repo] UpdateUserRoles: UserID=%s, Roles=%v", userID, roles)
	rolesJSON, err := json.Marshal(roles)
	if err != nil {
		logs.Errorf("[Repo] UpdateUserRoles Error Marshalling: %v", err)
		return err
	}

	result := r.db.WithContext(ctx).Model(&models.UserEntity{}).
		Where("id = ?", userID).
		Update("roles", datatypes.JSON(rolesJSON))

	if result.Error != nil {
		logs.Errorf("[Repo] UpdateUserRoles DB Error: %v", result.Error)
		return result.Error
	}
	logs.Infof("[Repo] UpdateUserRoles Success: RowsAffected=%d", result.RowsAffected)
	return nil
}

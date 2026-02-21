package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

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

func (r *userRepositoryDB) IsUserExistByID(id string) (bool, error) {
	var count int64
	if err := r.db.Table("user_entities").Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func applyJoins(condition *gorm.DB, optional map[string]interface{}, fieldTableMapping map[string]string) *gorm.DB {
	joined := false

	// 1. Join for Filters (Search/Sort)
	if utils.NeedsJoin(optional, fieldTableMapping, "department_entities") {
		condition = condition.Joins("LEFT JOIN department_entities ON department_entities.id = user_entities.department_id")
		joined = true
	}

	// 2. Join for Visibility
	if !joined && optional["visibility_allowed_depts"] != nil {
		condition = condition.Joins("LEFT JOIN department_entities ON department_entities.id = user_entities.department_id")
		joined = true
	}

	return condition
}

func (r *userRepositoryDB) GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]models.UserEntity, int, error) {
	fmt.Printf("DEBUG: repository.GetAll called with optional: %+v\n", optional)
	var users []models.UserEntity
	var totalRecords64 int64

	fieldTableMapping := map[string]string{
		"username":        "user_entities.username",
		"email":           "user_entities.email",
		"department_name": "department_entities.name",
		"department_code": "department_entities.code",
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	condition := r.db.WithContext(ctx).Table("user_entities")
	condition = applyJoins(condition, optional, fieldTableMapping)

	// --- Visibility Logic ---
	currentUserID, _ := optional["visibility_current_user_id"].(string)
	role, _ := optional["visibility_role"].(string)
	allowedDepts, _ := optional["visibility_allowed_depts"].([]string)

	logs.Info(fmt.Sprintf("Repo Visibility - User: %s, Role: %s, AllowedDepts: %v", currentUserID, role, allowedDepts))

	if role != "ADMIN" {
		// Not Admin: Filter by (Self OR Allowed Departments with Role Restrictions)
		visibilityQuery := "user_entities.id = ?"
		args := []interface{}{currentUserID}

		if len(allowedDepts) > 0 {
			if role == "OWNER" {
				// Owner: See Self OR Allowed Departments
				visibilityQuery += " OR department_entities.code IN ?"
				args = append(args, allowedDepts)
			} else if role == "DELEGATE" {
				// Delegate: See Self OR (Allowed Departments AND NOT (Admin OR Owner))
				visibilityQuery += " OR (department_entities.code IN ? AND COALESCE(user_entities.roles::text, '') NOT ILIKE '%ADMIN%' AND COALESCE(user_entities.roles::text, '') NOT ILIKE '%OWNER%')"
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
		if field == "search" && value != nil {
			searchStr := fmt.Sprintf("%%%v%%", value)
			condition = condition.Where("(username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?)", searchStr, searchStr, searchStr)
			continue
		}
		if column, ok := fieldTableMapping[field]; ok && value != nil {
			condition = utils.AddCondition(condition, value, column)
		}
	}

	if err := condition.Count(&totalRecords64).Error; err != nil {
		return nil, 0, fmt.Errorf("error counting user entities: %w", err)
	}

	// --- Sorting: Self at Top, then Username ---
	if currentUserID != "" {
		condition = condition.Order(fmt.Sprintf("CASE WHEN user_entities.id = '%s' THEN 0 ELSE 1 END", currentUserID))
	}
	condition = condition.Order("user_entities.username ASC")

	if err := condition.Offset(offset).Limit(size).Preload("Department").Preload("UserPermissions").Select("user_entities.*").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("error getting users: %w", err)
	}

	logs.Info(fmt.Sprintf("Repo Result - Total: %d, ResultCount: %d", totalRecords64, len(users)))
	for _, u := range users {
		logs.Debug(fmt.Sprintf("   - Visible User: %s (ID: %s)", u.Username, u.ID))
	}

	return users, int(totalRecords64), nil
}

func (r *userRepositoryDB) CreateUser(user *models.UserEntity) error {
	return r.db.Create(user).Error
}

func (r *userRepositoryDB) UpdateUser(user *models.UserEntity) error {
	return r.db.Model(user).Updates(user).Error
}

func (r *userRepositoryDB) GetUserContext(userID string) (*models.UserEntity, error) {
	var user models.UserEntity
	err := r.db.Preload("Department").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryDB) GetUserPermissions(userID string) ([]models.UserPermissionEntity, error) {
	var perms []models.UserPermissionEntity
	err := r.db.Where("user_id = ?", userID).Find(&perms).Error
	return perms, err
}

func (r *userRepositoryDB) SetUserPermissions(userID string, permissions []models.UserPermissionEntity) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete old
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserPermissionEntity{}).Error; err != nil {
			return err
		}
		// 2. Add new (if any)
		if len(permissions) > 0 {
			if err := tx.Create(&permissions).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *userRepositoryDB) ListDepartments() ([]models.DepartmentEntity, error) {
	var depts []models.DepartmentEntity
	err := r.db.Order("code asc").Find(&depts).Error
	return depts, err
}

func (r *userRepositoryDB) GetDepartmentByCode(code string) (*models.DepartmentEntity, error) {
	var dept models.DepartmentEntity
	err := r.db.Where("code = ?", code).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *userRepositoryDB) GetDepartmentByNavCode(navCode string) (*models.DepartmentEntity, error) {
	var mapping models.DepartmentMappingEntity
	// Find Mapping First
	err := r.db.Preload("Department").Where("nav_code = ?", navCode).First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return mapping.Department, nil
}

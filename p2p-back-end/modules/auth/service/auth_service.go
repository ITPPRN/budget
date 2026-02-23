package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/datatypes"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/errs"
)

type authService struct {
	keycloak *gocloak.GoCloak
	cfg      *configs.Config
	authRepo models.UserRepository
	Redis    *redis.Client
}

func filterRoles(roles []string) []string {
	var filtered []string
	validRoles := map[string]bool{
		"ADMIN":    true,
		"OWNER":    true,
		"DELEGATE": true,
	}

	for _, r := range roles {
		upperCaseRole := strings.ToUpper(r)
		// Exclude noise like default-roles-*, offline_access, uma_authorization, etc.
		if strings.HasPrefix(r, "default-roles-") || r == "offline_access" || r == "uma_authorization" {
			continue
		}

		// If it's a known app role, add it
		if validRoles[upperCaseRole] {
			filtered = append(filtered, upperCaseRole)
		} else {
			// Optionally keep other roles but this might include groups like 'IT'
			// The user said: "System role should only be admin, owner, delegate"
			// So I will strictly filter to these.
		}
	}

	if len(filtered) == 0 {
		return []string{"USER"}
	}
	return filtered
}

func NewAuthService(
	keycloak *gocloak.GoCloak,
	cfg *configs.Config,
	authRepo models.UserRepository,
	Redis *redis.Client,
) models.AuthService {
	return &authService{keycloak, cfg, authRepo, Redis}
}

func (s *authService) Login(req *models.LoginReq) (*gocloak.JWT, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("login_attempts:%s", req.Username)

	token, err := s.keycloak.Login(ctx, s.cfg.KeyCloak.ClientID, s.cfg.KeyCloak.ClientSecret, s.cfg.KeyCloak.RealmName, req.Username, req.Password)
	if err != nil {
		logs.Error(err)
		errStr := err.Error()
		if strings.Contains(errStr, "Account disabled") || strings.Contains(errStr, "Account temporarily disabled") {
			return nil, errors.New("Account Locked: บัญชีของคุณถูกระงับถาวร กรุณาติดต่อผู้ดูแลระบบ")
		}

		failCount, _ := s.Redis.Incr(ctx, redisKey).Result()
		if failCount == 1 {
			s.Redis.Expire(ctx, redisKey, 10*time.Minute)
		}

		if failCount == 3 {
			return nil, errors.New("Warning: คุณใส่รหัสผิด 3 ครั้งแล้ว โปรดตรวจสอบรหัสให้ดี")
		}

		if failCount >= 5 {
			return nil, errors.New("Account Locked: คุณใส่รหัสผิดเกิน 5 ครั้ง บัญชีถูกระงับถาวร กรุณาติดต่อผู้ดูแลระบบ")
		}
		return nil, errs.NewLoginFailedError()
	}

	// 4. Case: Login Success
	s.Redis.Del(ctx, redisKey)

	// --- Sync User to Local DB ---
	userInfo, err := s.keycloak.GetUserInfo(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName)
	if err == nil {
		safeStr := func(s *string) string {
			if s != nil {
				return *s
			}
			return ""
		}

		userID := safeStr(userInfo.Sub)
		username := safeStr(userInfo.PreferredUsername)
		email := safeStr(userInfo.Email)
		firstName := safeStr(userInfo.GivenName)
		lastName := safeStr(userInfo.FamilyName)

		if userID != "" {
			// Fetch Roles from Keycloak for Sync
			var rolesList []string
			adminToken, err := s.keycloak.LoginAdmin(ctx, s.cfg.KeyCloak.AdminUsername, s.cfg.KeyCloak.AdminPassword, "master")
			if err == nil {
				keycloakRoles, err := s.keycloak.GetRealmRolesByUserID(ctx, adminToken.AccessToken, s.cfg.KeyCloak.RealmName, userID)
				if err == nil {
					for _, r := range keycloakRoles {
						if r.Name != nil {
							rolesList = append(rolesList, *r.Name)
						}
					}
				}
			}

			rolesJSON, _ := json.Marshal(filterRoles(rolesList))

			existingUserPtr, _ := s.authRepo.GetUserContext(userID) // Use Context to get preloaded dept if needed, or simple get

			// Prepare update data
			userData := &models.UserEntity{
				ID:        userID,
				Username:  username,
				Email:     email,
				FirstName: firstName,
				LastName:  lastName,
				Roles:     datatypes.JSON(rolesJSON),
				IsActive:  true,
				UpdatedAt: time.Now(),
			}

			if existingUserPtr != nil {
				// Preserve existing DeptID if not updated by Keycloak (which it isn't)
				userData.DepartmentID = existingUserPtr.DepartmentID
				s.authRepo.UpdateUser(userData)
			} else {
				userData.CreatedAt = time.Now()
				s.authRepo.CreateUser(userData)
			}
		}
	}

	return token, nil
}

func (s *authService) Register(req *models.RegisterKCReq) (string, error) {
	ctx := context.Background()

	token, err := s.keycloak.LoginAdmin(ctx, s.cfg.KeyCloak.AdminUsername, s.cfg.KeyCloak.AdminPassword, "master")
	if err != nil {
		logs.Error(err)
		return "", errs.NewUnexpectedError()
	}

	user := gocloak.User{
		FirstName: gocloak.StringP(req.FirstName),
		LastName:  gocloak.StringP(req.LastName),
		Email:     gocloak.StringP(req.Email),
		Enabled:   gocloak.BoolP(true),
		Username:  gocloak.StringP(req.Username),
	}

	var rolesToAdd []gocloak.Role
	if len(req.Roles) > 0 {
		for _, r := range req.Roles {
			roleName := strings.ToLower(r)
			role, err := s.keycloak.GetRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, roleName)
			if err != nil && (strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Could not find role")) {
				newRole := gocloak.Role{
					Name:        gocloak.StringP(roleName),
					Description: gocloak.StringP(fmt.Sprintf("Auto-generated role: %s", roleName)),
				}
				s.keycloak.CreateRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, newRole)
				role, _ = s.keycloak.GetRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, roleName)
			}
			if role != nil {
				rolesToAdd = append(rolesToAdd, *role)
			}
		}
	}

	userID, err := s.keycloak.CreateUser(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, user)
	if err != nil {
		logs.Error(err)
		return "", errs.NewUnexpectedError()
	}

	s.keycloak.SetPassword(ctx, token.AccessToken, userID, s.cfg.KeyCloak.RealmName, req.Password, false)

	if len(rolesToAdd) > 0 {
		s.keycloak.AddRealmRoleToUser(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, userID, rolesToAdd)
	}

	var deptIDPtr *uuid.UUID
	if req.DepartmentID != "" {
		// Strict Validation: Must match a defined Master Department Code
		// Logic:
		// 1. "None" is explicitly rejected as per user requirement.
		// 2. Must exist in department_entities as a Code.
		// 3. No UUID parsing or NavCode fallback.

		if strings.EqualFold(req.DepartmentID, "None") {
			return "", errors.New("Registration Failed: 'None' is not a valid department.")
		}

		dept, err := s.authRepo.GetDepartmentByCode(req.DepartmentID)
		if err != nil || dept == nil {
			return "", fmt.Errorf("Registration Failed: Department '%s' not found or invalid.", req.DepartmentID)
		}
		deptIDPtr = &dept.ID
	} else {
		return "", errors.New("Registration Failed: Department is required.")
	}

	rolesJSON, _ := json.Marshal(filterRoles(req.Roles))

	newUser := &models.UserEntity{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DepartmentID: deptIDPtr,
		Roles:        datatypes.JSON(rolesJSON),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	logs.Info(fmt.Sprintf("Registering User: %s, DeptID: %v", req.Username, deptIDPtr))
	s.authRepo.CreateUser(newUser)

	return userID, nil
}

func (s *authService) RefreshToken(refreshToken string) (*gocloak.JWT, error) {
	ctx := context.Background()
	token, err := s.keycloak.RefreshToken(ctx, refreshToken, s.cfg.KeyCloak.ClientID, s.cfg.KeyCloak.ClientSecret, s.cfg.KeyCloak.RealmName)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *authService) ChangePassword(oldPassword, newPassword string, userInfo *models.UserInfo) error {
	ctx := context.Background()
	_, err := s.keycloak.Login(ctx, s.cfg.KeyCloak.ClientID, s.cfg.KeyCloak.ClientSecret, s.cfg.KeyCloak.RealmName, userInfo.UserName, oldPassword)
	if err != nil {
		return errors.New("รหัสผ่านเดิมไม่ถูกต้อง")
	}

	adminToken, err := s.keycloak.LoginAdmin(ctx, s.cfg.KeyCloak.AdminUsername, s.cfg.KeyCloak.AdminPassword, "master")
	if err != nil {
		return errors.New("ไม่สามารถเชื่อมต่อระบบจัดการผู้ใช้ได้ (Admin Login Failed)")
	}

	err = s.keycloak.SetPassword(ctx, adminToken.AccessToken, userInfo.UserId, s.cfg.KeyCloak.RealmName, newPassword, false)
	return err
}

func (s *authService) AdminResetUserPassword(targetUserID string, newPassword string) error {
	ctx := context.Background()
	adminToken, err := s.keycloak.LoginAdmin(ctx, s.cfg.KeyCloak.AdminUsername, s.cfg.KeyCloak.AdminPassword, s.cfg.KeyCloak.RealmName)
	if err != nil {
		return errors.New("Admin connection failed")
	}
	err = s.keycloak.SetPassword(ctx, adminToken.AccessToken, targetUserID, s.cfg.KeyCloak.RealmName, newPassword, true)
	return err
}

func (s *authService) GetUserProfile(userID string) (*models.UserInfo, error) {
	// 1. Fetch User from DB
	user, err := s.authRepo.GetUserContext(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found in local db: %v", err)
	}

	// 2. Fetch Permissions
	perms, err := s.authRepo.GetUserPermissions(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user permissions: %v", err)
	}

	var roles []string
	json.Unmarshal(user.Roles, &roles)
	roles = filterRoles(roles)

	// 3. Map to UserInfo
	userInfo := &models.UserInfo{
		UserId:   user.ID,
		UserName: user.Username,
		Name:     fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Email:    user.Email,
		Roles:    roles,
	}

	if user.Department != nil {
		userInfo.Department = user.Department.Name
		userInfo.DepartmentCode = user.Department.Code
	}

	userInfo.Permissions = make([]models.UserPermissionInfo, len(perms))
	for i, p := range perms {
		userInfo.Permissions[i] = models.UserPermissionInfo{
			DepartmentCode: p.DepartmentCode,
			Role:           p.Role,
			IsActive:       p.IsActive != nil && *p.IsActive,
		}
	}

	userInfo.Roles = mergeActiveRoles(userInfo.Roles, userInfo.Permissions)

	return userInfo, nil
}

func (s *authService) ListUsersForAdmin(optional map[string]interface{}, page, size int) ([]models.UserInfo, int, error) {
	// Force Admin Visibility
	optional["visibility_role"] = "ADMIN"
	delete(optional, "visibility_allowed_depts")

	return s.listUsersBase(optional, page, size)
}

func (s *authService) ListUsersForManagement(optional map[string]interface{}, page, size int) ([]models.UserInfo, int, error) {
	// Respect the optional filters (Already set by Controller's smart logic)
	return s.listUsersBase(optional, page, size)
}

func (s *authService) listUsersBase(optional map[string]interface{}, page, size int) ([]models.UserInfo, int, error) {
	offset := (page - 1) * size
	ctx := context.Background()

	users, total, err := s.authRepo.GetAll(optional, ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}

	infos := make([]models.UserInfo, len(users))
	for i, u := range users {
		var roles []string
		json.Unmarshal(u.Roles, &roles)
		roles = filterRoles(roles)

		infos[i] = models.UserInfo{
			UserId:   u.ID,
			UserName: u.Username,
			Name:     fmt.Sprintf("%s %s", u.FirstName, u.LastName),
			Email:    u.Email,
			Roles:    roles,
		}
		if u.Department != nil {
			infos[i].Department = u.Department.Name
			infos[i].DepartmentCode = u.Department.Code
		}

		if len(u.UserPermissions) > 0 {
			infos[i].Permissions = make([]models.UserPermissionInfo, len(u.UserPermissions))
			for j, p := range u.UserPermissions {
				infos[i].Permissions[j] = models.UserPermissionInfo{
					DepartmentCode: p.DepartmentCode,
					Role:           p.Role,
					IsActive:       p.IsActive != nil && *p.IsActive,
				}
			}
			infos[i].Roles = mergeActiveRoles(infos[i].Roles, infos[i].Permissions)
			logs.Info(fmt.Sprintf("User: %s, Permissions: %d, FinalRoles: %v", u.Username, len(u.UserPermissions), infos[i].Roles))
		}
	}

	return infos, total, nil
}

func (s *authService) GetUserPermissions(userID string) ([]models.UserPermissionInfo, error) {
	perms, err := s.authRepo.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	infos := make([]models.UserPermissionInfo, len(perms))
	for i, p := range perms {
		infos[i] = models.UserPermissionInfo{
			DepartmentCode: p.DepartmentCode,
			Role:           p.Role,
			IsActive:       p.IsActive != nil && *p.IsActive,
		}
	}
	return infos, nil
}

// Helper to merge local permissions into Keycloak roles
func mergeActiveRoles(currentRoles []string, perms []models.UserPermissionInfo) []string {
	roleMap := make(map[string]bool)
	for _, r := range currentRoles {
		roleMap[strings.ToUpper(r)] = true
	}

	for _, p := range perms {
		if p.IsActive {
			role := strings.ToUpper(p.Role)
			if !roleMap[role] {
				currentRoles = append(currentRoles, role)
				roleMap[role] = true
			}
		}
	}
	return currentRoles
}

func (s *authService) UpdateUserPermissions(userID string, perms []models.UserPermissionInfo) error {
	entities := make([]models.UserPermissionEntity, len(perms))
	for i, p := range perms {
		isActive := p.IsActive // Local copy to take address of
		entities[i] = models.UserPermissionEntity{
			UserID:         userID,
			DepartmentCode: p.DepartmentCode,
			Role:           p.Role,
			IsActive:       &isActive,
		}
	}

	return s.authRepo.SetUserPermissions(userID, entities)
}
func (s *authService) ListDepartments() ([]models.DepartmentEntity, error) {
	depts, err := s.authRepo.ListDepartments()
	if err != nil {
		return nil, err
	}
	// Filter out "None" for Admin Access Control usage
	var filtered []models.DepartmentEntity
	for _, d := range depts {
		if !strings.EqualFold(d.Code, "None") {
			filtered = append(filtered, d)
		}
	}
	return filtered, nil
}

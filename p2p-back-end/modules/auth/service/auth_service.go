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
		"ADMIN":       true,
		"SUPER_ADMIN": true,
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
		} 
		// else {
			// Optionally keep other roles but this might include groups like 'IT'
			// The user said: "System role should only be admin, owner, delegate"
			// So I will strictly filter to these.
		// }
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

func (s *authService) Login(ctx context.Context, req *models.LoginReq) (*gocloak.JWT, error) {
	redisKey := fmt.Sprintf("login_attempts:%s", req.Username)

	token, err := s.keycloak.Login(ctx, s.cfg.KeyCloak.ClientID, s.cfg.KeyCloak.ClientSecret, s.cfg.KeyCloak.RealmName, req.Username, req.Password)
	if err != nil {
		logs.Error(err)
		errStr := err.Error()
		if strings.Contains(errStr, "Account disabled") || strings.Contains(errStr, "Account temporarily disabled") {
			return nil, errors.New("account Locked: บัญชีของคุณถูกระงับถาวร กรุณาติดต่อผู้ดูแลระบบ")
		}

		failCount, _ := s.Redis.Incr(ctx, redisKey).Result()
		if failCount == 1 {
			s.Redis.Expire(ctx, redisKey, 10*time.Minute)
		}

		if failCount == 3 {
			return nil, errors.New("warning: คุณใส่รหัสผิด 3 ครั้งแล้ว โปรดตรวจสอบรหัสให้ดี")
		}

		if failCount >= 5 {
			return nil, errors.New("account Locked: คุณใส่รหัสผิดเกิน 5 ครั้ง บัญชีถูกระงับถาวร กรุณาติดต่อผู้ดูแลระบบ")
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

			existingUserPtr, _ := s.authRepo.GetUserContext(ctx, userID)
			if existingUserPtr == nil {
				// Fallback: Check if user exists by username (synced users have random UUID but correct username)
				existingUserPtr, _ = s.authRepo.FindByUsername(ctx, username)
			}

			// Merge existing local roles (Admin/Owner/Delegate) with Keycloak roles
			if existingUserPtr != nil {
				var existingRoles []string
				// json.Unmarshal(existingUserPtr.Roles, &existingRoles)
				_ = json.Unmarshal(existingUserPtr.Roles, &existingRoles)
				rolesList = append(rolesList, existingRoles...)
			}
			rolesJSON, _ := json.Marshal(filterRoles(rolesList))

			// Prepare update data
			userData := &models.UserEntity{
				ID:        userID, // Use Keycloak UUID as the source of truth
				Username:  username,
				Email:     email,
				FirstName: firstName,
				LastName:  lastName,
				Roles:     datatypes.JSON(rolesJSON),
				IsActive:  true,
				Deleted:   false, // Reactivate if previously deleted
				UpdatedAt: time.Now(),
			}

			if existingUserPtr != nil {
				// If ID was different, we need to handle the change (Unification)
				if existingUserPtr.ID != userID {
					logs.Infof("[Login] Unifying user: %s, Current ID: %s -> New ID: %s", username, existingUserPtr.ID, userID)
					if err := s.authRepo.UpdateUserID(ctx, existingUserPtr.ID, userID); err != nil {
						logs.Errorf("[Login] Failed to unify user ID: %v", err)
						// Proceeding might cause key violation if we don't return, but let's try to be resilient
					}
				}

				// Preserve existing NameTh/NameEn and DeptID
				userData.NameTh = existingUserPtr.NameTh
				userData.NameEn = existingUserPtr.NameEn
				userData.CentralID = existingUserPtr.CentralID
				userData.DepartmentID = existingUserPtr.DepartmentID
				userData.CompanyID = existingUserPtr.CompanyID
				userData.SectionID = existingUserPtr.SectionID
				userData.PositionID = existingUserPtr.PositionID
				_=s.authRepo.UpdateUser(ctx, userData)
				_=s.authRepo.ReactivateUser(ctx, userID) // Force reactivation
			} else {
				userData.CreatedAt = time.Now()
				_= s.authRepo.CreateUser(ctx, userData)
			}
		}
	}

	return token, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*gocloak.JWT, error) {
	token, err := s.keycloak.RefreshToken(ctx, refreshToken, s.cfg.KeyCloak.ClientID, s.cfg.KeyCloak.ClientSecret, s.cfg.KeyCloak.RealmName)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *authService) ChangePassword(ctx context.Context, oldPassword, newPassword string, userInfo *models.UserInfo) error {
	_, err := s.keycloak.Login(ctx, s.cfg.KeyCloak.ClientID, s.cfg.KeyCloak.ClientSecret, s.cfg.KeyCloak.RealmName, userInfo.Username, oldPassword)
	if err != nil {
		return errors.New("รหัสผ่านเดิมไม่ถูกต้อง")
	}

	adminToken, err := s.keycloak.LoginAdmin(ctx, s.cfg.KeyCloak.AdminUsername, s.cfg.KeyCloak.AdminPassword, "master")
	if err != nil {
		return errors.New("ไม่สามารถเชื่อมต่อระบบจัดการผู้ใช้ได้ (Admin Login Failed)")
	}

	err = s.keycloak.SetPassword(ctx, adminToken.AccessToken, userInfo.ID, s.cfg.KeyCloak.RealmName, newPassword, false)
	return err
}

func (s *authService) AdminResetUserPassword(ctx context.Context, targetUserID string, newPassword string) error {
	adminToken, err := s.keycloak.LoginAdmin(ctx, s.cfg.KeyCloak.AdminUsername, s.cfg.KeyCloak.AdminPassword, s.cfg.KeyCloak.RealmName)
	if err != nil {
		return errors.New("admin connection failed")
	}
	err = s.keycloak.SetPassword(ctx, adminToken.AccessToken, targetUserID, s.cfg.KeyCloak.RealmName, newPassword, true)
	return err
}

func (s *authService) ProvisionUser(ctx context.Context, userInfo *models.UserInfo) (*models.UserInfo, error) {
	// 1. Double check if user exists by ID (to avoid race conditions)
	existing, err := s.authRepo.GetUserContext(ctx, userInfo.ID)
	if err == nil && existing != nil {
		return s.GetUserProfile(ctx, userInfo.ID)
	}

	// 2. Fallback: Check if user exists by username (ID mismatch case)
	existingByUsername, _ := s.authRepo.FindByUsername(ctx, userInfo.Username)
	if existingByUsername != nil {
		logs.Infof("[Provision] Unifying user: %s, Current ID: %s -> New ID: %s", userInfo.Username, existingByUsername.ID, userInfo.ID)
		if err := s.authRepo.UpdateUserID(ctx, existingByUsername.ID, userInfo.ID); err != nil {
			logs.Errorf("[Provision] Failed to unify user ID: %v", err)
			return nil, fmt.Errorf("failed to provision user (id unification failed): %v", err)
		}
		// Refresh profile and reactivate
		userData := &models.UserEntity{
			ID:        userInfo.ID,
			Username:  userInfo.Username,
			Email:     userInfo.Email,
			IsActive:  true,
			Deleted:   false, // Reactivate upon auto-provisioning
			UpdatedAt: time.Now(),
		}
		_=s.authRepo.UpdateUser(ctx, userData)
		_=s.authRepo.ReactivateUser(ctx, userInfo.ID) // FORCE REACTIVATE (bypass zero-value update issue)
		return s.GetUserProfile(ctx, userInfo.ID)
	}

	// 3. Map to Entity (Pure New User)
	rolesJson, _ := json.Marshal([]string{"USER"})
	userEntity := &models.UserEntity{
		ID:        userInfo.ID,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		IsActive:  true,
		Roles:     datatypes.JSON(rolesJson),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 4. Create in DB
	err = s.authRepo.CreateUser(ctx, userEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to provision user: %v", err)
	}

	logs.Infof("✅ Auto-provisioned new user: %s (%s)", userInfo.Username, userInfo.ID)

	// 5. Return the new profile
	return s.GetUserProfile(ctx, userInfo.ID)
}

func (s *authService) GetUserProfile(ctx context.Context, userID string) (*models.UserInfo, error) {
	// 1. Fetch User from DB
	user, err := s.authRepo.GetUserContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found in local db: %v", err)
	}

	// 2. Fetch Permissions
	perms, err := s.authRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user permissions: %v", err)
	}

	var roles []string
	// json.Unmarshal(user.Roles, &roles)
	if err := json.Unmarshal(user.Roles, &roles); err != nil {
		logs.Errorf("[Service] Failed to unmarshal user roles: %v", err)
		// หรือ return nil, err หากฟังก์ชันนี้คืนค่า error ได้
	}
	roles = filterRoles(roles)

	// 3. Map to UserInfo
	userInfo := &models.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Name:     fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		NameTh:   user.NameTh,
		Email:    user.Email,
		Roles:    roles,
	}
	// Fallback to NameTh if FirstName is empty (Synced users might only have NameTh)
	if userInfo.Name == " " && user.NameTh != "" {
		userInfo.Name = user.NameTh
	}

	if user.Department != nil {
		userInfo.Department = user.Department.Name
		userInfo.DepartmentCode = user.Department.Code
		if user.Department.CodeMap != nil && *user.Department.CodeMap != "None" {
			userInfo.MappedDepartment = *user.Department.CodeMap
		}
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

func (s *authService) ListUsersForAdmin(ctx context.Context, optional map[string]interface{}, page, size int) ([]models.UserInfo, int, error) {
	// Force Admin Visibility
	optional["visibility_role"] = "ADMIN"
	delete(optional, "visibility_allowed_depts")

	return s.listUsersBase(ctx, optional, page, size)
}

func (s *authService) ListUsersForManagement(ctx context.Context, optional map[string]interface{}, page, size int) ([]models.UserInfo, int, error) {
	// Respect the optional filters (Already set by Controller's smart logic)
	return s.listUsersBase(ctx, optional, page, size)
}

func (s *authService) listUsersBase(ctx context.Context, optional map[string]interface{}, page, size int) ([]models.UserInfo, int, error) {
	offset := (page - 1) * size

	users, total, err := s.authRepo.GetAll(optional, ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}

	infos := make([]models.UserInfo, len(users))
	for i, u := range users {
		var roles []string
		// json.Unmarshal(u.Roles, &roles)
		if err := json.Unmarshal(u.Roles, &roles); err != nil {
			logs.Errorf("[Service] Failed to unmarshal roles for user %s: %v", u.Username, err)
			// เลือกจัดการ: จะให้ roles เป็น empty หรือจะ return error ออกไปเลยก็ได้
		}
		logs.Debugf("[Service] User: %s, Raw DB Roles: %v, Permissions Count: %d", u.Username, roles, len(u.UserPermissions))
		roles = filterRoles(roles)

		infos[i] = models.UserInfo{
			ID:       u.ID,
			Username: u.Username,
			Name:     fmt.Sprintf("%s %s", u.FirstName, u.LastName),
			NameTh:   u.NameTh,
			Email:    u.Email,
			Roles:    roles,
		}
		if infos[i].Name == " " && u.NameTh != "" {
			infos[i].Name = u.NameTh
		}
		if u.Department != nil {
			infos[i].Department = u.Department.Name
			infos[i].DepartmentCode = u.Department.Code
			if u.Department.CodeMap != nil && *u.Department.CodeMap != "None" {
				infos[i].MappedDepartment = *u.Department.CodeMap
			}
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

func (s *authService) GetUserPermissions(ctx context.Context, userID string) ([]models.UserPermissionInfo, error) {
	perms, err := s.authRepo.GetUserPermissions(ctx, userID)
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

func (s *authService) UpdateUserPermissions(ctx context.Context, userID string, perms []models.UserPermissionInfo, roles []string) error {
	// Update Department Permissions Entities
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

	return s.authRepo.UpdateUserPermissionsAndRoles(ctx, userID, entities, roles)
}
func (s *authService) ListDepartments(ctx context.Context, mappedOnly bool, user *models.UserInfo) ([]models.Departments, error) {
	var depts []models.Departments
	var err error

	if mappedOnly {
		// Used for Filter Dropdown: Only show Master categories that have users/granular data
		depts, err = s.authRepo.ListDepartments(ctx)
	} else {
		// Used for Access Control: Show ALL Master categories
		var masterDepts []models.DepartmentEntity
		masterDepts, err = s.authRepo.ListMasterDepartments(ctx)
		if err == nil {
			for _, md := range masterDepts {
				depts = append(depts, models.Departments{
					Code: md.Code,
					Name: md.Name,
				})
			}
		}
	}

	if err != nil {
		return nil, err
	}

	isAdmin := false
	isOwner := false
	ownedDepts := make(map[string]bool)

	for _, r := range user.Roles {
		if strings.EqualFold(r, "ADMIN") {
			isAdmin = true
		} else if strings.EqualFold(r, "OWNER") {
			isOwner = true
		}
	}

	// Collect departments where user is an active OWNER
	if isOwner {
		for _, p := range user.Permissions {
			if p.IsActive && strings.EqualFold(p.Role, "OWNER") {
				ownedDepts[p.DepartmentCode] = true
			}
		}
	}

	filtered := []models.Departments{}
	for _, d := range depts {
		if strings.EqualFold(d.Code, "None") {
			continue
		}

		if isAdmin {
			filtered = append(filtered, d)
		} else if isOwner {
			if ownedDepts[d.Code] {
				filtered = append(filtered, d)
			}
		}
	}
	return filtered, nil
}

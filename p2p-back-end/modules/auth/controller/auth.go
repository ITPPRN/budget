package authcontroller

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/middlewares"
	"p2p-back-end/pkg/utils"
)

type authController struct {
	authSrv          models.AuthService
	deptSrv          models.DepartmentService
	userSrv          models.UsersService
	branchCodeMapSrv models.CompanyBranchCodeMappingService
}

func NewUserController(router fiber.Router, authSrv models.AuthService, deptSrv models.DepartmentService, userSrv models.UsersService, branchCodeMapSrv models.CompanyBranchCodeMappingService) {
	controller := &authController{
		authSrv:          authSrv,
		deptSrv:          deptSrv,
		userSrv:          userSrv,
		branchCodeMapSrv: branchCodeMapSrv,
	}
	// --- Keycloak-dependent routes DISABLED (gateway handles auth flow) ---
	// router.Post("/login", middlewares.RequireGatewaySecret(), controller.login)
	// router.Post("/login-dev-test", middlewares.RequireGatewaySecret(), controller.loginDevTest)
	// router.Post("/refresh-token", middlewares.RequireGatewaySecret(), controller.refreshToken)
	// router.Post("/change-password", middlewares.JwtAuthentication(authSrv, controller.changePassword))

	// --- ADMIN Group (Strictly System Admin) ---
	admin := router.Group("/admin")
	// Keycloak-dependent (admin reset password) DISABLED
	// admin.Post("/users/:id/reset-password", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.adminResetUserPassword, models.RoleAdmin)))
	admin.Get("/users", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.adminListUsers, models.RoleAdmin)))

	// --- MANAGE Group (Shared Visibility Management for Admin, Owner, Delegate) ---
	manage := router.Group("/manage")
	manage.Get("/users", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.manageListUsers, models.RoleAdmin, models.RoleOwner, models.RoleDelegate, models.RoleBranchDelegate)))
	manage.Get("/users/:id/permissions", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.getUserPermissions, models.RoleAdmin, models.RoleOwner)))
	manage.Put("/users/:id/permissions", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.setUserPermissions, models.RoleAdmin, models.RoleOwner)))
	manage.Get("/departments", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.listDepartments, models.RoleAdmin, models.RoleOwner, models.RoleDelegate, models.RoleBranchDelegate)))
	// Keycloak-dependent (user sync from Keycloak) DISABLED
	// manage.Post("/sync-users", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.syncUsers, models.RoleAdmin, models.RoleOwner)))

	// --- Company Branch Code Mappings (Admin only) — drives BRANCH_DELEGATE scope ---
	manage.Get("/branch-code-mappings", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.listBranchCodeMappings, models.RoleAdmin)))
	manage.Put("/branch-code-mappings", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.upsertBranchCodeMapping, models.RoleAdmin)))
	manage.Delete("/branch-code-mappings/:id", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.deleteBranchCodeMapping, models.RoleAdmin)))
	manage.Post("/branch-code-mappings/import", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.importBranchCodeMappings, models.RoleAdmin)))
	manage.Get("/branch-code-mappings/template", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.downloadBranchCodeMappingTemplate, models.RoleAdmin)))
	manage.Get("/companies", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.listCompaniesForMapping, models.RoleAdmin)))
	manage.Get("/branch-codes", middlewares.JwtAuthentication(authSrv, middlewares.RolesGuard(controller.listAvailableBranchCodes, models.RoleAdmin)))

	// Logout: keeps only cookie-clear behavior; gateway handles Keycloak session termination
	router.Post("/logout", middlewares.RequireGatewaySecret(), middlewares.JwtAuthentication(authSrv, controller.logout))
	router.Get("/profile", middlewares.JwtAuthentication(authSrv, controller.getProfile))
	router.Get("/tcf", controller.test11)
}

func (h authController) test11(c *fiber.Ctx) error {
	m := "hello"
	return responseSuccess(c, m)
}


func (h authController) login(c *fiber.Ctx) error {
	var req models.LoginReq
	if err := c.BodyParser(&req); err != nil {
		logs.Info("Invalid request: " + err.Error())
		return badReqErrResponse(c, "Invalid request format")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return badReqErrResponse(c, err.Error())
	}
	token, err := h.authSrv.Login(c.Context(), &req)
	if err != nil {
		return responseWithError(c, err)
	}

	accessCookie := new(fiber.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = token.AccessToken
	accessCookie.HTTPOnly = true
	accessCookie.Expires = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	accessCookie.SameSite = "Lax"
	c.Cookie(accessCookie)

	refreshCookie := new(fiber.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = token.RefreshToken
	refreshCookie.HTTPOnly = true
	refreshCookie.Path = "/v1/auth"
	refreshCookie.Expires = time.Now().Add(time.Duration(token.RefreshExpiresIn) * time.Second)
	refreshCookie.SameSite = "Lax"
	c.Cookie(refreshCookie)

	return responseSuccess(c, "Login successful")
}

func (h *authController) refreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No refresh token found"})
	}

	newToken, err := h.authSrv.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token", "detail": err.Error()})
	}

	accessCookie := new(fiber.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = newToken.AccessToken
	accessCookie.HTTPOnly = true
	accessCookie.Expires = time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)
	accessCookie.SameSite = "Lax"
	c.Cookie(accessCookie)

	refreshCookie := new(fiber.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = newToken.RefreshToken
	refreshCookie.HTTPOnly = true
	refreshCookie.Path = "/v1/auth"
	refreshCookie.Expires = time.Now().Add(time.Duration(newToken.RefreshExpiresIn) * time.Second)
	refreshCookie.SameSite = "Lax"
	c.Cookie(refreshCookie)

	return responseSuccess(c, "Token refreshed")
}

func (h authController) loginDevTest(c *fiber.Ctx) error {
	var req models.LoginReq
	if err := c.BodyParser(&req); err != nil {
		logs.Info("Invalid request: " + err.Error())
		return badReqErrResponse(c, "Invalid request format")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return badReqErrResponse(c, err.Error())
	}
	m, err := h.authSrv.Login(c.Context(), &req)
	if err != nil {
		return responseWithError(c, err)
	}

	return responseSuccess(c, m)
}

func (h authController) getProfile(c *fiber.Ctx, userInfo *models.UserInfo) error {
	return responseSuccess(c, userInfo)
}

func (h *authController) changePassword(c *fiber.Ctx, userInfo *models.UserInfo) error {
	req := new(models.ChangePasswordReq)
	if err := c.BodyParser(req); err != nil {
		return badReqErrResponse(c, "Invalid request format")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return badReqErrResponse(c, err.Error())
	}
	if req.NewPassword != req.ConfirmPassword {
		return badReqErrResponse(c, "รหัสผ่านใหม่และการยืนยันไม่ตรงกัน")
	}

	err := h.authSrv.ChangePassword(c.Context(), req.OldPassword, req.NewPassword, userInfo)
	if err != nil {
		if err.Error() == "รหัสผ่านเดิมไม่ถูกต้อง" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return responseWithError(c, err)
	}

	return responseSuccess(c, "เปลี่ยนรหัสผ่านสำเร็จ")
}

func (h *authController) adminResetUserPassword(c *fiber.Ctx, userInfo *models.UserInfo) error {
	targetUserID := c.Params("id")
	if targetUserID == "" {
		return badReqErrResponse(c, "User ID is required")
	}

	req := new(models.AdminResetPasswordReq)
	if err := c.BodyParser(req); err != nil {
		return badReqErrResponse(c, "Invalid request format")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return badReqErrResponse(c, err.Error())
	}

	err := h.authSrv.AdminResetUserPassword(c.Context(), targetUserID, req.NewPassword)
	if err != nil {
		return responseWithError(c, err)
	}

	return responseSuccess(c, "รีเซ็ตรหัสผ่านสำเร็จ ผู้ใช้ต้องเปลี่ยนรหัสใหม่ในการเข้าสู่ระบบครั้งถัดไป")
}

func (h authController) logout(c *fiber.Ctx, userInfo *models.UserInfo) error {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Path:     "/v1/auth",
	})

	return responseSuccess(c, "Logged out successfully")
}

func (h authController) adminListUsers(ctx *fiber.Ctx, user *models.UserInfo) error {
	page := ctx.QueryInt("page", 1)
	size := ctx.QueryInt("size", 10)
	search := ctx.Query("search", "")
	status := ctx.Query("status", "ALL")
	roleFilter := ctx.Query("role", "ALL")
	deptCode := ctx.Query("department_code", "ALL")

	optional := make(map[string]interface{})
	if search != "" {
		optional["search"] = search
	}
	if status != "" && status != "ALL" {
		optional["status"] = status
	}
	if roleFilter != "" && roleFilter != "ALL" {
		optional["role"] = roleFilter
	}
	if deptCode != "" && deptCode != "ALL" {
		optional["department_code"] = deptCode
	}
	optional["visibility_current_user_id"] = user.ID

	users, total, err := h.authSrv.ListUsersForAdmin(ctx.Context(), optional, page, size)
	if err != nil {
		return responseWithError(ctx, err)
	}

	return responseSuccess(ctx, fiber.Map{
		"users": users,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (h authController) manageListUsers(ctx *fiber.Ctx, user *models.UserInfo) error {
	logs.Info(fmt.Sprintf("DEBUG: manageListUsers for User: %s (ID: %s)", user.Username, user.ID))

	page := ctx.QueryInt("page", 1)
	size := ctx.QueryInt("size", 10)
	search := ctx.Query("search", "")
	status := ctx.Query("status", "ALL")
	roleFilter := ctx.Query("role", "ALL")
	deptCode := ctx.Query("department_code", "ALL")

	isOwner := false
	isDelegate := false
	for _, r := range user.Roles {
		if strings.EqualFold(r, models.RoleOwner) {
			isOwner = true
		} else if strings.EqualFold(r, models.RoleDelegate) || strings.EqualFold(r, models.RoleBranchDelegate) {
			isDelegate = true
		}
	}

	optional := make(map[string]interface{})
	if search != "" {
		optional["search"] = search
	}
	if status != "" && status != "ALL" {
		optional["status"] = status
	}
	if roleFilter != "" && roleFilter != "ALL" {
		optional["role"] = roleFilter
	}
	if deptCode != "" && deptCode != "ALL" {
		optional["department_code"] = deptCode
	}

	optional["visibility_current_user_id"] = user.ID

	if isOwner {
		optional["visibility_role"] = "OWNER"
	} else if isDelegate {
		optional["visibility_role"] = "DELEGATE"
	} else {
		optional["visibility_role"] = "ADMIN"
	}

	var allowedDepts []string
	for _, p := range user.Permissions {
		if p.IsActive && (strings.EqualFold(p.Role, models.RoleOwner) || strings.EqualFold(p.Role, models.RoleDelegate) || strings.EqualFold(p.Role, models.RoleBranchDelegate)) {
			allowedDepts = append(allowedDepts, p.DepartmentCode)
		}
	}

	if len(allowedDepts) > 0 {
		optional["visibility_allowed_depts"] = allowedDepts
	}

	users, total, err := h.authSrv.ListUsersForManagement(ctx.Context(), optional, page, size)
	if err != nil {
		logs.Error(fmt.Sprintf("manageListUsers Service Error: %v", err))
		return responseWithError(ctx, err)
	}

	return responseSuccess(ctx, fiber.Map{
		"users": users,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (h *authController) getUserPermissions(ctx *fiber.Ctx, user *models.UserInfo) error {
	userID := ctx.Params("id")
	perms, err := h.authSrv.GetUserPermissions(ctx.Context(), userID)
	if err != nil {
		return responseWithError(ctx, err)
	}
	return responseSuccess(ctx, perms)
}

func (h *authController) setUserPermissions(ctx *fiber.Ctx, user *models.UserInfo) error {
	userID := ctx.Params("id")

	actorIsAdmin := false
	actorIsOwner := false
	for _, r := range user.Roles {
		if strings.EqualFold(r, models.RoleAdmin) {
			actorIsAdmin = true
		} else if strings.EqualFold(r, models.RoleOwner) {
			actorIsOwner = true
		}
	}

	targetProfile, _ := h.authSrv.GetUserProfile(ctx.Context(), userID)
	if targetProfile != nil {
		isTargetAdmin := false
		isTargetOwner := false
		for _, r := range targetProfile.Roles {
			upperR := strings.ToUpper(r)
			// เปลี่ยนจาก if-else เป็น switch เพื่อความชัดเจน
			switch upperR {
			case "ADMIN":
				isTargetAdmin = true
			case "OWNER":
				isTargetOwner = true
			}
		}
		// for _, r := range targetProfile.Roles {
		// 	upperR := strings.ToUpper(r)
		// 	if upperR == "ADMIN" {
		// 		isTargetAdmin = true
		// 	} else if upperR == "OWNER" {
		// 		isTargetOwner = true
		// 	}
		// }

		if actorIsOwner {
			if isTargetAdmin || isTargetOwner {
				return forbiddenErrResponse(ctx, "Owners are not allowed to modify permissions for other Owners or Admins.")
			}
		} else if actorIsAdmin {
			// RELAXED: Admin can manage other Admins (except themselves)
			// This addresses the user's request: "แอดมินก็สามารถปิดสิทธิแอดมินกันเองได้"
			if user.ID == userID {
				return forbiddenErrResponse(ctx, "You cannot modify your own administrative permissions.")
			}
			// Note: Admins can now manage Delegates (since they have full control)
		}
	}

	var req models.UpdatePermissionsReq
	if err := ctx.BodyParser(&req); err != nil {
		// Fallback for old/array-based requests (compatible with existing code during transition)
		var oldReq []models.UserPermissionInfo
		if err2 := ctx.BodyParser(&oldReq); err2 == nil {
			req.Permissions = oldReq
		} else {
			return badReqErrResponse(ctx, "Invalid permissions format")
		}
	}

	err := h.authSrv.UpdateUserPermissions(ctx.Context(), userID, req.Permissions, req.Roles)
	if err != nil {
		return responseWithError(ctx, err)
	}
	return responseSuccess(ctx, "Permissions and roles updated successfully")
}

func (h *authController) listDepartments(ctx *fiber.Ctx, user *models.UserInfo) error {
	mappedOnly := ctx.QueryBool("mapped_only", false)
	depts, err := h.authSrv.ListDepartments(ctx.Context(), mappedOnly, user)
	if err != nil {
		return responseWithError(ctx, err)
	}
	return responseSuccess(ctx, depts)
}

func (h *authController) syncUsers(ctx *fiber.Ctx, user *models.UserInfo) error {
	if err := h.userSrv.SyncAllUsersData(ctx.Context()); err != nil {
		return responseWithError(ctx, err)
	}
	return responseSuccess(ctx, "Users sync started successfully. Departments should be restored shortly.")
}

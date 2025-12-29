package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/middlewares"
)

type authController struct {
	authSrv models.AuthService
}

func NewUserController(router fiber.Router, authSrv models.AuthService) {
	controller := &authController{authSrv: authSrv}
	router.Post("/register", controller.register)
	router.Post("/login", controller.login)
	router.Post("/login-dev-test", controller.loginDevTest)
	router.Post("/refresh-token", controller.refreshToken)
	router.Post("/change-password", middlewares.JwtAuthentication(controller.changePassword))
	router.Post("/admin/users/:id/reset-password", middlewares.JwtAuthentication(middlewares.RolesGuard(controller.adminResetUserPassword, models.RoleAdmin)))
	router.Post("/logout", middlewares.JwtAuthentication(controller.logout))
	router.Get("/profile", middlewares.JwtAuthentication(controller.getProfile))
	router.Get("/tcf", controller.test11)
}

func (h authController) test11(c *fiber.Ctx) error {
	m := "hello"
	return responseSuccess(c, m)
}

// @Summary User registration
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param register body models.RegisterReq true "Registration request"
// @Success 200 {object} models.ResponseData{data=string}
// @Failure 400 {object} models.ResponseError
// @Router /v1/auth/register [post]
func (h authController) register(c *fiber.Ctx) error {
	var req models.RegisterKCReq
	if err := c.BodyParser(&req); err != nil {
		logs.Info("Invalid request: " + err.Error())
		return badReqErrResponse(c, "Invalid request: "+err.Error())
	}
	m, err := h.authSrv.Register(&req)
	if err != nil {
		return responseWithError(c, err)
	}

	return responseSuccess(c, m)
}

// @Summary Login
// @Description Login user and set HttpOnly Cookie
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body models.LoginReq true "Login Request"
// @Success 200 {object} models.ResponseData{data=models.UserInfo} "Login success (Token in Cookie)"
// @Failure 400 {object} models.ResponseData
// @Failure 401 {object} models.ResponseData
// @Router /v1/auth/login [post]
func (h authController) login(c *fiber.Ctx) error {
	var req models.LoginReq
	if err := c.BodyParser(&req); err != nil {
		logs.Info("Invalid request: " + err.Error())
		return badReqErrResponse(c, "Invalid request: "+err.Error())
	}
	token, err := h.authSrv.Login(&req)
	if err != nil {
		return responseWithError(c, err)
	}

	// สร้าง Cookie 1: Access Token (อายุสั้น 15 นาที ตาม Keycloak)
	accessCookie := new(fiber.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = token.AccessToken
	accessCookie.HTTPOnly = true
	accessCookie.Expires = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	accessCookie.SameSite = "Lax" // หรือ "None" ถ้าใช้ HTTPS
	c.Cookie(accessCookie)

	// สร้าง Cookie 2: Refresh Token (อายุนาน 1 วัน ตาม Keycloak)
	refreshCookie := new(fiber.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = token.RefreshToken
	refreshCookie.HTTPOnly = true
	refreshCookie.Path = "/v1/auth" // ส่งมาเฉพาะตอนจะ Login/Refresh ก็พอเพื่อความปลอดภัย
	refreshCookie.Expires = time.Now().Add(time.Duration(token.RefreshExpiresIn) * time.Second)
	refreshCookie.SameSite = "Lax"
	c.Cookie(refreshCookie)

	return responseSuccess(c, "Login successful")
}

// RefreshToken godoc
// @Summary      Refresh Access Token
// @Description  ตรวจสอบ Refresh Token จาก HttpOnly Cookie และออก Token ใหม่ให้ (ฝัง Cookie กลับไปอัตโนมัติ)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        refresh_token cookie string true "Refresh Token (ส่งมาอัตโนมัติผ่าน Browser Cookie)"
// @Success      200 {object} models.ResponseData "Token refreshed successfully"
// @Failure      401 {object} models.ResponseData "Unauthorized: No token or Invalid token"
// @Router       /v1/auth/refresh-token [post]
func (h *authController) refreshToken(c *fiber.Ctx) error {
	// ดึง Refresh Token จาก Cookie
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No refresh token found"})
	}

	// เรียก Service ไปขอ Token ใหม่
	newToken, err := h.authSrv.RefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token", "detail": err.Error()})
	}

	// Update Access Token
	accessCookie := new(fiber.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = newToken.AccessToken
	accessCookie.HTTPOnly = true
	accessCookie.Expires = time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)
	accessCookie.SameSite = "Lax"
	c.Cookie(accessCookie)

	// Update Refresh Token (Keycloak จะหมุน Refresh Token ใหม่มาให้ด้วยเสมอ)
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
		return badReqErrResponse(c, "Invalid request: "+err.Error())
	}
	m, err := h.authSrv.Login(&req)
	if err != nil {
		return responseWithError(c, err)
	}

	return responseSuccess(c, m)
}

// เพิ่มฟังก์ชันนี้ลงไปในไฟล์
// @Summary Get User Profile
// @Description Get current user info from HttpOnly Cookie
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} models.ResponseData{data=models.UserInfo}
// @Router /v1/auth/profile [get]
// @Security ApiKeyAuth
func (h authController) getProfile(c *fiber.Ctx, userInfo *models.UserInfo) error {
	return responseSuccess(c, userInfo)
}

// ChangePassword godoc
// @Summary      Change Password
// @Description  เปลี่ยนรหัสผ่าน (ต้อง Login ก่อน)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        req body models.ChangePasswordReq true "Password Info"
// @Success      200 {object} models.ResponseData "Success"
// @Failure      400,401 {object} models.ResponseData
// @Router       /v1/auth/change-password [post]
func (h *authController) changePassword(c *fiber.Ctx, userInfo *models.UserInfo) error {
	// 1. รับค่าจาก Body
	req := new(models.ChangePasswordReq)
	if err := c.BodyParser(req); err != nil {
		return badReqErrResponse(c, "Invalid request format")
	}

	// 2. Validate ข้อมูล (เช็คว่า New กับ Confirm ตรงกันไหม ฯลฯ)
	// ... (ถ้ามี Validator library) ...
	if req.NewPassword != req.ConfirmPassword {
		return badReqErrResponse(c, "รหัสผ่านใหม่และการยืนยันไม่ตรงกัน")
	}

	// 3. เรียก Service
	err := h.authSrv.ChangePassword(req.OldPassword, req.NewPassword, userInfo)
	if err != nil {
		// แยก Error ว่าเป็น 400 (รหัสผิด) หรือ 500 (ระบบพัง)
		if err.Error() == "รหัสผ่านเดิมไม่ถูกต้อง" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return responseWithError(c, err)
	}

	return responseSuccess(c, "เปลี่ยนรหัสผ่านสำเร็จ")
}

// AdminResetPassword godoc
// @Summary      Admin Reset User Password (สำหรับ Admin เท่านั้น)
// @Description  เปลี่ยนรหัสผ่านให้ User อื่น (โดยไม่ต้องรู้รหัสเดิม) และบังคับเปลี่ยนใหม่ตอน Login
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id   path      string                     true  "User ID (UUID)"
// @Param        req  body      models.AdminResetPasswordReq true  "New Password"
// @Success      200  {object}  models.ResponseData "Success"
// @Failure      403  {object}  models.ResponseData "Forbidden"
// @Router       /v1/admin/users/{id}/reset-password [post]
func (h *authController) adminResetUserPassword(c *fiber.Ctx, userInfo *models.UserInfo) error {
	// 1. รับ ID ของ User ที่จะแก้จาก URL
	targetUserID := c.Params("id")
	if targetUserID == "" {
		return badReqErrResponse(c, "User ID is required")
	}

	// 2. รับรหัสใหม่จาก Body
	req := new(models.AdminResetPasswordReq)
	if err := c.BodyParser(req); err != nil {
		return badReqErrResponse(c, "Invalid request")
	}

	// 3. เรียก Service
	err := h.authSrv.AdminResetUserPassword(targetUserID, req.NewPassword)
	if err != nil {
		return responseWithError(c, err)
	}

	return responseSuccess(c, "รีเซ็ตรหัสผ่านสำเร็จ ผู้ใช้ต้องเปลี่ยนรหัสใหม่ในการเข้าสู่ระบบครั้งถัดไป")
}

// @Summary Logout
// @Description Logout user and clear all token cookie
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} models.ResponseData "Logged out successfully"
// @Router /v1/auth/logout [post]
// @Security ApiKeyAuth
func (h authController) logout(c *fiber.Ctx, userInfo *models.UserInfo) error {

	// 1. ลบ Access Token (สำคัญที่สุด)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})

	// 2. ลบ Refresh Token (สำคัญรองลงมา)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Path:     "/v1/auth", // ต้องตรงกับตอนสร้าง
	})

	// 3. (Optional) ลบ csrf_ ก็ได้ถ้าต้องการคลีนๆ
	// c.Cookie(&fiber.Cookie{
	//     Name:    "csrf_",
	//     Value:   "",
	//     Expires: time.Now().Add(-time.Hour),
	// })

	return responseSuccess(c, "Logged out successfully")
}

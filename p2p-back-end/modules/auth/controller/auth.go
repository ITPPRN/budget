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
	router.Post("/logout",middlewares.JwtAuthentication(controller.logout))
	router.Get("/profile", middlewares.JwtAuthentication(controller.getProfile))
	router.Get("/tcf", controller.test11)
	// router.Get("/user/check-by-id", middlewares.JwtAuthentication(controller.checkUserByID))
	// router.Put("/user/reset-password", middlewares.JwtAuthentication(controller.resetPassword))
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
	var req models.RegisterReq
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

// // @Summary User login
// // @Description User login with username and password
// // @Tags User
// // @Accept json
// // @Produce json
// // @Param login body models.LoginReq true "Login request"
// // @Success 200 {object} models.ResponseData{data=string}
// // @Failure 400 {object} models.ResponseError
// // @Router /v1/auth/login [post]

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
	m, err := h.authSrv.Login(&req)
	if err != nil {
		return responseWithError(c, err)
	}
	// สร้าง Cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "access_token" // ชื่อ cookie
	cookie.Value = m
	cookie.HTTPOnly = true                          // สำคัญ! ป้องกัน JavaScript เข้าถึง
	cookie.Expires = time.Now().Add(24 * time.Hour) // ตั้งเวลาให้ตรงกับอายุ Token

	// Config ความปลอดภัยเพิ่มเติม
	// cookie.Secure = true // เปิดเมื่อใช้ HTTPS (Production)
	cookie.SameSite = "Lax" // หรือ "None" ถ้า Frontend/Backend คนละโดเมนและใช้ HTTPS

	// ฝัง Cookie ลงใน Response
	c.Cookie(cookie)

	// ส่ง Response กลับไปแค่ว่าสำเร็จ (ไม่ต้องส่ง Token ใน Data แล้ว)
	return responseSuccess(c, "Login successful")

	// return responseSuccess(c, m)
}

func (h authController) loginDevTest(c *fiber.Ctx)error{
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
	// Middleware (JwtAuthentication) ทำงานเสร็จแล้ว และส่ง userInfo มาให้เรา
	// เราแค่ส่งมันกลับไปหา Frontend
	return responseSuccess(c, userInfo)
}


// @Summary Logout
// @Description Logout user and clear access_token cookie
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} models.ResponseData "Logged out successfully"
// @Router /v1/auth/logout [post]
// @Security ApiKeyAuth
func (h authController) logout(c *fiber.Ctx, userInfo *models.UserInfo) error {
    
    cookie := new(fiber.Cookie)
    cookie.Name = "access_token" 
    cookie.Value = ""
    cookie.Expires = time.Now().Add(-time.Hour) 
    cookie.HTTPOnly = true
    
    c.Cookie(cookie)

    return responseSuccess(c, "Logged out successfully")
}

// // @Summary Check if user exists
// // @Description Check if a user exists by user ID
// // @Tags User
// // @Accept json
// // @Produce json
// // @Param Authorization header string true "Bearer {token}"
// // @Success 200 {object} models.ResponseData{data=bool}
// // @Failure 400 {object} models.ResponseError
// // @Router /v1/user/check-by-id [get]
// // @Security ApiKeyAuth
// func (h authController) checkUserByID(c *fiber.Ctx, userInfo *models.UserInfo) error {

// 	m, err := h.authSrv.IsUserExistByID(userInfo.UserId)
// 	if err != nil {
// 		return responseWithError(c, err)
// 	}

// 	return responseSuccess(c, m)
// }

// // @Summary Reset Password
// // @Description Reset the password for a user
// // @Tags User
// // @Accept json
// // @Produce json
// // @Param Authorization header string true "Bearer {token}"
// // @Param req body models.ChangePasswordReq true "Change Password Request"
// // @Success 200 {string} string "Password Changed Successfully"
// // @Failure 400 {object} models.ResponseError
// // @Router /v1/user/reset-password [put]
// // @Security ApiKeyAuth
// func (h authController) resetPassword(c *fiber.Ctx, userInfo *models.UserInfo) error {

// 	var req models.ChangePasswordReq
// 	if err := c.BodyParser(&req); err != nil {
// 		logs.Info("Invalid request: " + err.Error())
// 		return badReqErrResponse(c, "Invalid request: "+err.Error())
// 	}

// 	m, err := h.authSrv.ChangePassword(&req, userInfo)
// 	if err != nil {
// 		return responseWithError(c, err)
// 	}
// 	return responseSuccess(c, m)
// }

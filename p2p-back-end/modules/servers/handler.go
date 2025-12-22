package servers

import (
	"time"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/utils"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"p2p-back-end/logs"
	_authCon "p2p-back-end/modules/auth/controller"
	_authSer "p2p-back-end/modules/auth/service"
	_authRe "p2p-back-end/modules/users/repository"
	"p2p-back-end/pkg/middlewares"
)

func (s *server) Handlers() error {

	v1 := s.App.Group("/v1")
	// Register swagger handler
	v1.Get("/swagger/*", fiberSwagger.WrapHandler)
	v1.Use(middlewares.NewCorsOriginMiddleWare())
	v1.Use(csrf.New(csrf.Config{
		// 1. Frontend ต้องส่ง Token กลับมาทาง Header นี้
		KeyLookup: "header:X-CSRF-Token",

		// 2. ชื่อ Cookie ที่จะใช้เก็บ Token (คนละตัวกับ access_token นะ)
		CookieName: "csrf_",

		// 3. ความปลอดภัยของ Cookie
		CookieSameSite: "Lax",                       // แนะนำ Lax สำหรับเว็บทั่วไป
		CookieSecure:   s.Cfg.App.Mode == "release", // True เมื่อเป็น Production (HTTPS)

		// ⚠️ สำคัญ: ต้องเป็น False เพื่อให้ Frontend (JS) อ่านค่าจาก Cookie
		// แล้วเอาไปใส่ใน Header 'X-CSRF-Token' ได้
		CookieHTTPOnly: false,

		Expiration:   1 * time.Hour,
		KeyGenerator: utils.UUIDv4, // ใช้ UUID สร้าง Token ที่เดายาก
	}))

	v1.Use(logs.LogHttp)

	if s.Cfg.App.Mode == "release" {
		s.App.Use(fiberzap.New(fiberzap.Config{Logger: logs.Logger}))
	} else {
		s.App.Use(middlewares.NewLoggerMiddleWare())
	}

	userRepo := _authRe.NewUserRepositoryDB(s.Db)
	authSrv := _authSer.NewAuthService(s.Keycloak, s.Cfg, userRepo, s.Redis)
	_authCon.NewUserController(v1.Group("/auth"), authSrv)

	s.App.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"status":      fiber.ErrInternalServerError.Message,
			"status_code": fiber.ErrInternalServerError.Code,
			"message":     "error, end point not found",
			"result":      nil,
		})
	})

	return nil
}

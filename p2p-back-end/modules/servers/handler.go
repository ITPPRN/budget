package servers

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	// "github.com/gofiber/fiber/v2/middleware/csrf"

	// fiberSwagger "github.com/swaggo/fiber-swagger"

	"p2p-back-end/logs"
	_authCon "p2p-back-end/modules/auth/controller"
	_authSer "p2p-back-end/modules/auth/service"
	_budgetCon "p2p-back-end/modules/budgets/controller"
	_budgetRe "p2p-back-end/modules/budgets/repository"
	_budgetSer "p2p-back-end/modules/budgets/service"
	_capexCon "p2p-back-end/modules/capex/controller"
	_capexRe "p2p-back-end/modules/capex/repository"
	_capexSer "p2p-back-end/modules/capex/service"
	_deptRe "p2p-back-end/modules/organization/repository"
	_deptSer "p2p-back-end/modules/organization/service"
	_ownerCon "p2p-back-end/modules/owner/controller"
	_ownerRe "p2p-back-end/modules/owner/repository"
	_ownerSer "p2p-back-end/modules/owner/service"
	_authRe "p2p-back-end/modules/users/repository"
	"p2p-back-end/pkg/middlewares"
)

func (s *server) Handlers() error {
	s.App.Use(recover.New())

	v1 := s.App.Group("/v1")
	// Register swagger handler
	// v1.Get("/swagger/*", fiberSwagger.WrapHandler)
	v1.Use(middlewares.NewCorsOriginMiddleWare())
	// v1.Use(csrf.New(csrf.Config{
	// 	// 1. Frontend ต้องส่ง Token กลับมาทาง Header นี้
	// 	KeyLookup: "header:X-CSRF-Token",

	// 	// 2. ชื่อ Cookie ที่จะใช้เก็บ Token (คนละตัวกับ access_token)
	// 	CookieName: "csrf_",

	// 	// 3. ความปลอดภัยของ Cookie
	// 	CookieSameSite: "Lax",                       // แนะนำ Lax สำหรับเว็บทั่วไป
	// 	CookieSecure:   s.Cfg.App.Mode == "release", // True เมื่อเป็น Production (HTTPS)

	// 	// ⚠️ สำคัญ: ต้องเป็น False เพื่อให้ Frontend (JS) อ่านค่าจาก Cookie
	// 	// แล้วเอาไปใส่ใน Header 'X-CSRF-Token' ได้
	// 	CookieHTTPOnly: false,

	// 	Expiration:   1 * time.Hour,
	// 	KeyGenerator: utils.UUIDv4, // ใช้ UUID สร้าง Token ที่เดายาก
	// }))

	v1.Use(logs.LogHttp)

	if s.Cfg.App.Mode == "release" {
		s.App.Use(fiberzap.New(fiberzap.Config{Logger: logs.Logger}))
	} else {
		s.App.Use(middlewares.NewLoggerMiddleWare())
	}

	// Organization / Department Module (Seeding)
	deptRepo := _deptRe.NewDepartmentRepositoryDB(s.Db)
	deptSrv := _deptSer.NewDepartmentService(deptRepo)

	// Owner Module
	ownerRepo := _ownerRe.NewOwnerRepositoryDB(s.Db)
	ownerSrv := _ownerSer.NewOwnerService(ownerRepo)

	// Chain Seeding & Auto-Sync (Sequential)
	go func() {
		logs.Info("System: Starting Department Data Seeding...")
		if err := deptSrv.ManageDepartments(); err != nil {
			logs.Error("Failed to Seed Departments: " + err.Error())
			return // Stop if seeding fails
		}
		logs.Info("System: Department Data Seeding Completed Successfully")

		// Run Auto-Sync AFTER Seeding
		logs.Info("Details: Starting Auto-Sync Owner Actuals...")
		if err := ownerSrv.AutoSyncOwnerActuals(); err != nil {
			logs.Error("Failed to Auto-Sync Owner Actuals: " + err.Error())
		} else {
			logs.Info("Details: Auto-Sync Owner Actuals Completed Successfully")
		}
	}()

	userRepo := _authRe.NewUserRepositoryDB(s.Db)
	authSrv := _authSer.NewAuthService(s.Keycloak, s.Cfg, userRepo, s.Redis)
	_authCon.NewUserController(v1.Group("/auth"), authSrv, deptSrv)

	// Budget Module
	budgetRepo := _budgetRe.NewBudgetRepositoryDB(s.Db)
	budgetSrv := _budgetSer.NewBudgetService(budgetRepo, deptSrv)
	_budgetCon.NewBudgetController(v1.Group("/budgets"), budgetSrv)

	// Capex Module
	capexRepo := _capexRe.NewCapexRepositoryDB(s.Db)
	capexSrv := _capexSer.NewCapexService(capexRepo)
	_capexCon.NewCapexController(v1, capexSrv)

	// Create group with Middleware
	ownerGroup := v1.Group("/owner")
	ownerGroup.Use(middlewares.JwtAuthentication(nil))
	_ownerCon.NewOwnerController(ownerGroup, ownerSrv)

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

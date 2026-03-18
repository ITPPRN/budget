package servers

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"p2p-back-end/logs"
	_authCon "p2p-back-end/modules/auth/controller"
	_budgetCon "p2p-back-end/modules/budgets/controller"
	_capexCon "p2p-back-end/modules/capex/controller"
	_ownerCon "p2p-back-end/modules/owner/controller"
	"p2p-back-end/pkg/middlewares"

	"github.com/gofiber/swagger"
	_ "p2p-back-end/docs"
)

func (s *server) Handlers() error {
	s.App.Use(recover.New())

	v1 := s.App.Group("/v1")
	v1.Use(middlewares.NewCorsOriginMiddleWare())

	v1.Use(logs.LogHttp)

	v1.Get("/swagger/*", swagger.HandlerDefault)

	if s.Cfg.App.Mode == "release" {
		s.App.Use(fiberzap.New(fiberzap.Config{Logger: logs.Logger}))
	} else {
		s.App.Use(middlewares.NewLoggerMiddleWare())
	}

	// --- Messaging Verification (RabbitMQ) ---
	if s.MqChannel != nil {
		logs.Info("🐰 RabbitMQ Channel is active and ready for use")
	}

	// --- Start Background Tasks (Cron) ---
	s.StartCronJob()

	// --- Controller Registration ---
	_authCon.NewUserController(v1.Group("/auth"), s.Shd.AuthService, s.Shd.DepartmentService, s.Shd.UserService)

	budgetGroup := v1.Group("/budgets")
	budgetGroup.Use(middlewares.JwtAuthentication(s.Shd.AuthService, nil))
	_budgetCon.NewBudgetController(budgetGroup, s.Shd.PLBudgetService, s.Shd.CapexService, s.Shd.ActualService, s.Shd.MasterDataService, s.Shd.DashboardService)

	_capexCon.NewCapexController(v1, s.Shd.CapexService)

	ownerGroup := v1.Group("/owner")
	ownerGroup.Use(middlewares.JwtAuthentication(s.Shd.AuthService, nil))
	_ownerCon.NewOwnerController(ownerGroup, s.Shd.OwnerService)

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

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

	// Export Modules (Admin)
	_bdEC "p2p-back-end/modules/exports/budget_detail_export/controller"
	_bdES "p2p-back-end/modules/exports/budget_detail_export/service"
	_bdER "p2p-back-end/modules/exports/budget_detail_export/repository"

	_adEC "p2p-back-end/modules/exports/actual_detail_export/controller"
	_adES "p2p-back-end/modules/exports/actual_detail_export/service"
	_adER "p2p-back-end/modules/exports/actual_detail_export/repository"

	_dbsEC "p2p-back-end/modules/exports/department_budget_status_export_admin/controller"
	_dbsES "p2p-back-end/modules/exports/department_budget_status_export_admin/service"
	_dbsER "p2p-back-end/modules/exports/department_budget_status_export_admin/repository"

	_bvaEC "p2p-back-end/modules/exports/budget_vs_actual_export_admin/controller"
	_bvaES "p2p-back-end/modules/exports/budget_vs_actual_export_admin/service"
	_bvaER "p2p-back-end/modules/exports/budget_vs_actual_export_admin/repository"

	_cdsEC "p2p-back-end/modules/exports/capex_department_status_export_admin/controller"
	_cdsES "p2p-back-end/modules/exports/capex_department_status_export_admin/service"
	_cdsER "p2p-back-end/modules/exports/capex_department_status_export_admin/repository"

	_cvaEC "p2p-back-end/modules/exports/capex_budget_vs_actual_export_admin/controller"
	_cvaES "p2p-back-end/modules/exports/capex_budget_vs_actual_export_admin/service"
	_cvaER "p2p-back-end/modules/exports/capex_budget_vs_actual_export_admin/repository"

	// Export Modules (Owner)
	_bvaoEC "p2p-back-end/modules/exports/budgetvsactual_export_owner/controller"
	_bvaoES "p2p-back-end/modules/exports/budgetvsactual_export_owner/service"
	_bvaoER "p2p-back-end/modules/exports/budgetvsactual_export_owner/repository"

	_cbeEC "p2p-back-end/modules/exports/capex_budget_export_owner/controller"
	_cbeES "p2p-back-end/modules/exports/capex_budget_export_owner/service"
	_cbeER "p2p-back-end/modules/exports/capex_budget_export_owner/repository"

	_bdeoEC "p2p-back-end/modules/exports/budget_detail_export_owner/controller"
	_bdeoES "p2p-back-end/modules/exports/budget_detail_export_owner/service"
	_bdeoER "p2p-back-end/modules/exports/budget_detail_export_owner/repository"

	_adeoEC "p2p-back-end/modules/exports/actual_detail_export_owner/controller"
	_adeoES "p2p-back-end/modules/exports/actual_detail_export_owner/service"
	_adeoER "p2p-back-end/modules/exports/actual_detail_export_owner/repository"
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
	_budgetCon.NewAuditController(budgetGroup, s.Shd.AuditService)

	_capexCon.NewCapexController(v1, s.Shd.CapexService)

	ownerGroup := v1.Group("/owner")
	ownerGroup.Use(middlewares.JwtAuthentication(s.Shd.AuthService, nil))
	_ownerCon.NewOwnerController(ownerGroup, s.Shd.OwnerService)

	// --- Export Module Initialization ---
	exportGroup := v1.Group("")
	exportGroup.Use(middlewares.JwtAuthentication(s.Shd.AuthService, nil))
	_bdEC.NewExportController(exportGroup, _bdES.NewService(_bdER.NewRepository(s.Db)))
	_adEC.NewExportController(exportGroup, _adES.NewService(_adER.NewRepository(s.Db)))
	_dbsEC.NewExportController(v1, _dbsES.NewService(_dbsER.NewRepository(s.Db)))
	_bvaEC.NewExportController(v1, _bvaES.NewService(_bvaER.NewRepository(s.Db)))
	_cdsEC.NewExportController(v1, _cdsES.NewService(_cdsER.NewRepository(s.Db)))
	_cvaEC.NewExportController(v1, _cvaES.NewService(_cvaER.NewRepository(s.Db)))

	_bvaoEC.NewExportController(ownerGroup, _bvaoES.NewService(_bvaoER.NewRepository(s.Db), s.Shd.OwnerService))
	_cbeEC.NewExportController(ownerGroup, _cbeES.NewService(_cbeER.NewRepository(s.Db), s.Shd.OwnerService))

	// Owner Detail Exports
	_bdeoEC.NewExportController(ownerGroup, _bdeoES.NewService(_bdeoER.NewRepository(s.Db), s.Shd.OwnerService))
	_adeoEC.NewExportController(ownerGroup, _adeoES.NewService(_adeoER.NewRepository(s.Db), s.Shd.OwnerService))

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

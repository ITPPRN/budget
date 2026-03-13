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
	_conCon "p2p-back-end/modules/consumer/controller"
	_conSer "p2p-back-end/modules/consumer/service"
	"p2p-back-end/modules/entities/models"
	_masterRe "p2p-back-end/modules/master/repository/local"
	_masterSrcRe "p2p-back-end/modules/master/repository/source"
	_masterSer "p2p-back-end/modules/master/service"
	_deptRe "p2p-back-end/modules/organization/repository"
	_deptSer "p2p-back-end/modules/organization/service"
	_ownerCon "p2p-back-end/modules/owner/controller"
	_ownerRe "p2p-back-end/modules/owner/repository"
	_ownerSer "p2p-back-end/modules/owner/service"
	_proRe "p2p-back-end/modules/producer/repository"
	_proSer "p2p-back-end/modules/producer/service"
	_authRe "p2p-back-end/modules/users/repository"
	"p2p-back-end/pkg/middlewares"

	amqp "github.com/rabbitmq/amqp091-go"
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

	// --- Messaging Setup (RabbitMQ) ---
	if s.Cfg.RabbitMQ.URL != "" {
		conn, err := amqp.Dial(s.Cfg.RabbitMQ.URL)
		if err != nil {
			logs.Errorf("Failed to connect to RabbitMQ: %v", err)
		} else {
			ch, err := conn.Channel()
			if err != nil {
				logs.Errorf("Failed to open a channel: %v", err)
			} else {
				s.MqChannel = ch
				logs.Info("Successfully connected to RabbitMQ")
			}
		}
	}

	// Messaging Services
	var eventProducer models.EvenProducer
	if s.MqChannel != nil {
		eventProducer = _proRe.NewEventProducer(s.MqChannel)
		s.ProducerSrv = _proSer.NewProducerService(eventProducer)
	}

	// --- Service Initialization ---

	// Master Module (Refactored from senior's code)
	masterRepo := _masterRe.NewMasterRepositoryDB(s.Db)
	// Assuming Postgres2 is the source DB for master data if configured, otherwise fallback to Db
	sourceDb := s.Db
	if s.Cfg.Postgres2.Host != "" {
		// In a real scenario, this would be a separate connection.
		// For now, we'll assume s.Db can access both or is the same for dev.
	}
	masterSrcRepo := _masterSrcRe.NewSourceMasterRepositoryDB(sourceDb)
	s.MasterSrv = _masterSer.NewMasterService(masterRepo, masterSrcRepo, s.ProducerSrv)

	// Organization / Department Module
	deptRepo := _deptRe.NewDepartmentRepositoryDB(s.Db)
	s.DeptSrv = _deptSer.NewDepartmentService(deptRepo)

	// Auth Module
	userRepo := _authRe.NewUserRepositoryDB(s.Db)
	s.AuthSrv = _authSer.NewAuthService(s.Keycloak, s.Cfg, userRepo, s.Redis)

	// Budget Module Domains
	plRepo := _budgetRe.NewPLBudgetRepository(s.Db)
	actualRepo := _budgetRe.NewActualRepository(s.Db)
	masterDataRepo := _budgetRe.NewMasterDataRepository(s.Db)
	dashRepo := _budgetRe.NewDashboardRepository(s.Db)

	s.PLBudgetSrv = _budgetSer.NewPLBudgetService(plRepo, s.DeptSrv)
	s.MasterDataSrv = _budgetSer.NewMasterDataService(masterDataRepo)
	s.DashboardSrv = _budgetSer.NewDashboardService(dashRepo, s.DeptSrv)
	s.ActualSrv = _budgetSer.NewActualService(actualRepo, s.MasterDataSrv, s.DashboardSrv, s.DeptSrv)

	// Capex Module
	capexRepo := _capexRe.NewCapexRepositoryDB(s.Db)
	s.CapexSrv = _capexSer.NewCapexService(capexRepo)

	// Owner Module
	ownerRepo := _ownerRe.NewOwnerRepository(s.Db)
	s.OwnerSrv = _ownerSer.NewOwnerService(ownerRepo, s.AuthSrv, s.CapexSrv)

	// --- Start Background Tasks (Cron) ---
	s.StartCronJob()

	// --- Start RabbitMQ Consumer ---
	if s.MqChannel != nil {
		consumerSrv := _conSer.NewConsumerService(s.AuthSrv, s.MasterSrv)
		consumerCtrl := _conCon.NewConsumerController(consumerSrv)
		go func() {
			msgs, err := s.MqChannel.Consume(
				"p2p_queue", // Queue name (Should be unique for P2P)
				"",          // consumer
				false,       // auto-ack (We use manual Ack)
				false,       // exclusive
				false,       // no-local
				false,       // no-wait
				nil,         // args
			)
			if err != nil {
				logs.Errorf("Failed to register a consumer: %v", err)
				return
			}
			for d := range msgs {
				consumerCtrl.HandleMessage(d)
			}
		}()
	}

	// --- Controller Registration ---
	_authCon.NewUserController(v1.Group("/auth"), s.AuthSrv, s.DeptSrv)

	budgetGroup := v1.Group("/budgets")
	budgetGroup.Use(middlewares.JwtAuthentication(s.AuthSrv, nil))
	_budgetCon.NewBudgetController(budgetGroup, s.PLBudgetSrv, s.CapexSrv, s.ActualSrv, s.MasterDataSrv, s.DashboardSrv)

	_capexCon.NewCapexController(v1, s.CapexSrv)


	ownerGroup := v1.Group("/owner")
	ownerGroup.Use(middlewares.JwtAuthentication(s.AuthSrv, nil))
	_ownerCon.NewOwnerController(ownerGroup, s.OwnerSrv)

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

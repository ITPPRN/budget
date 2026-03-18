package servers

import (
	_authSer "p2p-back-end/modules/auth/service"
	_budgetRe "p2p-back-end/modules/budgets/repository"
	_budgetSer "p2p-back-end/modules/budgets/service"
	_capexRe "p2p-back-end/modules/capex/repository"
	_capexSer "p2p-back-end/modules/capex/service"
	_consumeContr "p2p-back-end/modules/consumer/controller"
	"p2p-back-end/modules/entities/models"
	_masLoRepo "p2p-back-end/modules/master/repository"
	_masterSer "p2p-back-end/modules/master/service"
	_deptRe "p2p-back-end/modules/organization/repository"
	_deptSer "p2p-back-end/modules/organization/service"
	_ownerRe "p2p-back-end/modules/owner/repository"
	_ownerSer "p2p-back-end/modules/owner/service"
	_pubHan "p2p-back-end/modules/producer/controller"
	_produser "p2p-back-end/modules/producer/service"
	_userReLo "p2p-back-end/modules/users/repository/local"
	_userReSou "p2p-back-end/modules/users/repository/source"
	_userSer "p2p-back-end/modules/users/service"
)

type SharedDeps struct {
	AuthService        models.AuthService
	MasterService      models.MasterService
	PLBudgetService    models.PLBudgetService
	ActualService      models.ActualService
	MasterDataService  models.MasterDataService
	DashboardService   models.DashboardService
	CapexService       models.CapexService
	DepartmentService  models.DepartmentService
	ProducerService    models.ProducerService
	OwnerService       models.OwnerService
	ConsumerController models.ConsumerController
	UserService        models.UsersService
}

func initSharedDeps(s *server) *SharedDeps {
	// --- Messaging Services (Producer) ---
	var producerService models.ProducerService
	if s.MqChannel != nil {
		publisherHandler := _pubHan.NewEventProducer(s.MqChannel)
		producerService = _produser.NewProducerService(publisherHandler)
	}

	// --- Master Module ---
	masterRepo := _masLoRepo.NewMasterRepositoryDB(s.Db)
	masterService := _masterSer.NewMasterService(masterRepo, producerService)

	// --- Organization Module ---
	deptRepo := _deptRe.NewDepartmentRepositoryDB(s.Db)
	deptService := _deptSer.NewDepartmentService(deptRepo)

	// --- Users Module ---
	userRepo := _userReLo.NewUserRepositoryDB(s.Db)
	userReSource := _userReSou.NewSourceUsersRepositoryDB(s.Db)
	userService := _userSer.NewUsersService(userRepo, userReSource, producerService, masterRepo, deptService)

	// --- Auth Module ---
	authService := _authSer.NewAuthService(s.Keycloak, s.Cfg, userRepo, s.Redis)

	// --- Budget Module ---
	plRepo := _budgetRe.NewPLBudgetRepository(s.Db)
	actualRepo := _budgetRe.NewActualRepository(s.Db)
	masterDataRepo := _budgetRe.NewMasterDataRepository(s.Db)
	dashRepo := _budgetRe.NewDashboardRepository(s.Db)

	plBudgetService := _budgetSer.NewPLBudgetService(plRepo, deptService)
	masterDataService := _budgetSer.NewMasterDataService(masterDataRepo)
	dashboardService := _budgetSer.NewDashboardService(dashRepo, deptService)
	actualService := _budgetSer.NewActualService(actualRepo, masterDataService, dashboardService, deptService)

	// --- Capex Module ---
	capexRepo := _capexRe.NewCapexRepositoryDB(s.Db)
	capexService := _capexSer.NewCapexService(capexRepo)

	// --- Owner Module ---
	ownerRepo := _ownerRe.NewOwnerRepository(s.Db)
	ownerService := _ownerSer.NewOwnerService(ownerRepo, authService, capexService)

	// --- Consumer Module ---
	var consumerController models.ConsumerController
	if userService != nil && masterService != nil {
		consumerController = _consumeContr.NewConsumerController(userService, masterService)
	}

	return &SharedDeps{
		AuthService:        authService,
		MasterService:      masterService,
		PLBudgetService:    plBudgetService,
		ActualService:      actualService,
		MasterDataService:  masterDataService,
		DashboardService:   dashboardService,
		CapexService:       capexService,
		DepartmentService:  deptService,
		ProducerService:    producerService,
		OwnerService:       ownerService,
		ConsumerController: consumerController,
		UserService:        userService,
	}
}

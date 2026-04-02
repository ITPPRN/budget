package models

import (
	"context"
	"mime/multipart"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/shopspring/decimal"

	"p2p-back-end/modules/entities/events"
)

// --- Auth & User ---

type AuthService interface {
	Login(ctx context.Context, req *LoginReq) (*gocloak.JWT, error)
	RefreshToken(ctx context.Context, refreshToken string) (*gocloak.JWT, error)
	ChangePassword(ctx context.Context, oldPassword, newPassword string, userInfo *UserInfo) error
	AdminResetUserPassword(ctx context.Context, targetUserID string, newPassword string) error
	GetUserProfile(ctx context.Context, userID string) (*UserInfo, error)
	ProvisionUser(ctx context.Context, user *UserInfo) (*UserInfo, error)
	ListUsersForAdmin(ctx context.Context, optional map[string]interface{}, page, size int) ([]UserInfo, int, error)
	ListUsersForManagement(ctx context.Context, optional map[string]interface{}, page, size int) ([]UserInfo, int, error)
	GetUserPermissions(ctx context.Context, userID string) ([]UserPermissionInfo, error)
	UpdateUserPermissions(ctx context.Context, userID string, perms []UserPermissionInfo, roles []string) error
	ListDepartments(ctx context.Context, mappedOnly bool, user *UserInfo) ([]Departments, error)
}

type UserRepository interface {
	GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]UserEntity, int, error)
	IsUserExistByID(ctx context.Context, id string) (bool, error)
	CreateUser(ctx context.Context, user *UserEntity) error
	UpdateUser(ctx context.Context, user *UserEntity) error
	ReactivateUser(ctx context.Context, userID string) error
	GetUserContext(ctx context.Context, userID string) (*UserEntity, error)
	GetUserPermissions(ctx context.Context, userID string) ([]UserPermissionEntity, error)
	UpdateUserPermissionsAndRoles(ctx context.Context, userID string, permissions []UserPermissionEntity, roles []string) error
	UpdateUserID(ctx context.Context, oldID, newID string) error
	ListDepartments(ctx context.Context) ([]Departments, error)
	ListMasterDepartments(ctx context.Context) ([]DepartmentEntity, error)
	GetDepartmentByCode(ctx context.Context, code string) (*DepartmentEntity, error)
	GetDepartmentByNavCode(ctx context.Context, navCode string) (*DepartmentEntity, error)
	FindByUsername(ctx context.Context, username string) (*UserEntity, error)
	SyncUsers(ctx context.Context, users []UserEntity) ([]UserEntity, error)
	GetUsers(ctx context.Context, lastID uint, limit int) ([]UserEntity, error)
}

type SourceUserRepository interface {
	GetUsers(ctx context.Context, lastID uint, limit int) ([]CentralUser, error)
	FindByUsername(ctx context.Context, username string) (*CentralUser, error)
}

type UsersService interface {
	SyncAllUsersData(ctx context.Context) error
	SyncUserByUserName(ctx context.Context, username string) (*UserResponse, error)
	BroadcastAllLocalUsers(ctx context.Context) error
	SyncUsersFromEvent(ctx context.Context, users []events.UserEvent) error
}

type TokenHandler func(c *fiber.Ctx, user *UserInfo) error

// --- Master ---

type MasterService interface {
	BroadcastAllLocalCompanies(ctx context.Context) error
	BroadcastAllLocalDepartments(ctx context.Context) error
	BroadcastAllLocalSections(ctx context.Context) error
	BroadcastAllLocalPositions(ctx context.Context) error
	BroadcastAllData(ctx context.Context)
	SyncCompaniesFromEvent(ctx context.Context, companies []events.CompanyEvent) error
	SyncDepartmentsFromEvent(ctx context.Context, departments []events.DepartmentEvent) error
	SyncSectionsFromEvent(ctx context.Context, sections []events.SectionEvent) error
	SyncPositionsFromEvent(ctx context.Context, positions []events.PositionEvent) error
}

type MasterRepository interface {
	SyncCompany(ctx context.Context, companies []Companies) ([]Companies, error)
	GetCompanies(ctx context.Context, lastID uint, limit int) ([]Companies, error)
	SyncDepartment(ctx context.Context, departments []Departments) ([]Departments, error)
	GetDepartments(ctx context.Context, lastID uint, limit int) ([]Departments, error)
	SyncSection(ctx context.Context, sections []Sections) ([]Sections, error)
	GetSections(ctx context.Context, lastID uint, limit int) ([]Sections, error)
	SyncPosition(ctx context.Context, positions []Positions) ([]Positions, error)
	GetPositions(ctx context.Context, lastID uint, limit int) ([]Positions, error)

	FindCompanyUUID(ctx context.Context, centralID uint) (*uuid.UUID, error)
	FindDeptUUID(ctx context.Context, centralID uint) (*uuid.UUID, error)
	FindSectionUUID(ctx context.Context, centralID uint) (*uuid.UUID, error)
	FindPositionUUID(ctx context.Context, centralID uint) (*uuid.UUID, error)
}

// --- Messaging (RabbitMQ) ---

type EvenProducer interface {
	Producer(event events.Event) error
}

type ProducerService interface {
	UserChange(event *events.MessageUserEvent) error
	CompanyChange(event *events.MessageCompaniesEvent) error
	DepartmentChange(event *events.MessageDepartmentEvent) error
	SectionChange(event *events.MessageSectionEvent) error
	PositionChange(event *events.MessagePositionEvent) error
	UserBegin(event *events.MessageUserBeginEvent) error
	CompanyBegin(event *events.MessageCompaniesBeginEvent) error
	DepartmentBegin(event *events.MessageDepartmentBeginEvent) error
	SectionBegin(event *events.MessageSectionBeginEvent) error
	PositionBegin(event *events.MessagePositionBeginEvent) error
}

type ConsumerController interface {
	HandleMessage(d amqp.Delivery)
}

type ConsumerService interface {
	ProcessUserChange(body []byte) error
	ProcessCompanyChange(body []byte) error
	ProcessDepartmentChange(body []byte) error
	ProcessSectionChange(body []byte) error
	ProcessPositionChange(body []byte) error
}

// --- Clean Architecture Domains (Budgets) ---

// 1. PL Budget Domain
type PLBudgetRepository interface {
	WithTrx(trxHandle func(repo PLBudgetRepository) error) error
	CreateFileBudget(ctx context.Context, file *FileBudgetEntity) error
	CreateBudgetFacts(ctx context.Context, facts []BudgetFactEntity) error
	ListFileBudgets(ctx context.Context) ([]FileBudgetEntity, error)
	GetFileBudget(ctx context.Context, id string) (*FileBudgetEntity, error)
	DeleteFileBudget(ctx context.Context, id string) error
	DeleteAllBudgetFacts(ctx context.Context) error
	DeleteBudgetFactsByFileID(ctx context.Context, fileID string) error
	UpdateFileBudget(ctx context.Context, id string, filename string) error
}

type PLBudgetService interface {
	ImportBudget(ctx context.Context, file *multipart.FileHeader, userID string, versionName string) error
	SyncBudget(ctx context.Context, id string) error
	ClearBudget(ctx context.Context) error
	ListBudgetFiles(ctx context.Context) ([]FileBudgetEntity, error)
	DeleteBudgetFile(ctx context.Context, id string) error
	RenameBudgetFile(ctx context.Context, id string, newName string) error
}

// 2. Capex Domain
type CapexRepository interface {
	WithTrx(trxHandle func(repo CapexRepository) error) error
	CreateFileCapexBudget(ctx context.Context, file *FileCapexBudgetEntity) error
	CreateFileCapexActual(ctx context.Context, file *FileCapexActualEntity) error
	CreateCapexBudgetFacts(ctx context.Context, facts []CapexBudgetFactEntity) error
	CreateCapexActualFacts(ctx context.Context, facts []CapexActualFactEntity) error
	ListFileCapexBudgets(ctx context.Context) ([]FileCapexBudgetEntity, error)
	ListFileCapexActuals(ctx context.Context) ([]FileCapexActualEntity, error)
	GetFileCapexBudget(ctx context.Context, id string) (*FileCapexBudgetEntity, error)
	GetFileCapexActual(ctx context.Context, id string) (*FileCapexActualEntity, error)
	DeleteFileCapexBudget(ctx context.Context, id string) error
	DeleteFileCapexActual(ctx context.Context, id string) error
	DeleteAllCapexBudgetFacts(ctx context.Context) error
	DeleteAllCapexActualFacts(ctx context.Context) error
	DeleteCapexBudgetFactsByFileID(ctx context.Context, fileID string) error
	DeleteCapexActualFactsByFileID(ctx context.Context, fileID string) error
	UpdateFileCapexBudget(ctx context.Context, id string, filename string) error
	UpdateFileCapexActual(ctx context.Context, id string, filename string) error
	GetCapexDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*DashboardSummaryDTO, error)
}

type CapexService interface {
	ImportCapexBudget(ctx context.Context, file *multipart.FileHeader, userID string, versionName string) error
	ImportCapexActual(ctx context.Context, file *multipart.FileHeader, userID string, versionName string) error
	SyncCapexBudget(ctx context.Context, id string) error
	SyncCapexActual(ctx context.Context, id string) error
	ClearCapexBudget(ctx context.Context) error
	ClearCapexActual(ctx context.Context) error
	ListCapexBudgetFiles(ctx context.Context) ([]FileCapexBudgetEntity, error)
	ListCapexActualFiles(ctx context.Context) ([]FileCapexActualEntity, error)
	DeleteCapexBudgetFile(ctx context.Context, id string) error
	DeleteCapexActualFile(ctx context.Context, id string) error
	RenameCapexBudgetFile(ctx context.Context, id string, newName string) error
	RenameCapexActualFile(ctx context.Context, id string, newName string) error
	GetCapexDashboardSummary(ctx context.Context, filter map[string]interface{}) (*DashboardSummaryDTO, error)
}

// 3. Actuals Domain
type ActualRepository interface {
	WithTrx(trxHandle func(repo ActualRepository) error) error
	CreateActualFacts(ctx context.Context, facts []ActualFactEntity) error
	DeleteAllActualFacts(ctx context.Context) error
	DeleteActualFactsByYear(ctx context.Context, year string) error
	DeleteActualFactsByMonth(ctx context.Context, year string, month string) error
	DeleteAllActualTransactions(ctx context.Context) error
	DeleteActualTransactionsByYear(ctx context.Context, year string) error
	DeleteActualTransactionsByMonth(ctx context.Context, year string, month string) error
	GetAllAchHmwGle(ctx context.Context) ([]AchHmwGleEntity, error)
	GetAggregatedHMW(ctx context.Context, year string, months []string) ([]ActualAggregatedDTO, error)
	GetRawTransactionsHMW(ctx context.Context, year string, months []string) ([]ActualTransactionDTO, error)
	GetAllClikGle(ctx context.Context) ([]ClikGleEntity, error)
	GetAggregatedCLIK(ctx context.Context, year string, months []string) ([]ActualAggregatedDTO, error)
	GetRawTransactionsCLIK(ctx context.Context, year string, months []string) ([]ActualTransactionDTO, error)
	CreateActualTransactions(ctx context.Context, txs []ActualTransactionEntity) error
	GetRawDate(ctx context.Context) (string, error)
	RefreshDataInventory(ctx context.Context) error
}

type ActualService interface {
	SyncActuals(ctx context.Context, year string, months []string) error
	DeleteActualFacts(ctx context.Context, year string) error
	GetRawDate(ctx context.Context) (string, error)
	RefreshDataInventory(ctx context.Context) error
}

// 6. External Sync Domain (NAV/DW)
type ExternalSyncRepository interface {
	FetchHMWInBatches(ctx context.Context, year int, month int, batchSize int, handle func([]AchHmwGleEntity) error) error
	FetchCLIKInBatches(ctx context.Context, year int, month int, batchSize int, handle func([]ClikGleEntity) error) error
	UpsertHMWLocal(ctx context.Context, data []AchHmwGleEntity) error
	UpsertCLIKLocal(ctx context.Context, data []ClikGleEntity) error
}

type ExternalSyncService interface {
	SyncFromDW(ctx context.Context) error
}

// 4. Master Data Domain
type MasterDataRepository interface {
	WithTrx(trxHandle func(repo MasterDataRepository) error) error

	// Unified GL Grouping
	ListGLGroupings(ctx context.Context) ([]GlGroupingEntity, error)
	GetGLGroupingByID(ctx context.Context, id string) (*GlGroupingEntity, error)
	CreateGLGrouping(ctx context.Context, g *GlGroupingEntity) error
	UpdateGLGrouping(ctx context.Context, g *GlGroupingEntity) error
	DeleteGLGrouping(ctx context.Context, id string) error
	GetGLGroupingInfo(ctx context.Context, entity string, entityGL string, target *GlGroupingEntity) error

	// User Config
	GetUserConfigs(ctx context.Context, userID string) ([]UserConfigEntity, error)
	UpdateUserConfig(ctx context.Context, config *UserConfigEntity) error
}

type MasterDataService interface {
	GetBudgetStructureTree(ctx context.Context) (interface{}, error)
	ListGLGroupings(ctx context.Context) ([]GlGroupingEntity, error)
	GetGLGroupingByID(ctx context.Context, id string) (*GlGroupingEntity, error)
	CreateGLGrouping(ctx context.Context, g *GlGroupingEntity) error
	UpdateGLGrouping(ctx context.Context, g *GlGroupingEntity) error
	DeleteGLGrouping(ctx context.Context, id string) error
	ImportGLGrouping(ctx context.Context, file *multipart.FileHeader) error

	// User Config
	GetUserConfigs(ctx context.Context, userID string) (map[string]string, error)
	SetUserConfig(ctx context.Context, userID string, key string, value string) error
}

// 5. Dashboard Domain
type DashboardRepository interface {
	GetBudgetFilterOptions(ctx context.Context) ([]BudgetFactEntity, error)
	GetOrganizationStructure(ctx context.Context) ([]BudgetFactEntity, error)
	GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]BudgetDetailDTO, error)
	GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]ActualFactEntity, error)
	GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*DashboardSummaryDTO, error)
	GetActualYears(ctx context.Context) ([]string, error)
	GetAvailableMonths(ctx context.Context, year string) ([]string, error)
}

type DashboardService interface {
	GetFilterOptions(ctx context.Context) ([]FilterOptionDTO, error)
	GetRawFilterOptions(ctx context.Context) ([]BudgetFactEntity, error)
	GetOrganizationStructure(ctx context.Context) ([]OrganizationDTO, error)
	GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]BudgetDetailDTO, error)
	GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]ActualFactEntity, error)
	GetDashboardSummary(ctx context.Context, filter map[string]interface{}) (*DashboardSummaryDTO, error)
	GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetActualYears(ctx context.Context) ([]string, error)
	GetAvailableMonths(ctx context.Context, year string) ([]string, error)
}

// --- Owner ---

type OwnerRepository interface {
	GetBudgetFilterOptions(ctx context.Context) ([]BudgetFactEntity, error)
	GetOrganizationStructure(ctx context.Context) ([]BudgetFactEntity, error)
	GetDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*DashboardSummaryDTO, error)
	GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]ActualFactEntity, error)
	GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]BudgetFactEntity, error)
	GetActualYears(ctx context.Context) ([]string, error)
}

type OwnerService interface {
	GetDashboardSummary(ctx context.Context, user *UserInfo, filter map[string]interface{}) (*OwnerDashboardSummaryDTO, error)
	GetActualTransactions(ctx context.Context, user *UserInfo, filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetActualDetails(ctx context.Context, user *UserInfo, filter map[string]interface{}) ([]ActualFactEntity, error)
	GetBudgetDetails(ctx context.Context, user *UserInfo, filter map[string]interface{}) ([]BudgetFactEntity, error)
	GetFilterOptions(ctx context.Context, user *UserInfo) (interface{}, error)
	GetOrganizationStructure(ctx context.Context, user *UserInfo) ([]OrganizationDTO, error)
	GetOwnerFilterLists(ctx context.Context, user *UserInfo) (*OwnerFilterListsDTO, error)
	GetActualYears(ctx context.Context, user *UserInfo) ([]string, error)
	InjectPermissions(ctx context.Context, user *UserInfo, filter map[string]interface{}) map[string]interface{}
}

// 7. Audit Log Domain
type AuditRepository interface {
	WithTrx(trxHandle func(repo AuditRepository) error) error
	SaveAuditLog(ctx context.Context, log *AuditLogEntity) error
	SaveRejectedItems(ctx context.Context, items []AuditLogRejectedItemEntity) error
	GetAuditLogs(ctx context.Context, filter map[string]interface{}) ([]AuditLogEntity, error)
	GetRejectedItemsByLogID(ctx context.Context, logID string) ([]AuditLogRejectedItemEntity, error)
	GetTransactionsByIDs(ctx context.Context, ids []uuid.UUID) ([]ActualTransactionEntity, error)
	GetTransactionsByFilter(ctx context.Context, filter map[string]interface{}) ([]ActualTransactionEntity, error)
	UpdateTransactionsStatus(ctx context.Context, ids []uuid.UUID, status string) error
	MarkRestAsComplete(ctx context.Context, department, year, month string, excludedIDs []uuid.UUID) error
}

type AuditService interface {
	Approve(ctx context.Context, user *UserInfo, payload map[string]interface{}) error
	Report(ctx context.Context, user *UserInfo, payload map[string]interface{}) error
	ListLogs(ctx context.Context, filter map[string]interface{}) ([]AuditLogEntity, error)
	GetRejectedItemDetails(ctx context.Context, logID string) ([]AuditLogRejectedItemEntity, error)
	GetReportableTransactions(ctx context.Context, user *UserInfo, payload map[string]interface{}) ([]ActualTransactionEntity, error)
}

// --- Organization ---

type DepartmentService interface {
	ManageDepartments(ctx context.Context) error
	GetMasterDepartment(ctx context.Context, navCode, entity string) (*DepartmentEntity, error)
}

// --- DTOs ---

type FilterOptionDTO struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Level    int               `json:"level"`
	Children []FilterOptionDTO `json:"children,omitempty"`
}

type OrganizationDTO struct {
	Entity   string      `json:"entity"`
	Branches []BranchDTO `json:"branches"`
}

type BranchDTO struct {
	Name        string   `json:"name"`
	Departments []string `json:"departments"`
}

type DashboardSummaryDTO struct {
	TotalBudget     decimal.Decimal     `json:"total_budget"`
	TotalActual     decimal.Decimal     `json:"total_actual"`
	DepartmentData  []DepartmentStatDTO `json:"department_data"`
	ChartData       []MonthlyStatDTO    `json:"chart_data"`
	TopExpenses     []TopExpenseDTO     `json:"top_expenses"`
	TotalCount      int64               `json:"total_count"`
	Page            int                 `json:"page"`
	Limit           int                 `json:"limit"`
	OverBudgetCount int                 `json:"over_budget_count"`
	NearLimitCount  int                 `json:"near_limit_count"`
}

type DepartmentStatDTO struct {
	Department string          `json:"department"`
	Budget     decimal.Decimal `json:"budget"`
	Actual     decimal.Decimal `json:"actual"`
}

type MonthlyStatDTO struct {
	Month  string          `json:"month"`
	Budget decimal.Decimal `json:"budget"`
	Actual decimal.Decimal `json:"actual"`
}

type ActualAggregatedDTO struct {
	Company       string          `json:"company" gorm:"column:company"`
	Branch        string          `json:"branch" gorm:"column:branch"`
	Department    string          `json:"department" gorm:"column:department"`
	GLAccountNo   string          `json:"gl_account_no" gorm:"column:gl_account_no"`
	GLAccountName string          `json:"gl_account_name" gorm:"column:gl_account_name"`
	Month         string          `json:"month" gorm:"column:month"`
	VendorName    string          `json:"vendor_name" gorm:"column:vendor_name"`
	TotalAmount   decimal.Decimal `json:"total_amount" gorm:"column:total_amount"`
}

type ActualTransactionDTO struct {
	Source        string          `json:"source" gorm:"column:source"`
	PostingDate   string          `json:"posting_date" gorm:"column:posting_date"`
	DocNo         string          `json:"document_no" gorm:"column:doc_no"`
	Vendor        string          `json:"vendor" gorm:"column:vendor"`
	Description   string          `json:"description" gorm:"column:description"`
	EntityGL      string          `json:"entity_gl" gorm:"column:entity_gl"`
	ConsoGL       string          `json:"conso_gl" gorm:"column:conso_gl"`
	GLAccountName string          `json:"gl_account_name" gorm:"column:gl_account_name"`
	Amount        decimal.Decimal `json:"amount" gorm:"column:amount"`
	Department    string          `json:"department" gorm:"column:department"`
	Company       string          `json:"company" gorm:"column:company"`
	Branch        string          `json:"branch" gorm:"column:branch"`
	Status        string          `json:"status" gorm:"column:status"`
}

type BudgetDetailDTO struct {
	ConsoGL       string            `json:"conso_gl"`
	GLName        string            `json:"gl_name"`
	YearTotal     decimal.Decimal   `json:"year_total"`
	BudgetAmounts []BudgetAmountDTO `json:"budget_amounts"`
}

type BudgetAmountDTO struct {
	Month  string          `json:"month"`
	Amount decimal.Decimal `json:"amount"`
}

type PaginatedActualTransactionDTO struct {
	Data       []ActualTransactionDTO `json:"data"`
	TotalCount int64                  `json:"total_count"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
}

type OwnerDashboardSummaryDTO struct {
	DashboardSummaryDTO
	CapexBudget decimal.Decimal `json:"capex_budget"`
	CapexActual decimal.Decimal `json:"capex_actual"`
}

type OwnerFilterListsDTO struct {
	Companies []string `json:"companies"`
	Branches  []string `json:"branches"`
	Years     []string `json:"years"`
}

type TopExpenseDTO struct {
	Name   string          `json:"name"`
	Amount decimal.Decimal `json:"amount"`
}

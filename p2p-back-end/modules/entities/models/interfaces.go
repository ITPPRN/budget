package models

import (
	"context"
	"mime/multipart"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/shopspring/decimal"

	"p2p-back-end/modules/entities/events"
)

// --- Auth & User ---

type AuthService interface {
	Register(*RegisterKCReq) (string, error)
	Login(req *LoginReq) (*gocloak.JWT, error)
	RefreshToken(refreshToken string) (*gocloak.JWT, error)
	ChangePassword(oldPassword, newPassword string, userInfo *UserInfo) error
	AdminResetUserPassword(targetUserID string, newPassword string) error
	GetUserProfile(userID string) (*UserInfo, error)
	ListUsersForAdmin(optional map[string]interface{}, page, size int) ([]UserInfo, int, error)
	ListUsersForManagement(optional map[string]interface{}, page, size int) ([]UserInfo, int, error)
	GetUserPermissions(userID string) ([]UserPermissionInfo, error)
	UpdateUserPermissions(userID string, perms []UserPermissionInfo) error
	ListDepartments(user *UserInfo) ([]DepartmentEntity, error)
}

type UserRepository interface {
	GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]UserEntity, int, error)
	IsUserExistByID(string) (bool, error)
	CreateUser(user *UserEntity) error
	UpdateUser(user *UserEntity) error
	GetUserContext(userID string) (*UserEntity, error)
	GetUserPermissions(userID string) ([]UserPermissionEntity, error)
	SetUserPermissions(userID string, permissions []UserPermissionEntity) error
	ListDepartments() ([]DepartmentEntity, error)
	GetDepartmentByCode(code string) (*DepartmentEntity, error)
	GetDepartmentByNavCode(navCode string) (*DepartmentEntity, error)
}

type TokenHandler func(c *fiber.Ctx, user *UserInfo) error

// --- Master ---

type MasterService interface {
	SyncAllData()
	SyncAllCompaniesData() error
	SyncAllDepartmentData() error
	SyncAllSectionData() error
	SyncAllPositionData() error
	BroadcastAllLocalCompanies() error
	BroadcastAllLocalDepartments() error
	BroadcastAllLocalSections() error
	BroadcastAllLocalPositions() error
	BroadcastAllData()
}

type MasterRepository interface {
	SyncCompany(companies []Companies) ([]Companies, error)
	GetCompanies(lastID uint, limit int) ([]Companies, error)
	SyncDepartment(departments []Departments) ([]Departments, error)
	GetDepartments(lastID uint, limit int) ([]Departments, error)
	SyncSection(sections []Sections) ([]Sections, error)
	GetSections(lastID uint, limit int) ([]Sections, error)
	SyncPosition(positions []Positions) ([]Positions, error)
	GetPositions(lastID uint, limit int) ([]Positions, error)
}

type SourceMasterRepository interface {
	GetCompanies(lastID uint, limit int) ([]CentralCompany, error)
	GetDepartments(lastID uint, limit int) ([]CentralDepartment, error)
	GetSections(lastID uint, limit int) ([]CentralSection, error)
	GetPositions(lastID uint, limit int) ([]CentralPosition, error)
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
	RequestCompanySync() error
	RequestDepartmentSync() error
	RequestUserSync() error
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
	CreateFileBudget(file *FileBudgetEntity) error
	CreateBudgetFacts(facts []BudgetFactEntity) error
	ListFileBudgets() ([]FileBudgetEntity, error)
	GetFileBudget(id string) (*FileBudgetEntity, error)
	DeleteFileBudget(id string) error
	DeleteAllBudgetFacts() error
	DeleteBudgetFactsByFileID(fileID string) error
	UpdateFileBudget(id string, filename string) error
}

type PLBudgetService interface {
	ImportBudget(file *multipart.FileHeader, userID string, versionName string) error
	SyncBudget(id string) error
	ClearBudget() error
	ListBudgetFiles() ([]FileBudgetEntity, error)
	DeleteBudgetFile(id string) error
	RenameBudgetFile(id string, newName string) error
}

// 2. Capex Domain
type CapexRepository interface {
	WithTrx(trxHandle func(repo CapexRepository) error) error
	CreateFileCapexBudget(file *FileCapexBudgetEntity) error
	CreateFileCapexActual(file *FileCapexActualEntity) error
	CreateCapexBudgetFacts(facts []CapexBudgetFactEntity) error
	CreateCapexActualFacts(facts []CapexActualFactEntity) error
	ListFileCapexBudgets() ([]FileCapexBudgetEntity, error)
	ListFileCapexActuals() ([]FileCapexActualEntity, error)
	GetFileCapexBudget(id string) (*FileCapexBudgetEntity, error)
	GetFileCapexActual(id string) (*FileCapexActualEntity, error)
	DeleteFileCapexBudget(id string) error
	DeleteFileCapexActual(id string) error
	DeleteAllCapexBudgetFacts() error
	DeleteAllCapexActualFacts() error
	DeleteCapexBudgetFactsByFileID(fileID string) error
	DeleteCapexActualFactsByFileID(fileID string) error
	UpdateFileCapexBudget(id string, filename string) error
	UpdateFileCapexActual(id string, filename string) error
	GetCapexDashboardAggregates(filter map[string]interface{}) (*DashboardSummaryDTO, error)
}

type CapexService interface {
	ImportCapexBudget(file *multipart.FileHeader, userID string, versionName string) error
	ImportCapexActual(file *multipart.FileHeader, userID string, versionName string) error
	SyncCapexBudget(id string) error
	SyncCapexActual(id string) error
	ClearCapexBudget() error
	ClearCapexActual() error
	ListCapexBudgetFiles() ([]FileCapexBudgetEntity, error)
	ListCapexActualFiles() ([]FileCapexActualEntity, error)
	DeleteCapexBudgetFile(id string) error
	DeleteCapexActualFile(id string) error
	RenameCapexBudgetFile(id string, newName string) error
	RenameCapexActualFile(id string, newName string) error
	GetCapexDashboardSummary(filter map[string]interface{}) (*DashboardSummaryDTO, error)
}

// 3. Actuals Domain
type ActualRepository interface {
	WithTrx(trxHandle func(repo ActualRepository) error) error
	CreateActualFacts(facts []ActualFactEntity) error
	DeleteAllActualFacts() error
	DeleteActualFactsByYear(year string) error
	DeleteActualFactsByMonth(year string, month string) error
	DeleteAllActualTransactions() error
	DeleteActualTransactionsByYear(year string) error
	DeleteActualTransactionsByMonth(year string, month string) error
	GetAllAchHmwGle() ([]AchHmwGleEntity, error)
	GetAggregatedHMW(year string, months []string) ([]ActualAggregatedDTO, error)
	GetRawTransactionsHMW(year string, months []string) ([]ActualTransactionDTO, error)
	GetAllClikGle() ([]ClikGleEntity, error)
	GetAggregatedCLIK(year string, months []string) ([]ActualAggregatedDTO, error)
	GetRawTransactionsCLIK(year string, months []string) ([]ActualTransactionDTO, error)
	CreateActualTransactions(txs []ActualTransactionEntity) error
	GetRawDate() (string, error)
}

type ActualService interface {
	SyncActuals(year string, months []string) error
	DeleteActualFacts(year string) error
	GetRawDate() (string, error)
}

// 4. Master Data Domain
type MasterDataRepository interface {
	WithTrx(trxHandle func(repo MasterDataRepository) error) error
	ListGLMappings() ([]GlMappingEntity, error)
	GetGLMappingByID(id string) (*GlMappingEntity, error)
	CreateGLMapping(m *GlMappingEntity) error
	UpdateGLMapping(m *GlMappingEntity) error
	DeleteGLMapping(id string) error
	GetGLInfo(entity string, entityGL string, target *GlMappingEntity) error
	CheckExactGLMapping(entity, entityGL, consoGL, accountName string) (bool, error)

	GetBudgetStructure() ([]BudgetStructureEntity, error)
	GetBudgetStructureByID(id uint) (*BudgetStructureEntity, error)
	CreateBudgetStructure(entity *BudgetStructureEntity) error
	UpdateBudgetStructure(entity *BudgetStructureEntity) error
	DeleteBudgetStructure(id uint) error
	InsertBudgetStructures(entities []BudgetStructureEntity) error
	DeleteAllBudgetStructures() error

	// User Config
	GetUserConfigs(userID string) ([]UserConfigEntity, error)
	UpdateUserConfig(config *UserConfigEntity) error
}

type MasterDataService interface {
	ListGLMappings() ([]GlMappingEntity, error)
	GetGLMappingByID(id string) (*GlMappingEntity, error)
	CreateGLMapping(m *GlMappingEntity) error
	UpdateGLMapping(m *GlMappingEntity) error
	DeleteGLMapping(id string) error
	ImportGLMapping(file *multipart.FileHeader) error

	GetBudgetStructureTree() (interface{}, error)
	ListBudgetStructure() ([]BudgetStructureEntity, error)
	GetBudgetStructureByID(id uint) (*BudgetStructureEntity, error)
	CreateBudgetStructure(entity *BudgetStructureEntity) error
	UpdateBudgetStructure(entity *BudgetStructureEntity) error
	DeleteBudgetStructure(id uint) error

	// User Config
	GetUserConfigs(userID string) (map[string]string, error)
	SetUserConfig(userID string, key string, value string) error
}

// 5. Dashboard Domain
type DashboardRepository interface {
	GetBudgetFilterOptions() ([]BudgetFactEntity, error)
	GetOrganizationStructure() ([]BudgetFactEntity, error)
	GetBudgetDetails(filter map[string]interface{}) ([]BudgetDetailDTO, error)
	GetActualDetails(filter map[string]interface{}) ([]ActualFactEntity, error)
	GetActualTransactions(filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetDashboardAggregates(filter map[string]interface{}) (*DashboardSummaryDTO, error)
	GetActualYears() ([]string, error)
}

type DashboardService interface {
	GetFilterOptions() ([]FilterOptionDTO, error)
	GetRawFilterOptions() ([]BudgetFactEntity, error)
	GetOrganizationStructure() ([]OrganizationDTO, error)
	GetBudgetDetails(filter map[string]interface{}) ([]BudgetDetailDTO, error)
	GetActualDetails(filter map[string]interface{}) ([]ActualFactEntity, error)
	GetDashboardSummary(filter map[string]interface{}) (*DashboardSummaryDTO, error)
	GetActualTransactions(filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetActualYears() ([]string, error)
}

// --- Owner ---

type OwnerRepository interface {
	GetBudgetFilterOptions() ([]BudgetFactEntity, error)
	GetOrganizationStructure() ([]BudgetFactEntity, error)
	GetDashboardAggregates(filter map[string]interface{}) (*DashboardSummaryDTO, error)
	GetActualTransactions(filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetActualDetails(filter map[string]interface{}) ([]ActualFactEntity, error)
	GetBudgetDetails(filter map[string]interface{}) ([]BudgetFactEntity, error)
	GetActualYears() ([]string, error)
}

type OwnerService interface {
	GetDashboardSummary(user *UserInfo, filter map[string]interface{}) (*OwnerDashboardSummaryDTO, error)
	GetActualTransactions(user *UserInfo, filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetActualDetails(user *UserInfo, filter map[string]interface{}) ([]ActualFactEntity, error)
	GetBudgetDetails(user *UserInfo, filter map[string]interface{}) ([]BudgetFactEntity, error)
	GetFilterOptions(user *UserInfo) (interface{}, error)
	GetOrganizationStructure(user *UserInfo) ([]OrganizationDTO, error)
	GetOwnerFilterLists(user *UserInfo) (*OwnerFilterListsDTO, error)
}

// --- Organization ---

type DepartmentService interface {
	ManageDepartments() error
	GetMasterDepartment(navCode, entity string) (*DepartmentEntity, error)
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
	TotalBudget     float64             `json:"total_budget"`
	TotalActual     float64             `json:"total_actual"`
	DepartmentData  []DepartmentStatDTO `json:"department_data"`
	ChartData       []MonthlyStatDTO    `json:"chart_data"`
	TotalCount      int64               `json:"total_count"`
	Page            int                 `json:"page"`
	Limit           int                 `json:"limit"`
	OverBudgetCount int                 `json:"over_budget_count"`
	NearLimitCount  int                 `json:"near_limit_count"`
}

type DepartmentStatDTO struct {
	Department string  `json:"department"`
	Budget     float64 `json:"budget"`
	Actual     float64 `json:"actual"`
}

type MonthlyStatDTO struct {
	Month  string  `json:"month"`
	Budget float64 `json:"budget"`
	Actual float64 `json:"actual"`
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
	GLAccountNo   string          `json:"gl_account_no" gorm:"column:gl_account_no"`
	GLAccountName string          `json:"gl_account_name" gorm:"column:gl_account_name"`
	Amount        decimal.Decimal `json:"amount" gorm:"column:amount"`
	Department    string          `json:"department" gorm:"column:department"`
	Company       string          `json:"company" gorm:"column:company"`
	Branch        string          `json:"branch" gorm:"column:branch"`
}

type BudgetDetailDTO struct {
	ConsoGL       string            `json:"conso_gl"`
	GLName        string            `json:"gl_name"`
	YearTotal     float64           `json:"year_total"`
	BudgetAmounts []BudgetAmountDTO `json:"budget_amounts"`
}

type BudgetAmountDTO struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

type PaginatedActualTransactionDTO struct {
	Data       []ActualTransactionDTO `json:"data"`
	TotalCount int64                  `json:"total_count"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
}

type OwnerDashboardSummaryDTO struct {
	DashboardSummaryDTO
	TopExpenses []TopExpenseDTO `json:"top_expenses"`
	CapexBudget float64         `json:"capex_budget"`
	CapexActual float64         `json:"capex_actual"`
}

type OwnerFilterListsDTO struct {
	Companies []string `json:"companies"`
	Branches  []string `json:"branches"`
	Years     []string `json:"years"`
}

type TopExpenseDTO struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

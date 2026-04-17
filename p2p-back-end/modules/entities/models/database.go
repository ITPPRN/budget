package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UserEntity struct {
	ID                 string                 `gorm:"primaryKey;type:varchar(36);not null" json:"id"`
	CentralID          uint                   `gorm:"index" json:"central_id"`
	Username           string                 `gorm:"uniqueIndex;not null" json:"username"`
	Email              string                 `gorm:"index" json:"email"`
	FirstName          string                 `json:"first_name"`
	LastName           string                 `json:"last_name"`
	NameTh             string                 `json:"name_th"`
	NameEn             string                 `json:"name_en"`
	SignatureURL       string                 `json:"signature_url"`
	PdpaAcknowledgedAt *time.Time             `json:"pdpa_acknowledged_at"`
	IsActive           bool                   `gorm:"default:true" json:"is_active"`
	CompanyID          *uuid.UUID             `gorm:"type:uuid;index" json:"company_id"`
	Company            *Companies             `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	DepartmentID       *uuid.UUID             `gorm:"type:uuid;index" json:"department_id"`
	Department         *Departments           `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	SectionID          *uuid.UUID             `gorm:"type:uuid;index" json:"section_id"`
	Section            *Sections              `gorm:"foreignKey:SectionID" json:"section,omitempty"`
	PositionID         *uuid.UUID             `gorm:"type:uuid;index" json:"position_id"`
	Position           *Positions             `gorm:"foreignKey:PositionID" json:"position,omitempty"`
	Roles              datatypes.JSON         `gorm:"type:jsonb" json:"roles"`
	UserPermissions    []UserPermissionEntity `gorm:"foreignKey:UserID" json:"user_permissions,omitempty"`

	UpdateBy  *string        `gorm:"type:varchar(36)" json:"update_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Deleted   bool           `gorm:"default:false;not null" json:"deleted"`
}

func (UserEntity) TableName() string { return "user_entities" }

type Companies struct {
	ID           uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CentralID    uint           `gorm:"uniqueIndex" json:"central_id"`
	Name         string         `json:"name"`
	BranchName   string         `json:"branch_name"`
	BranchNameEn string         `json:"branch_name_en"`
	BranchNo     string         `json:"branch_no"`
	Address      string         `json:"address"`
	TaxID        string         `gorm:"column:taxid" json:"taxid"`
	Province     string         `json:"province"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Companies) TableName() string { return "companies" }

type Departments struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CentralID uint           `gorm:"uniqueIndex" json:"central_id"`
	Name      string         `json:"name"`
	Code      string         `json:"code"`
	CodeMap   *string        `gorm:"index" json:"code_map"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Departments) TableName() string { return "departments" }

type Sections struct {
	ID           uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CentralID    uint           `gorm:"uniqueIndex" json:"central_id"`
	Name         string         `json:"name"`
	Code         string         `json:"code"`
	DepartmentID *uuid.UUID     `gorm:"type:uuid;index" json:"department_id"`
	Department   *Departments   `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Sections) TableName() string { return "sections" }

type Positions struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CentralID uint           `gorm:"uniqueIndex" json:"central_id"`
	Name      string         `json:"name"`
	Code      string         `json:"code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Positions) TableName() string { return "positions" }

// [New] Explicit User Permissions managed by Admin
type UserPermissionEntity struct {
	ID             uuid.UUID   `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID         string      `gorm:"type:varchar(36);index" json:"user_id"`
	User           *UserEntity `gorm:"foreignKey:UserID" json:"-"`
	DepartmentCode string      `gorm:"index" json:"department_code"`
	Role           string      `json:"role"` // e.g. "OWNER", "DELEGATE"
	IsActive       *bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func (UserPermissionEntity) TableName() string { return "user_permission_entities" }

// --- Department Master & Mapping ---

// DepartmentEntity now represents MASTER Departments only (e.g., "ACC", "IT")
type DepartmentEntity struct {
	gorm.Model
	ID        uuid.UUID   `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Code      string      `gorm:"uniqueIndex;not null" json:"code"` // Master Code
	Name      string      `gorm:"not null" json:"name"`
	UpdateBy  *uuid.UUID  `gorm:"type:uuid" json:"update_by"`
	ManagerID *string     `gorm:"type:varchar(36)" json:"manager_id"`
	Manager   *UserEntity `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
}

func (DepartmentEntity) TableName() string { return "department_entities" }

// DepartmentMappingEntity maps specific Entity+NavCode to a Master Department
type DepartmentMappingEntity struct {
	ID           uuid.UUID         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	DepartmentID uuid.UUID         `gorm:"type:uuid;index;not null" json:"department_id"` // FK to Master
	Entity       string            `gorm:"index:idx_dept_mapping_nav_entity;not null" json:"entity"`
	NavCode      string            `gorm:"index:idx_dept_mapping_nav_entity;not null" json:"nav_code"`
	Department   *DepartmentEntity `gorm:"foreignKey:DepartmentID" json:"department"`
}

func (DepartmentMappingEntity) TableName() string { return "department_mapping_entities" }

// GlGroupingEntity represents the UNIFIED GL Mapping and Hierarchy structure
type GlGroupingEntity struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Entity      string    `gorm:"index:idx_gl_grouping_entity_gl;not null" json:"entity"`
	EntityGL    string    `gorm:"index:idx_gl_grouping_entity_gl;not null" json:"entity_gl"`
	ConsoGL     string    `gorm:"index;not null" json:"conso_gl"`
	AccountName string    `json:"account_name"`
	Group1      string    `gorm:"type:varchar(255)" json:"group1"`
	Group2      string    `gorm:"type:varchar(255)" json:"group2"`
	Group3      string    `gorm:"type:varchar(255)" json:"group3"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (GlGroupingEntity) TableName() string {
	return "gl_grouping_entities"
}

// --- FILE ENTITIES (3 Distinct Tables) ---
type FileBudgetEntity struct {
	gorm.Model
	ID       uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	FileName string         `json:"file_name"`
	UploadAt time.Time      `json:"upload_at"`
	UserID   string         `gorm:"type:varchar(36)" json:"user_id"`
	User     *UserEntity    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Data     datatypes.JSON `gorm:"type:jsonb" json:"-"` // Store parsed Excel content
}

func (FileBudgetEntity) TableName() string { return "file_budget_entities" }

type FileCapexBudgetEntity struct {
	gorm.Model
	ID       uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	FileName string         `json:"file_name"`
	Year     string         `json:"year"` // Added Year
	UploadAt time.Time      `json:"upload_at"`
	UserID   string         `gorm:"type:varchar(36)" json:"user_id"`
	User     *UserEntity    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Data     datatypes.JSON `gorm:"type:jsonb" json:"-"`
}

func (FileCapexBudgetEntity) TableName() string { return "file_capex_budget_entities" }

type FileCapexActualEntity struct {
	gorm.Model
	ID       uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	FileName string         `json:"file_name"`
	Year     string         `json:"year"` // Added Year
	UploadAt time.Time      `json:"upload_at"`
	UserID   string         `gorm:"type:varchar(36)" json:"user_id"`
	User     *UserEntity    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Data     datatypes.JSON `gorm:"type:jsonb" json:"-"`
}

func (FileCapexActualEntity) TableName() string { return "file_capex_actual_entities" }

// --- FACT TABLES (Flattened: Header + Detail) ---

// 1. Budget (PL)
type BudgetFactEntity struct { // HEADER
	gorm.Model
	ID           uuid.UUID         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	FileBudgetID uuid.UUID         `gorm:"type:uuid;index" json:"file_budget_id"`
	FileBudget   *FileBudgetEntity `gorm:"foreignKey:FileBudgetID"`

	// Dimensions
	Entity     string `gorm:"index:idx_budget_year_entity_branch_dept;index:idx_budget_year_entity" json:"entity"`
	Branch     string `gorm:"index:idx_budget_year_entity_branch_dept" json:"branch"`
	Group      string `json:"group"`
	EntityGL   string `json:"entity_gl"`
	ConsoGL    string `json:"conso_gl"`
	GLName     string `json:"gl_name"`
	Department string `gorm:"index:idx_budget_year_entity_branch_dept" json:"department"`
	NavCode    string `gorm:"index" json:"nav_code"`
	Year       string `gorm:"index:idx_budget_year_entity_branch_dept;index:idx_budget_year_entity" json:"year"`

	// Summary
	YearTotal decimal.Decimal `gorm:"type:decimal(18,2)" json:"year_total"`

	// Amounts (1 Header -> Many Monthly Amounts)
	BudgetAmounts []BudgetAmountEntity `gorm:"foreignKey:BudgetFactID" json:"budget_amounts,omitempty"`
}

func (BudgetFactEntity) TableName() string { return "budget_fact_entities" }

type BudgetAmountEntity struct { // DETAIL
	gorm.Model
	ID           uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	BudgetFactID uuid.UUID       `gorm:"type:uuid;index" json:"budget_fact_id"`
	Month        string          `json:"month"` // JAN, FEB...
	Amount       decimal.Decimal `gorm:"type:decimal(18,2)" json:"amount"`
}

func (BudgetAmountEntity) TableName() string { return "budget_amount_entities" }

// 2. Capex Budget (Plan)
type CapexBudgetFactEntity struct { // HEADER
	gorm.Model
	ID                uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	FileCapexBudgetID uuid.UUID              `gorm:"type:uuid;index" json:"file_capex_budget_id"`
	FileCapexBudget   *FileCapexBudgetEntity `gorm:"foreignKey:FileCapexBudgetID"`

	// Dimensions
	Year          string `gorm:"index" json:"year"`
	Entity        string `gorm:"index" json:"entity"`
	Branch        string `json:"branch"`
	Department    string `json:"department"`
	CapexNo       string `json:"capex_no"`
	CapexName     string `json:"capex_name"`
	CapexCategory string `json:"capex_category"`

	// Summary
	YearTotal decimal.Decimal `gorm:"type:decimal(18,2)" json:"year_total"`

	// Amounts
	CapexBudgetAmounts []CapexBudgetAmountEntity `gorm:"foreignKey:CapexBudgetFactID" json:"capex_budget_amounts,omitempty"`
}

func (CapexBudgetFactEntity) TableName() string { return "capex_budget_fact_entities" }

type CapexBudgetAmountEntity struct { // DETAIL
	gorm.Model
	ID                uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CapexBudgetFactID uuid.UUID       `gorm:"type:uuid;index" json:"capex_budget_fact_id"`
	Month             string          `json:"month"`
	Amount            decimal.Decimal `gorm:"type:decimal(18,2)" json:"amount"`
}

func (CapexBudgetAmountEntity) TableName() string { return "capex_budget_amount_entities" }

// 3. Capex Actual
type CapexActualFactEntity struct { // HEADER
	gorm.Model
	ID                uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	FileCapexActualID uuid.UUID              `gorm:"type:uuid;index" json:"file_capex_actual_id"`
	FileCapexActual   *FileCapexActualEntity `gorm:"foreignKey:FileCapexActualID"`

	// Dimensions
	Year          string `gorm:"index" json:"year"`
	Entity        string `gorm:"index" json:"entity"`
	Branch        string `json:"branch"`
	Department    string `json:"department"`
	CapexNo       string `json:"capex_no"`
	CapexName     string `json:"capex_name"`
	CapexCategory string `json:"capex_category"`

	// Summary
	YearTotal decimal.Decimal `gorm:"type:decimal(18,2)" json:"year_total"`

	// Amounts
	CapexActualAmounts []CapexActualAmountEntity `gorm:"foreignKey:CapexActualFactID" json:"capex_actual_amounts,omitempty"`
}

func (CapexActualFactEntity) TableName() string { return "capex_actual_fact_entities" }

type CapexActualAmountEntity struct { // DETAIL
	gorm.Model
	ID                uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CapexActualFactID uuid.UUID       `gorm:"type:uuid;index" json:"capex_actual_fact_id"`
	Month             string          `json:"month"`
	Amount            decimal.Decimal `gorm:"type:decimal(18,2)" json:"amount"`
}

func (CapexActualAmountEntity) TableName() string { return "capex_actual_amount_entities" }

// 4. Central Actuals (Unified Source of Truth)
type ActualFactEntity struct { // HEADER
	gorm.Model
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`

	// Dimensions
	Entity     string `gorm:"index:idx_actual_composite" json:"entity"`
	Branch     string `gorm:"index:idx_actual_composite" json:"branch"`
	Department string `gorm:"index:idx_actual_composite" json:"department"`
	NavCode    string `gorm:"index" json:"nav_code"`
	Group      string `json:"group"`
	EntityGL   string `json:"entity_gl"`
	ConsoGL    string `json:"conso_gl"`
	GLName     string `json:"gl_name"`
	VendorName string `json:"vendor_name"`
	Year       string `gorm:"index:idx_actual_composite" json:"year"`

	// Summary
	YearTotal decimal.Decimal `gorm:"type:decimal(18,2)" json:"year_total"`

	// Amounts
	ActualAmounts []ActualAmountEntity `gorm:"foreignKey:ActualFactID" json:"actual_amounts,omitempty"`

	// Status
	IsValid bool `gorm:"default:true" json:"is_valid"`
}

func (ActualFactEntity) TableName() string { return "actual_fact_entities" }

type ActualAmountEntity struct { // DETAIL
	gorm.Model
	ID           uuid.UUID         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ActualFactID uuid.UUID         `gorm:"type:uuid;index" json:"actual_fact_id"`
	Month        string            `json:"month"` // JAN, FEB
	Amount       decimal.Decimal   `gorm:"type:decimal(18,2)" json:"amount"`
	ActualFact   *ActualFactEntity `gorm:"foreignKey:ActualFactID" json:"actual_fact"`
}

func (ActualAmountEntity) TableName() string { return "actual_amount_entities" }

// ActualTransactionEntity stores detailed transaction records for reporting
type ActualTransactionEntity struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Source Details
	Source      string          `gorm:"index" json:"source"`       // HMW, CLIK
	PostingDate string          `gorm:"index" json:"posting_date"` // YYYY-MM-DD
	DocNo       string          `json:"doc_no"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `gorm:"type:decimal(20,2)" json:"amount"`
	VendorName  string          `json:"vendor"`
	GLAccountName string        `json:"gl_account_name"`

	// Dimensional Data (Mapped/Centralized)
	Entity     string `gorm:"index" json:"entity"`
	Branch     string `gorm:"index" json:"branch"`
	Department string `gorm:"index" json:"department"`
	EntityGL   string `gorm:"index;column:entity_gl" json:"entity_gl"`
	ConsoGL    string `gorm:"index;column:conso_gl" json:"conso_gl"`
	Year       string `gorm:"index" json:"year"`

	// Status
	IsValid bool   `gorm:"default:true" json:"is_valid"`
	Status  string `gorm:"type:varchar(20);default:'PENDING';index" json:"status"`
}

const (
	TxStatusPending  = "PENDING"
	TxStatusDraft    = "DRAFT"
	TxStatusReported = "REPORTED"
	TxStatusComplete = "COMPLETE"
)

func (ActualTransactionEntity) TableName() string { return "actual_transaction_entities" }

// UserConfigEntity stores personal settings (e.g., active budget files) per user
type UserConfigEntity struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID    string    `gorm:"index:idx_user_config_user_key;not null" json:"user_id"`
	ConfigKey string    `gorm:"index:idx_user_config_user_key;not null" json:"config_key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserConfigEntity) TableName() string { return "user_config_entities" }

// DataInventoryEntity tracks years and months that have data in staging tables
type DataInventoryEntity struct {
	Year      string    `gorm:"primaryKey;index" json:"year"`
	Month     string    `gorm:"primaryKey;index" json:"month"` // JAN, FEB...
	UpdatedAt time.Time `json:"updated_at"`
}

func (DataInventoryEntity) TableName() string { return "data_inventory_entities" }

// --- Source (Central) Entities & DTOs ---

type CentralCompany struct {
	CompanyID    uint   `gorm:"column:company_id;primaryKey"`
	Name         string `gorm:"column:name"`
	BranchName   string `gorm:"column:branch_name"`
	BranchNameEn string `gorm:"column:branch_name_en"`
	BranchNo     string `gorm:"column:branch_no"`
	Address      string `gorm:"column:address"`
	TaxID        string `gorm:"column:taxid"`
	Province     string `gorm:"column:province"`
}

type CentralDepartment struct {
	DeptID uint   `gorm:"column:department_id;primaryKey"`
	Name   string `gorm:"column:name"`
	Code   string `gorm:"column:code"`
}

type CentralSection struct {
	SectionID    uint   `gorm:"column:section_id;primaryKey"`
	Name         string `gorm:"column:name"`
	Code         string `gorm:"column:code"`
	DepartmentID uint   `gorm:"column:department_id"`
}

type CentralPosition struct {
	PositionID uint   `gorm:"column:position_id;primaryKey"`
	Name       string `gorm:"column:name"`
	Code       string `gorm:"column:code"`
}

type CentralUser struct {
	UserID       uint   `gorm:"column:id;primaryKey"`
	Username     string `gorm:"column:username"`
	NameTh       string `gorm:"column:name_th"`
	NameEn       string `gorm:"column:name_en"`
	CompanyID    uint   `gorm:"column:company_id"`
	DepartmentID uint   `gorm:"column:department_id"`
	SectionID    uint   `gorm:"column:section_id"`
	PositionID   uint   `gorm:"column:position_id"`
	Deleted      bool   `gorm:"column:deleted;default:false;not null"`
}

func (CentralUser) TableName() string { return "users" }

type UserResponse struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	NameTh       string   `json:"name_th"`
	NameEn       string   `json:"name_en"`
	CompanyID    uint     `json:"company_id"`
	DepartmentID uint     `json:"department_id"`
	SectionID    uint     `json:"section_id"`
	PositionID   uint     `json:"position_id"`
	Roles        []string `json:"roles"`
	Deleted      bool     `json:"deleted"`
}

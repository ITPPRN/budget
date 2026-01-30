package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// --- Users (Keep Existing) ---
type UserEntity struct {
	ID                 string            `gorm:"primaryKey;type:varchar(36);not null" json:"id"`
	Username           string            `gorm:"uniqueIndex;not null" json:"username"`
	Email              string            `gorm:"uniqueIndex;not null" json:"email"`
	FirstName          string            `json:"first_name"`
	LastName           string            `json:"last_name"`
	Roles              datatypes.JSON    `gorm:"type:jsonb" json:"roles"`
	SignatureURL       string            `json:"signature_url"`
	PdpaAcknowledgedAt *time.Time        `json:"pdpa_acknowledged_at"`
	IsActive           bool              `gorm:"default:true" json:"is_active"`
	DepartmentID       *uuid.UUID        `gorm:"type:uuid;index" json:"department_id"`
	Department         *DepartmentEntity `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	UpdateBy           *string           `gorm:"type:varchar(36)" json:"update_by"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	DeletedAt          gorm.DeletedAt    `gorm:"index" json:"deleted_at"`
}

func (UserEntity) TableName() string { return "user_entities" }

type DepartmentEntity struct {
	gorm.Model
	ID        uuid.UUID   `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Code      string      `gorm:"uniqueIndex;not null" json:"code"`
	Name      string      `gorm:"not null" json:"name"`
	UpdateBy  *uuid.UUID  `gorm:"type:uuid" json:"update_by"`
	ManagerID *string     `gorm:"type:varchar(36)" json:"manager_id"`
	Manager   *UserEntity `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
}

func (DepartmentEntity) TableName() string { return "department_entities" }

// --- P2P Entities (Keep Existing) ---
type VendorEntity struct {
	gorm.Model
	ID          uuid.UUID  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	VendorCode  string     `gorm:"uniqueIndex" json:"vendor_code"`
	Name        string     `json:"name"`
	TaxID       string     `json:"tax_id"`
	Address     string     `json:"address"`
	PaymentTerm string     `json:"payment_term"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	UpdateBy    *uuid.UUID `gorm:"type:uuid" json:"update_by"`
}

func (VendorEntity) TableName() string { return "vendor_entities" }

type ProductEntity struct {
	gorm.Model
	ID            uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProductCode   string          `gorm:"uniqueIndex" json:"product_code"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Unit          string          `json:"unit"`
	StandardPrice decimal.Decimal `gorm:"type:decimal(18,2)" json:"standard_price"`
	Category      string          `json:"category"`
	UpdateBy      *uuid.UUID      `gorm:"type:uuid" json:"update_by"`
}

func (ProductEntity) TableName() string { return "product_entities" }

type PurchaseRequestEntity struct {
	gorm.Model
	ID             uuid.UUID         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PrNumber       string            `gorm:"uniqueIndex" json:"pr_number"`
	RequesterID    string            `gorm:"type:varchar(36)" json:"requester_id"`
	DepartmentID   uuid.UUID         `gorm:"type:uuid" json:"department_id"`
	FlowType       string            `json:"flow_type"`
	ExternalRefDoc string            `json:"external_ref_doc"`
	RequiredDate   time.Time         `json:"required_date"`
	Status         string            `gorm:"default:'DRAFT'" json:"status"`
	RejectReason   string            `json:"reject_reason"`
	UpdateBy       *uuid.UUID        `gorm:"type:uuid" json:"update_by"`
	Requester      *UserEntity       `gorm:"foreignKey:RequesterID;references:ID" json:"requester"`
	Department     *DepartmentEntity `gorm:"foreignKey:DepartmentID" json:"department"`
	Items          []PrItemEntity    `gorm:"foreignKey:PrID" json:"items"`
}

func (PurchaseRequestEntity) TableName() string { return "purchase_request_entities" }

type PrItemEntity struct {
	gorm.Model
	ID                 uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PrID               uuid.UUID       `gorm:"type:uuid" json:"pr_id"`
	ProductID          *uuid.UUID      `gorm:"type:uuid" json:"product_id"`
	ItemDescription    string          `json:"item_description"`
	Quantity           decimal.Decimal `gorm:"type:decimal(18,2)" json:"quantity"`
	EstimatedUnitPrice decimal.Decimal `gorm:"type:decimal(18,2)" json:"estimated_unit_price"`
	TotalPrice         decimal.Decimal `gorm:"type:decimal(18,2)" json:"total_price"`
	UpdateBy           *uuid.UUID      `gorm:"type:uuid" json:"update_by"`
	Product            *ProductEntity  `gorm:"foreignKey:ProductID" json:"product"`
}

func (PrItemEntity) TableName() string { return "pr_item_entities" }

type PurchaseOrderEntity struct {
	gorm.Model
	ID               uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PoNumberSystem   string                 `gorm:"uniqueIndex" json:"po_number_system"`
	ExternalPoNumber string                 `json:"external_po_number"`
	TargetSystem     string                 `json:"target_system"`
	PrID             uuid.UUID              `gorm:"type:uuid" json:"pr_id"`
	VendorID         uuid.UUID              `gorm:"type:uuid" json:"vendor_id"`
	PurchaserID      string                 `gorm:"type:varchar(36)" json:"purchaser_id"`
	PoDate           time.Time              `json:"po_date"`
	Status           string                 `json:"status"`
	TotalAmount      decimal.Decimal        `gorm:"type:decimal(18,2)" json:"total_amount"`
	VatAmount        decimal.Decimal        `gorm:"type:decimal(18,2)" json:"vat_amount"`
	GrandTotal       decimal.Decimal        `gorm:"type:decimal(18,2)" json:"grand_total"`
	UpdateBy         *uuid.UUID             `gorm:"type:uuid" json:"update_by"`
	PurchaseRequest  *PurchaseRequestEntity `gorm:"foreignKey:PrID" json:"purchase_request"`
	Vendor           *VendorEntity          `gorm:"foreignKey:VendorID" json:"vendor"`
	Purchaser        *UserEntity            `gorm:"foreignKey:PurchaserID;references:ID" json:"purchaser"`
}

func (PurchaseOrderEntity) TableName() string { return "purchase_order_entities" }

type GoodsReceiptEntity struct {
	gorm.Model
	ID                uuid.UUID            `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	GrNumber          string               `gorm:"uniqueIndex" json:"gr_number"`
	PoID              uuid.UUID            `gorm:"type:uuid" json:"po_id"`
	ReceivedByID      string               `gorm:"type:varchar(36)" json:"received_by_id"`
	VendorDeliveryDoc string               `json:"vendor_delivery_doc"`
	ReceivedDate      time.Time            `json:"received_date"`
	InspectionStatus  string               `json:"inspection_status"`
	Photos            datatypes.JSON       `gorm:"type:jsonb" json:"photos"`
	Remark            string               `json:"remark"`
	UpdateBy          *uuid.UUID           `gorm:"type:uuid" json:"update_by"`
	PurchaseOrder     *PurchaseOrderEntity `gorm:"foreignKey:PoID" json:"purchase_order"`
	ReceivedBy        *UserEntity          `gorm:"foreignKey:ReceivedByID;references:ID" json:"received_by"`
}

func (GoodsReceiptEntity) TableName() string { return "goods_receipt_entities" }

type ApVoucherEntity struct {
	gorm.Model
	ID                uuid.UUID            `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PoID              uuid.UUID            `gorm:"type:uuid" json:"po_id"`
	InvoiceNumber     string               `json:"invoice_number"`
	InvoiceDate       time.Time            `json:"invoice_date"`
	ExternalVoucherNo string               `json:"external_voucher_no"`
	InvoiceAmount     decimal.Decimal      `gorm:"type:decimal(18,2)" json:"invoice_amount"`
	VatAmount         decimal.Decimal      `gorm:"type:decimal(18,2)" json:"vat_amount"`
	WhtAmount         decimal.Decimal      `gorm:"type:decimal(18,2)" json:"wht_amount"`
	NetPayAmount      decimal.Decimal      `gorm:"type:decimal(18,2)" json:"net_pay_amount"`
	Status            string               `json:"status"`
	UpdateBy          *uuid.UUID           `gorm:"type:uuid" json:"update_by"`
	PurchaseOrder     *PurchaseOrderEntity `gorm:"foreignKey:PoID" json:"purchase_order"`
}

func (ApVoucherEntity) TableName() string { return "ap_voucher_entities" }

type PaymentEntity struct {
	gorm.Model
	ID               uuid.UUID        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ApVoucherID      uuid.UUID        `gorm:"type:uuid" json:"ap_voucher_id"`
	PaidByID         string           `gorm:"type:varchar(36)" json:"paid_by_id"`
	PaymentDate      time.Time        `json:"payment_date"`
	PaymentMethod    string           `json:"payment_method"`
	RefTransactionID string           `json:"ref_transaction_id"`
	AmountPaid       decimal.Decimal  `gorm:"type:decimal(18,2)" json:"amount_paid"`
	UpdateBy         *uuid.UUID       `gorm:"type:uuid" json:"update_by"`
	ApVoucher        *ApVoucherEntity `gorm:"foreignKey:ApVoucherID" json:"ap_voucher"`
}

func (PaymentEntity) TableName() string { return "payment_entities" }

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
	Entity     string `gorm:"index" json:"entity"`
	Branch     string `gorm:"index" json:"branch"`
	Group      string `json:"group"`
	EntityGL   string `json:"entity_gl"`
	ConsoGL    string `json:"conso_gl"`
	GLName     string `json:"gl_name"`
	Department string `json:"department"`

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
	Entity        string `gorm:"index" json:"entity"`
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
	Entity        string `gorm:"index" json:"entity"`
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

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// --- Users ---
type UserEntity struct {
    // 1. Primary Key: ใช้ ID เดียวกับ Keycloak (UUID String)
    ID          string         `gorm:"primaryKey;type:varchar(36);not null" json:"id"` 
    
    // 2. ข้อมูล User
    Username    string         `gorm:"uniqueIndex;not null" json:"username"`
    Email       string         `gorm:"uniqueIndex;not null" json:"email"`
    FirstName   string         `json:"first_name"`
    LastName    string         `json:"last_name"`
    
    // 3. ข้อมูลเสริม
    Roles       datatypes.JSON `gorm:"type:jsonb" json:"roles"` // เก็บเป็น ["employee", "manager"]
    SignatureURL string        `json:"signature_url"`
    PdpaAcknowledgedAt *time.Time `json:"pdpa_acknowledged_at"`
    IsActive    bool           `gorm:"default:true" json:"is_active"`
    
    // 4. Relation (แผนก)
    // ใช้ *uuid.UUID เพื่อให้เป็น NULL ได้ (เผื่อ User ใหม่ยังไม่มีแผนก)
    DepartmentID *uuid.UUID     `gorm:"type:uuid;index" json:"department_id"` 
    Department   *DepartmentEntity `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`

    // 5. Timestamps (แทน gorm.Model)
    UpdateBy    *string        `gorm:"type:varchar(36)" json:"update_by"` // คนแก้ก็เก็บเป็น Keycloak ID
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"` // รองรับ Soft Delete
}

func (UserEntity) TableName() string { return "user_entities" }

// --- Departments ---
type DepartmentEntity struct {
	gorm.Model
	ID       uuid.UUID  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Code     string     `gorm:"uniqueIndex;not null" json:"code"`
	Name     string     `gorm:"not null" json:"name"`
	UpdateBy *uuid.UUID `gorm:"type:uuid" json:"update_by"`

	// Relation - เชื่อมไปที่ UserID (string)
	ManagerID *string `gorm:"type:varchar(36)" json:"manager_id"`

	Manager *UserEntity `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
}

func (DepartmentEntity) TableName() string { return "department_entities" }

// --- Vendors ---
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

// --- Products ---
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

// --- Purchase Requests ---
type PurchaseRequestEntity struct {
	gorm.Model
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PrNumber       string     `gorm:"uniqueIndex" json:"pr_number"`
	RequesterID    string     `gorm:"type:varchar(36)" json:"requester_id"`
	DepartmentID   uuid.UUID  `gorm:"type:uuid" json:"department_id"`
	FlowType       string     `json:"flow_type"`
	ExternalRefDoc string     `json:"external_ref_doc"`
	RequiredDate   time.Time  `json:"required_date"`
	Status         string     `gorm:"default:'DRAFT'" json:"status"`
	RejectReason   string     `json:"reject_reason"`
	UpdateBy       *uuid.UUID `gorm:"type:uuid" json:"update_by"`

	Requester  *UserEntity       `gorm:"foreignKey:RequesterID;references:ID" json:"requester"`
	Department *DepartmentEntity `gorm:"foreignKey:DepartmentID" json:"department"`
	Items      []PrItemEntity    `gorm:"foreignKey:PrID" json:"items"`
}

func (PurchaseRequestEntity) TableName() string { return "purchase_request_entities" }

// --- PR Items ---
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

	Product *ProductEntity `gorm:"foreignKey:ProductID" json:"product"`
}

func (PrItemEntity) TableName() string { return "pr_item_entities" }

// --- Purchase Orders ---
type PurchaseOrderEntity struct {
	gorm.Model
	ID               uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PoNumberSystem   string          `gorm:"uniqueIndex" json:"po_number_system"`
	ExternalPoNumber string          `json:"external_po_number"`
	TargetSystem     string          `json:"target_system"`
	PrID             uuid.UUID       `gorm:"type:uuid" json:"pr_id"`
	VendorID         uuid.UUID       `gorm:"type:uuid" json:"vendor_id"`
	PurchaserID      string          `gorm:"type:varchar(36)" json:"purchaser_id"`
	PoDate           time.Time       `json:"po_date"`
	Status           string          `json:"status"`
	TotalAmount      decimal.Decimal `gorm:"type:decimal(18,2)" json:"total_amount"`
	VatAmount        decimal.Decimal `gorm:"type:decimal(18,2)" json:"vat_amount"`
	GrandTotal       decimal.Decimal `gorm:"type:decimal(18,2)" json:"grand_total"`
	UpdateBy         *uuid.UUID      `gorm:"type:uuid" json:"update_by"`

	PurchaseRequest *PurchaseRequestEntity `gorm:"foreignKey:PrID" json:"purchase_request"`
	Vendor          *VendorEntity          `gorm:"foreignKey:VendorID" json:"vendor"`
	Purchaser       *UserEntity            `gorm:"foreignKey:PurchaserID;references:ID" json:"purchaser"`
}

func (PurchaseOrderEntity) TableName() string { return "purchase_order_entities" }

// --- Goods Receipts ---
type GoodsReceiptEntity struct {
	gorm.Model
	ID                uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	GrNumber          string         `gorm:"uniqueIndex" json:"gr_number"`
	PoID              uuid.UUID      `gorm:"type:uuid" json:"po_id"`
	ReceivedByID      string         `gorm:"type:varchar(36)" json:"received_by_id"`
	VendorDeliveryDoc string         `json:"vendor_delivery_doc"`
	ReceivedDate      time.Time      `json:"received_date"`
	InspectionStatus  string         `json:"inspection_status"`
	Photos            datatypes.JSON `gorm:"type:jsonb" json:"photos"`
	Remark            string         `json:"remark"`
	UpdateBy          *uuid.UUID     `gorm:"type:uuid" json:"update_by"`

	PurchaseOrder *PurchaseOrderEntity `gorm:"foreignKey:PoID" json:"purchase_order"`
	ReceivedBy    *UserEntity          `gorm:"foreignKey:ReceivedByID;references:ID" json:"received_by"`
}

func (GoodsReceiptEntity) TableName() string { return "goods_receipt_entities" }

// --- AP Vouchers ---
type ApVoucherEntity struct {
	gorm.Model
	ID                uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PoID              uuid.UUID       `gorm:"type:uuid" json:"po_id"`
	InvoiceNumber     string          `json:"invoice_number"`
	InvoiceDate       time.Time       `json:"invoice_date"`
	ExternalVoucherNo string          `json:"external_voucher_no"`
	InvoiceAmount     decimal.Decimal `gorm:"type:decimal(18,2)" json:"invoice_amount"`
	VatAmount         decimal.Decimal `gorm:"type:decimal(18,2)" json:"vat_amount"`
	WhtAmount         decimal.Decimal `gorm:"type:decimal(18,2)" json:"wht_amount"`
	NetPayAmount      decimal.Decimal `gorm:"type:decimal(18,2)" json:"net_pay_amount"`
	Status            string          `json:"status"`
	UpdateBy          *uuid.UUID      `gorm:"type:uuid" json:"update_by"`

	PurchaseOrder *PurchaseOrderEntity `gorm:"foreignKey:PoID" json:"purchase_order"`
}

func (ApVoucherEntity) TableName() string { return "ap_voucher_entities" }

// --- Payments ---
type PaymentEntity struct {
	gorm.Model
	ID               uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ApVoucherID      uuid.UUID       `gorm:"type:uuid" json:"ap_voucher_id"`
	PaidByID         string          `gorm:"type:varchar(36)" json:"paid_by_id"`
	PaymentDate      time.Time       `json:"payment_date"`
	PaymentMethod    string          `json:"payment_method"`
	RefTransactionID string          `json:"ref_transaction_id"`
	AmountPaid       decimal.Decimal `gorm:"type:decimal(18,2)" json:"amount_paid"`
	UpdateBy         *uuid.UUID      `gorm:"type:uuid" json:"update_by"`

	ApVoucher *ApVoucherEntity `gorm:"foreignKey:ApVoucherID" json:"ap_voucher"`
}

func (PaymentEntity) TableName() string { return "payment_entities" }

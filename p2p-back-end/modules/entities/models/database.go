package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// --- Users (Path: /v1/users) ---
type UserEntity struct {
	ID        string `gorm:"primaryKey;type:varchar(36)" json:"id"` // ตรงกับ Keycloak ID
	Username  string `gorm:"uniqueIndex;not null" json:"username"`
	Email     string `gorm:"uniqueIndex;not null" json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`

	// เก็บ Array เป็น JSON ใน Postgres
	Roles datatypes.JSON `gorm:"type:jsonb" json:"roles"`

	DepartmentID       *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	SignatureURL       string     `json:"signature_url"`
	PdpaAcknowledgedAt *time.Time `json:"pdpa_acknowledged_at"`
	IsActive           bool       `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdateBy  *uuid.UUID `gorm:"type:uuid" json:"update_by"`

	// Relation: User สังกัดแผนกอะไร
	Department *DepartmentEntity `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
}

// --- Departments (Path: /v1/departments) ---
type DepartmentEntity struct {
	ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Code string    `gorm:"uniqueIndex;not null" json:"code"`
	Name string    `gorm:"not null" json:"name"`

	ManagerID *string `gorm:"type:varchar(36)" json:"manager_id"` // Link ไปหา User ID

	// Relation: ใครเป็น Manager
	Manager   *UserEntity `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	UpdatedAt time.Time   `json:"updated_at"`
	UpdateBy  *uuid.UUID  `gorm:"type:uuid" json:"update_by"`
}

// --- Vendors (Path: /v1/vendors) ---
type VendorEntity struct {
	ID          uuid.UUID  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	VendorCode  string     `gorm:"uniqueIndex" json:"vendor_code"` // รหัสใน NAV/IDS
	Name        string     `json:"name"`
	TaxID       string     `json:"tax_id"`
	Address     string     `json:"address"`
	PaymentTerm string     `json:"payment_term"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	UpdatedAt   time.Time  `json:"updated_at"`
	UpdateBy    *uuid.UUID `gorm:"type:uuid" json:"update_by"`
}

// --- Products (Path: /v1/products) ---
type ProductEntity struct {
	ID            uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProductCode   string          `gorm:"uniqueIndex" json:"product_code"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Unit          string          `json:"unit"` // หน่วยนับ
	StandardPrice decimal.Decimal `gorm:"type:decimal(18,2)" json:"standard_price"`
	Category      string          `json:"category"` // Stock, Asset, Expense
	UpdatedAt     time.Time       `json:"updated_at"`
	UpdateBy      *uuid.UUID      `gorm:"type:uuid" json:"update_by"`
}

// --- Purchase Requests (Path: /v1/purchase-requests) ---
type PurchaseRequestEntity struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PrNumber string    `gorm:"uniqueIndex" json:"pr_number"` // PR-YYYYMM-XXXX

	RequesterID  string    `gorm:"type:varchar(36)" json:"requester_id"`
	DepartmentID uuid.UUID `gorm:"type:uuid" json:"department_id"`

	FlowType       string    `json:"flow_type"`        // NEW_CAR, SERVICE, STOCK, etc.
	ExternalRefDoc string    `json:"external_ref_doc"` // ใบแจ้งซ่อม PPL / Job ID
	RequiredDate   time.Time `json:"required_date"`
	Status         string    `gorm:"default:'DRAFT'" json:"status"`
	RejectReason   string    `json:"reject_reason"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdateBy  *uuid.UUID `gorm:"type:uuid" json:"update_by"`

	// Relations
	Requester  *UserEntity       `gorm:"foreignKey:RequesterID" json:"requester"`
	Department *DepartmentEntity `gorm:"foreignKey:DepartmentID" json:"department"`
	Items      []PrItemEntity    `gorm:"foreignKey:PrID" json:"items"` // Has Many Items
}

// --- PR Items (อยู่ภายใน PR) ---
type PrItemEntity struct {
	ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PrID uuid.UUID `gorm:"type:uuid" json:"pr_id"`

	ProductID          *uuid.UUID      `gorm:"type:uuid" json:"product_id"` // Null ได้
	ItemDescription    string          `json:"item_description"`
	Quantity           decimal.Decimal `gorm:"type:decimal(18,2)" json:"quantity"`
	EstimatedUnitPrice decimal.Decimal `gorm:"type:decimal(18,2)" json:"estimated_unit_price"`
	TotalPrice         decimal.Decimal `gorm:"type:decimal(18,2)" json:"total_price"`
	UpdatedAt          time.Time       `json:"updated_at"`
	UpdateBy           *uuid.UUID      `gorm:"type:uuid" json:"update_by"`

	Product *ProductEntity `gorm:"foreignKey:ProductID" json:"product"`
}

// --- Purchase Orders (Path: /v1/purchase-orders) ---
type PurchaseOrderEntity struct {
	ID               uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PoNumberSystem   string    `gorm:"uniqueIndex" json:"po_number_system"`
	ExternalPoNumber string    `json:"external_po_number"` // จาก NAV/IDS
	TargetSystem     string    `json:"target_system"`      // NAV, IDS

	PrID        uuid.UUID `gorm:"type:uuid" json:"pr_id"`
	VendorID    uuid.UUID `gorm:"type:uuid" json:"vendor_id"`
	PurchaserID string    `gorm:"type:varchar(36)" json:"purchaser_id"` // คนออก PO

	PoDate time.Time `json:"po_date"`
	Status string    `json:"status"` // OPEN, RECEIVED, CLOSED

	TotalAmount decimal.Decimal `gorm:"type:decimal(18,2)" json:"total_amount"`
	VatAmount   decimal.Decimal `gorm:"type:decimal(18,2)" json:"vat_amount"`
	GrandTotal  decimal.Decimal `gorm:"type:decimal(18,2)" json:"grand_total"`

	// Relations
	PurchaseRequest *PurchaseRequestEntity `gorm:"foreignKey:PrID" json:"purchase_request"`
	Vendor          *VendorEntity          `gorm:"foreignKey:VendorID" json:"vendor"`
	Purchaser       *UserEntity            `gorm:"foreignKey:PurchaserID" json:"purchaser"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdateBy  *uuid.UUID `gorm:"type:uuid" json:"update_by"`
}

// --- Goods Receipts (Path: /v1/goods-receipts) ---
type GoodsReceiptEntity struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	GrNumber string    `gorm:"uniqueIndex" json:"gr_number"`

	PoID         uuid.UUID `gorm:"type:uuid" json:"po_id"`
	ReceivedByID string    `gorm:"type:varchar(36)" json:"received_by_id"`

	VendorDeliveryDoc string    `json:"vendor_delivery_doc"`
	ReceivedDate      time.Time `json:"received_date"`
	InspectionStatus  string    `json:"inspection_status"` // PASS, FAIL

	// เก็บ Array รูปภาพเป็น JSON
	Photos datatypes.JSON `gorm:"type:jsonb" json:"photos"`
	Remark string         `json:"remark"`

	// Relations
	PurchaseOrder *PurchaseOrderEntity `gorm:"foreignKey:PoID" json:"purchase_order"`
	ReceivedBy    *UserEntity          `gorm:"foreignKey:ReceivedByID" json:"received_by"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdateBy  *uuid.UUID `gorm:"type:uuid" json:"update_by"`
}

// --- AP Vouchers (Path: /v1/finance/ap-vouchers) ---
type ApVoucherEntity struct {
	ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	PoID uuid.UUID `gorm:"type:uuid" json:"po_id"`

	InvoiceNumber     string    `json:"invoice_number"`
	InvoiceDate       time.Time `json:"invoice_date"`
	ExternalVoucherNo string    `json:"external_voucher_no"` // จาก NAV/IDS

	InvoiceAmount decimal.Decimal `gorm:"type:decimal(18,2)" json:"invoice_amount"`
	VatAmount     decimal.Decimal `gorm:"type:decimal(18,2)" json:"vat_amount"`
	WhtAmount     decimal.Decimal `gorm:"type:decimal(18,2)" json:"wht_amount"`
	NetPayAmount  decimal.Decimal `gorm:"type:decimal(18,2)" json:"net_pay_amount"`

	Status    string    `json:"status"` // PENDING_PAYMENT, PAID
	CreatedAt time.Time `json:"created_at"`

	PurchaseOrder *PurchaseOrderEntity `gorm:"foreignKey:PoID" json:"purchase_order"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdateBy  *uuid.UUID `gorm:"type:uuid" json:"update_by"`
}

// --- Payments (Path: /v1/finance/payments) ---
type PaymentEntity struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ApVoucherID uuid.UUID `gorm:"type:uuid" json:"ap_voucher_id"`
	PaidByID    string    `gorm:"type:varchar(36)" json:"paid_by_id"`

	PaymentDate      time.Time       `json:"payment_date"`
	PaymentMethod    string          `json:"payment_method"` // TRANSFER, CHEQUE
	RefTransactionID string          `json:"ref_transaction_id"`
	AmountPaid       decimal.Decimal `gorm:"type:decimal(18,2)" json:"amount_paid"`

	ApVoucher *ApVoucherEntity `gorm:"foreignKey:ApVoucherID" json:"ap_voucher"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdateBy  *uuid.UUID `gorm:"type:uuid" json:"update_by"`
}

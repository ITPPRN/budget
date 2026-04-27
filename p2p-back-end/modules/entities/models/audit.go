package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AuditLogEntity records a confirmation or rejection by an Owner
type AuditLogEntity struct {
	ID            uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Entity        string         `gorm:"index" json:"entity"`
	Branch        string         `gorm:"index" json:"branch"`
	Department    string         `gorm:"index" json:"department"`
	Year          string         `gorm:"index" json:"year"`
	Month         string         `gorm:"index" json:"month"`
	Status        string         `gorm:"index" json:"status"` // CONFIRMED, REJECTED
	RejectedCount int            `json:"rejected_count"`      // Number of rejected line items
	CreatedBy     string         `json:"created_by"`          // Name of the user
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Link to specific rejected items
	RejectedItems []AuditLogRejectedItemEntity `gorm:"foreignKey:AuditLogID" json:"-"`
}

func (AuditLogEntity) TableName() string { return "audit_log_entities" }

// AuditLogRejectedItemEntity links a log entry to specific transaction records
type AuditLogRejectedItemEntity struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	AuditLogID uuid.UUID `gorm:"type:uuid;index" json:"audit_log_id"`

	// Reference to ActualTransactionEntity
	TransactionID uuid.UUID `gorm:"type:uuid;index" json:"transaction_id"`

	// Snapshot Data (recorded at the time of report)
	ConsoGL       string          `json:"conso_gl"`
	GLAccountName string          `json:"gl_account_name"`
	Amount        decimal.Decimal `gorm:"type:decimal(20,2)" json:"amount"`
	Vendor        string          `json:"vendor"`
	DocNo         string          `json:"doc_no"`
	Description   string          `json:"description"`
	PostingDate   string          `json:"posting_date"`
	Note          string          `gorm:"type:text" json:"note"`
}

func (AuditLogRejectedItemEntity) TableName() string { return "audit_log_rejected_item_entities" }

type AuditRejectBasket struct {
    ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
    TransactionID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_tx" json:"transaction_id"`
    UserID        uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_tx" json:"user_id"` // OWNER ที่เป็นเจ้าของตะกร้านี้
    AddedBy       uuid.UUID `gorm:"type:uuid;index" json:"added_by"`                  // ผู้กดเพิ่ม (OWNER เอง / DELEGATE / BRANCH_DELEGATE)
    Note          string    `gorm:"type:text" json:"note"`
    CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AuditRejectBasket) TableName() string { return  "audit_rejection_baskets"}

// BasketItemView is the wire-format DTO for /audit/basket/list — actual transaction
// fields plus the user-supplied rejection note and added_by from the basket join.
type BasketItemView struct {
	ActualTransactionEntity
	Note    string    `json:"note"`
	AddedBy uuid.UUID `json:"added_by"`
}

// BasketAddItem is the per-row payload for POST /audit/basket/add.
type BasketAddItem struct {
	TransactionID string `json:"transaction_id"`
	Note          string `json:"note"`
}


type YearMonth struct {
    Year  string
    Month string
}
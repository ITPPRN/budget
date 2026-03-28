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
	ID         uuid.UUID  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	AuditLogID uuid.UUID  `gorm:"type:uuid;index" json:"audit_log_id"`
	
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
}

func (AuditLogRejectedItemEntity) TableName() string { return "audit_log_rejected_item_entities" }

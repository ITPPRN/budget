package models

import (
	"time"

	"github.com/google/uuid"
)

// SyncRunEntity — บันทึกประวัติการ sync ทุกรอบ (DW, Tier 1, Tier 2, Manual)
// ใช้สำหรับ observability + retry + admin monitoring
// ID ถูกสร้างใน CreateRun (uuid.New) ไม่พึ่ง DB default เพื่อให้ compatible กับ SQLite tests
type SyncRunEntity struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid" json:"id"`
	JobType      string     `gorm:"size:32;index" json:"job_type"`       // DW_SYNC | TIER1_FAST | TIER2_FULL | MANUAL
	Year         string     `gorm:"size:8;index" json:"year"`            // "2026"
	Month        string     `gorm:"size:4;index" json:"month"`           // "APR" หรือ "" ถ้า full-year
	Status       string     `gorm:"size:16;index" json:"status"`         // RUNNING | SUCCESS | FAILED | PARTIAL
	StartedAt    time.Time  `gorm:"index" json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	DurationMs   int64      `json:"duration_ms"`
	RowsFetched  int64      `json:"rows_fetched"`
	RowsInserted int64      `json:"rows_inserted"`
	RowsSkipped  int64      `json:"rows_skipped"`
	ErrorMessage string     `gorm:"size:2000" json:"error_message,omitempty"`
	RetryCount   int        `json:"retry_count"`
	TriggeredBy  string     `gorm:"size:64" json:"triggered_by"` // CRON | STARTUP | ADMIN:<user_id>
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (SyncRunEntity) TableName() string { return "sync_runs" }

// Sync job type constants
const (
	SyncJobDW         = "DW_SYNC"
	SyncJobTier1Fast  = "TIER1_FAST"
	SyncJobTier2Full  = "TIER2_FULL"
	SyncJobManual     = "MANUAL"
	SyncJobActualFact = "ACTUAL_FACT" // internal SyncActuals
)

// Sync status constants
const (
	SyncStatusRunning  = "RUNNING"
	SyncStatusSuccess  = "SUCCESS"
	SyncStatusFailed   = "FAILED"
	SyncStatusPartial  = "PARTIAL"
	SyncStatusCanceled = "CANCELED" // ผู้ดูแลกด "ยกเลิกคิวทั้งหมด" — retry job จะข้าม
)

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

// SyncTrackingRepository — จัดการบันทึก sync_runs สำหรับ observability
type SyncTrackingRepository interface {
	CreateRun(ctx context.Context, run *models.SyncRunEntity) error
	CompleteRun(ctx context.Context, id uuid.UUID, status string, rowsFetched, rowsInserted, rowsSkipped int64, errMsg string) error
	IncrementRetry(ctx context.Context, id uuid.UUID) error
	GetLatest(ctx context.Context, jobType string) (*models.SyncRunEntity, error)
	GetRecent(ctx context.Context, limit int) ([]models.SyncRunEntity, error)
	GetRunsByJobAndStatus(ctx context.Context, jobType, status string, limit int) ([]models.SyncRunEntity, error)
	GetReconciliation(ctx context.Context, year string) (*ReconciliationResult, error)
	GetFailedRunsForRetry(ctx context.Context, within time.Duration, maxRetries int) ([]models.SyncRunEntity, error)
	ClearStaleRunningRuns(ctx context.Context, olderThan time.Duration) (int64, error)
	DeleteOldRunsByJobType(ctx context.Context, jobType string, olderThan time.Duration) (int64, error)
}

type syncTrackingRepository struct {
	db *gorm.DB
}

func NewSyncTrackingRepository(db *gorm.DB) SyncTrackingRepository {
	// Auto-migrate sync_runs table
	_ = db.AutoMigrate(&models.SyncRunEntity{})
	return &syncTrackingRepository{db: db}
}

func (r *syncTrackingRepository) CreateRun(ctx context.Context, run *models.SyncRunEntity) error {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now()
	}
	if run.Status == "" {
		run.Status = models.SyncStatusRunning
	}
	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return fmt.Errorf("syncTrackingRepo.CreateRun: %w", err)
	}
	return nil
}

func (r *syncTrackingRepository) CompleteRun(
	ctx context.Context, id uuid.UUID, status string,
	rowsFetched, rowsInserted, rowsSkipped int64, errMsg string,
) error {
	now := time.Now()
	// Truncate error message to field size (2000)
	if len(errMsg) > 2000 {
		errMsg = errMsg[:1997] + "..."
	}
	// Compute duration from StartedAt
	var run models.SyncRunEntity
	if err := r.db.WithContext(ctx).Select("started_at").Where("id = ?", id).First(&run).Error; err != nil {
		return fmt.Errorf("syncTrackingRepo.CompleteRun.Read: %w", err)
	}
	duration := now.Sub(run.StartedAt).Milliseconds()

	updates := map[string]interface{}{
		"status":        status,
		"finished_at":   now,
		"duration_ms":   duration,
		"rows_fetched":  rowsFetched,
		"rows_inserted": rowsInserted,
		"rows_skipped":  rowsSkipped,
		"error_message": errMsg,
	}
	if err := r.db.WithContext(ctx).
		Model(&models.SyncRunEntity{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("syncTrackingRepo.CompleteRun.Update: %w", err)
	}
	return nil
}

func (r *syncTrackingRepository) IncrementRetry(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&models.SyncRunEntity{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error; err != nil {
		return fmt.Errorf("syncTrackingRepo.IncrementRetry: %w", err)
	}
	return nil
}

func (r *syncTrackingRepository) GetLatest(ctx context.Context, jobType string) (*models.SyncRunEntity, error) {
	var run models.SyncRunEntity
	q := r.db.WithContext(ctx).Order("started_at DESC").Limit(1)
	if jobType != "" {
		q = q.Where("job_type = ?", jobType)
	}
	if err := q.First(&run).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("syncTrackingRepo.GetLatest: %w", err)
	}
	return &run, nil
}

func (r *syncTrackingRepository) GetRecent(ctx context.Context, limit int) ([]models.SyncRunEntity, error) {
	if limit <= 0 {
		limit = 50
	}
	var runs []models.SyncRunEntity
	if err := r.db.WithContext(ctx).
		Order("started_at DESC").
		Limit(limit).
		Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("syncTrackingRepo.GetRecent: %w", err)
	}
	return runs, nil
}

func (r *syncTrackingRepository) GetRunsByJobAndStatus(
	ctx context.Context, jobType, status string, limit int,
) ([]models.SyncRunEntity, error) {
	if limit <= 0 {
		limit = 50
	}
	var runs []models.SyncRunEntity
	q := r.db.WithContext(ctx).Order("started_at DESC").Limit(limit)
	if jobType != "" {
		q = q.Where("job_type = ?", jobType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("syncTrackingRepo.GetRunsByJobAndStatus: %w", err)
	}
	return runs, nil
}

// ReconciliationResult — สรุป row counts สำหรับตรวจสุขภาพ sync
type ReconciliationResult struct {
	Year                 string           `json:"year"`
	RawHMWCount          int64            `json:"raw_hmw_count"`
	RawCLIKCount         int64            `json:"raw_clik_count"`
	RawTotal             int64            `json:"raw_total"`
	FactTransactionCount int64            `json:"fact_transaction_count"`
	FactAmountCount      int64            `json:"fact_amount_count"`
	MonthlyFactCounts    map[string]int64 `json:"monthly_fact_counts"`
	IsHealthy            bool             `json:"is_healthy"`
	Warnings             []string         `json:"warnings"`
}

// GetReconciliation — ตรวจสอบความสอดคล้องระหว่าง raw tables กับ fact tables
// ใช้สำหรับ admin endpoint + auto-check หลัง sync
func (r *syncTrackingRepository) GetReconciliation(ctx context.Context, year string) (*ReconciliationResult, error) {
	result := &ReconciliationResult{
		Year:              year,
		MonthlyFactCounts: make(map[string]int64),
		IsHealthy:         true,
		Warnings:          []string{},
	}

	// 1. Raw HMW count
	if err := r.db.WithContext(ctx).
		Table("achhmw_gle_api").
		Where(`EXTRACT(YEAR FROM "Posting_Date") = ?`, year).
		Count(&result.RawHMWCount).Error; err != nil {
		return nil, fmt.Errorf("count HMW: %w", err)
	}

	// 2. Raw CLIK count
	if err := r.db.WithContext(ctx).
		Table("general_ledger_entries_clik").
		Where(`EXTRACT(YEAR FROM "Posting_Date") = ?`, year).
		Count(&result.RawCLIKCount).Error; err != nil {
		return nil, fmt.Errorf("count CLIK: %w", err)
	}

	result.RawTotal = result.RawHMWCount + result.RawCLIKCount

	// 3. Fact transaction count
	if err := r.db.WithContext(ctx).
		Table("actual_transaction_entities").
		Where("year = ?", year).
		Count(&result.FactTransactionCount).Error; err != nil {
		return nil, fmt.Errorf("count actual_transactions: %w", err)
	}

	// 4. Fact amount rows (aggregated)
	if err := r.db.WithContext(ctx).
		Table("actual_amount_entities aa").
		Joins("JOIN actual_fact_entities af ON aa.actual_fact_id = af.id").
		Where("af.year = ?", year).
		Count(&result.FactAmountCount).Error; err != nil {
		return nil, fmt.Errorf("count actual_amounts: %w", err)
	}

	// 5. Monthly fact transaction counts
	type monthCount struct {
		Month string
		Count int64
	}
	var monthlyRows []monthCount
	if err := r.db.WithContext(ctx).
		Table("actual_transaction_entities").
		Select(`UPPER(TO_CHAR(posting_date::DATE, 'MON')) as month, COUNT(*) as count`).
		Where("year = ?", year).
		Group(`UPPER(TO_CHAR(posting_date::DATE, 'MON'))`).
		Scan(&monthlyRows).Error; err != nil {
		return nil, fmt.Errorf("monthly counts: %w", err)
	}
	for _, m := range monthlyRows {
		result.MonthlyFactCounts[m.Month] = m.Count
	}

	// 6. Health checks
	if result.RawTotal > 0 && result.FactTransactionCount == 0 {
		result.IsHealthy = false
		result.Warnings = append(result.Warnings, "Fact transactions empty while raw tables have data — sync may have failed")
	}
	if result.FactTransactionCount > 0 && result.FactAmountCount == 0 {
		result.IsHealthy = false
		result.Warnings = append(result.Warnings, "Fact transactions exist but amount aggregations are empty — fact build may have failed")
	}
	// Mapping filter drops rows; ratio < 1 is normal but extremely low ratio is suspicious
	if result.RawTotal > 0 && result.FactTransactionCount > 0 {
		ratio := float64(result.FactTransactionCount) / float64(result.RawTotal)
		if ratio < 0.01 {
			result.IsHealthy = false
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Fact/Raw ratio very low (%.4f) — most transactions filtered; check GL mapping coverage", ratio))
		}
	}

	return result, nil
}

// GetFailedRunsForRetry — FAILED runs ใน {within} ชั่วโมงย้อนหลังที่ retry_count < maxRetries
// ใช้โดย retry cron เพื่อจัดคิว re-run
func (r *syncTrackingRepository) GetFailedRunsForRetry(ctx context.Context, within time.Duration, maxRetries int) ([]models.SyncRunEntity, error) {
	cutoff := time.Now().Add(-within)
	var runs []models.SyncRunEntity
	if err := r.db.WithContext(ctx).
		Where("status = ? AND started_at >= ? AND retry_count < ?",
			models.SyncStatusFailed, cutoff, maxRetries).
		Order("started_at ASC").
		Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("syncTrackingRepo.GetFailedRunsForRetry: %w", err)
	}
	return runs, nil
}

// DeleteOldRunsByJobType — ลบ sync_runs ของ jobType ที่เก่ากว่า olderThan
// ใช้กับ TIER1_FAST (รันทุก 5 นาที = 288 แถว/วัน) ป้องกัน table โต disk เต็ม
func (r *syncTrackingRepository) DeleteOldRunsByJobType(ctx context.Context, jobType string, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	res := r.db.WithContext(ctx).
		Where("job_type = ? AND started_at < ?", jobType, cutoff).
		Delete(&models.SyncRunEntity{})
	if res.Error != nil {
		return 0, fmt.Errorf("syncTrackingRepo.DeleteOldRunsByJobType: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// ClearStaleRunningRuns — clean up RUNNING rows ที่ค้างจาก server crash
// (ใช้ตอน server startup เพื่อ detect + mark เป็น FAILED)
func (r *syncTrackingRepository) ClearStaleRunningRuns(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	res := r.db.WithContext(ctx).
		Model(&models.SyncRunEntity{}).
		Where("status = ? AND started_at < ?", models.SyncStatusRunning, cutoff).
		Updates(map[string]interface{}{
			"status":        models.SyncStatusFailed,
			"error_message": "marked FAILED by startup cleanup (likely server crashed mid-run)",
			"finished_at":   time.Now(),
		})
	if res.Error != nil {
		return 0, fmt.Errorf("syncTrackingRepo.ClearStaleRunningRuns: %w", res.Error)
	}
	return res.RowsAffected, nil
}

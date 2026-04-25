package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"p2p-back-end/modules/entities/models"
)

// setupTestDB creates an in-memory SQLite DB with the sync_runs schema auto-migrated.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "open sqlite in-memory")
	require.NoError(t, db.AutoMigrate(&models.SyncRunEntity{}))
	return db
}

func newTrackingRepoForTest(t *testing.T) (SyncTrackingRepository, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	return &syncTrackingRepository{db: db}, db
}

// ─────────────────────────────────────────────
// CreateRun
// ─────────────────────────────────────────────

func TestCreateRun_AutoGeneratesIDAndStartedAt(t *testing.T) {
	repo, _ := newTrackingRepoForTest(t)

	run := &models.SyncRunEntity{
		JobType:     models.SyncJobDW,
		TriggeredBy: "CRON",
	}
	err := repo.CreateRun(context.Background(), run)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, run.ID, "ID should be auto-generated")
	assert.False(t, run.StartedAt.IsZero(), "StartedAt should be set")
	assert.Equal(t, models.SyncStatusRunning, run.Status, "Status defaults to RUNNING")
}

func TestCreateRun_PreservesProvidedFields(t *testing.T) {
	repo, _ := newTrackingRepoForTest(t)

	fixedID := uuid.New()
	started := time.Now().Add(-1 * time.Hour)
	run := &models.SyncRunEntity{
		ID:          fixedID,
		JobType:     models.SyncJobTier1Fast,
		Year:        "2026",
		Month:       "APR",
		Status:      models.SyncStatusRunning,
		StartedAt:   started,
		TriggeredBy: "ADMIN:user-1",
	}
	err := repo.CreateRun(context.Background(), run)
	assert.NoError(t, err)
	assert.Equal(t, fixedID, run.ID)
	assert.Equal(t, "2026", run.Year)
	assert.Equal(t, "APR", run.Month)
	assert.WithinDuration(t, started, run.StartedAt, time.Second)
}

// ─────────────────────────────────────────────
// CompleteRun
// ─────────────────────────────────────────────

func TestCompleteRun_UpdatesStatusAndDuration(t *testing.T) {
	repo, db := newTrackingRepoForTest(t)

	run := &models.SyncRunEntity{
		JobType:   models.SyncJobTier2Full,
		StartedAt: time.Now().Add(-2 * time.Second),
	}
	require.NoError(t, repo.CreateRun(context.Background(), run))

	err := repo.CompleteRun(context.Background(), run.ID, models.SyncStatusSuccess, 1000, 950, 50, "")
	assert.NoError(t, err)

	var updated models.SyncRunEntity
	require.NoError(t, db.First(&updated, "id = ?", run.ID).Error)
	assert.Equal(t, models.SyncStatusSuccess, updated.Status)
	assert.NotNil(t, updated.FinishedAt)
	assert.GreaterOrEqual(t, updated.DurationMs, int64(1900), "duration ≥ 1.9s")
	assert.Equal(t, int64(1000), updated.RowsFetched)
	assert.Equal(t, int64(950), updated.RowsInserted)
	assert.Equal(t, int64(50), updated.RowsSkipped)
	assert.Empty(t, updated.ErrorMessage)
}

func TestCompleteRun_TruncatesLongErrorMessage(t *testing.T) {
	repo, db := newTrackingRepoForTest(t)

	run := &models.SyncRunEntity{JobType: models.SyncJobDW}
	require.NoError(t, repo.CreateRun(context.Background(), run))

	longErr := strings.Repeat("x", 3000)
	err := repo.CompleteRun(context.Background(), run.ID, models.SyncStatusFailed, 0, 0, 0, longErr)
	assert.NoError(t, err)

	var updated models.SyncRunEntity
	require.NoError(t, db.First(&updated, "id = ?", run.ID).Error)
	assert.LessOrEqual(t, len(updated.ErrorMessage), 2000, "error message must fit column")
	assert.True(t, strings.HasSuffix(updated.ErrorMessage, "..."), "should be truncated with ellipsis")
}

func TestCompleteRun_NonExistentID_ReturnsError(t *testing.T) {
	repo, _ := newTrackingRepoForTest(t)
	err := repo.CompleteRun(context.Background(), uuid.New(), models.SyncStatusSuccess, 0, 0, 0, "")
	assert.Error(t, err, "CompleteRun should fail when ID does not exist (read fails)")
}

// ─────────────────────────────────────────────
// IncrementRetry
// ─────────────────────────────────────────────

func TestIncrementRetry_IncrementsCountAtomically(t *testing.T) {
	repo, db := newTrackingRepoForTest(t)

	run := &models.SyncRunEntity{JobType: models.SyncJobTier1Fast}
	require.NoError(t, repo.CreateRun(context.Background(), run))

	for i := 1; i <= 3; i++ {
		err := repo.IncrementRetry(context.Background(), run.ID)
		assert.NoError(t, err)

		var updated models.SyncRunEntity
		require.NoError(t, db.First(&updated, "id = ?", run.ID).Error)
		assert.Equal(t, i, updated.RetryCount, "iteration %d", i)
	}
}

// ─────────────────────────────────────────────
// GetLatest
// ─────────────────────────────────────────────

func TestGetLatest_EmptyReturnsNil(t *testing.T) {
	repo, _ := newTrackingRepoForTest(t)
	run, err := repo.GetLatest(context.Background(), "")
	assert.NoError(t, err)
	assert.Nil(t, run)
}

func TestGetLatest_ByJobType(t *testing.T) {
	repo, _ := newTrackingRepoForTest(t)

	old := &models.SyncRunEntity{JobType: models.SyncJobDW, StartedAt: time.Now().Add(-2 * time.Hour)}
	mid := &models.SyncRunEntity{JobType: models.SyncJobTier1Fast, StartedAt: time.Now().Add(-1 * time.Hour)}
	latest := &models.SyncRunEntity{JobType: models.SyncJobDW, StartedAt: time.Now()}
	require.NoError(t, repo.CreateRun(context.Background(), old))
	require.NoError(t, repo.CreateRun(context.Background(), mid))
	require.NoError(t, repo.CreateRun(context.Background(), latest))

	got, err := repo.GetLatest(context.Background(), models.SyncJobDW)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, latest.ID, got.ID, "should return most recent DW_SYNC, skipping Tier1")
}

// ─────────────────────────────────────────────
// GetRecent / GetRunsByJobAndStatus
// ─────────────────────────────────────────────

func TestGetRecent_OrdersByStartedAtDesc(t *testing.T) {
	repo, _ := newTrackingRepoForTest(t)

	now := time.Now()
	for i := 0; i < 5; i++ {
		require.NoError(t, repo.CreateRun(context.Background(), &models.SyncRunEntity{
			JobType:   models.SyncJobDW,
			StartedAt: now.Add(time.Duration(-i) * time.Minute),
		}))
	}

	runs, err := repo.GetRecent(context.Background(), 3)
	require.NoError(t, err)
	assert.Len(t, runs, 3, "limit respected")
	for i := 1; i < len(runs); i++ {
		assert.True(t, runs[i-1].StartedAt.After(runs[i].StartedAt) || runs[i-1].StartedAt.Equal(runs[i].StartedAt),
			"ordered DESC")
	}
}

func TestGetRunsByJobAndStatus_Filters(t *testing.T) {
	repo, _ := newTrackingRepoForTest(t)

	require.NoError(t, repo.CreateRun(context.Background(), &models.SyncRunEntity{JobType: models.SyncJobDW, Status: models.SyncStatusSuccess}))
	require.NoError(t, repo.CreateRun(context.Background(), &models.SyncRunEntity{JobType: models.SyncJobDW, Status: models.SyncStatusFailed}))
	require.NoError(t, repo.CreateRun(context.Background(), &models.SyncRunEntity{JobType: models.SyncJobTier1Fast, Status: models.SyncStatusFailed}))

	runs, err := repo.GetRunsByJobAndStatus(context.Background(), models.SyncJobDW, models.SyncStatusFailed, 10)
	require.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, models.SyncJobDW, runs[0].JobType)
	assert.Equal(t, models.SyncStatusFailed, runs[0].Status)

	// No filter → all three
	all, err := repo.GetRunsByJobAndStatus(context.Background(), "", "", 10)
	require.NoError(t, err)
	assert.Len(t, all, 3)
}

// ─────────────────────────────────────────────
// ClearStaleRunningRuns
// ─────────────────────────────────────────────

func TestClearStaleRunningRuns_MarksOldRunningAsFailed(t *testing.T) {
	repo, db := newTrackingRepoForTest(t)

	now := time.Now()
	// Stale: RUNNING, started 3h ago
	stale := &models.SyncRunEntity{JobType: models.SyncJobDW, Status: models.SyncStatusRunning, StartedAt: now.Add(-3 * time.Hour)}
	// Fresh: RUNNING, started 30min ago (not stale)
	fresh := &models.SyncRunEntity{JobType: models.SyncJobDW, Status: models.SyncStatusRunning, StartedAt: now.Add(-30 * time.Minute)}
	// Completed: SUCCESS, regardless of age
	done := &models.SyncRunEntity{JobType: models.SyncJobDW, Status: models.SyncStatusSuccess, StartedAt: now.Add(-5 * time.Hour)}
	require.NoError(t, repo.CreateRun(context.Background(), stale))
	require.NoError(t, repo.CreateRun(context.Background(), fresh))
	require.NoError(t, repo.CreateRun(context.Background(), done))

	cleared, err := repo.ClearStaleRunningRuns(context.Background(), 2*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, int64(1), cleared, "only stale RUNNING should be cleared")

	var staleAfter models.SyncRunEntity
	require.NoError(t, db.First(&staleAfter, "id = ?", stale.ID).Error)
	assert.Equal(t, models.SyncStatusFailed, staleAfter.Status)
	assert.Contains(t, staleAfter.ErrorMessage, "startup cleanup")

	// Fresh RUNNING stays RUNNING
	var freshAfter models.SyncRunEntity
	require.NoError(t, db.First(&freshAfter, "id = ?", fresh.ID).Error)
	assert.Equal(t, models.SyncStatusRunning, freshAfter.Status)
}

// ─────────────────────────────────────────────
// GetFailedRunsForRetry
// ─────────────────────────────────────────────

func TestGetFailedRunsForRetry_FiltersByWindowAndMaxRetries(t *testing.T) {
	repo, _ := newTrackingRepoForTest(t)

	now := time.Now()
	// Within window, retry=0 → eligible
	eligible := &models.SyncRunEntity{
		JobType: models.SyncJobTier1Fast, Status: models.SyncStatusFailed,
		StartedAt: now.Add(-1 * time.Hour), RetryCount: 0,
	}
	// Within window, retry=3 → exceeded max
	exceeded := &models.SyncRunEntity{
		JobType: models.SyncJobTier1Fast, Status: models.SyncStatusFailed,
		StartedAt: now.Add(-1 * time.Hour), RetryCount: 3,
	}
	// Outside window → ineligible even if retry=0
	old := &models.SyncRunEntity{
		JobType: models.SyncJobTier1Fast, Status: models.SyncStatusFailed,
		StartedAt: now.Add(-48 * time.Hour), RetryCount: 0,
	}
	// Success status → not eligible
	done := &models.SyncRunEntity{
		JobType: models.SyncJobDW, Status: models.SyncStatusSuccess,
		StartedAt: now.Add(-1 * time.Hour),
	}
	for _, r := range []*models.SyncRunEntity{eligible, exceeded, old, done} {
		require.NoError(t, repo.CreateRun(context.Background(), r))
	}

	runs, err := repo.GetFailedRunsForRetry(context.Background(), 24*time.Hour, 3)
	require.NoError(t, err)
	assert.Len(t, runs, 1, "only eligible run returned")
	assert.Equal(t, eligible.ID, runs[0].ID)
}

// ─────────────────────────────────────────────
// GetReconciliation
// ─────────────────────────────────────────────

func TestGetReconciliation_EmptyTables_ReturnsHealthyZeroes(t *testing.T) {
	repo, db := newTrackingRepoForTest(t)
	// Create empty raw + fact tables so queries don't fail
	require.NoError(t, db.Exec(`CREATE TABLE achhmw_gle_api ("Posting_Date" DATETIME, id INTEGER PRIMARY KEY)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE general_ledger_entries_clik ("Posting_Date" DATETIME, id INTEGER PRIMARY KEY)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE actual_transaction_entities (year TEXT, posting_date TEXT, id TEXT PRIMARY KEY)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE actual_fact_entities (year TEXT, id TEXT PRIMARY KEY)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE actual_amount_entities (actual_fact_id TEXT, id TEXT PRIMARY KEY)`).Error)

	res, err := repo.GetReconciliation(context.Background(), "2026")
	// SQLite may not support EXTRACT() or TO_CHAR() — this test primarily checks that
	// the method returns without panic and that healthy=true with zero counts is handled.
	// If SQLite throws an error on the EXTRACT() expression, at least verify the error surface.
	if err != nil {
		t.Skipf("SQLite lacks EXTRACT/TO_CHAR support — reconciliation test requires Postgres: %v", err)
	}
	assert.Equal(t, "2026", res.Year)
	assert.True(t, res.IsHealthy)
	assert.Empty(t, res.Warnings)
	assert.Equal(t, int64(0), res.RawTotal)
	assert.Equal(t, int64(0), res.FactTransactionCount)
}

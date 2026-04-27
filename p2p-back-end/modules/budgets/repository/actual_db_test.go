package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

func TestMain(m *testing.M) {
	logs.Loginit()
	os.Exit(m.Run())
}

// setupTestDB returns an in-memory SQLite DB with hand-written schemas that
// match the production GORM models. We don't use AutoMigrate because the model
// tags include Postgres-specific defaults (`uuid_generate_v4()`) that SQLite
// rejects.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	stmts := []string{
		`CREATE TABLE actual_transaction_entities (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			source TEXT,
			posting_date TEXT,
			doc_no TEXT,
			description TEXT,
			amount REAL,
			vendor_name TEXT,
			gl_account_name TEXT,
			entity TEXT,
			branch TEXT,
			department TEXT,
			entity_gl TEXT,
			conso_gl TEXT,
			year TEXT,
			is_valid BOOLEAN DEFAULT 1,
			status TEXT DEFAULT 'PENDING'
		)`,
		`CREATE INDEX idx_atx_posting_date ON actual_transaction_entities(posting_date)`,
		`CREATE INDEX idx_atx_year ON actual_transaction_entities(year)`,
		`CREATE INDEX idx_atx_status ON actual_transaction_entities(status)`,

		`CREATE TABLE actual_fact_entities (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			entity TEXT,
			branch TEXT,
			department TEXT,
			nav_code TEXT,
			"group" TEXT,
			entity_gl TEXT,
			conso_gl TEXT,
			gl_name TEXT,
			vendor_name TEXT,
			year TEXT,
			year_total REAL DEFAULT 0,
			is_valid BOOLEAN DEFAULT 1
		)`,

		`CREATE TABLE actual_amount_entities (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			actual_fact_id TEXT,
			month TEXT,
			amount REAL DEFAULT 0
		)`,
	}
	for _, s := range stmts {
		require.NoError(t, db.Exec(s).Error)
	}
	return db
}

func newTransaction(year, postingDate, entity, branch, dept, gl, doc, status string, amount float64) models.ActualTransactionEntity {
	return models.ActualTransactionEntity{
		ID:          uuid.New(),
		Source:      "TEST",
		PostingDate: postingDate,
		DocNo:       doc,
		Amount:      decimal.NewFromFloat(amount),
		Entity:      entity,
		Branch:      branch,
		Department:  dept,
		EntityGL:    gl,
		ConsoGL:     gl,
		Year:        year,
		Status:      status,
	}
}

// ─────────────────────────── DeleteActualTransactionsByYear ───────────────────────────

// Regression test for the "Detail vs Dashboard mismatch" bug:
// DeleteActualTransactionsByYear used to delete only PENDING rows, leaving stale
// REPORTED/COMPLETE rows that no longer matched raw → caused phantom row inflation.
// After the fix, ALL transactions for the year are deleted; status is preserved
// separately via GetNonPendingTransactionKeys → RestoreTransactionStatuses.
func TestDeleteActualTransactionsByYear_DeletesAllStatuses(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)

	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{
		newTransaction("2026", "2026-01-15", "ACG", "HQ", "ACC", "5100", "DOC-1", models.TxStatusPending, 100),
		newTransaction("2026", "2026-02-15", "ACG", "HQ", "ACC", "5100", "DOC-2", models.TxStatusReported, 200),
		newTransaction("2026", "2026-03-15", "ACG", "HQ", "ACC", "5100", "DOC-3", models.TxStatusComplete, 300),
		newTransaction("2025", "2025-12-15", "ACG", "HQ", "ACC", "5100", "DOC-OLD", models.TxStatusPending, 50),
	}).Error)

	require.NoError(t, r.DeleteActualTransactionsByYear(context.Background(), "2026"))

	var remaining []models.ActualTransactionEntity
	require.NoError(t, db.Find(&remaining).Error)
	require.Len(t, remaining, 1, "year=2025 row must survive; all year=2026 rows deleted regardless of status")
	assert.Equal(t, "DOC-OLD", remaining[0].DocNo)
}

// ─────────────────────────── DeleteActualTransactionsByMonth ───────────────────────────

func TestDeleteActualTransactionsByMonth_DeletesAllStatusesForMonth(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)

	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{
		newTransaction("2026", "2026-04-01", "ACG", "HQ", "ACC", "5100", "APR-1", models.TxStatusPending, 100),
		newTransaction("2026", "2026-04-15", "ACG", "HQ", "ACC", "5100", "APR-2", models.TxStatusReported, 200),
		newTransaction("2026", "2026-04-20", "ACG", "HQ", "ACC", "5100", "APR-3", models.TxStatusComplete, 300),
		newTransaction("2026", "2026-03-01", "ACG", "HQ", "ACC", "5100", "MAR-1", models.TxStatusPending, 400),
	}).Error)

	require.NoError(t, r.DeleteActualTransactionsByMonth(context.Background(), "2026", "APR"))

	var remaining []models.ActualTransactionEntity
	require.NoError(t, db.Find(&remaining).Error)
	require.Len(t, remaining, 1)
	assert.Equal(t, "MAR-1", remaining[0].DocNo, "only March transactions survive")
}

func TestDeleteActualTransactionsByMonth_InvalidMonth(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)
	err := r.DeleteActualTransactionsByMonth(context.Background(), "2026", "FOO")
	assert.Error(t, err)
}

// ─────────────────────────── RestoreTransactionStatuses ───────────────────────────

// Regression: previously this routine ran a "DELETE … LIMIT 1" cleanup keyed only
// on (entity, entity_gl, doc_no, posting_date), which silently destroyed legitimate
// multi-line documents (same doc / GL with different branch or department).
// The new implementation only UPDATEs status and never deletes — concurrency is
// prevented by SyncMutex / queue, not by per-row de-duplication.
func TestRestoreTransactionStatuses_PromotesAllMatchingPendingRows(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)

	// Two PENDING rows share (entity, entity_gl, doc_no, posting_date) but have different branches.
	// Both must be promoted to REPORTED — neither deleted.
	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{
		newTransaction("2026", "2026-04-01", "ACG", "HQ", "ACC", "5100", "DOC-1", models.TxStatusPending, 100),
		newTransaction("2026", "2026-04-01", "ACG", "BR1", "SAL", "5100", "DOC-1", models.TxStatusPending, 200),
	}).Error)

	statusMap := map[string]string{
		"ACG|5100|DOC-1|2026-04-01": models.TxStatusReported,
	}
	require.NoError(t, r.RestoreTransactionStatuses(context.Background(), statusMap))

	var rows []models.ActualTransactionEntity
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 2, "both legit multi-line rows must remain — none deleted")
	for _, row := range rows {
		assert.Equal(t, models.TxStatusReported, row.Status)
	}
}

func TestRestoreTransactionStatuses_EmptyMapNoOp(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)
	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{
		newTransaction("2026", "2026-04-01", "ACG", "HQ", "ACC", "5100", "DOC-1", models.TxStatusPending, 100),
	}).Error)

	require.NoError(t, r.RestoreTransactionStatuses(context.Background(), nil))
	require.NoError(t, r.RestoreTransactionStatuses(context.Background(), map[string]string{}))

	var rows []models.ActualTransactionEntity
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 1)
	assert.Equal(t, models.TxStatusPending, rows[0].Status)
}

func TestRestoreTransactionStatuses_OnlyTargetsPending(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)

	// A row already at COMPLETE must NOT be touched even if its key is in the map.
	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{
		newTransaction("2026", "2026-04-01", "ACG", "HQ", "ACC", "5100", "DOC-1", models.TxStatusComplete, 100),
		newTransaction("2026", "2026-04-01", "ACG", "BR1", "ACC", "5100", "DOC-1", models.TxStatusPending, 200),
	}).Error)

	statusMap := map[string]string{
		"ACG|5100|DOC-1|2026-04-01": models.TxStatusReported,
	}
	require.NoError(t, r.RestoreTransactionStatuses(context.Background(), statusMap))

	var rows []models.ActualTransactionEntity
	require.NoError(t, db.Order("amount").Find(&rows).Error)
	require.Len(t, rows, 2)
	assert.Equal(t, models.TxStatusComplete, rows[0].Status, "existing COMPLETE row stays COMPLETE")
	assert.Equal(t, models.TxStatusReported, rows[1].Status, "PENDING row promoted to REPORTED")
}

// ─────────────────────────── DeleteActualFactsByMonth (year_total=0 bug) ───────────────────────────

// Regression: the cleanup step used to run "DELETE … WHERE year_total = 0", which
// also wiped facts whose JAN+FEB+MAR amounts coincidentally cancelled to 0
// (e.g. credit/debit pairs). After deletion, the surviving JAN/FEB/MAR amount
// rows became orphans not joinable from the dashboard query. The new logic uses
// "WHERE NOT EXISTS amounts" so only truly empty facts are removed.
func TestDeleteActualFactsByMonth_KeepsFactsWhoseRemainingAmountsSumToZero(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)

	factID := uuid.New()
	require.NoError(t, db.Create(&models.ActualFactEntity{
		ID: factID, Entity: "ACG", Branch: "HQ", Department: "ACC",
		EntityGL: "5100", ConsoGL: "5100", Year: "2026",
		YearTotal: decimal.Zero,
	}).Error)

	// Non-APR amounts that sum to 0 (legit credit/debit cancellation).
	require.NoError(t, db.Create(&[]models.ActualAmountEntity{
		{ID: uuid.New(), ActualFactID: factID, Month: "JAN", Amount: decimal.NewFromInt(100)},
		{ID: uuid.New(), ActualFactID: factID, Month: "FEB", Amount: decimal.NewFromInt(-100)},
		{ID: uuid.New(), ActualFactID: factID, Month: "APR", Amount: decimal.NewFromInt(50)},
	}).Error)

	// Recompute year_total to mirror what production does.
	require.NoError(t, db.Exec(`UPDATE actual_fact_entities SET year_total = 50 WHERE id = ?`, factID).Error)

	require.NoError(t, r.DeleteActualFactsByMonth(context.Background(), "2026", "APR"))

	// Fact must STILL EXIST: it has remaining JAN/FEB amounts even though they sum to 0.
	var fact models.ActualFactEntity
	err := db.First(&fact, "id = ?", factID).Error
	require.NoError(t, err, "fact entity must survive — bug previously deleted it because year_total=0")

	// APR amount removed; JAN/FEB amounts intact.
	var amounts []models.ActualAmountEntity
	require.NoError(t, db.Where("actual_fact_id = ?", factID).Find(&amounts).Error)
	require.Len(t, amounts, 2)
	for _, a := range amounts {
		assert.NotEqual(t, "APR", a.Month, "APR amount should be removed")
	}
}

func TestDeleteActualFactsByMonth_RemovesEmptiedFact(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)

	factID := uuid.New()
	require.NoError(t, db.Create(&models.ActualFactEntity{
		ID: factID, Entity: "ACG", Branch: "HQ", Department: "ACC",
		EntityGL: "5100", ConsoGL: "5100", Year: "2026",
		YearTotal: decimal.NewFromInt(99),
	}).Error)

	// Only APR amount — after deleting APR, fact has no amounts left at all.
	require.NoError(t, db.Create(&models.ActualAmountEntity{
		ID: uuid.New(), ActualFactID: factID, Month: "APR", Amount: decimal.NewFromInt(99),
	}).Error)

	require.NoError(t, r.DeleteActualFactsByMonth(context.Background(), "2026", "APR"))

	// Fact must be gone — it has no amount rows referencing it.
	var fact models.ActualFactEntity
	err := db.First(&fact, "id = ?", factID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "fact with no remaining amounts should be cleaned up")
}

// ─────────────────────────── GetNonPendingTransactionKeys ───────────────────────────

func TestGetNonPendingTransactionKeys_ReturnsOnlyNonPending(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)

	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{
		newTransaction("2026", "2026-04-01", "ACG", "HQ", "ACC", "5100", "DOC-PENDING", models.TxStatusPending, 100),
		newTransaction("2026", "2026-04-02", "ACG", "HQ", "ACC", "5100", "DOC-REPORTED", models.TxStatusReported, 100),
		newTransaction("2026", "2026-04-03", "ACG", "HQ", "ACC", "5100", "DOC-COMPLETE", models.TxStatusComplete, 100),
	}).Error)

	keys, err := r.GetNonPendingTransactionKeys(context.Background(), "2026", []string{"APR"})
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "ACG|5100|DOC-REPORTED|2026-04-02")
	assert.Contains(t, keys, "ACG|5100|DOC-COMPLETE|2026-04-03")
	for k := range keys {
		assert.NotContains(t, k, "DOC-PENDING")
	}
}

// ─────────────────────────── End-to-end: preserve-then-restore round trip ───────────────────────────

// This is the critical contract: after a sync cycle, an item that was REPORTED
// before the sync must be REPORTED again — even though we now hard-delete the
// row in step 2. The test simulates the full sequence used in SyncActuals.
func TestPreserveDeleteRestore_EndToEnd(t *testing.T) {
	db := setupTestDB(t)
	r := NewActualRepository(db)
	ctx := context.Background()

	// Initial: one REPORTED row.
	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{
		newTransaction("2026", "2026-04-01", "ACG", "HQ", "ACC", "5100", "DOC-1", models.TxStatusReported, 100),
	}).Error)

	// Step 1: capture preserved statuses BEFORE deletion.
	preserved, err := r.GetNonPendingTransactionKeys(ctx, "2026", []string{"APR"})
	require.NoError(t, err)
	require.Len(t, preserved, 1)

	// Step 2: hard-delete (simulates DeleteActualTransactionsByMonth in fresh sync).
	require.NoError(t, r.DeleteActualTransactionsByMonth(ctx, "2026", "APR"))

	var afterDelete []models.ActualTransactionEntity
	require.NoError(t, db.Find(&afterDelete).Error)
	require.Len(t, afterDelete, 0, "all APR rows deleted")

	// Step 3: re-insert from raw (simulates CreateActualTransactions).
	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{
		newTransaction("2026", "2026-04-01", "ACG", "HQ", "ACC", "5100", "DOC-1", models.TxStatusPending, 100),
	}).Error)

	// Step 4: restore status from preserved map.
	require.NoError(t, r.RestoreTransactionStatuses(ctx, preserved))

	var final []models.ActualTransactionEntity
	require.NoError(t, db.Find(&final).Error)
	require.Len(t, final, 1)
	assert.Equal(t, models.TxStatusReported, final[0].Status, "REPORTED status must be preserved across the sync cycle")
}

// Sanity test: NewActualRepository must drop the legacy too-narrow unique index
// if it's hanging around from before the fix. Without this, multi-line documents
// would silently collapse. We simulate by creating the bad index then re-running
// the constructor and asserting the index is gone.
func TestNewActualRepository_DropsLegacyNarrowUniqueIndex(t *testing.T) {
	db := setupTestDB(t)

	// Create the bad legacy index manually.
	err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uniq_actual_txn_business_key ON actual_transaction_entities (entity, entity_gl, doc_no, posting_date)`).Error
	require.NoError(t, err)

	// Constructor should drop it.
	_ = NewActualRepository(db)

	// SQLite reports indexes via PRAGMA index_list — the legacy name should be absent.
	type idxRow struct {
		Seq  int
		Name string
	}
	var indexes []idxRow
	require.NoError(t, db.Raw(`PRAGMA index_list(actual_transaction_entities)`).Scan(&indexes).Error)
	for _, idx := range indexes {
		assert.NotEqual(t, "uniq_actual_txn_business_key", idx.Name, "legacy narrow unique index must be dropped")
	}

	// And to prove it: a multi-line insert that the legacy index would have rejected must succeed.
	tx := newTransaction("2026", "2026-04-01", "ACG", "HQ", "ACC", "5100", "DOC-1", models.TxStatusPending, 100)
	tx2 := newTransaction("2026", "2026-04-01", "ACG", "BR1", "ACC", "5100", "DOC-1", models.TxStatusPending, 200) // same key, different branch
	require.NoError(t, db.Create(&[]models.ActualTransactionEntity{tx, tx2}).Error)
}

// TestActualTransactionEntity_TableNameMatches confirms the constants used in
// our raw SQL still match the GORM model — guards against silent breakage if
// someone renames the table.
func TestActualTransactionEntity_TableNameMatches(t *testing.T) {
	assert.Equal(t, "actual_transaction_entities", models.ActualTransactionEntity{}.TableName())
}

// Used to silence unused import warnings in older Go toolchains during
// incremental development. Safe to delete once additional time-based tests are added.
var _ = time.Now

package repository

import (
	"context"
	"database/sql/driver"
	"testing"

	"p2p-back-end/modules/entities/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mockSql, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mockSql
}

// --- UpsertHMWLocal Tests ---

func TestUpsertHMWLocal_EmptyData_ReturnsNil(t *testing.T) {
	localDb, _ := setupMockDB(t)
	dwDb, _ := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	err := repo.UpsertHMWLocal(context.Background(), []models.AchHmwGleEntity{})

	assert.NoError(t, err)
}

func TestUpsertHMWLocal_WithData_ExecutesUpsert(t *testing.T) {
	localDb, mockSql := setupMockDB(t)
	dwDb, _ := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	data := []models.AchHmwGleEntity{
		{
			ID:            1,
			EntryNo:       100,
			GLAccountNo:   "5100",
			GLAccountName: "Cost of Sales",
			Amount:        decimal.NewFromFloat(1500.50),
			Company:       "ACH",
			Branch:        "BKK",
		},
	}

	mockSql.ExpectBegin()
	// 44 args: HMW has 45 columns total but ID is reset to 0 by UpsertHMWLocal,
	// so GORM excludes ID from INSERT (auto-increment) → 44 args go to the query
	mockSql.ExpectQuery("INSERT INTO \"achhmw_gle_api\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mockSql.ExpectCommit()

	err := repo.UpsertHMWLocal(context.Background(), data)

	assert.NoError(t, err)
	assert.NoError(t, mockSql.ExpectationsWereMet())
}

func TestUpsertHMWLocal_DBError_ReturnsWrappedError(t *testing.T) {
	localDb, mockSql := setupMockDB(t)
	dwDb, _ := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	data := []models.AchHmwGleEntity{
		{ID: 1, GLAccountNo: "5100", Company: "ACH"},
	}

	mockSql.ExpectBegin()
	mockSql.ExpectQuery("INSERT INTO \"achhmw_gle_api\"").
		WillReturnError(assert.AnError)
	mockSql.ExpectRollback()

	err := repo.UpsertHMWLocal(context.Background(), data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "extSyncRepo.UpsertHMWLocal")
}

func TestUpsertHMWLocal_MultipleBatches(t *testing.T) {
	localDb, mockSql := setupMockDB(t)
	dwDb, _ := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	// Build nil data slice to verify empty guard
	err := repo.UpsertHMWLocal(context.Background(), nil)
	assert.NoError(t, err)
	assert.NoError(t, mockSql.ExpectationsWereMet())
}

// --- UpsertCLIKLocal Tests ---

func TestUpsertCLIKLocal_EmptyData_ReturnsNil(t *testing.T) {
	localDb, _ := setupMockDB(t)
	dwDb, _ := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	err := repo.UpsertCLIKLocal(context.Background(), []models.ClikGleEntity{})

	assert.NoError(t, err)
}

func TestUpsertCLIKLocal_SetsCompanyToCLIK(t *testing.T) {
	localDb, mockSql := setupMockDB(t)
	dwDb, _ := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	data := []models.ClikGleEntity{
		{ID: 1, GLAccountNo: "6100", Company: ""},
		{ID: 2, GLAccountNo: "6200", Company: "WRONG"},
	}

	mockSql.ExpectBegin()
	// 82 args = 2 records × 41 columns. CLIK has 42 columns total, but ID is reset to 0
	// by UpsertCLIKLocal so GORM excludes it from INSERT (auto-increment).
	args := make([]driver.Value, 82)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	mockSql.ExpectQuery("INSERT INTO \"general_ledger_entries_clik\"").
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
	mockSql.ExpectCommit()

	err := repo.UpsertCLIKLocal(context.Background(), data)

	assert.NoError(t, err)
	// Verify company was overwritten to CLIK
	assert.Equal(t, "CLIK", data[0].Company)
	assert.Equal(t, "CLIK", data[1].Company)
	assert.NoError(t, mockSql.ExpectationsWereMet())
}

func TestUpsertCLIKLocal_DBError_ReturnsWrappedError(t *testing.T) {
	localDb, mockSql := setupMockDB(t)
	dwDb, _ := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	data := []models.ClikGleEntity{
		{ID: 1, GLAccountNo: "6100"},
	}

	mockSql.ExpectBegin()
	mockSql.ExpectQuery("INSERT INTO \"general_ledger_entries_clik\"").
		WillReturnError(assert.AnError)
	mockSql.ExpectRollback()

	err := repo.UpsertCLIKLocal(context.Background(), data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "extSyncRepo.UpsertCLIKLocal")
}

func TestUpsertCLIKLocal_NilData_ReturnsNil(t *testing.T) {
	localDb, mockSql := setupMockDB(t)
	dwDb, _ := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	err := repo.UpsertCLIKLocal(context.Background(), nil)
	assert.NoError(t, err)
	assert.NoError(t, mockSql.ExpectationsWereMet())
}

// --- FetchHMWInBatches Tests ---

func TestFetchHMWInBatches_DBError_ReturnsWrappedError(t *testing.T) {
	localDb, _ := setupMockDB(t)
	dwDb, mockSql := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	mockSql.ExpectQuery("SELECT .+ FROM \"achhmw_gle_api\"").
		WillReturnError(assert.AnError)

	handler := func(data []models.AchHmwGleEntity) error {
		return nil
	}

	err := repo.FetchHMWInBatches(context.Background(), 2026, 1, 2000, handler)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "extSyncRepo.FetchHMWInBatches")
}

// --- FetchCLIKInBatches Tests ---

func TestFetchCLIKInBatches_DBError_ReturnsWrappedError(t *testing.T) {
	localDb, _ := setupMockDB(t)
	dwDb, mockSql := setupMockDB(t)

	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	mockSql.ExpectQuery("SELECT .+ FROM \"general_ledger_entries_clik\"").
		WillReturnError(assert.AnError)

	handler := func(data []models.ClikGleEntity) error {
		return nil
	}

	err := repo.FetchCLIKInBatches(context.Background(), 2026, 3, 2000, handler)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "extSyncRepo.FetchCLIKInBatches")
}

// TestFetchHMWInBatches_CursorPaginationIteratesAllPages confirms the manual cursor
// pagination introduced after FindInBatches errored out with "primary key required".
// Each page must use `id > <last>` as the resume cursor; iteration must stop on a
// short page (rows < batchSize).
func TestFetchHMWInBatches_CursorPaginationIteratesAllPages(t *testing.T) {
	localDb, _ := setupMockDB(t)
	dwDb, mockSql := setupMockDB(t)
	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	// Page 1: 2 rows (id=1, id=2). batchSize=2, so iteration continues.
	mockSql.ExpectQuery(`SELECT .+ FROM "achhmw_gle_api"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 0, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Amount"}).
			AddRow(1, decimal.NewFromInt(100)).
			AddRow(2, decimal.NewFromInt(200)))

	// Page 2: 1 row (id=3). Short page → loop exits.
	mockSql.ExpectQuery(`SELECT .+ FROM "achhmw_gle_api"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 2, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Amount"}).
			AddRow(3, decimal.NewFromInt(300)))

	var collected []int
	err := repo.FetchHMWInBatches(context.Background(), 2026, 1, 2, func(batch []models.AchHmwGleEntity) error {
		for _, row := range batch {
			collected = append(collected, row.ID)
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, collected, "cursor must iterate all rows in id order")
	assert.NoError(t, mockSql.ExpectationsWereMet(), "second page must use id>2 as cursor")
}

// TestFetchHMWInBatches_EmptyResultStopsImmediately ensures we don't loop forever
// when the queried month has no rows (start of empty range).
func TestFetchHMWInBatches_EmptyResultStopsImmediately(t *testing.T) {
	localDb, _ := setupMockDB(t)
	dwDb, mockSql := setupMockDB(t)
	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	mockSql.ExpectQuery(`SELECT .+ FROM "achhmw_gle_api"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Amount"}))

	called := 0
	err := repo.FetchHMWInBatches(context.Background(), 2026, 1, 100, func(batch []models.AchHmwGleEntity) error {
		called++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, called, "handler must not be called for empty page")
	assert.NoError(t, mockSql.ExpectationsWereMet())
}

// TestFetchHMWInBatches_HandlerErrorAborts confirms a returning handler error
// terminates iteration immediately (no further DB queries).
func TestFetchHMWInBatches_HandlerErrorAborts(t *testing.T) {
	localDb, _ := setupMockDB(t)
	dwDb, mockSql := setupMockDB(t)
	repo := &externalSyncRepository{localDb: localDb, dwDb: dwDb}

	mockSql.ExpectQuery(`SELECT .+ FROM "achhmw_gle_api"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Amount"}).
			AddRow(1, decimal.NewFromInt(100)).
			AddRow(2, decimal.NewFromInt(200)))

	handlerErr := assert.AnError
	err := repo.FetchHMWInBatches(context.Background(), 2026, 1, 2, func(batch []models.AchHmwGleEntity) error {
		return handlerErr
	})

	assert.ErrorIs(t, err, handlerErr)
	// Only 1 query must have been made — the second batch must NOT be fetched.
	assert.NoError(t, mockSql.ExpectationsWereMet())
}

// silence unused import on driver — kept for future tests that expect typed args.
var _ = driver.Value(nil)

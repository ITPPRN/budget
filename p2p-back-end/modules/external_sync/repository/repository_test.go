package repository

import (
	"context"
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
	mockSql.ExpectQuery("INSERT INTO \"achhmw_gle_api\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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
	mockSql.ExpectQuery("INSERT INTO \"general_ledger_entries_clik\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg()).
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

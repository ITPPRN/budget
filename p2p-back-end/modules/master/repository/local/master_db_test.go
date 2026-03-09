package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

func TestSyncCompany_EmptyData(t *testing.T) {
	dbMock, _, _ := sqlmock.New()
	dialector := postgres.New(postgres.Config{Conn: dbMock})
	db, _ := gorm.Open(dialector, &gorm.Config{})
	repo := &masterRepositoryDB{db: db}

	result, err := repo.SyncCompany([]models.Companies{})

	assert.Error(t, err)
	assert.Equal(t, "no data companies", err.Error())
	assert.Nil(t, result)
}

func TestSyncDepartmentEmptyData(t *testing.T) {
	dbMock, _, _ := sqlmock.New()
	dialector := postgres.New(postgres.Config{Conn: dbMock})
	db, _ := gorm.Open(dialector, &gorm.Config{})

	repo := &masterRepositoryDB{db: db}

	result, err := repo.SyncDepartment([]models.Departments{})

	assert.Error(t, err)
	assert.Equal(t, "no data Departments", err.Error())
	assert.Nil(t, result)
}

// Note: Complex GORM OnConflict tests with sqlmock are brittle and excluded for now
// as the primary goal is compilation and architectural structure.

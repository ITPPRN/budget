package models

import (
	"mime/multipart"
)

// CapexRepository Interface
type CapexRepository interface {
	// Transaction Helper
	WithTrx(trxHandle func(repo CapexRepository) error) error

	// Create Files
	CreateFileCapexBudget(file *FileCapexBudgetEntity) error
	CreateFileCapexActual(file *FileCapexActualEntity) error

	// Create Facts
	CreateCapexBudgetFacts(facts []CapexBudgetFactEntity) error
	CreateCapexActualFacts(facts []CapexActualFactEntity) error

	// List Files
	ListFileCapexBudgets() ([]FileCapexBudgetEntity, error)
	ListFileCapexActuals() ([]FileCapexActualEntity, error)

	// Get Single File
	GetFileCapexBudget(id string) (*FileCapexBudgetEntity, error)
	GetFileCapexActual(id string) (*FileCapexActualEntity, error)

	// Delete Files
	DeleteFileCapexBudget(id string) error
	DeleteFileCapexActual(id string) error

	// Delete Facts
	DeleteAllCapexBudgetFacts() error
	DeleteAllCapexActualFacts() error

	// Update Files (Rename)
	UpdateFileCapexBudget(id string, filename string) error
	UpdateFileCapexActual(id string, filename string) error

	// Dashboard Aggregation
	GetCapexDashboardAggregates(filter map[string]interface{}) (*DashboardSummaryDTO, error)
}

// CapexService Interface
type CapexService interface {
	// Import
	ImportCapexBudget(file *multipart.FileHeader, userID string, versionName string) error
	ImportCapexActual(file *multipart.FileHeader, userID string, versionName string) error

	// Sync
	SyncCapexBudget(id string) error
	SyncCapexActual(id string) error

	// List Files
	ListCapexBudgetFiles() ([]FileCapexBudgetEntity, error)
	ListCapexActualFiles() ([]FileCapexActualEntity, error)

	// Delete
	DeleteCapexBudgetFile(id string) error
	DeleteCapexActualFile(id string) error

	// Rename
	RenameCapexBudgetFile(id string, newName string) error
	RenameCapexActualFile(id string, newName string) error

	// Dashboard
	GetCapexDashboardSummary(filter map[string]interface{}) (*DashboardSummaryDTO, error)
}

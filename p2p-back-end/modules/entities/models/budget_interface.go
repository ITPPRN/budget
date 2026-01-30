package models

import (
	"mime/multipart"
)

// BudgetRepository Interface สำหรับจัดการฐานข้อมูล
type BudgetRepository interface {
	// Transaction Helper
	WithTrx(trxHandle func(repo BudgetRepository) error) error

	// Create Files (3 Distinct Tables)
	CreateFileBudget(file *FileBudgetEntity) error
	CreateFileCapexBudget(file *FileCapexBudgetEntity) error
	CreateFileCapexActual(file *FileCapexActualEntity) error

	// Create Facts (Flattened Data)
	CreateBudgetFacts(facts []BudgetFactEntity) error
	CreateCapexBudgetFacts(facts []CapexBudgetFactEntity) error
	CreateCapexActualFacts(facts []CapexActualFactEntity) error

	// List Files (For Frontend Dropdown)
	ListFileBudgets() ([]FileBudgetEntity, error)
	ListFileCapexBudgets() ([]FileCapexBudgetEntity, error)
	ListFileCapexActuals() ([]FileCapexActualEntity, error)

	// Get Single File (For Sync)
	GetFileBudget(id string) (*FileBudgetEntity, error)
	GetFileCapexBudget(id string) (*FileCapexBudgetEntity, error)
	GetFileCapexActual(id string) (*FileCapexActualEntity, error)

	// Dashboard / Detail View
	GetBudgetFilterOptions() ([]BudgetFactEntity, error)
	GetBudgetDetails(groups []string, departments []string, entityGLs []string, consoGLs []string) ([]BudgetFactEntity, error)

	// Delete Files
	DeleteFileBudget(id string) error
	DeleteFileCapexBudget(id string) error
	DeleteFileCapexActual(id string) error

	// Delete All Facts (For Sync)
	DeleteAllBudgetFacts() error
	DeleteAllCapexBudgetFacts() error
	DeleteAllCapexActualFacts() error

	// Update Files (Rename)
	UpdateFileBudget(id string, filename string) error
	UpdateFileCapexBudget(id string, filename string) error
	UpdateFileCapexActual(id string, filename string) error
}

// BudgetService Interface สำหรับ Business Logic
type BudgetService interface {
	// Import (Upload Only)
	ImportBudget(fileHeader *multipart.FileHeader, userID string, versionName string) error
	ImportCapexBudget(fileHeader *multipart.FileHeader, userID string, versionName string) error
	ImportCapexActual(fileHeader *multipart.FileHeader, userID string, versionName string) error

	// Sync (Process Logic)
	SyncBudget(fileID string) error
	SyncCapexBudget(fileID string) error
	SyncCapexActual(fileID string) error

	// Dashboard Service
	GetFilterOptions() ([]FilterOptionDTO, error)
	GetBudgetDetails(groups []string, departments []string, entityGLs []string, consoGLs []string) ([]BudgetFactEntity, error)

	// List Files (For UI)
	ListBudgetFiles() ([]FileBudgetEntity, error)
	ListCapexBudgetFiles() ([]FileCapexBudgetEntity, error)
	ListCapexActualFiles() ([]FileCapexActualEntity, error)

	// Delete
	DeleteBudgetFile(id string) error
	DeleteCapexBudgetFile(id string) error
	DeleteCapexActualFile(id string) error

	// Rename
	RenameBudgetFile(id string, newName string) error
	RenameCapexBudgetFile(id string, newName string) error
	RenameCapexActualFile(id string, newName string) error
}

// DTOs
type FilterOptionDTO struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Level    int               `json:"level"`
	Children []FilterOptionDTO `json:"children,omitempty"`
}

package models

import (
	"mime/multipart"

	"github.com/shopspring/decimal"
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
	CreateActualFacts(facts []ActualFactEntity) error // New

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
	GetOrganizationStructure() ([]BudgetFactEntity, error)
	GetBudgetDetails(filter map[string]interface{}) ([]BudgetFactEntity, error)
	GetActualDetails(filter map[string]interface{}) ([]ActualFactEntity, error)          // New
	GetActualTransactions(filter map[string]interface{}) ([]ActualTransactionDTO, error) // New Transaction View

	// Aggregation
	GetDashboardAggregates(filter map[string]interface{}) (*DashboardSummaryDTO, error)

	// Delete Files
	DeleteFileBudget(id string) error
	DeleteFileCapexBudget(id string) error
	DeleteFileCapexActual(id string) error

	// Delete All Facts (For Sync)
	DeleteAllBudgetFacts() error
	DeleteAllCapexBudgetFacts() error
	DeleteAllCapexActualFacts() error
	DeleteAllActualFacts() error               // Keep for full reset
	DeleteActualFactsByYear(year string) error // New

	// Sync From DB (External Tables)
	GetAllAchHmwGle() ([]AchHmwGleEntity, error)
	GetAggregatedHMW(year string) ([]ActualAggregatedDTO, error)
	GetAllClikGle() ([]ClikGleEntity, error)
	GetAggregatedCLIK(year string) ([]ActualAggregatedDTO, error)
	GetRawDate() (string, error) // Debugging

	// Update Files (Rename)
	UpdateFileBudget(id string, filename string) error
	UpdateFileCapexBudget(id string, filename string) error
	UpdateFileCapexActual(id string, filename string) error
}

// BudgetService Interface สำหรับ Business Logic
type BudgetService interface {
	// Import (Upload Only)
	ImportBudget(file *multipart.FileHeader, userID string, versionName string) error
	ImportCapexBudget(file *multipart.FileHeader, userID string, versionName string) error
	ImportCapexActual(file *multipart.FileHeader, userID string, versionName string) error

	// Sync (Process Logic)
	SyncBudget(id string) error
	SyncCapexBudget(id string) error
	SyncCapexActual(id string) error

	// Dashboard Service
	GetFilterOptions() ([]FilterOptionDTO, error)
	GetOrganizationStructure() ([]OrganizationDTO, error)
	GetBudgetDetails(filter map[string]interface{}) ([]BudgetFactEntity, error)
	GetActualDetails(filter map[string]interface{}) ([]ActualFactEntity, error) // New

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

	// Actuals
	SyncActuals(year string) error

	// Dashboard Optimized
	GetDashboardSummary(filter map[string]interface{}) (*DashboardSummaryDTO, error)
	GetActualTransactions(filter map[string]interface{}) ([]ActualTransactionDTO, error) // New
	GetRawDate() (string, error)                                                         // Debugging
}

// DTOs
type FilterOptionDTO struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Level    int               `json:"level"`
	Children []FilterOptionDTO `json:"children,omitempty"`
}

type OrganizationDTO struct {
	Entity   string   `json:"entity"`
	Branches []string `json:"branches"`
}

// Dashboard Aggregation DTOs
type DashboardSummaryDTO struct {
	TotalBudget    float64             `json:"total_budget"`
	TotalActual    float64             `json:"total_actual"`
	DepartmentData []DepartmentStatDTO `json:"department_data"`
	ChartData      []MonthlyStatDTO    `json:"chart_data"`

	// Pagination
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
}

type DepartmentStatDTO struct {
	Department string  `json:"department"`
	Budget     float64 `json:"budget"`
	Actual     float64 `json:"actual"`
}

type MonthlyStatDTO struct {
	Month  string  `json:"month"`
	Budget float64 `json:"budget"`
	Actual float64 `json:"actual"`
}

type ActualAggregatedDTO struct {
	Company       string          `json:"company"`
	Branch        string          `json:"branch"`
	Department    string          `json:"department"`
	GLAccountNo   string          `json:"gl_account_no"`
	GLAccountName string          `json:"gl_account_name"`
	Month         string          `json:"month"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
}

type ActualTransactionDTO struct {
	Source        string          `json:"source" gorm:"column:source"`
	PostingDate   string          `json:"posting_date" gorm:"column:posting_date"`
	DocNo         string          `json:"document_no" gorm:"column:doc_no"`
	Vendor        string          `json:"vendor" gorm:"column:vendor"`
	Description   string          `json:"description" gorm:"column:description"`
	GLAccountNo   string          `json:"gl_account_no" gorm:"column:gl_account_no"`
	GLAccountName string          `json:"gl_account_name" gorm:"column:gl_account_name"`
	Amount        decimal.Decimal `json:"amount" gorm:"column:amount"`
	Department    string          `json:"department" gorm:"column:department"`
}

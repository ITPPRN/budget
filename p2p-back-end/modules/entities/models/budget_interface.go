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
}

// BudgetService Interface สำหรับ Business Logic
type BudgetService interface {
	ImportBudget(fileHeader *multipart.FileHeader, userID string) error
	ImportCapexBudget(fileHeader *multipart.FileHeader, userID string) error
	ImportCapexActual(fileHeader *multipart.FileHeader, userID string) error
}

package models

// OwnerService Interface สำหรับ Business Logic ของฝั่ง Owner/Staff
type OwnerService interface {
	// Dashboard
	GetDashboardSummary(user *UserInfo, filter map[string]interface{}) (*OwnerDashboardSummaryDTO, error)
	GetActualTransactions(user *UserInfo, filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetActualDetails(user *UserInfo, filter map[string]interface{}) ([]OwnerActualFactEntity, error)
	GetBudgetDetails(user *UserInfo, filter map[string]interface{}) ([]BudgetFactEntity, error)

	// Filter Options (Scoped to Owner)
	GetFilterOptions(user *UserInfo) ([]FilterOptionDTO, error)
	GetOrganizationStructure(user *UserInfo) ([]OrganizationDTO, error)
	GetOwnerFilterLists(user *UserInfo) (*OwnerFilterListsDTO, error)
	AutoSyncOwnerActuals() error // New Auto Sync
}

// OwnerDashboardSummaryDTO extends the base DTO with specific Owner fields
type OwnerDashboardSummaryDTO struct {
	DashboardSummaryDTO                 // Embed Base DTO
	TopExpenses         []TopExpenseDTO `json:"top_expenses"`
	CapexBudget         float64         `json:"capex_budget"`
	CapexActual         float64         `json:"capex_actual"`
}

type OwnerFilterListsDTO struct {
	Companies []string `json:"companies"`
	Branches  []string `json:"branches"`
	Years     []string `json:"years"`
}

type TopExpenseDTO struct {
	Name   string  `json:"name"`   // GL Name or Category
	Amount float64 `json:"amount"` // Actual Spending
}

// OwnerRepository Interface สำหรับจัดการฐานข้อมูลฝั่ง Owner
type OwnerRepository interface {
	GetUserContext(userID string) (*UserEntity, error) // Restored
	CreateUser(user *UserEntity) error                 // New: For Lazy Sync
	GetDashboardAggregates(filter map[string]interface{}) (*OwnerDashboardSummaryDTO, error)
	GetBudgetDetails(filter map[string]interface{}) ([]BudgetFactEntity, error)
	GetActualDetails(filter map[string]interface{}) ([]OwnerActualFactEntity, error)
	GetActualTransactions(filter map[string]interface{}) (*PaginatedActualTransactionDTO, error)
	GetBudgetFilterOptions(filter map[string]interface{}) ([]BudgetFactEntity, error)
	GetOwnerFilterLists(filter map[string]interface{}) (*OwnerFilterListsDTO, error)
	AutoSyncOwnerActuals() error // New Auto Sync
	GetUserPermissions(userID string) ([]UserPermissionEntity, error)
	GetNavCodesByMasterDepts(masterCodes []string) ([]string, error)
}

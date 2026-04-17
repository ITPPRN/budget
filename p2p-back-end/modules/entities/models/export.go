package models

import "github.com/shopspring/decimal"

type BudgetExportDTO struct {
	Entity     string          `gorm:"column:entity"`
	Branch     string          `gorm:"column:branch"`
	Department string          `gorm:"column:department"`
	Group      string          `gorm:"column:group"`
	Group2     string          `gorm:"column:group2"`
	Group3     string          `gorm:"column:group3"`
	ConsoGL    string          `gorm:"column:conso_gl"`
	GLName     string          `gorm:"column:gl_name"`
	Month      string          `gorm:"column:month"`
	Amount     decimal.Decimal `gorm:"column:amount"`

	// Flattened fields for summary
	MonthsAmounts map[string]interface{} `gorm:"-"`
	YearTotal     decimal.Decimal        `gorm:"-"`
}

type ActualExportDTO struct {
	Entity      string          `gorm:"column:entity"`
	Branch      string          `gorm:"column:branch"`
	Department  string          `gorm:"column:department"`
	Group       string          `gorm:"column:group"`
	Group2      string          `gorm:"column:group2"`
	Group3      string          `gorm:"column:group3"`
	ConsoGL     string          `gorm:"column:conso_gl"`
	GLName      string          `gorm:"column:gl_name"`
	DocumentNo  string          `gorm:"column:doc_no"`
	Amount      decimal.Decimal `gorm:"column:amount"`
	VendorName  string          `gorm:"column:vendor_name"`
	Description string          `gorm:"column:description"`
	PostingDate string          `gorm:"column:posting_date"`
	Status      string          `gorm:"column:status"`
}

type DeptBudgetStatusDTO struct {
	Status     string          `json:"status"`
	Department string          `gorm:"column:department"`
	Budget     decimal.Decimal `gorm:"column:budget"`
	Spend      decimal.Decimal `gorm:"column:spend"`
	Remaining  decimal.Decimal `gorm:"column:remaining"`
	Percentage float64         `gorm:"column:percentage"`
}

type BudgetVsActualExportDTO struct {
	Entity        string                 `gorm:"column:entity"`
	Branch        string                 `gorm:"column:branch"`
	Department    string                 `gorm:"column:department"`
	Type          string                 `json:"type"` // "Budget" or "Actual"
	Group         string                 `gorm:"column:group"`
	Group2        string                 `gorm:"column:group2"`
	Group3        string                 `gorm:"column:group3"`
	ConsoGL       string                 `gorm:"column:conso_gl"`
	GLName        string                 `gorm:"column:gl_name"`
	MonthsAmounts map[string]interface{} `gorm:"-"`
	YearTotal     decimal.Decimal        `gorm:"-"`
}

type CapexDeptStatusDTO struct {
	Status      string          `json:"status"`
	Department  string          `gorm:"column:department"`
	CapexBudget decimal.Decimal `gorm:"column:capex_bg"`
	Spend       decimal.Decimal `gorm:"column:spend"`
	Remaining   decimal.Decimal `gorm:"column:remaining"`
	Percentage  float64         `gorm:"column:percentage"`
}

type CapexVsActualExportDTO struct {
	Entity        string                 `gorm:"column:entity"`
	Branch        string                 `gorm:"column:branch"`
	Department    string                 `gorm:"column:department"`
	CapexNo       string                 `gorm:"column:capex_no"`
	CapexName     string                 `gorm:"column:capex_name"`
	CapexCategory string                 `gorm:"column:capex_category"`
	Type          string                 `json:"type"` // "Budget" or "Actual"
	MonthsAmounts map[string]interface{} `gorm:"-"`
	YearTotal     decimal.Decimal        `gorm:"-"`
}

type OwnerBudgetVsActualDTO struct {
	Month    string          `json:"month"`
	Budget   decimal.Decimal `json:"budget"`
	Actual   decimal.Decimal `json:"actual"`
	Variance decimal.Decimal `json:"variance"`
	Ratio    float64         `json:"ratio"`
}

type TopExpenseExportDTO struct {
	ConsoGL string          `gorm:"column:conso_gl"`
	GLName  string          `gorm:"column:gl_name"`
	Amount  decimal.Decimal `gorm:"column:amount"`
}

type OwnerCapexBudgetExportDTO struct {
	Entity        string          `gorm:"column:entity"`
	Branch        string          `gorm:"column:branch"`
	Department    string          `gorm:"column:department"`
	CapexNo       string          `gorm:"column:capex_no"`
	CapexName     string          `gorm:"column:capex_name"`
	CapexCategory string          `gorm:"column:capex_category"`
	Budget        decimal.Decimal `gorm:"column:budget"`
	Actual        decimal.Decimal `gorm:"column:actual"`
	Remaining     decimal.Decimal `gorm:"column:remaining"`
	Percentage    float64         `gorm:"column:percentage"`
}

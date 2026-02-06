package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Source Table 1: ACH & HMW
type AchHmwGleEntity struct {
	ID                   int             `gorm:"primaryKey;column:id"` // integer id from sample
	EntryNo              int             `gorm:"column:Entry_No"`
	PostingDate          time.Time       `gorm:"column:Posting_Date"`
	DocumentNo           string          `gorm:"column:Document_No"`
	GLAccountNo          string          `gorm:"column:G_L_Account_No"`
	GLAccountName        string          `gorm:"column:G_L_Account_Name"`
	Description          string          `gorm:"column:Description"`
	GlobalDimension1Code string          `gorm:"column:Global_Dimension_1_Code"` // Department?
	GenBusPostingGroup   string          `gorm:"column:Gen_Bus_Posting_Group"`
	GenProdPostingGroup  string          `gorm:"column:Gen_Prod_Posting_Group"`
	Amount               decimal.Decimal `gorm:"column:Amount;type:decimal(18,2)"`
	Company              string          `gorm:"column:company"` // HMW, ACG
	Branch               string          `gorm:"column:branch"`  // HQ, AVN...
}

func (AchHmwGleEntity) TableName() string { return "achhmw_gle_api" }

// Source Table 2: CLIK
type ClikGleEntity struct {
	ID                   int             `gorm:"primaryKey;column:id"`
	EntryNo              int             `gorm:"column:Entry_No"`
	PostingDate          time.Time       `gorm:"column:Posting_Date"`
	DocumentNo           string          `gorm:"column:Document_No"`
	GLAccountNo          string          `gorm:"column:G_L_Account_No"`
	GLAccountName        string          `gorm:"column:G_L_Account_Name"`
	Description          string          `gorm:"column:Description"`
	GlobalDimension1Code string          `gorm:"column:Global_Dimension_1_Code"` // Department?
	GlobalDimension2Code string          `gorm:"column:Global_Dimension_2_Code"` // Branch
	Amount               decimal.Decimal `gorm:"column:Amount;type:decimal(18,2)"`
}

func (ClikGleEntity) TableName() string { return "general_ledger_entries_clik" }

package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Source Table 1: ACH & HMW
// Source Table 1: ACH & HMW
type AchHmwGleEntity struct {
	ID                        int             `gorm:"primaryKey;column:id"`
	EntryNo                   int             `gorm:"column:Entry_No"`
	SystemCreatedEntry        bool            `gorm:"column:System_Created_Entry"`
	PostingDate               time.Time       `gorm:"column:Posting_Date;type:date"`
	DocumentType              string          `gorm:"column:Document_Type"`
	DocumentNo                string          `gorm:"column:Document_No"`
	GLAccountNo               string          `gorm:"column:G_L_Account_No"`
	GLAccountName             string          `gorm:"column:G_L_Account_Name"`
	Description               string          `gorm:"column:Description"`
	Description2              string          `gorm:"column:Description_2"`
	JobNo                     string          `gorm:"column:Job_No"`
	GlobalDimension1Code      string          `gorm:"column:Global_Dimension_1_Code"`
	GlobalDimension2Code      string          `gorm:"column:Global_Dimension_2_Code"`
	VendorBankAccount         string          `gorm:"column:Vendor_Bank_Account"`
	ICPartnerCode             string          `gorm:"column:IC_Partner_Code"`
	GenPostingType            string          `gorm:"column:Gen_Posting_Type"`
	GenBusPostingGroup        string          `gorm:"column:Gen_Bus_Posting_Group"`
	GenProdPostingGroup       string          `gorm:"column:Gen_Prod_Posting_Group"`
	Quantity                  decimal.Decimal `gorm:"column:Quantity;type:decimal(18,4)"`
	Amount                    decimal.Decimal `gorm:"column:Amount;type:decimal(18,4)"`
	DebitAmount               decimal.Decimal `gorm:"column:Debit_Amount;type:decimal(18,4)"`
	CreditAmount              decimal.Decimal `gorm:"column:Credit_Amount;type:decimal(18,4)"`
	AdditionalCurrencyAmount  decimal.Decimal `gorm:"column:Additional_Currency_Amount;type:decimal(18,4)"`
	VATAmount                 decimal.Decimal `gorm:"column:VAT_Amount;type:decimal(18,4)"`
	BalAccountType            string          `gorm:"column:Bal_Account_Type"`
	BalAccountNo              string          `gorm:"column:Bal_Account_No"`
	UserID                    string          `gorm:"column:User_ID"`
	SourceCode                string          `gorm:"column:Source_Code"`
	ReasonCode                string          `gorm:"column:Reason_Code"`
	Reversed                  bool            `gorm:"column:Reversed"`
	ReversedByEntryNo         int             `gorm:"column:Reversed_by_Entry_No"`
	CarID                     string          `gorm:"column:Car_ID"`
	ReversedEntryNo           int             `gorm:"column:Reversed_Entry_No"`
	FAEntryType               string          `gorm:"column:FA_Entry_Type"`
	FAEntryNo                 string          `gorm:"column:FA_Entry_No"`
	DimensionSetID            int             `gorm:"column:Dimension_Set_ID"`
	CustomerNo                string          `gorm:"column:Cutomer_No"`
	CustomerName              string          `gorm:"column:Cutomer_Name"`
	VendorName                string          `gorm:"column:Vendor_Name"`
	SerialNo                  string          `gorm:"column:Serial_No"`
	SerialDescription         string          `gorm:"column:Serial_Description"`
	Company                   string          `gorm:"column:company"`
	Branch                    string          `gorm:"column:branch"`
	MongoID                   string          `gorm:"column:_id"`
	MongoV                    int             `gorm:"column:__v"`
}

func (AchHmwGleEntity) TableName() string { return "achhmw_gle_api" }

// Source Table 2: CLIK
type ClikGleEntity struct {
	ID                        int             `gorm:"primaryKey;column:id"`
	EntryNo                   int             `gorm:"column:Entry_No"`
	SystemCreatedEntry        bool            `gorm:"column:System_Created_Entry"`
	PostingDate               time.Time       `gorm:"column:Posting_Date;type:date"`
	DocumentType              string          `gorm:"column:Document_Type"`
	DocumentNo                string          `gorm:"column:Document_No"`
	GLAccountNo               string          `gorm:"column:G_L_Account_No"`
	GLAccountName             string          `gorm:"column:G_L_Account_Name"`
	Description               string          `gorm:"column:Description"`
	Description2              string          `gorm:"column:Description_2"`
	JobNo                     string          `gorm:"column:Job_No"`
	GlobalDimension1Code      string          `gorm:"column:Global_Dimension_1_Code"`
	GlobalDimension2Code      string          `gorm:"column:Global_Dimension_2_Code"`
	VendorBankAccount         string          `gorm:"column:Vendor_Bank_Account"`
	ICPartnerCode             string          `gorm:"column:IC_Partner_Code"`
	GenPostingType            string          `gorm:"column:Gen_Posting_Type"`
	GenBusPostingGroup        string          `gorm:"column:Gen_Bus_Posting_Group"`
	GenProdPostingGroup       string          `gorm:"column:Gen_Prod_Posting_Group"`
	Quantity                  decimal.Decimal `gorm:"column:Quantity;type:decimal(18,4)"`
	Amount                    decimal.Decimal `gorm:"column:Amount;type:decimal(18,4)"`
	DebitAmount               decimal.Decimal `gorm:"column:Debit_Amount;type:decimal(18,4)"`
	CreditAmount              decimal.Decimal `gorm:"column:Credit_Amount;type:decimal(18,4)"`
	AdditionalCurrencyAmount  decimal.Decimal `gorm:"column:Additional_Currency_Amount;type:decimal(18,4)"`
	VATAmount                 decimal.Decimal `gorm:"column:VAT_Amount;type:decimal(18,4)"`
	BalAccountType            string          `gorm:"column:Bal_Account_Type"`
	BalAccountNo              string          `gorm:"column:Bal_Account_No"`
	UserID                    string          `gorm:"column:User_ID"`
	SourceCode                string          `gorm:"column:Source_Code"`
	ReasonCode                string          `gorm:"column:Reason_Code"`
	Reversed                  bool            `gorm:"column:Reversed"`
	ReversedByEntryNo         int             `gorm:"column:Reversed_by_Entry_No"`
	CarID                     string          `gorm:"column:Car_ID"`
	ReversedEntryNo           int             `gorm:"column:Reversed_Entry_No"`
	FAEntryType               string          `gorm:"column:FA_Entry_Type"`
	FAEntryNo                 string          `gorm:"column:FA_Entry_No"`
	DimensionSetID            int             `gorm:"column:Dimension_Set_ID"`
	CustomerName              string          `gorm:"column:Cutomer_Name"`
	VendorName                string          `gorm:"column:Vendor_Name"`
	SerialNo                  string          `gorm:"column:Serial_No"`
	CustomerNo                string          `gorm:"column:Cutomer_No"`
	SerialDescription         string          `gorm:"column:Serial_Description"`
}

func (ClikGleEntity) TableName() string { return "general_ledger_entries_clik" }

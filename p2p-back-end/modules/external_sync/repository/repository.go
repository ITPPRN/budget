package repository

import (
	"context"
	"fmt"
	"p2p-back-end/modules/entities/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type externalSyncRepository struct {
	localDb *gorm.DB
	dwDb    *gorm.DB
}

func NewExternalSyncRepository(localDb *gorm.DB, dwDb *gorm.DB) models.ExternalSyncRepository {
	// Robustness: Ensure local tables have correct schema/PK.
	// We previously dropped these once to clear legacy ghost columns.
	localDb.AutoMigrate(&models.AchHmwGleEntity{}, &models.ClikGleEntity{})

	return &externalSyncRepository{
		localDb: localDb,
		dwDb:    dwDb,
	}
}

func (r *externalSyncRepository) FetchHMWInBatches(ctx context.Context, year int, month int, batchSize int, handle func([]models.AchHmwGleEntity) error) error {
	var results []models.AchHmwGleEntity
	// We use CAST(Posting_Date AS DATE) in SELECT to make it scannable to time.Time
	err := r.dwDb.WithContext(ctx).
		Table("achhmw_gle_api").
		Select(`
			id, "Entry_No", "System_Created_Entry", CAST("Posting_Date" AS DATE) AS "Posting_Date", 
			"Document_Type", "Document_No", "G_L_Account_No", "G_L_Account_Name", "Description", 
			"Description_2", "Job_No", "Global_Dimension_1_Code", "Global_Dimension_2_Code", 
			"Vendor_Bank_Account", "IC_Partner_Code", "Gen_Posting_Type", "Gen_Bus_Posting_Group", 
			"Gen_Prod_Posting_Group", "Quantity", "Amount", "Debit_Amount", "Credit_Amount", 
			"Additional_Currency_Amount", "VAT_Amount", "Bal_Account_Type", "Bal_Account_No", 
			"User_ID", "Source_Code", "Reason_Code", "Reversed", "Reversed_by_Entry_No", 
			"Car_ID", "Reversed_Entry_No", "FA_Entry_Type", "FA_Entry_No", "Dimension_Set_ID", 
			"Cutomer_No", "Cutomer_Name", "Vendor_Name", "Serial_No", "Serial_Description", 
			company, branch, _id, __v
		`).
		Where("EXTRACT(YEAR FROM CAST(\"Posting_Date\" AS DATE)) = ? AND EXTRACT(MONTH FROM CAST(\"Posting_Date\" AS DATE)) = ?", year, month).
		Order("\"achhmw_gle_api\".\"id\"").
		FindInBatches(&results, 2000, func(tx *gorm.DB, batchCount int) error {
			return handle(results)
		}).Error

	if err != nil {
		return fmt.Errorf("extSyncRepo.FetchHMWInBatches: %w", err)
	}
	return nil
}

func (r *externalSyncRepository) FetchCLIKInBatches(ctx context.Context, year int, month int, batchSize int, handle func([]models.ClikGleEntity) error) error {
	var results []models.ClikGleEntity
	err := r.dwDb.WithContext(ctx).
		Table("general_ledger_entries_clik").
		Select(`
			id, "Entry_No", "System_Created_Entry", CAST("Posting_Date" AS DATE) AS "Posting_Date", 
			"Document_Type", "Document_No", "G_L_Account_No", "G_L_Account_Name", "Description", 
			"Description_2", "Job_No", "Global_Dimension_1_Code", "Global_Dimension_2_Code", 
			"Vendor_Bank_Account", "IC_Partner_Code", "Gen_Posting_Type", "Gen_Bus_Posting_Group", 
			"Gen_Prod_Posting_Group", "Quantity", "Amount", "Debit_Amount", "Credit_Amount", 
			"Additional_Currency_Amount", "VAT_Amount", "Bal_Account_Type", "Bal_Account_No", 
			"User_ID", "Source_Code", "Reason_Code", "Reversed", "Reversed_by_Entry_No", 
			"Car_ID", "Reversed_Entry_No", "FA_Entry_Type", "FA_Entry_No", "Dimension_Set_ID", 
			"Cutomer_No", "Cutomer_Name", "Vendor_Name", "Serial_No", "Serial_Description"
		`).
		Where("EXTRACT(YEAR FROM CAST(\"Posting_Date\" AS DATE)) = ? AND EXTRACT(MONTH FROM CAST(\"Posting_Date\" AS DATE)) = ?", year, month).
		Order("\"general_ledger_entries_clik\".\"id\"").
		FindInBatches(&results, 2000, func(tx *gorm.DB, batchCount int) error {
			return handle(results)
		}).Error

	if err != nil {
		return fmt.Errorf("extSyncRepo.FetchCLIKInBatches: %w", err)
	}
	return nil
}

func (r *externalSyncRepository) UpsertHMWLocal(ctx context.Context, data []models.AchHmwGleEntity) error {
	if len(data) == 0 {
		return nil
	}
	err := r.localDb.WithContext(ctx).
		Table("achhmw_gle_api").
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		CreateInBatches(data, 500).Error
	if err != nil {
		return fmt.Errorf("extSyncRepo.UpsertHMWLocal: %w", err)
	}
	return nil
}

func (r *externalSyncRepository) UpsertCLIKLocal(ctx context.Context, data []models.ClikGleEntity) error {
	if len(data) == 0 {
		return nil
	}
	err := r.localDb.WithContext(ctx).
		Table("general_ledger_entries_clik").
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		CreateInBatches(data, 500).Error
	if err != nil {
		return fmt.Errorf("extSyncRepo.UpsertCLIKLocal: %w", err)
	}
	return nil
}

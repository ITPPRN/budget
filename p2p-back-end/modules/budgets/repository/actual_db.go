package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

type actualRepository struct {
	db *gorm.DB
}

func NewActualRepository(db *gorm.DB) models.ActualRepository {
	return &actualRepository{db: db}
}

func (r *actualRepository) WithTrx(trxHandle func(repo models.ActualRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewActualRepository(tx)
		return trxHandle(repo)
	})
}

func (r *actualRepository) CreateActualFacts(ctx context.Context, headers []models.ActualFactEntity) error {
	// 1. บันทึกส่วนหัว (Insert Headers)
	if err := r.db.WithContext(ctx).Omit("ActualAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return fmt.Errorf("actualRepo.CreateActualFacts.Headers: %w", err)
	}
	// 2. รวบรวมยอดเงิน (Collect Amounts)
	var allAmounts []models.ActualAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.ActualAmounts...)
	}
	// 3. บันทึกยอดเงิน (Insert Amounts)
	if len(allAmounts) > 0 {
		if err := r.db.WithContext(ctx).CreateInBatches(&allAmounts, 1000).Error; err != nil {
			return fmt.Errorf("actualRepo.CreateActualFacts.Amounts: %w", err)
		}
	}
	return nil
}

func (r *actualRepository) DeleteAllActualFacts(ctx context.Context) error {
	// 1. ลบยอดเงิน (Delete Amounts)
	if err := r.db.WithContext(ctx).Exec("DELETE FROM actual_amount_entities").Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteAllActualFacts.Amounts: %w", err)
	}
	// 2. ลบส่วนหัว (Delete Headers)
	if err := r.db.WithContext(ctx).Exec("DELETE FROM actual_fact_entities").Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteAllActualFacts.Headers: %w", err)
	}
	return nil
}

func (r *actualRepository) DeleteActualFactsByYear(ctx context.Context, year string) error {
	if err := r.db.WithContext(ctx).Exec(`
		DELETE FROM actual_amount_entities 
		WHERE actual_fact_id IN (SELECT id FROM actual_fact_entities WHERE year = ?)
	`, year).Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteActualFactsByYear.Amounts: %w", err)
	}

	if err := r.db.WithContext(ctx).Unscoped().Where("year = ?", year).Delete(&models.ActualFactEntity{}).Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteActualFactsByYear.Headers: %w", err)
	}
	return nil
}

func (r *actualRepository) DeleteActualFactsByMonth(ctx context.Context, year string, month string) error {
	if err := r.db.WithContext(ctx).Exec(`
		DELETE FROM actual_amount_entities 
		WHERE month = ? AND actual_fact_id IN (SELECT id FROM actual_fact_entities WHERE year = ?)
	`, month, year).Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteActualFactsByMonth.Amounts: %w", err)
	}

	if err := r.db.WithContext(ctx).Exec(`
		UPDATE actual_fact_entities f
		SET year_total = COALESCE((SELECT SUM(amount) FROM actual_amount_entities a WHERE a.actual_fact_id = f.id), 0)
		WHERE f.year = ?
	`, year).Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteActualFactsByMonth.UpdateTotal: %w", err)
	}

	if err := r.db.WithContext(ctx).Unscoped().
		Where("year = ? AND year_total = 0", year).
		Delete(&models.ActualFactEntity{}).Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteActualFactsByMonth.Cleanup: %w", err)
	}
	return nil
}

func (r *actualRepository) DeleteAllActualTransactions(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Unscoped().Where("1=1").Delete(&models.ActualTransactionEntity{}).Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteAllActualTransactions: %w", err)
	}
	return nil
}

func (r *actualRepository) DeleteActualTransactionsByYear(ctx context.Context, year string) error {
	// 🛡️ CRITICAL: Only delete PENDING transactions to preserve Owner work (Drafts, Reported, Complete)
	if err := r.db.WithContext(ctx).Unscoped().
		Where("year = ? AND status = ?", year, models.TxStatusPending).
		Delete(&models.ActualTransactionEntity{}).Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteActualTransactionsByYear: %w", err)
	}
	return nil
}

func (r *actualRepository) DeleteActualTransactionsByMonth(ctx context.Context, year string, month string) error {
	monthMap := map[string]string{
		"JAN": "01", "FEB": "02", "MAR": "03", "APR": "04", "MAY": "05", "JUN": "06",
		"JUL": "07", "AUG": "08", "SEP": "09", "OCT": "10", "NOV": "11", "DEC": "12",
	}
	mCode, ok := monthMap[month]
	if !ok {
		return fmt.Errorf("actualRepo.DeleteActualTransactionsByMonth: invalid month: %s", month)
	}
	pattern := fmt.Sprintf("%s-%s-%%", year, mCode)
	// 🛡️ CRITICAL: Only delete PENDING transactions to preserve Owner work
	if err := r.db.WithContext(ctx).Unscoped().
		Where("year = ? AND posting_date LIKE ? AND status = ?", year, pattern, models.TxStatusPending).
		Delete(&models.ActualTransactionEntity{}).Error; err != nil {
		return fmt.Errorf("actualRepo.DeleteActualTransactionsByMonth: %w", err)
	}
	return nil
}

func (r *actualRepository) GetAllAchHmwGle(ctx context.Context) ([]models.AchHmwGleEntity, error) {
	var results []models.AchHmwGleEntity
	if err := r.db.WithContext(ctx).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("actualRepo.GetAllAchHmwGle: %w", err)
	}
	return results, nil
}

func (r *actualRepository) GetAggregatedHMW(ctx context.Context, year string, months []string) ([]models.ActualAggregatedDTO, error) {
	var results []models.ActualAggregatedDTO
	query := r.db.WithContext(ctx).Table("achhmw_gle_api").
		Select(`
			company, 
			branch, 
			"Global_Dimension_1_Code" as department, 
			"G_L_Account_No" as gl_account_no, 
			"G_L_Account_Name" as gl_account_name, 
			"Vendor_Name" as vendor_name,
			UPPER(TO_CHAR("Posting_Date"::DATE, 'MON')) as month, 
			SUM("Amount") as total_amount
		`).
		Where("TO_CHAR(\"Posting_Date\"::DATE, 'YYYY') = ?", year)

	if len(months) > 0 {
		query = query.Where("UPPER(TO_CHAR(\"Posting_Date\"::DATE, 'MON')) IN ?", months)
	}

	if err := query.Group(`company, branch, "Global_Dimension_1_Code", "G_L_Account_No", "G_L_Account_Name", "Vendor_Name", UPPER(TO_CHAR("Posting_Date"::DATE, 'MON'))`).
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("actualRepo.GetAggregatedHMW: %w", err)
	}
	return results, nil
}

func (r *actualRepository) GetRawTransactionsHMW(ctx context.Context, year string, months []string) ([]models.ActualTransactionDTO, error) {
	var results []models.ActualTransactionDTO
	query := r.db.WithContext(ctx).Table("achhmw_gle_api").
		Select(`
			'HMW' as source,
			TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date,
			"Document_No" as doc_no, 
			"Description" as description, 
			"G_L_Account_No" as entity_gl,
			"G_L_Account_Name" as gl_account_name,
			"Global_Dimension_1_Code" as department,
			"Amount" as amount,
			"Vendor_Name" as vendor,
			company,
			branch
		`).
		Where("TO_CHAR(\"Posting_Date\"::DATE, 'YYYY') = ?", year)

	if len(months) > 0 {
		query = query.Where("UPPER(TO_CHAR(\"Posting_Date\"::DATE, 'MON')) IN ?", months)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("actualRepo.GetRawTransactionsHMW: %w", err)
	}
	return results, nil
}

func (r *actualRepository) GetAllClikGle(ctx context.Context) ([]models.ClikGleEntity, error) {
	var results []models.ClikGleEntity
	if err := r.db.WithContext(ctx).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("actualRepo.GetAllClikGle: %w", err)
	}
	return results, nil
}

func (r *actualRepository) GetAggregatedCLIK(ctx context.Context, year string, months []string) ([]models.ActualAggregatedDTO, error) {
	var results []models.ActualAggregatedDTO
	query := r.db.WithContext(ctx).Table("general_ledger_entries_clik").
		Select(`
			company,
			"Global_Dimension_2_Code" as branch, 
			"Global_Dimension_1_Code" as department, 
			"G_L_Account_No" as gl_account_no, 
			"G_L_Account_Name" as gl_account_name, 
			"Vendor_Name" as vendor_name,
			UPPER(TO_CHAR("Posting_Date"::DATE, 'MON')) as month, 
			SUM("Amount") as total_amount
		`).
		Where("TO_CHAR(\"Posting_Date\"::DATE, 'YYYY') = ?", year)

	if len(months) > 0 {
		query = query.Where("UPPER(TO_CHAR(\"Posting_Date\"::DATE, 'MON')) IN ?", months)
	}

	if err := query.Group(`"Global_Dimension_2_Code", "Global_Dimension_1_Code", "G_L_Account_No", "G_L_Account_Name", "Vendor_Name", UPPER(TO_CHAR("Posting_Date"::DATE, 'MON'))`).
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("actualRepo.GetAggregatedCLIK: %w", err)
	}
	return results, nil
}

func (r *actualRepository) GetRawTransactionsCLIK(ctx context.Context, year string, months []string) ([]models.ActualTransactionDTO, error) {
	var results []models.ActualTransactionDTO
	query := r.db.WithContext(ctx).Table("general_ledger_entries_clik").
		Select(`
			'CLIK' as source,
			TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date,
			"Document_No" as doc_no, 
			"Description" as description,
			"G_L_Account_No" as entity_gl,
			"G_L_Account_Name" as gl_account_name,
			"Global_Dimension_1_Code" as department,
			"Amount" as amount,
			"Vendor_Name" as vendor,
			company,
			"Global_Dimension_2_Code" as branch
		`).
		Where("TO_CHAR(\"Posting_Date\"::DATE, 'YYYY') = ?", year)

	if len(months) > 0 {
		query = query.Where("UPPER(TO_CHAR(\"Posting_Date\"::DATE, 'MON')) IN ?", months)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("actualRepo.GetRawTransactionsCLIK: %w", err)
	}
	return results, nil
}

func (r *actualRepository) CreateActualTransactions(ctx context.Context, txs []models.ActualTransactionEntity) error {
	if len(txs) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).CreateInBatches(txs, 500).Error; err != nil {
		return fmt.Errorf("actualRepo.CreateActualTransactions: %w", err)
	}
	return nil
}

func (r *actualRepository) GetRawDate(ctx context.Context) (string, error) {
	var rawDate string
	// Try HMW first
	if err := r.db.WithContext(ctx).Table("achhmw_gle_api").Select("\"Posting_Date\"").Limit(1).Scan(&rawDate).Error; err == nil && rawDate != "" {
		return fmt.Sprintf("HMW Date: %s", rawDate), nil
	}

	// Try CLIK if HMW empty
	if err := r.db.WithContext(ctx).Table("general_ledger_entries_clik").Select("\"Posting_Date\"").Limit(1).Scan(&rawDate).Error; err != nil {
		return "", fmt.Errorf("actualRepo.GetRawDate: %w", err)
	}
	return fmt.Sprintf("CLIK Date: %s", rawDate), nil
}

func (r *actualRepository) RefreshDataInventory(ctx context.Context) error {
	// 1. Get all unique year-month pairs from staging tables
	query := `
		SELECT DISTINCT TO_CHAR("Posting_Date", 'YYYY') as year, UPPER(TO_CHAR("Posting_Date", 'MON')) as month FROM achhmw_gle_api
		UNION
		SELECT DISTINCT TO_CHAR("Posting_Date", 'YYYY') as year, UPPER(TO_CHAR("Posting_Date", 'MON')) as month FROM general_ledger_entries_clik
	`
	type yearMonth struct {
		Year  string
		Month string
	}
	var pairs []yearMonth
	if err := r.db.WithContext(ctx).Raw(query).Scan(&pairs).Error; err != nil {
		return fmt.Errorf("actualRepo.RefreshDataInventory.Scan: %w", err)
	}

	// 2. Clear old inventory and bulk insert new pairs
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// We use a full refresh approach for simplicity and accuracy
		if err := tx.Exec("TRUNCATE TABLE data_inventory_entities").Error; err != nil {
			return fmt.Errorf("actualRepo.RefreshDataInventory.Truncate: %w", err)
		}

		if len(pairs) > 0 {
			var entities []models.DataInventoryEntity
			now := time.Now()
			for _, p := range pairs {
				if p.Year == "" || p.Month == "" {
					continue
				}
				entities = append(entities, models.DataInventoryEntity{
					Year:      p.Year,
					Month:     p.Month,
					UpdatedAt: now,
				})
			}
			if len(entities) > 0 {
				if err := tx.CreateInBatches(entities, 500).Error; err != nil {
					return fmt.Errorf("actualRepo.RefreshDataInventory.Create: %w", err)
				}
			}
		}
		return nil
	})
}

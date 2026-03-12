package repository

import (
	"fmt"

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

func (r *actualRepository) CreateActualFacts(headers []models.ActualFactEntity) error {
	// 1. บันทึกส่วนหัว (Insert Headers)
	if err := r.db.Omit("ActualAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}
	// 2. รวบรวมยอดเงิน (Collect Amounts)
	var allAmounts []models.ActualAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.ActualAmounts...)
	}
	// 3. บันทึกยอดเงิน (Insert Amounts)
	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

func (r *actualRepository) DeleteAllActualFacts() error {
	// 1. ลบยอดเงิน (Delete Amounts)
	if err := r.db.Exec("DELETE FROM actual_amount_entities").Error; err != nil {
		return err
	}
	// 2. ลบส่วนหัว (Delete Headers)
	return r.db.Exec("DELETE FROM actual_fact_entities").Error
}

func (r *actualRepository) DeleteActualFactsByYear(year string) error {
	// 1. ลบยอดเงินที่เชื่อมโยงกับ Header ของปีนั้นๆ
	// (ต้องใช้ subquery เพราะ Amount ไม่มี Year เก็บไว้โดยตรง)
	// Subquery delete: DELETE FROM actual_amount_entities WHERE actual_fact_id IN (SELECT id FROM actual_fact_entities WHERE year = ?)
	// Use Unscoped to force HARD DELETE
	if err := r.db.Exec(`
		DELETE FROM actual_amount_entities 
		WHERE actual_fact_id IN (SELECT id FROM actual_fact_entities WHERE year = ?)
	`, year).Error; err != nil {
		return err
	}

	// 2. Delete Headers (Hard Delete)
	return r.db.Unscoped().Where("year = ?", year).Delete(&models.ActualFactEntity{}).Error
}

func (r *actualRepository) DeleteActualFactsByMonth(year string, month string) error {
	// 1. ลบยอดเงินเฉพาะเดือนที่ระบุ (และปีที่ระบุด้วยการเชื่อมโยง)
	if err := r.db.Exec(`
		DELETE FROM actual_amount_entities 
		WHERE month = ? AND actual_fact_id IN (SELECT id FROM actual_fact_entities WHERE year = ?)
	`, month, year).Error; err != nil {
		return err
	}

	// 2. อัปเดตยอดรวมรายปี (YearTotal) ของ Header
	// ตั้งค่าเป็น 0 หากไม่มีข้อมูล Amount เหลืออยู่เลย
	if err := r.db.Exec(`
		UPDATE actual_fact_entities f
		SET year_total = COALESCE((SELECT SUM(amount) FROM actual_amount_entities a WHERE a.actual_fact_id = f.id), 0)
		WHERE f.year = ?
	`, year).Error; err != nil {
		return err
	}

	// 3. ล้าง Header ที่ไม่มีข้อมูลเหลืออยู่ออกไป (Optional cleanup)
	return r.db.Unscoped().
		Where("year = ? AND year_total = 0", year).
		Delete(&models.ActualFactEntity{}).Error
}

func (r *actualRepository) DeleteAllActualTransactions() error {
	return r.db.Unscoped().Where("1=1").Delete(&models.ActualTransactionEntity{}).Error
}

func (r *actualRepository) DeleteActualTransactionsByYear(year string) error {
	return r.db.Unscoped().Where("year = ?", year).Delete(&models.ActualTransactionEntity{}).Error
}

func (r *actualRepository) DeleteActualTransactionsByMonth(year string, month string) error {
	monthMap := map[string]string{
		"JAN": "01", "FEB": "02", "MAR": "03", "APR": "04", "MAY": "05", "JUN": "06",
		"JUL": "07", "AUG": "08", "SEP": "09", "OCT": "10", "NOV": "11", "DEC": "12",
	}
	mCode, ok := monthMap[month]
	if !ok {
		return fmt.Errorf("invalid month: %s", month)
	}
	pattern := fmt.Sprintf("%s-%s-%%", year, mCode)
	return r.db.Unscoped().
		Where("year = ? AND posting_date LIKE ?", year, pattern).
		Delete(&models.ActualTransactionEntity{}).Error
}

func (r *actualRepository) GetAllAchHmwGle() ([]models.AchHmwGleEntity, error) {
	var results []models.AchHmwGleEntity
	err := r.db.Find(&results).Error
	return results, err
}

func (r *actualRepository) GetAggregatedHMW(year string, months []string) ([]models.ActualAggregatedDTO, error) {
	var results []models.ActualAggregatedDTO
	// การเพิ่มประสิทธิภาพ: Group โดย Database

	query := r.db.Table("achhmw_gle_api").
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

	err := query.Group(`company, branch, "Global_Dimension_1_Code", "G_L_Account_No", "G_L_Account_Name", "Vendor_Name", UPPER(TO_CHAR("Posting_Date"::DATE, 'MON'))`).
		Scan(&results).Error
	return results, err
}

func (r *actualRepository) GetRawTransactionsHMW(year string, months []string) ([]models.ActualTransactionDTO, error) {
	var results []models.ActualTransactionDTO
	query := r.db.Table("achhmw_gle_api").
		Select(`
			'HMW' as source,
			TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date,
			"Document_No" as doc_no,
			"Description" as description, 
			"G_L_Account_No" as gl_account_no,
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

	err := query.Scan(&results).Error
	return results, err
}

func (r *actualRepository) GetAllClikGle() ([]models.ClikGleEntity, error) {
	var results []models.ClikGleEntity
	err := r.db.Find(&results).Error
	return results, err
}

func (r *actualRepository) GetAggregatedCLIK(year string, months []string) ([]models.ActualAggregatedDTO, error) {
	var results []models.ActualAggregatedDTO
	// CLIK uses Global_Dimension_2_Code for Branch
	query := r.db.Table("general_ledger_entries_clik").
		Select(`
			'CLIK' as company,
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

	err := query.Group(`"Global_Dimension_2_Code", "Global_Dimension_1_Code", "G_L_Account_No", "G_L_Account_Name", "Vendor_Name", UPPER(TO_CHAR("Posting_Date"::DATE, 'MON'))`).
		Scan(&results).Error
	return results, err
}

func (r *actualRepository) GetRawTransactionsCLIK(year string, months []string) ([]models.ActualTransactionDTO, error) {
	var results []models.ActualTransactionDTO
	query := r.db.Table("general_ledger_entries_clik").
		Select(`
			'CLIK' as source,
			TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date,
			"Document_No" as doc_no, 
			"Description" as description,
			"G_L_Account_No" as gl_account_no,
			"G_L_Account_Name" as gl_account_name,
			"Global_Dimension_1_Code" as department,
			"Amount" as amount,
			"Vendor_Name" as vendor,
			'CLIK' as company,
			"Global_Dimension_2_Code" as branch
		`).
		Where("TO_CHAR(\"Posting_Date\"::DATE, 'YYYY') = ?", year)

	if len(months) > 0 {
		query = query.Where("UPPER(TO_CHAR(\"Posting_Date\"::DATE, 'MON')) IN ?", months)
	}

	err := query.Scan(&results).Error
	return results, err
}

func (r *actualRepository) CreateActualTransactions(txs []models.ActualTransactionEntity) error {
	if len(txs) == 0 {
		return nil
	}
	// Bulk insert with 500 records per batch for performance
	return r.db.CreateInBatches(txs, 500).Error
}

func (r *actualRepository) GetRawDate() (string, error) {
	var rawDate string
	// Try HMW first
	if err := r.db.Table("achhmw_gle_api").Select("\"Posting_Date\"").Limit(1).Scan(&rawDate).Error; err != nil {
		return "", err
	}
	if rawDate != "" {
		return fmt.Sprintf("HMW Date: %s", rawDate), nil
	}

	// Try CLIK if HMW empty
	if err := r.db.Table("general_ledger_entries_clik").Select("\"Posting_Date\"").Limit(1).Scan(&rawDate).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("CLIK Date: %s", rawDate), nil
}

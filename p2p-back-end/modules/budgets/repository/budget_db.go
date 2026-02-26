package repository

import (
	"fmt"
	"p2p-back-end/modules/entities/models"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type budgetRepositoryDB struct {
	db *gorm.DB
}

func NewBudgetRepositoryDB(db *gorm.DB) models.BudgetRepository {
	return &budgetRepositoryDB{db: db}
}

// ตัวช่วยสำหรับ Transaction
func (r *budgetRepositoryDB) WithTrx(trxHandle func(repo models.BudgetRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewBudgetRepositoryDB(tx)
		return trxHandle(repo)
	})
}

// ---------------------------------------------------------
// ฟังก์ชันสร้างไฟล์ (File Create Methods)
// ---------------------------------------------------------

func (r *budgetRepositoryDB) CreateFileBudget(file *models.FileBudgetEntity) error {
	return r.db.Create(file).Error
}

func (r *budgetRepositoryDB) CreateFileCapexBudget(file *models.FileCapexBudgetEntity) error {
	return r.db.Create(file).Error
}

func (r *budgetRepositoryDB) CreateFileCapexActual(file *models.FileCapexActualEntity) error {
	return r.db.Create(file).Error
}

// ---------------------------------------------------------
// ฟังก์ชันสร้างข้อมูล (Fact Create Methods - บันทึกแบบกลุ่ม + ความสัมพันธ์)
// ---------------------------------------------------------

// 1. Budget (PL)
func (r *budgetRepositoryDB) CreateBudgetFacts(headers []models.BudgetFactEntity) error {
	// GORM CreateInBatches ไม่บันทึก Association (Amounts) โดยอัตโนมัติ
	// เราต้องแยกบันทึก Header และ Amount เองเพื่อประสิทธิภาพ 100%

	// 1.1 บันทึกส่วนหัว (Headers)
	if err := r.db.Omit("BudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	// 1.2 รวบรวมข้อมูลยอดเงินทั้งหมด (Amounts)
	var allAmounts []models.BudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.BudgetAmounts...)
	}

	// 1.3 บันทึกยอดเงิน (Insert Amounts)
	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// 2. Capex Budget
func (r *budgetRepositoryDB) CreateCapexBudgetFacts(headers []models.CapexBudgetFactEntity) error {
	if err := r.db.Omit("CapexBudgetAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	// 1.2 รวบรวมข้อมูลยอดเงินทั้งหมด (Amounts)
	var allAmounts []models.CapexBudgetAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.CapexBudgetAmounts...)
	}

	// 1.3 บันทึกยอดเงิน (Insert Amounts)
	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// 3. Capex Actual
func (r *budgetRepositoryDB) CreateCapexActualFacts(headers []models.CapexActualFactEntity) error {
	// 1.1 บันทึกส่วนหัว (Headers)
	if err := r.db.Omit("CapexActualAmounts").CreateInBatches(&headers, 1000).Error; err != nil {
		return err
	}

	// 1.2 รวบรวมข้อมูลยอดเงินทั้งหมด (Amounts)
	var allAmounts []models.CapexActualAmountEntity
	for _, h := range headers {
		allAmounts = append(allAmounts, h.CapexActualAmounts...)
	}

	// 1.3 บันทึกยอดเงิน (Insert Amounts)
	if len(allAmounts) > 0 {
		return r.db.CreateInBatches(&allAmounts, 1000).Error
	}
	return nil
}

// ---------------------------------------------------------
// ฟังก์ชันดึงรายการไฟล์ (File List Methods)
// ---------------------------------------------------------

func (r *budgetRepositoryDB) ListFileBudgets() ([]models.FileBudgetEntity, error) {
	var files []models.FileBudgetEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *budgetRepositoryDB) ListFileCapexBudgets() ([]models.FileCapexBudgetEntity, error) {
	var files []models.FileCapexBudgetEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *budgetRepositoryDB) ListFileCapexActuals() ([]models.FileCapexActualEntity, error) {
	var files []models.FileCapexActualEntity
	err := r.db.Order("upload_at desc").Find(&files).Error
	return files, err
}

func (r *budgetRepositoryDB) GetFileBudget(id string) (*models.FileBudgetEntity, error) {
	var file models.FileBudgetEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *budgetRepositoryDB) GetFileCapexBudget(id string) (*models.FileCapexBudgetEntity, error) {
	var file models.FileCapexBudgetEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *budgetRepositoryDB) GetFileCapexActual(id string) (*models.FileCapexActualEntity, error) {
	var file models.FileCapexActualEntity
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// ---------------------------------------------------------------------
// ส่วนแสดงผล Dashboard / รายละเอียด (Dashboard / Detail View)
// ---------------------------------------------------------------------

func (r *budgetRepositoryDB) GetBudgetFilterOptions() ([]models.BudgetFactEntity, error) {
	fmt.Println("[DEBUG] Repo: GetBudgetFilterOptions START")
	var results []models.BudgetFactEntity
	// เลือกข้อมูลที่ไม่ซ้ำกันเพื่อสร้างลำดับชั้น (Hierarchy)
	err := r.db.Model(&models.BudgetFactEntity{}).
		Distinct("\"group\"", "department", "entity_gl", "conso_gl", "gl_name").
		Order("\"group\", department, entity_gl, conso_gl").
		Find(&results).Error
	fmt.Printf("[DEBUG] Repo: GetBudgetFilterOptions END - Count: %d, Err: %v\n", len(results), err)
	return results, err
}

func (r *budgetRepositoryDB) GetOrganizationStructure() ([]models.BudgetFactEntity, error) {
	var results []models.BudgetFactEntity
	// รวม Entity และ Branch ที่ไม่ซ้ำกันจากทั้งตาราง Budget และ Actual
	// GORM ไม่รองรับ UNION ในการ Scan struct โดยตรงได้ง่ายๆ
	// เราจึงใช้ Raw SQL เพื่อความชัดเจนและประสิทธิภาพ

	query := `
        SELECT DISTINCT entity, branch, department FROM budget_fact_entities WHERE entity != ''
        UNION
        SELECT DISTINCT entity, branch, department FROM actual_fact_entities WHERE entity != ''
        ORDER BY entity, branch, department
    `

	err := r.db.Raw(query).Scan(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetBudgetDetails(filter map[string]interface{}) ([]models.BudgetDetailDTO, error) {
	var results []models.BudgetDetailDTO
	query := r.db.Model(&models.BudgetFactEntity{})

	// Dynamic Filtering Helper
	applyFilter := func(key string, dbCol string) {
		if val, ok := filter[key]; ok {
			var strs []string
			if s, ok := val.([]string); ok {
				strs = s
			} else if s, ok := val.([]interface{}); ok {
				for _, item := range s {
					strs = append(strs, fmt.Sprintf("%v", item))
				}
			}
			if len(strs) > 0 {
				query = query.Where(fmt.Sprintf("budget_fact_entities.%s IN ?", dbCol), strs)
			}
		}
	}

	applyFilter("groups", "\"group\"")
	applyFilter("departments", "department")
	applyFilter("entity_gls", "entity_gl")
	applyFilter("conso_gls", "conso_gl")
	applyFilter("entities", "entity")
	applyFilter("branches", "branch")

	type flatResult struct {
		ConsoGL string
		GLName  string
		Month   string
		Amount  float64
	}

	var flatData []flatResult
	err := query.
		Joins("JOIN budget_amount_entities ba ON ba.budget_fact_id = budget_fact_entities.id AND ba.deleted_at IS NULL").
		Select("budget_fact_entities.conso_gl, MAX(budget_fact_entities.gl_name) as gl_name, ba.month, SUM(ba.amount) as amount").
		Group("budget_fact_entities.conso_gl, ba.month").
		Scan(&flatData).Error

	if err != nil {
		return nil, err
	}

	// Map flat results to nested DTO
	budgetMap := make(map[string]*models.BudgetDetailDTO)
	for _, row := range flatData {
		dto, exists := budgetMap[row.ConsoGL]
		if !exists {
			dto = &models.BudgetDetailDTO{
				ConsoGL:       row.ConsoGL,
				GLName:        row.GLName,
				YearTotal:     0,
				BudgetAmounts: []models.BudgetAmountDTO{},
			}
			budgetMap[row.ConsoGL] = dto
		}
		dto.BudgetAmounts = append(dto.BudgetAmounts, models.BudgetAmountDTO{
			Month:  row.Month,
			Amount: row.Amount,
		})
		dto.YearTotal += row.Amount
	}

	// Convert map to slice
	for _, dto := range budgetMap {
		results = append(results, *dto)
	}

	return results, nil
}

func (r *budgetRepositoryDB) GetActualDetails(filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	var results []models.ActualFactEntity
	query := r.db.Model(&models.ActualFactEntity{}).Preload("ActualAmounts")

	// Dynamic Filtering Helper
	applyFilter := func(key string, dbCol string) {
		if val, ok := filter[key]; ok {
			var strs []string
			if s, ok := val.([]string); ok {
				strs = s
			} else if s, ok := val.([]interface{}); ok {
				for _, item := range s {
					strs = append(strs, fmt.Sprintf("%v", item))
				}
			}
			if len(strs) > 0 {
				query = query.Where(fmt.Sprintf("%s IN ?", dbCol), strs)
			}
		}
	}

	// มิติข้อมูลสำหรับ Actuals: Entity, Branch, Department, ConsoGL (Code), GLName
	// หมายเหตุ: Actuals อาจจะยังไม่มี "Group" หรือ "EntityGL" ถ้ายังไม่ได้ Mapping
	applyFilter("departments", "department")
	applyFilter("conso_gls", "conso_gl")
	applyFilter("entities", "entity")
	applyFilter("branches", "branch")

	// เรียงลำดับข้อมูล
	err := query.Order("department, conso_gl, gl_name").Find(&results).Error
	return results, err
}

// ---------------------------------------------------------
// ฟังก์ชันลบไฟล์ (File Delete Methods)
// ---------------------------------------------------------

func (r *budgetRepositoryDB) DeleteFileBudget(id string) error {
	return r.db.Delete(&models.FileBudgetEntity{}, "id = ?", id).Error
}

func (r *budgetRepositoryDB) DeleteFileCapexBudget(id string) error {
	return r.db.Delete(&models.FileCapexBudgetEntity{}, "id = ?", id).Error
}

func (r *budgetRepositoryDB) DeleteFileCapexActual(id string) error {
	return r.db.Delete(&models.FileCapexActualEntity{}, "id = ?", id).Error
}

// ---------------------------------------------------------------------
// 4. Delete All Facts (For Sync)
// ---------------------------------------------------------------------
// ---------------------------------------------------------------------
// ---------------------------------------------------------------------
// 4. ลบข้อมูล Fact ทั้งหมด (สำหรับการ Sync)
// ---------------------------------------------------------------------
func (r *budgetRepositoryDB) DeleteAllBudgetFacts() error {
	// 1. ลบยอดเงิน (ลูก)
	// Use Raw SQL to guarantee execution and avoid Soft Delete confusion
	if err := r.db.Exec("DELETE FROM budget_amount_entities").Error; err != nil {
		return err
	}
	// 2. ลบส่วนหัว (แม่)
	return r.db.Exec("DELETE FROM budget_fact_entities").Error
}

func (r *budgetRepositoryDB) DeleteBudgetFactsByFileID(fileID string) error {
	// 1. Delete Amounts (Subquery or Join)
	if err := r.db.Exec(`
		DELETE FROM budget_amount_entities 
		WHERE budget_fact_id IN (SELECT id FROM budget_fact_entities WHERE file_budget_id = ?)
	`, fileID).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Where("file_budget_id = ?", fileID).Delete(&models.BudgetFactEntity{}).Error
}

func (r *budgetRepositoryDB) DeleteAllCapexBudgetFacts() error {
	// 1. Delete Amounts
	if err := r.db.Exec("DELETE FROM capex_budget_amount_entities").Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Exec("DELETE FROM capex_budget_fact_entities").Error
}

func (r *budgetRepositoryDB) DeleteCapexBudgetFactsByFileID(fileID string) error {
	// 1. Delete Amounts
	if err := r.db.Exec(`
		DELETE FROM capex_budget_amount_entities 
		WHERE capex_budget_fact_id IN (SELECT id FROM capex_budget_fact_entities WHERE file_capex_budget_id = ?)
	`, fileID).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Where("file_capex_budget_id = ?", fileID).Delete(&models.CapexBudgetFactEntity{}).Error
}

func (r *budgetRepositoryDB) DeleteAllCapexActualFacts() error {
	// 1. Delete Amounts
	if err := r.db.Exec("DELETE FROM capex_actual_amount_entities").Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Exec("DELETE FROM capex_actual_fact_entities").Error
}

func (r *budgetRepositoryDB) DeleteCapexActualFactsByFileID(fileID string) error {
	// 1. Delete Amounts
	if err := r.db.Exec(`
		DELETE FROM capex_actual_amount_entities 
		WHERE capex_actual_fact_id IN (SELECT id FROM capex_actual_fact_entities WHERE file_capex_actual_id = ?)
	`, fileID).Error; err != nil {
		return err
	}
	// 2. Delete Headers
	return r.db.Unscoped().Where("file_capex_actual_id = ?", fileID).Delete(&models.CapexActualFactEntity{}).Error
}

// ---------------------------------------------------------------------
// ---------------------------------------------------------------------
// 5. อัปเดตไฟล์ (เปลี่ยนชื่อ) - Implementation
// ---------------------------------------------------------------------

func (r *budgetRepositoryDB) UpdateFileBudget(id string, filename string) error {
	return r.db.Model(&models.FileBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

func (r *budgetRepositoryDB) UpdateFileCapexBudget(id string, filename string) error {
	return r.db.Model(&models.FileCapexBudgetEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

func (r *budgetRepositoryDB) UpdateFileCapexActual(id string, filename string) error {
	return r.db.Model(&models.FileCapexActualEntity{}).Where("id = ?", id).Update("file_name", filename).Error
}

// ---------------------------------------------------------------------
// ---------------------------------------------------------------------
// 6. ข้อมูลจริง (Actuals/Operational) - Sync Implementation
// ---------------------------------------------------------------------

func (r *budgetRepositoryDB) CreateActualFacts(headers []models.ActualFactEntity) error {
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

func (r *budgetRepositoryDB) DeleteAllActualFacts() error {
	// 1. ลบยอดเงิน (Delete Amounts)
	if err := r.db.Exec("DELETE FROM actual_amount_entities").Error; err != nil {
		return err
	}
	// 2. ลบส่วนหัว (Delete Headers)
	return r.db.Exec("DELETE FROM actual_fact_entities").Error
}

func (r *budgetRepositoryDB) DeleteActualFactsByYear(year string) error {
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

func (r *budgetRepositoryDB) GetAggregatedHMW(year string, months []string) ([]models.ActualAggregatedDTO, error) {
	var results []models.ActualAggregatedDTO
	// การเพิ่มประสิทธิภาพ: Group โดย Database

	query := r.db.Table("achhmw_gle_api").
		Select(`
			company, 
			branch, 
			"Global_Dimension_1_Code" as department, 
			"G_L_Account_No" as gl_account_no, 
			"G_L_Account_Name" as gl_account_name, 
			UPPER(TO_CHAR("Posting_Date"::DATE, 'MON')) as month, 
			SUM("Credit_Amount") as total_amount
		`).
		Where("LEFT(\"Posting_Date\", 4) = ?", year)

	if len(months) > 0 {
		query = query.Where("UPPER(TO_CHAR(\"Posting_Date\"::DATE, 'MON')) IN ?", months)
	}

	err := query.Group(`company, branch, "Global_Dimension_1_Code", "G_L_Account_No", "G_L_Account_Name", UPPER(TO_CHAR("Posting_Date"::DATE, 'MON'))`).
		Scan(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetAllAchHmwGle() ([]models.AchHmwGleEntity, error) {
	var results []models.AchHmwGleEntity
	err := r.db.Find(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetAggregatedCLIK(year string, months []string) ([]models.ActualAggregatedDTO, error) {
	var results []models.ActualAggregatedDTO
	// CLIK uses Global_Dimension_2_Code for Branch
	query := r.db.Table("general_ledger_entries_clik").
		Select(`
			'CLIK' as company,
			"Global_Dimension_2_Code" as branch, 
			"Global_Dimension_1_Code" as department, 
			"G_L_Account_No" as gl_account_no, 
			"G_L_Account_Name" as gl_account_name, 
			UPPER(TO_CHAR("Posting_Date"::DATE, 'MON')) as month, 
			SUM("Credit_Amount") as total_amount
		`).
		Where("LEFT(\"Posting_Date\", 4) = ?", year)

	if len(months) > 0 {
		query = query.Where("UPPER(TO_CHAR(\"Posting_Date\"::DATE, 'MON')) IN ?", months)
	}

	err := query.Group(`"Global_Dimension_2_Code", "Global_Dimension_1_Code", "G_L_Account_No", "G_L_Account_Name", UPPER(TO_CHAR("Posting_Date"::DATE, 'MON'))`).
		Scan(&results).Error
	return results, err
}

func (r *budgetRepositoryDB) GetAllClikGle() ([]models.ClikGleEntity, error) {
	var results []models.ClikGleEntity
	err := r.db.Find(&results).Error
	return results, err
}

// ---------------------------------------------------------------------
// ---------------------------------------------------------------------
// 7. การรวมข้อมูลสำหรับ Dashboard (Dashboard Aggregation - Optimized)
// ---------------------------------------------------------------------
func (r *budgetRepositoryDB) GetActualTransactions(filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	// Filtering Logic
	whereClause := "1=1"
	var args []interface{}

	// ความต้องการผู้ใช้: Map ชื่อต้นทาง (Source Names) -> รหัส (Codes) สำหรับเก็บลงฐานข้อมูล
	// เราต้อง Map ย้อนกลับ (Code -> Source Name) เพื่อกรองตารางต้นทางโดยตรง
	// แผนผัง Entity (Code -> Source Name)
	// อัปเดต: ผู้ใช้ยืนยันว่า Source ใช้ตัวพิมพ์ใหญ่ผสมเล็ก (Title Case) เช่น "Honda Maliwan"
	entityCodeToNameMap := map[string][]string{
		"HMW":  {"HONDA MALIWAN", "HMW", "Honda Maliwan"},
		"ACG":  {"AUTOCORP HOLDING", "ACG", "Autocorp Holding"},
		"CLIK": {"CLIK"},
	}

	fmt.Printf("[DEBUG] GetActualTransactions Check Filter: %+v\n", filter)

	// แผนผัง Branch (Code -> Source Name(s)) - HMW/ACG/CLIK รวมกัน
	// หมายเหตุ: รหัส Branch ใน UI (HOF, VEE, ฯลฯ) อาจ Map ไปยังชื่อต้นทางหลายชื่อ หรือชื่อเดียวกัน
	// เราต้องใช้ตัวช่วยเพื่อหาชื่อต้นทางทั้งหมดจากรหัสที่ได้รับ
	// เนื่องจาก Logic ตอน Sync คือ: Normalize(Source) -> Map[Source] -> Code
	// เราต้องการ: Code -> [รายการชื่อต้นทางที่ Map มาหา Code นี้]
	// การดูแลสองที่มันยุ่งยาก แต่อนาคตควรทำเป็นตารางกลาง
	// ตอนนี้ยอมให้ใช้ทั้ง Code และชื่อทั่วไป
	// หรือดีกว่านั้น: ใน `achhmw_gle_api`, branch น่าจะเป็นชื่อต้นทาง (เช่น "HEAD OFFICE")
	// ใน `actual_fact_entities` มันคือ "HOF"
	// ถ้า User กรอง "HOF" เราต้องไปค้นหา "HEAD OFFICE" ในต้นทาง

	// Hardcoded Reverse Map อ้างอิงจาก budget_service.go
	// อัปเดต: เพิ่มแบบ Title Case ตามที่ User ส่งรูปมาให้ดู
	branchCodeToNameMap := map[string][]string{
		"HOF":      {"HEAD OFFICE", "AUTOCORP HEAD OFFICE", "HEADOFFICE", "HOF", "Head Office", "Headoffice"},
		"BUR":      {"BURIRUM", "BUR", "Burirum"},
		"KBI":      {"KRABI", "KBI", "Krabi"},
		"MSR":      {"MINI_SURIN", "MSR", "Mini_Surin"},
		"MKB":      {"MUEANG KRABI", "MKB", "Mueang Krabi"},
		"NAK":      {"NAKA", "NAK", "Naka"},
		"AVN":      {"NANGRONG", "AVN", "Nangrong"},
		"PHC":      {"PHACHA", "PHC", "Phacha"},
		"PRA":      {"PHUKET", "PRA", "Phuket"},
		"SUR":      {"SURIN", "SUR", "Surin"},
		"VEE":      {"VEERAWAT", "VEE", "Veerawat"},
		"HQ":       {"AUTOCORP HEAD OFFICE", "HQ", "Autocorp Head Office"},
		"Branch00": {"", "Branch00"},
		// Add Branch01..15 if needed
	}
	// Add Branch01-15 loop
	for i := 1; i <= 15; i++ {
		key := fmt.Sprintf("Branch%02d", i)
		branchCodeToNameMap[key] = []string{fmt.Sprintf("BRANCH%02d", i), fmt.Sprintf("Branch%02d", i)}
	}

	// ตัวกรอง Entity
	var hmwEntities []string
	var clikEntities []string
	var selectedEntities []string // เก็บ UI Code ที่เลือกไว้เพื่อไปกรอง actual_fact_entities
	if val, ok := filter["entities"]; ok {
		var entities []string
		if s, ok := val.([]string); ok {
			entities = s
		} else if s, ok := val.([]interface{}); ok {
			for _, item := range s {
				entities = append(entities, fmt.Sprintf("%v", item))
			}
		}

		if len(entities) > 0 {
			selectedEntities = entities // เก็บ UI Code ที่เลือกไว้
			// แปลง UI Codes (HMW, ACG) ไปเป็นชื่อต้นทาง (Source Names)
			for _, e := range entities {
				if names, ok := entityCodeToNameMap[e]; ok {
					hmwEntities = append(hmwEntities, names...)
					clikEntities = append(clikEntities, names...) // CLIK อาจจะใช้ Logic เดียวกัน หรือใช้ 'CLIK' เฉยๆ
				} else {
					// Fallback: Use the code itself
					hmwEntities = append(hmwEntities, e)
					clikEntities = append(clikEntities, e)
				}
			}
		}
	}

	// ตัวกรอง Branch
	var hmwBranches []string
	var clikBranches []string
	if val, ok := filter["branches"]; ok {
		var branches []string
		if s, ok := val.([]string); ok {
			branches = s
		} else if s, ok := val.([]interface{}); ok {
			for _, item := range s {
				branches = append(branches, fmt.Sprintf("%v", item))
			}
		}

		if len(branches) > 0 {
			for _, b := range branches {
				if names, ok := branchCodeToNameMap[b]; ok {
					hmwBranches = append(hmwBranches, names...)
					clikBranches = append(clikBranches, names...)
				} else {
					hmwBranches = append(hmwBranches, b)
					clikBranches = append(clikBranches, b)
				}
			}
		}
	}

	// กรองโดย GL Account No (จำเป็นสำหรับการ Drill Down)
	// ความต้องการผู้ใช้: Map Conso GL (Filter) -> Entity GL (Source)
	if val, ok := filter["conso_gls"]; ok {
		var consoGLs []string
		if s, ok := val.([]string); ok {
			consoGLs = s
		} else if s, ok := val.([]interface{}); ok {
			for _, item := range s {
				consoGLs = append(consoGLs, fmt.Sprintf("%v", item))
			}
		}

		if len(consoGLs) > 0 && len(consoGLs) < 200 {
			// ขั้นตอนที่ 1: หา Entity GLs ที่ตรงกันจากตาราง Mapping
			// ใช้ BUDGET FACTS เป็นแหล่งข้อมูลหลัก (Source of Truth) สำหรับการ Mapping
			var mappedEntityGLs []string
			mappingQuery := r.db.Model(&models.BudgetFactEntity{}).
				Where("conso_gl IN ?", consoGLs).
				Distinct("entity_gl")

			if len(selectedEntities) > 0 {
				mappingQuery = mappingQuery.Where("entity IN ?", selectedEntities)
			}

			mappingQuery.Pluck("entity_gl", &mappedEntityGLs)
			// fmt.Printf("[DEBUG] GL Mapping (Budget): Conso %v -> EntityGLs %v\n", consoGLs, mappedEntityGLs)

			// รวม mapped + original
			finalGLs := append(mappedEntityGLs, consoGLs...)

			// Remove duplicates
			uniqueGLs := make(map[string]bool)
			var list []string
			for _, item := range finalGLs {
				if _, value := uniqueGLs[item]; !value {
					uniqueGLs[item] = true
					list = append(list, item)
				}
			}

			// ขั้นตอนที่ 2: ใช้รายการที่รวมแล้วไปกรองตารางต้นทาง
			if len(list) > 0 {
				whereClause += " AND \"G_L_Account_No\" IN ?"
				args = append(args, list)
			}
		} else if len(consoGLs) >= 200 {
			// หากส่งมาเยอะเกินไป (เช่น ส่งมาทั้งหมด) ให้ถือว่าไม่กรอง GL เพื่อความเร็ว
			// ไม่ต้องเพิ่ม WHERE G_L_Account_No เข้าไป
		}
	}

	// Date Filtering
	if val, ok := filter["start_date"]; ok {
		if startDate, ok := val.(string); ok && startDate != "" {
			whereClause += " AND \"Posting_Date\" >= ?"
			args = append(args, startDate)
		}
	}
	if val, ok := filter["end_date"]; ok {
		if endDate, ok := val.(string); ok && endDate != "" {
			whereClause += " AND \"Posting_Date\" <= ?"
			args = append(args, endDate)
		}
	}

	// Department Filtering (New)
	if val, ok := filter["departments"]; ok {
		var depts []string
		if s, ok := val.([]string); ok {
			depts = s
		} else if s, ok := val.([]interface{}); ok {
			for _, item := range s {
				depts = append(depts, fmt.Sprintf("%v", item))
			}
		}
		if len(depts) > 0 {
			// Both HMW and CLIK use "Global_Dimension_1_Code" for Department
			whereClause += " AND \"Global_Dimension_1_Code\" IN ?"
			args = append(args, depts)
		}
	}

	// ✅ อัปเดต Logic การกรอง:
	// แสดงเฉพาะ Transaction ของปีที่ "Sync" แล้วเท่านั้น (Active ใน actual_fact_entities)
	var activeYears []string
	r.db.Model(&models.ActualFactEntity{}).Distinct("year").Pluck("year", &activeYears)
	fmt.Printf("[DEBUG] Active Years in Facts: %v\n", activeYears)

	// ✅ Optimized Year Filter using ranges (for Index support)
	if len(activeYears) > 0 {
		var yearClauses []string
		for _, y := range activeYears {
			yearClauses = append(yearClauses, fmt.Sprintf("(\"Posting_Date\" >= '%s-01-01' AND \"Posting_Date\" <= '%s-12-31')", y, y))
		}
		whereClause += " AND (" + strings.Join(yearClauses, " OR ") + ")"
	} else {
		whereClause += " AND 1=0"
	}

	fmt.Printf("[DEBUG] WhereClause: %s\n", whereClause)
	fmt.Printf("[DEBUG] Args: %v\n", args)

	// No applyLimit for the queries so we can get accurate count
	// Pagination parameters
	page := 0
	limit := 10
	if p, ok := filter["page"].(float64); ok {
		page = int(p)
	}
	if l, ok := filter["limit"].(float64); ok {
		limit = int(l)
	}
	offset := page * limit

	// Construct the Union query as a raw SQL string or using GORM's Table
	// We'll use a subquery approach for accurate total count and sorting
	hmwSQL := r.db.Table("achhmw_gle_api").
		Select(`
			'HMW' as source,
			TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date,
			"Document_No" as doc_no,
			"Description" as description, 
			"G_L_Account_No" as gl_account_no,
			"G_L_Account_Name" as gl_account_name,
			"Global_Dimension_1_Code" as department,
			"Credit_Amount" as amount,
			company,
			branch
		`).
		Where(whereClause, args...)

	if len(hmwEntities) > 0 {
		hmwSQL = hmwSQL.Where("company IN ?", hmwEntities)
	}
	if len(hmwBranches) > 0 {
		hmwSQL = hmwSQL.Where("branch IN ?", hmwBranches)
	}

	clikSQL := r.db.Table("general_ledger_entries_clik").
		Select(`
			'CLIK' as source,
			TO_CHAR("Posting_Date"::DATE, 'YYYY-MM-DD') as posting_date,
			"Document_No" as doc_no, 
			"Description" as description,
			"G_L_Account_No" as gl_account_no,
			"G_L_Account_Name" as gl_account_name,
			"Global_Dimension_1_Code" as department,
			"Credit_Amount" as amount,
			'CLIK' as company,
			"Global_Dimension_2_Code" as branch
		`).
		Where(whereClause, args...)

	if len(hmwEntities) > 0 {
		hasClik := false
		for _, e := range hmwEntities {
			if e == "CLIK" {
				hasClik = true
				break
			}
		}
		if !hasClik {
			clikSQL = clikSQL.Where("1=0")
		}
	}
	if len(clikBranches) > 0 {
		clikSQL = clikSQL.Where("\"Global_Dimension_2_Code\" IN ?", clikBranches)
	}

	// 1. Get Total Count (Simpler and faster)
	var totalCount int64
	var countHMW, countCLIK int64

	// Construct dedicated count queries without the expensive Select(...) part
	hmwCountQuery := r.db.Table("achhmw_gle_api").Where(whereClause, args...)
	if len(hmwEntities) > 0 {
		hmwCountQuery = hmwCountQuery.Where("company IN ?", hmwEntities)
	}
	if len(hmwBranches) > 0 {
		hmwCountQuery = hmwCountQuery.Where("branch IN ?", hmwBranches)
	}

	clikCountQuery := r.db.Table("general_ledger_entries_clik").Where(whereClause, args...)
	if len(hmwEntities) > 0 {
		hasClik := false
		for _, e := range hmwEntities {
			if e == "CLIK" {
				hasClik = true
				break
			}
		}
		if !hasClik {
			clikCountQuery = clikCountQuery.Where("1=0")
		}
	}
	if len(clikBranches) > 0 {
		clikCountQuery = clikCountQuery.Where("\"Global_Dimension_2_Code\" IN ?", clikBranches)
	}

	// Use channels to wait for result
	type countResult struct {
		count int64
		err   error
	}
	hmwChan := make(chan countResult, 1)
	clikChan := make(chan countResult, 1)

	go func() {
		var c int64
		err := hmwCountQuery.Count(&c).Error
		hmwChan <- countResult{c, err}
	}()

	go func() {
		var c int64
		err := clikCountQuery.Count(&c).Error
		clikChan <- countResult{c, err}
	}()

	resHMW := <-hmwChan
	if resHMW.err != nil {
		return nil, resHMW.err
	}
	resCLIK := <-clikChan
	if resCLIK.err != nil {
		return nil, resCLIK.err
	}

	countHMW = resHMW.count
	countCLIK = resCLIK.count
	totalCount = countHMW + countCLIK

	// 2. Get Paginated Data (Double-Sort Optimization)
	var pagedData []models.ActualTransactionDTO

	// Apply ORDER BY and LIMIT to EACH branch first.
	// This makes the Union extremely fast because it only combines a few rows.
	hmwSQL_sorted := hmwSQL.Order("posting_date ASC").Limit(limit + offset)
	clikSQL_sorted := clikSQL.Order("posting_date ASC").Limit(limit + offset)

	err := r.db.Raw(`
		SELECT * FROM (
			SELECT * FROM (?) AS hmw_sub
			UNION ALL
			SELECT * FROM (?) AS clik_sub
		) AS u
		ORDER BY posting_date ASC
		LIMIT ? OFFSET ?
	`, hmwSQL_sorted, clikSQL_sorted, limit, offset).Scan(&pagedData).Error

	if err != nil {
		return nil, err
	}

	return &models.PaginatedActualTransactionDTO{
		Data:       pagedData,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (r *budgetRepositoryDB) GetDashboardAggregates(filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	summary := &models.DashboardSummaryDTO{
		DepartmentData: []models.DepartmentStatDTO{},
		ChartData:      []models.MonthlyStatDTO{},
	}

	// ตัวช่วยกรองข้อมูลแบบไดนามิก (คืนค่า GORM Scope)
	applyFilter := func(tx *gorm.DB, tableName string) *gorm.DB {
		// Helper to safely convert interface{} to []string
		toStringSlice := func(val interface{}) []string {
			if strs, ok := val.([]string); ok {
				return strs
			}
			if interfaces, ok := val.([]interface{}); ok {
				var strs []string
				for _, i := range interfaces {
					if s, ok := i.(string); ok {
						strs = append(strs, s)
					}
				}
				return strs
			}
			return nil
		}

		if val, ok := filter["entities"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".entity IN ?", strs)
			}
		}
		if val, ok := filter["branches"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".branch IN ?", strs)
			}
		}
		// Added: Departments Filter
		if val, ok := filter["departments"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".department IN ?", strs)
			}
		}
		// Added: NavCodes Filter (For Sub-Department Drill-down)
		if val, ok := filter["nav_codes"]; ok {
			if strs := toStringSlice(val); len(strs) > 0 {
				tx = tx.Where(tableName+".nav_code IN ?", strs)
			}
		}
		// Added: Year Filter (Only for Actuals, Budget ignores year)
		if val, ok := filter["year"]; ok {
			if s, ok := val.(string); ok && s != "" {
				s = strings.ReplaceAll(s, "FY", "")
				if !strings.Contains(tableName, "budget") {
					tx = tx.Where(tableName+".year = ?", s)
				}
			}
		}
		return tx
	}

	// Determine Grouping Strategy (Drill-down vs Summary)
	groupBy := "department"
	selectCol := "department"

	// Helper to safely convert interface{} to []string
	toStringSlice := func(val interface{}) []string {
		if strs, ok := val.([]string); ok {
			return strs
		}
		if interfaces, ok := val.([]interface{}); ok {
			var strs []string
			for _, i := range interfaces {
				if s, ok := i.(string); ok {
					strs = append(strs, s)
				}
			}
			return strs
		}
		return nil
	}

	if val, ok := filter["departments"]; ok {
		if strs := toStringSlice(val); len(strs) > 0 {
			// If filtered by Department, we likely want to drill down to NavCode
			groupBy = "nav_code"
			selectCol = "nav_code as department"
		}
	}

	// 1. รวมยอดตาม Department (Department Aggregation)
	// Budget
	type DeptResult struct {
		Department string
		Total      float64
	}
	var budgetDept []DeptResult
	tx1 := r.db.Table("budget_fact_entities").Select(selectCol + ", SUM(year_total) as total")
	tx1 = applyFilter(tx1, "budget_fact_entities")
	if err := tx1.Group(groupBy).Scan(&budgetDept).Error; err != nil {
		return nil, err
	}

	// Actual
	// Actual
	var actualDept []DeptResult
	tx2 := r.db.Table("actual_fact_entities").Select(selectCol + ", SUM(year_total) as total")
	tx2 = applyFilter(tx2, "actual_fact_entities")

	// Debug: Count Actual Records before aggregation
	var count int64
	r.db.Model(&models.ActualFactEntity{}).Count(&count)
	fmt.Printf("[DEBUG] GetDashboardAggregates: ActualFactEntity Count = %d\n", count)

	if err := tx2.Group(groupBy).Scan(&actualDept).Error; err != nil {
		// Log error but maybe continue? No, return error
		return nil, err
	}

	// รวมข้อมูล (Merge Department Data)
	deptMap := make(map[string]*models.DepartmentStatDTO)
	for _, b := range budgetDept {
		deptMap[b.Department] = &models.DepartmentStatDTO{Department: b.Department, Budget: b.Total}
		summary.TotalBudget += b.Total
	}
	for _, a := range actualDept {
		if _, ok := deptMap[a.Department]; !ok {
			deptMap[a.Department] = &models.DepartmentStatDTO{Department: a.Department}
		}
		deptMap[a.Department].Actual += a.Total
		summary.TotalActual += a.Total
	}

	// แปลง Map เป็น Slice (Flatten Map)
	var allDepts []models.DepartmentStatDTO
	for _, v := range deptMap {
		allDepts = append(allDepts, *v)
	}

	// ตรรกะการเรียงลำดับ (Sort Logic)
	sortBy := "actual" // Default
	sortOrder := "desc"
	if val, ok := filter["sort_by"]; ok {
		if s, ok := val.(string); ok && s != "" {
			sortBy = s
		}
	}
	if val, ok := filter["sort_order"]; ok {
		if s, ok := val.(string); ok && s != "" {
			sortOrder = s
		}
	}

	// คำนวณสถานะภาพรวม (ก่อนแบ่งหน้า)
	var overBudgetCount, nearLimitCount int
	for _, d := range allDepts {
		budget := d.Budget
		actual := d.Actual
		remaining := budget - actual

		// เกินงบ: (Budget=0 & Actual>0) หรือ (คงเหลือ < 0)
		if (budget == 0 && actual > 0) || remaining < 0 {
			overBudgetCount++
		} else if budget > 0 {
			// ใกล้เต็ม: คงเหลือ < 20%
			ratio := remaining / budget
			if ratio <= 0.2 {
				nearLimitCount++
			}
		}
	}
	summary.OverBudgetCount = overBudgetCount
	summary.NearLimitCount = nearLimitCount

	sort.Slice(allDepts, func(i, j int) bool {
		var valI, valJ float64
		switch sortBy {
		case "budget":
			valI, valJ = allDepts[i].Budget, allDepts[j].Budget
		case "actual": // Spend
			valI, valJ = allDepts[i].Actual, allDepts[j].Actual
		case "remaining":
			valI = allDepts[i].Budget - allDepts[i].Actual
			valJ = allDepts[j].Budget - allDepts[j].Actual
		case "remaining_pct":
			if allDepts[i].Budget != 0 {
				valI = (allDepts[i].Budget - allDepts[i].Actual) / allDepts[i].Budget
			} else {
				valI = -999999 // Treat no budget as lowest priority or handled specifically
			}
			if allDepts[j].Budget != 0 {
				valJ = (allDepts[j].Budget - allDepts[j].Actual) / allDepts[j].Budget
			} else {
				valJ = -999999
			}
		default:
			valI, valJ = allDepts[i].Actual, allDepts[j].Actual
		}

		if sortOrder == "asc" {
			return valI < valJ
		}
		return valI > valJ
	})

	// Pagination Logic
	page := 1
	limit := 10
	if val, ok := filter["page"]; ok {
		if p, ok := val.(float64); ok { // JSON unmarshal makes numbers float64
			page = int(p)
		} else if p, ok := val.(int); ok {
			page = p
		}
	}
	if val, ok := filter["limit"]; ok {
		if l, ok := val.(float64); ok {
			limit = int(l)
		} else if l, ok := val.(int); ok {
			limit = l
		}
	}

	summary.TotalCount = int64(len(allDepts))
	summary.Page = page
	summary.Limit = limit

	start := (page - 1) * limit
	end := start + limit

	if start > len(allDepts) {
		summary.DepartmentData = []models.DepartmentStatDTO{}
	} else {
		if end > len(allDepts) {
			end = len(allDepts)
		}
		summary.DepartmentData = allDepts[start:end]
	}

	// 2. Monthly Aggregation for Chart
	// This is trickier because we need to join with Amount tables.
	// Budget Amounts
	type MonthResult struct {
		Month string
		Total float64
	}
	var budgetMonth []MonthResult
	// Join Header to filter -> Sum Amount
	tx3 := r.db.Table("budget_amount_entities").
		Select("budget_amount_entities.month, SUM(budget_amount_entities.amount) as total").
		Joins("JOIN budget_fact_entities ON budget_amount_entities.budget_fact_id = budget_fact_entities.id")
	tx3 = applyFilter(tx3, "budget_fact_entities")
	if err := tx3.Group("budget_amount_entities.month").Scan(&budgetMonth).Error; err != nil {
		return nil, err
	}

	// Actual Amounts
	var actualMonth []MonthResult
	tx4 := r.db.Table("actual_amount_entities").
		Select("actual_amount_entities.month, SUM(actual_amount_entities.amount) as total").
		Joins("JOIN actual_fact_entities ON actual_amount_entities.actual_fact_id = actual_fact_entities.id")
	tx4 = applyFilter(tx4, "actual_fact_entities")
	if err := tx4.Group("actual_amount_entities.month").Scan(&actualMonth).Error; err != nil {
		return nil, err
	}

	// Merge Chart Data
	monthMap := make(map[string]*models.MonthlyStatDTO)
	// Initialize 12 months? or just map what we have
	// Let's rely on what we have, frontend usually handles ordering or we fix it.
	// We'll normalize keys to JAN,FEB...
	for _, m := range budgetMonth {
		// m.Month might be "January", "JAN", etc. Assumed stored as JAN,FEB from Import.
		monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month, Budget: m.Total}
	}
	for _, m := range actualMonth {
		if _, ok := monthMap[m.Month]; !ok {
			monthMap[m.Month] = &models.MonthlyStatDTO{Month: m.Month}
		}
		monthMap[m.Month].Actual += m.Total
	}

	// Ensure logical order if possible, or just slice
	// Order: JAN, FEB, MAR...
	monthsOrder := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	for _, mon := range monthsOrder {
		if val, ok := monthMap[mon]; ok {
			summary.ChartData = append(summary.ChartData, *val)
		} else {
			summary.ChartData = append(summary.ChartData, models.MonthlyStatDTO{Month: mon, Budget: 0, Actual: 0})
		}
	}

	return summary, nil
}

// Debugging
func (r *budgetRepositoryDB) GetRawDate() (string, error) {
	var rawDate string
	// Try HMW first
	if err := r.db.Table("achhmw_gle_api").Select("\"Posting_Date\"").Limit(1).Scan(&rawDate).Error; err != nil {
		return "", err
	}
	if rawDate != "" {
		return fmt.Sprintf("HMW Date: %s", rawDate), nil
	}

	// Try CLIK if HMW empty
	if err := r.db.Table("acclik_gle_api").Select("\"Posting_Date\"").Limit(1).Scan(&rawDate).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("CLIK Date: %s", rawDate), nil
}

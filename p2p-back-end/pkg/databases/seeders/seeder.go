package seeders

import (
	"fmt"
	"p2p-back-end/modules/entities/models"
	"strings"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// SeedGLMappings reads GL Groupping _ Mapping.xlsx and populates the gl_mapping_entities table.
// It uses an idempotent approach: only inserts if (Entity, EntityGL) does not exist.
func SeedGLMappings(db *gorm.DB) error {
	fmt.Println("Syncing GL Mappings from GL Groupping _ Mapping.xlsx (Idempotent)...")

	filePath := "pkg/databases/seed_data/GL Groupping _ Mapping.xlsx"
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		// Try alternative path if run from different CWD
		filePath = "../pkg/databases/seed_data/GL Groupping _ Mapping.xlsx"
		f, err = excelize.OpenFile(filePath)
		if err != nil {
			return fmt.Errorf("could not find GL Groupping _ Mapping.xlsx to seed database: %v", err)
		}
	}
	defer f.Close()

	sheetName := "Total Mapping"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to read rows from excel: %v", err)
	}

	if len(rows) < 3 {
		return fmt.Errorf("excel file does not have enough data (expected header at row 2)")
	}

	// Data begins at index 2 (Row 3)
	for i := 2; i < len(rows); i++ {
		row := rows[i]
		// Pad row to ensure we can access indexes safely
		paddedRow := make([]string, 7)
		copy(paddedRow, row)

		// Column mapping based on user confirmation:
		// [0] Group1, [1] Group2, [2] Group3, [3] Entity, [4] Entity GL, [5] Conso GL, [6] Account Name
		entity := strings.ToUpper(strings.TrimSpace(paddedRow[3]))
		entityGL := strings.TrimSpace(paddedRow[4])
		consoGL := strings.TrimSpace(paddedRow[5])
		accountName := strings.TrimSpace(paddedRow[6])

		if entity == "" || entityGL == "" || consoGL == "" {
			continue // skip incomplete rows
		}

		// Idempotent Check: Check if Entity + EntityGL already exists
		var existing models.GlMappingEntity
		err := db.Where("entity = ? AND entity_gl = ?", entity, entityGL).First(&existing).Error
		if err == nil {
			// Record exists, check if conso_gl or account_name changed (Optional Update)
			if existing.ConsoGL != consoGL || existing.AccountName != accountName {
				db.Model(&existing).Updates(models.GlMappingEntity{
					ConsoGL:     consoGL,
					AccountName: accountName,
				})
			}
			continue
		}

		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// Create new record
		newMapping := models.GlMappingEntity{
			Entity:      entity,
			EntityGL:    entityGL,
			ConsoGL:     consoGL,
			AccountName: accountName,
			IsActive:    true,
		}

		if err := db.Create(&newMapping).Error; err != nil {
			return err
		}
	}

	return nil
}

// SeedBudgetStructure reads GL Groupping _ Mapping.xlsx and populates the budget_structure_entities table.
func SeedBudgetStructure(db *gorm.DB) error {
	fmt.Println("Syncing Budget Structure (Filter Pane) from GL Groupping _ Mapping.xlsx...")

	filePath := "pkg/databases/seed_data/GL Groupping _ Mapping.xlsx"
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		filePath = "../pkg/databases/seed_data/GL Groupping _ Mapping.xlsx"
		f, err = excelize.OpenFile(filePath)
		if err != nil {
			return fmt.Errorf("could not find GL Groupping _ Mapping.xlsx for budget structure: %v", err)
		}
	}
	defer f.Close()

	sheetName := "Total Mapping"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to read rows: %v", err)
	}

	uniqueKeys := make(map[string]bool)

	// Data begins at index 2 (Row 3)
	for i := 2; i < len(rows); i++ {
		row := rows[i]
		paddedRow := make([]string, 7)
		copy(paddedRow, row)

		g1 := strings.TrimSpace(paddedRow[0])
		g2 := strings.TrimSpace(paddedRow[1])
		g3 := strings.TrimSpace(paddedRow[2])
		consoGL := strings.TrimSpace(paddedRow[5])
		accName := strings.TrimSpace(paddedRow[6])

		if g1 == "" && g2 == "" && g3 == "" {
			continue
		}

		// Create a unique key for deduplication from Excel itself first
		key := fmt.Sprintf("%s|%s|%s|%s", g1, g2, g3, consoGL)
		if uniqueKeys[key] {
			continue
		}
		uniqueKeys[key] = true

		// Check if exists in DB
		var existing models.BudgetStructureEntity
		err := db.Where("group1 = ? AND group2 = ? AND group3 = ? AND conso_gl = ?", g1, g2, g3, consoGL).First(&existing).Error
		if err == nil {
			// Already exists, maybe update name if changed?
			if existing.AccountName != accName {
				db.Model(&existing).Update("account_name", accName)
			}
			continue
		}

		// Insert new
		newEntry := models.BudgetStructureEntity{
			Group1:      g1,
			Group2:      g2,
			Group3:      g3,
			ConsoGL:     consoGL,
			AccountName: accName,
		}
		db.Create(&newEntry)
	}

	return nil
}

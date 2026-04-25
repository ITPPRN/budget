package seeders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

func findSeedFile() (string, error) {
	searchPaths := []string{"pkg/databases/seed_data", "../pkg/databases/seed_data"}
	for _, p := range searchPaths {
		files, err := os.ReadDir(p)
		if err != nil {
			continue
		}
		for _, f := range files {
			if !f.IsDir() && (filepath.Ext(f.Name()) == ".xlsx" || filepath.Ext(f.Name()) == ".xls") {
				return filepath.Join(p, f.Name()), nil
			}
		}
	}
	return "", fmt.Errorf("no Excel file found in seed_data directory")
}

// SeedGLGrouping reads any Excel in seed_data and populates the UNIFIED gl_grouping_entities table.
func SeedGLGrouping(db *gorm.DB) error {
	fmt.Println("Syncing UNIFIED GL Grouping...")

	filePath, err := findSeedFile()
	if err != nil {
		return fmt.Errorf("SeedGLGrouping: %v", err)
	}
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("SeedGLGrouping: failed to open %s: %v", filePath, err)
	}
	// defer f.Close()
	defer func() { _ = f.Close() }()
	sheetName := "Total Mapping"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("SeedGLGrouping: failed to read rows: %v", err)
	}

	for i := 2; i < len(rows); i++ {
		row := rows[i]
		paddedRow := make([]string, 7)
		copy(paddedRow, row)

		g1 := strings.TrimSpace(paddedRow[0])
		g2 := strings.TrimSpace(paddedRow[1])
		g3 := strings.TrimSpace(paddedRow[2])
		entity := strings.ToUpper(strings.TrimSpace(paddedRow[3]))
		entityGL := strings.TrimSpace(paddedRow[4])
		consoGL := strings.TrimSpace(paddedRow[5])
		accName := strings.TrimSpace(paddedRow[6])

		if entity == "" || entityGL == "" || consoGL == "" {
			continue
		}

		// Idempotent Check: Entity + EntityGL
		var groupings []models.GlGroupingEntity
		err := db.Where("entity = ? AND entity_gl = ?", entity, entityGL).Limit(1).Find(&groupings).Error
		if err == nil && len(groupings) > 0 {
			existing := groupings[0]
			// Update if any fields changed
			if existing.ConsoGL != consoGL || existing.AccountName != accName ||
				existing.Group1 != g1 || existing.Group2 != g2 || existing.Group3 != g3 {
				db.Model(&existing).Updates(models.GlGroupingEntity{
					ConsoGL:     consoGL,
					AccountName: accName,
					Group1:      g1,
					Group2:      g2,
					Group3:      g3,
				})
			}
			continue
		}

		// Insert new
		newEntry := models.GlGroupingEntity{
			Entity:      entity,
			EntityGL:    entityGL,
			ConsoGL:     consoGL,
			AccountName: accName,
			Group1:      g1,
			Group2:      g2,
			Group3:      g3,
		}
		db.Create(&newEntry)
	}

	return nil
}

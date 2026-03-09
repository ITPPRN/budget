package postgres

import (
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/databases/seeders"
	"p2p-back-end/pkg/utils"
)

func NewPostgresConnection(cfg *configs.Config) (*gorm.DB, error) {
	dsn, err := utils.UrlBuilder("postgres", cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Error),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		logs.Error("Failed to connect to database: ", zap.Error(err))
		return nil, err
	}

	// 1. สร้าง Extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	// 2. Auto Migrate (สร้าง Table อัตโนมัติ)
	err = db.AutoMigrate(
		// Auth & Base
		&models.UserEntity{},
		&models.DepartmentEntity{},
		&models.DepartmentMappingEntity{},
		&models.UserPermissionEntity{},

		// Budget & Capex (Flattened Type 2: Header + Detail)
		&models.FileBudgetEntity{},
		&models.FileCapexBudgetEntity{},
		&models.FileCapexActualEntity{},

		&models.BudgetFactEntity{},
		&models.BudgetAmountEntity{}, // New Detail Table

		&models.CapexBudgetFactEntity{},
		&models.CapexBudgetAmountEntity{}, // New Detail Table

		&models.CapexActualFactEntity{},
		&models.CapexActualAmountEntity{}, // New Detail Table

		// Actual (Operational / P2P) - Centralized
		&models.ActualFactEntity{},
		&models.ActualAmountEntity{}, // Detail Table

		// GL Mapping (Whitelisting & Consolidation)
		&models.GlMappingEntity{},

		// Budget Structure Hierarchy
		&models.BudgetStructureEntity{},

		// Centralized Transaction Table
		&models.ActualTransactionEntity{},
		&models.GeneralLedgerEntriesClik{},
	)

	if err != nil {
		logs.Error("Migration Error: ", zap.Error(err))
		return nil, err
	}

	// 3. Seed GL Mappings (Whitelist)
	if err := seeders.SeedGLMappings(db); err != nil {
		logs.Error("Failed to seed GL mappings: ", zap.Error(err))
		// Continue even if seeding fails (log error)
	}

	// 4. Seed Budget Structure
	if err := seeders.SeedBudgetStructure(db); err != nil {
		logs.Error("Failed to seed budget structure: ", zap.Error(err))
	}

	logs.Info("Database connected and migrated successfully with Base Practice 🐘")
	return db, nil
}

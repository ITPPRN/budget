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

func NewPostgresConnection(cfg *configs.Config, connType string) (*gorm.DB, error) {
	dsn, err := utils.UrlBuilder(connType, cfg)
	if err != nil {
		logs.Error("Failed to build DSN: ", zap.Error(err))
		return nil, err
	}

	logs.Info("🐘 GORM: Opening connection...", zap.String("type", connType))
	_ = logs.Sync()

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

	// 2. Auto Migrate (สร้าง Table อัตโนมัติ) - เฉพาะ Local Postgres
	if connType == "postgres" {
		modelsToMigrate := getModelsToMigrate()
		err = db.AutoMigrate(modelsToMigrate...)
		if err != nil {
			logs.Error("Migration Error: ", zap.Error(err))
			return nil, err
		}

		// 3. Seed GL Mappings (Whitelist)
		if err := seeders.SeedGLMappings(db); err != nil {
			logs.Error("Failed to seed GL mappings: ", zap.Error(err))
		}

		// 4. Seed Budget Structure
		if err := seeders.SeedBudgetStructure(db); err != nil {
			logs.Error("Failed to seed budget structure: ", zap.Error(err))
		}
	}

	logs.Info("Database connected successfully 🐘", zap.String("type", connType))
	return db, nil
}

func getModelsToMigrate() []interface{} {
	return []interface{}{
		// Auth & Base
		&models.UserEntity{},
		&models.DepartmentEntity{},
		&models.DepartmentMappingEntity{},
		&models.UserPermissionEntity{},

		// Master Data (Synced via RabbitMQ / Source)
		&models.Companies{},
		&models.Departments{},
		&models.Sections{},
		&models.Positions{},

		// Budget & Capex (Flattened Type 2: Header + Detail)
		&models.FileBudgetEntity{},
		&models.FileCapexBudgetEntity{},
		&models.FileCapexActualEntity{},

		&models.BudgetFactEntity{},
		&models.BudgetAmountEntity{},

		&models.CapexBudgetFactEntity{},
		&models.CapexBudgetAmountEntity{},

		&models.CapexActualFactEntity{},
		&models.CapexActualAmountEntity{},

		// Actual (Operational / P2P) - Centralized
		&models.ActualFactEntity{},
		&models.ActualAmountEntity{},

		// GL Mapping (Whitelisting & Consolidation)
		&models.GlMappingEntity{},

		// Budget Structure Hierarchy
		&models.BudgetStructureEntity{},

		// User Configuration (Personalized)
		&models.UserConfigEntity{},

		// Centralized Transaction Table
		&models.ActualTransactionEntity{},
		&models.AchHmwGleEntity{},
		&models.ClikGleEntity{},
	}
}

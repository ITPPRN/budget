package postgres

import (
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
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

		// Actual (Operational / P2P)
		&models.ActualFactEntity{},
		&models.ActualAmountEntity{}, // New Detail Table

		// Owner (Denormalized)
		&models.OwnerActualFactEntity{},
		&models.OwnerActualAmountEntity{}, // New Detail Table
	)

	if err != nil {
		logs.Error("Failed to migrate database: ", zap.Error(err))
		return nil, err
	}

	logs.Info("Database connected and migrated successfully with Base Practice 🐘")
	return db, nil
}

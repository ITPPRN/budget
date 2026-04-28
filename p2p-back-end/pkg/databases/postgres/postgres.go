package postgres

import (
	"time"

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

	// Connection pool settings เพื่อป้องกัน "driver: bad connection"
	sqlDB, err := db.DB()
	if err != nil {
		logs.Error("Failed to get underlying sql.DB: ", zap.Error(err))
		return nil, err
	}
	if connType == "central_postgres" {
		// DW connection: aggressive recycling. Used rarely (sync jobs only) and the
		// external DW server is more likely to drop idle connections silently than
		// our own DB, so we keep idle pool small and recycle frequently to avoid
		// reusing a stale connection that hangs on read.
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(10 * time.Minute)
		sqlDB.SetConnMaxIdleTime(2 * time.Minute)
	} else {
		// Local DB: serves the API + many concurrent dashboard/export queries,
		// so we keep a larger pool with longer lifetime for amortized connect cost.
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
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

		// --- Explicit Backfill: Set deleted = false for existing users ---
		if err := db.Exec("UPDATE user_entities SET deleted = false WHERE deleted IS NULL").Error; err != nil {
			logs.Warn("Failed to backfill deleted flag for users: ", zap.Error(err))
		}

		// --- Drop legacy single-column unique on company_branch_code_mappings.company_id
		//     so the new composite (company_id, branch_code) unique can take effect.
		//     Safe to run repeatedly — IF EXISTS guards.
		if err := db.Exec(`ALTER TABLE company_branch_code_mappings
			DROP CONSTRAINT IF EXISTS company_branch_code_mappings_company_id_key`).Error; err != nil {
			logs.Warn("Drop legacy company_id unique constraint failed: ", zap.Error(err))
		}
		if err := db.Exec(`DROP INDEX IF EXISTS idx_company_branch_code_mappings_company_id`).Error; err != nil {
			logs.Warn("Drop legacy company_id unique index failed: ", zap.Error(err))
		}

		// --- 🚀 FORCE REACTIVATE ADMIN (Don't remove until confirm) ---
		if err := db.Table("user_entities").Where("username = ?", "admin").Update("deleted", false).Error; err != nil {
			logs.Errorf("CRITICAL: Failed to reactivate admin user: %v", err)
		} else {
			logs.Info("✅ [STARTUP] Admin account has been force-reactivated.")
		}

		// 3. Seed GL Mappings (Unified)
		if err := seeders.SeedGLGrouping(db); err != nil {
			logs.Error("Failed to seed GL grouping: ", zap.Error(err))
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

		// Branch-scope mapping for BRANCH_DELEGATE role
		&models.CompanyBranchCodeMappingEntity{},

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

		// Budget Structure Hierarchy (Unified)
		&models.GlGroupingEntity{},

		// User Configuration (Personalized)
		&models.UserConfigEntity{},

		// Centralized Transaction Table
		&models.ActualTransactionEntity{},
		&models.AchHmwGleEntity{},
		&models.ClikGleEntity{},
		&models.DataInventoryEntity{},

		// Audit Logs (Owner Approval & Reporting)
		&models.AuditLogEntity{},
		&models.AuditLogRejectedItemEntity{},

		&models.AuditRejectBasket{},
	}
}

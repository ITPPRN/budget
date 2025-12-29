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
		Logger: logger.Default.LogMode(logger.Error),
        // Logger: logger.Default.LogMode(logger.Info), // เปิด Info เพื่อดู SQL ที่เกิดขึ้นจริง
		DisableForeignKeyConstraintWhenMigrating: true,
    })

    if err != nil {
        logs.Error("Failed to connect to database: ", zap.Error(err))
        return nil, err
    }

    // 1. สร้าง Extension
    db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

    // 2. สั่ง AutoMigrate ทั้งหมด (ตอนนี้จะผ่านฉลุยเพราะมันจะสร้างแค่ตาราง ไม่เช็ค FK)
    err = db.AutoMigrate(
        &models.DepartmentEntity{},
        &models.UserEntity{},
        &models.VendorEntity{},
        &models.ProductEntity{},
        &models.PurchaseRequestEntity{},
        &models.PrItemEntity{},
        &models.PurchaseOrderEntity{},
        &models.GoodsReceiptEntity{},
        &models.ApVoucherEntity{},
        &models.PaymentEntity{},
    )

    if err != nil {
        logs.Error("Critical: AutoMigrate failed: ", zap.Error(err))
        return nil, err
    }

    // 3. ✅ (Optional) ถ้าต้องการให้มี Foreign Key ใน Database จริงๆ 
    // หลังจาก Migrate ตารางเสร็จแล้ว ให้เปิดการสร้าง FK แล้วสั่ง Migrate ซ้ำอีกรอบ
    db.Config.DisableForeignKeyConstraintWhenMigrating = false
    db.AutoMigrate(&models.DepartmentEntity{}, &models.UserEntity{})

    logs.Info("Database connected and migrated successfully with Base Practice 🐘")
    return db, nil
}
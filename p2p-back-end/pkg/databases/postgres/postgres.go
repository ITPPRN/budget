// package databases

// import (
//
//
//
//

// 	"go.uber.org/zap"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"

// 	"p2p-back-end/configs"
// 	"p2p-back-end/logs"
// 	"p2p-back-end/pkg/utils"
// )

// func NewPostgresConnection(cfg *configs.Config) (*gorm.DB, error) {

// 	dsn, err := utils.UrlBuilder("postgres", cfg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	db, err := gorm.Open(postgres.New(postgres.Config{
// 		// DriverName:           cfg.Postgres.DriverName,
// 		DSN:                  dsn,
// 		PreferSimpleProtocol: true,
// 	}), &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Silent),
// 	})

//		i	f err != nil {
//				logs.Error("Failed to connect to database: ", zap.Error(err))
//				return nil, err
//			}
//			logs.Info("postgreSQL database has been connected 🐘")
//			return db, nil
//	}
package databases

import (
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/pkg/utils"

	// ✅ 1. อย่าลืม Import models เข้ามาด้ยนะครับ
	"p2p-back-end/modules/entities/models"
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
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		logs.Error("Failed to connect to database: ", zap.Error(err))
		return nil, err
	}

	// -------------------------------------------------------------
	// ✅ 2. เริ่มต้นโซน Migration (ใส่หลังจากต่อ DB ติดแล้ว)
	// -------------------------------------------------------------

	// 2.1 เปิดใช้งาน UUID Extension (จำเป็นสำหรับ uuid_generate_v4())
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		logs.Error("Failed to create uuid extension: ", zap.Error(err))
		return nil, err
	}

	// 2.2 สั่งสร้างตารางอัตโนมัติ (ใส่ Struct ให้ครบทุกอันที่มี)
	err = db.AutoMigrate(
		// Group 1: Organization
		&models.DepartmentEntity{},
		&models.UserEntity{},

		// Group 2: Master Data
		&models.VendorEntity{},
		&models.ProductEntity{},

		// Group 3: Purchasing
		&models.PurchaseRequestEntity{},
		&models.PrItemEntity{},
		&models.PurchaseOrderEntity{},
		&models.GoodsReceiptEntity{},

		// Group 4: Finance
		&models.ApVoucherEntity{},
		&models.PaymentEntity{},
	)

	if err != nil {
		logs.Error("Failed to AutoMigrate: ", zap.Error(err))
		return nil, err
	}

	// -------------------------------------------------------------

	logs.Info("postgreSQL database has been connected 🐘")
	return db, nil
}

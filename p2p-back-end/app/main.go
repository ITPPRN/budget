package main

import (
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/modules/servers"
	databases "p2p-back-end/pkg/databases/postgres"
	redis "p2p-back-end/pkg/databases/redis"
	keycloak "p2p-back-end/pkg/keycloak"
	"p2p-back-end/pkg/middlewares"
	rabbitmq "p2p-back-end/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func init() {
	initTimeZone()
}

// @title P2P Back-End API
// @version 1.0
// @description This is the API documentation for the P2P application back-end.
// @host localhost:8080
// @BasePath /v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	logs.Loginit()

	_ = godotenv.Load("../.env")

	cfg := new(configs.Config)
	configs.LoadConfigs(cfg)

	// Test logs
	logs.Info("Application is starting...", zap.String("mode", cfg.App.Mode))

	// Internal Security Setup
	if cfg.App.InternalSecret == "" {
		logs.Warn("Warning: INTERNAL_SECRET is missing. Internal auth will fail.")
	}
	middlewares.InitInternalSecret(cfg.App.InternalSecret)
	logs.Info("✅ Internal Security initialized.")
	_ = logs.Sync()

	// Keycloak Validator Setup (JWKS)
	middlewares.InitKeycloakValidator(
		cfg.KeyCloak.Host,
		cfg.KeyCloak.Port,
		cfg.KeyCloak.RealmName,
		cfg.KeyCloak.ClientID,
	)
	logs.Info("✅ Keycloak Validator initialized.")
	_ = logs.Sync() // Force flush logs

	// Local Database Setup
	logs.Info("🐘 Starting Local Database Setup...")
	_ = logs.Sync()

	db, err := databases.NewPostgresConnection(cfg, "postgres")
	if err != nil {
		logs.Fatalf("❌ Failed to connect to local database: %v", err)
	}
	logs.Info("✅ Local Database connected and ready.")
	_ = logs.Sync()

	// Central Database Setup (Optional)
	var db2 *gorm.DB
	if cfg.Postgres2.Host != "" {
		logs.Info("🐘 Starting Central Database Setup...")
		_ = logs.Sync()
		db2, err = databases.NewPostgresConnection(cfg, "central_postgres")
		if err != nil {
			logs.Warn("Failed to connect to central database: " + err.Error())
		}
	}

	redisClient := redis.NewRedisClient(cfg)
	keycloakClient := keycloak.NewKeyCloakClient(cfg)

	// RabbitMQ Setup (Optional)
	var mqConn *amqp.Connection
	var mqCh *amqp.Channel
	if cfg.RabbitMQ.Host != "" {
		mqConn, err = rabbitmq.NewRabbitMQConnection(cfg)
		if err != nil {
			logs.Warn("RabbitMQ Connection Warning: " + err.Error())
		} else {
			defer func() {
				if err := mqConn.Close(); err != nil {
					logs.Error("Failed to close RabbitMQ connection", zap.Error(err))
				}
			}()

			mqCh, err = rabbitmq.NewRabbitMQChannel(mqConn)
			if err != nil {
				logs.Warn("RabbitMQ Channel Warning: " + err.Error())
			} else {
				defer func() {
					if err := mqCh.Close(); err != nil {
						logs.Error("Failed to close RabbitMQ channel", zap.Error(err))
					}
				}()
			}
		}
	} else {
		logs.Warn("RabbitMQ is not configured. Messaging features will be disabled.")
	}

	// Start Server
	server := servers.NewServer(cfg, db, db2, redisClient, keycloakClient, mqConn, mqCh)

	server.Start()
}

func initTimeZone() {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	time.Local = ict
}

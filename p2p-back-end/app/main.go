package main

import (
	"time"

	"github.com/joho/godotenv"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/modules/servers"
	databases "p2p-back-end/pkg/databases/postgres"
	redis "p2p-back-end/pkg/databases/redis"
	keycloak "p2p-back-end/pkg/keycloak"
	"p2p-back-end/pkg/middlewares"
)

func init() {
	initTimeZone()
	logs.Loginit()

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
	if err := godotenv.Load("../.env"); err != nil {
		logs.Error(err)
	}

	cfg := new(configs.Config)

	configs.LoadConfigs(cfg)

	middlewares.InitKeycloakValidator(
		cfg.KeyCloak.Host,
		cfg.KeyCloak.Port,
		cfg.KeyCloak.RealmName,
		cfg.KeyCloak.ClientID,
	)

	db, err := databases.NewPostgresConnection(cfg)
	if err != nil {
		logs.Error(err.Error())
	}

	redis := redis.NewRedisClient(cfg)

	keycloak := keycloak.NewKeyCloakClient(cfg)

	server := servers.NewServer(cfg, db, redis, keycloak)

	server.Start()
}

func initTimeZone() {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	time.Local = ict
}

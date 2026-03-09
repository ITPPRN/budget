package configs

import (
	"context"
	"fmt"
	"os"

	infisical "github.com/infisical/go-sdk"

	"p2p-back-end/logs"
)

func LoadConfigs(cfg *Config) {
	infisicalURL := os.Getenv("INFISICAL_URL")
	clientID := os.Getenv("INFISICAL_CLIENT_ID")
	clientSecret := os.Getenv("INFISICAL_CLIENT_SECRET")
	projectID := os.Getenv("INFISICAL_PROJECT_ID")

	client := infisical.NewInfisicalClient(
		context.Background(),
		infisical.Config{
			SiteUrl: infisicalURL, // ตัวเลือก, ค่าเริ่มต้นคือ https://app.infisical.com
		})

	_, err := client.Auth().UniversalAuthLogin(clientID, clientSecret)
	if err != nil {
		fmt.Printf("การยืนยันตัวตนล้มเหลว: %v", err)
		os.Exit(1)
	}

	apiKeySecrets, err := client.Secrets().List(
		infisical.ListSecretsOptions{
			ProjectID:   projectID,
			SecretPath:  "/backend",
			Environment: "dev",
		},
	)
	if err != nil {
		logs.Error(err)
	}

	setData := func(key CfgKey) string {
		for _, secret := range apiKeySecrets {
			if secret.SecretKey == string(key) {
				return secret.SecretValue
			}
		}
		logs.Error(fmt.Sprintf("ไม่พบคีย์ %s ใน Secrets ของ Infisical", key))
		return ""
	}

	// setJWTPublicKEY := func() {
	// 	cfg.KeyCloak.PublicKey = setData(PublicKey)
	// 	key := string(PublicKey)
	// 	err := os.Setenv(key, cfg.KeyCloak.PublicKey)
	// 	if err != nil {
	// 		fmt.Println("Error setting environment variable:", err)
	// 		return
	// 	}
	// }
	// setJWTPublicKEY()

	// การตั้งค่าสำหรับแอปพลิเคชัน
	cfg.App.Port = setData(FiberPort)
	cfg.App.Mode = setData(FiberMode)

	// การตั้งค่าสำหรับ PostgreSQL
	cfg.Postgres.Host = setData(PostgresHost)
	cfg.Postgres.Port = setData(PostgresPort)
	cfg.Postgres.Username = setData(PostgresUsername)
	cfg.Postgres.Password = setData(PostgresPassword)
	cfg.Postgres.DatabaseName = setData(PostgresDatabase)
	cfg.Postgres.Schema = setData(PostgresSchema)
	cfg.Postgres.SslMode = setData(PostgresSslMode)

	cfg.Redis.Host = setData(RedisHost)
	cfg.Redis.Port = setData(RedisPort)
	cfg.Redis.Password = setData(RedisPassword)

	cfg.KeyCloak.Host = setData(KeyCloakHost)
	cfg.KeyCloak.Port = setData(KeyCloakPort)
	cfg.KeyCloak.AdminUsername = setData(AdminUsername)
	cfg.KeyCloak.AdminPassword = setData(AdminPassword)
	cfg.KeyCloak.ClientID = setData(ClientID)
	cfg.KeyCloak.ClientSecret = setData(ClientSecret)
	cfg.KeyCloak.RealmName = setData(RealmName)

	cfg.RabbitMQ.URL = setData(RabbitMQURL)

	printLog(cfg)
}

func printLog(cfg *Config) {
	fields := map[CfgKey]interface{}{
		FiberPort:        cfg.App.Port,
		FiberMode:        cfg.App.Mode,
		RedisHost:        cfg.Redis.Host,
		RedisPort:        cfg.Redis.Port,
		RedisPassword:    cfg.Redis.Password,
		PostgresHost:     cfg.Postgres.Host,
		PostgresPort:     cfg.Postgres.Port,
		PostgresUsername: cfg.Postgres.Username,
		PostgresPassword: cfg.Postgres.Password,
		PostgresDatabase: cfg.Postgres.DatabaseName,
		PostgresSslMode:  cfg.Postgres.SslMode,
		PostgresSchema:   cfg.Postgres.Schema,
		KeyCloakHost:     cfg.KeyCloak.Host,
		KeyCloakPort:     cfg.KeyCloak.Port,
		ClientID:         cfg.KeyCloak.ClientID,
		ClientSecret:     cfg.KeyCloak.ClientSecret,
		RealmName:        cfg.KeyCloak.RealmName,
		AdminUsername:    cfg.KeyCloak.AdminUsername,
		AdminPassword:    cfg.KeyCloak.AdminPassword,
		RabbitMQURL:      cfg.RabbitMQ.URL,
	}

	for key, value := range fields {
		logs.Debugf("%s: %v", key, value)
	}

}

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

	infisicalEnv := os.Getenv("INFISICAL_ENV")
	if infisicalEnv == "" {
		infisicalEnv = "dev"
	}

	var apiKeySecrets []infisical.Secret
	var err error

	// 1. พยายามเชื่อมต่อ Infisical (เฉพาะเมื่อมี Credentials ครบ)
	if clientID != "" && clientSecret != "" && projectID != "" {
		client := infisical.NewInfisicalClient(context.Background(), infisical.Config{
			SiteUrl: infisicalURL,
		})

		_, err = client.Auth().UniversalAuthLogin(clientID, clientSecret)
		if err != nil {
			logs.Warn(fmt.Sprintf("Infisical Auth Failed: %v. Falling back to System Env.", err))
		} else {
			apiKeySecrets, err = client.Secrets().List(infisical.ListSecretsOptions{
				ProjectID:   projectID,
				SecretPath:  "/backend",
				Environment: infisicalEnv,
			})
			if err != nil {
				logs.Warn(fmt.Sprintf("Could not list secrets from Infisical: %v", err))
			}
		}
	} else {
		logs.Info("Infisical credentials not found. Using System Environment Variables only.")
	}

	setData := func(key CfgKey) string {
		// ชั้นที่ 1: หาใน Infisical
		for _, secret := range apiKeySecrets {
			if secret.SecretKey == string(key) {
				return secret.SecretValue
			}
		}

		// ชั้นที่ 2: ถ้าไม่เจอ หรือ Infisical ใช้ไม่ได้ ให้หาใน Environment Variable (จาก .env หรือ Docker -e)
		val := os.Getenv(string(key))
		if val != "" {
			return val
		}

		logs.Warn(fmt.Sprintf("Key [%s] not found in Infisical or System Env", key))
		return ""
	}

	// การตั้งค่าสำหรับแอปพลิเคชัน
	cfg.App.Port = setData(FiberPort)
	cfg.App.Mode = setData(FiberMode)
	cfg.App.InternalSecret = setData(InternalSecret)
	cfg.App.GatewaySecret = setData(GatewaySecret)

	// การตั้งค่าสำหรับ PostgreSQL (Database หลัก)
	cfg.Postgres.Host = setData(PostgresHost)
	cfg.Postgres.Port = setData(PostgresPort)
	cfg.Postgres.Username = setData(PostgresUsername)
	cfg.Postgres.Password = setData(PostgresPassword)
	cfg.Postgres.DatabaseName = setData(PostgresDatabase)
	cfg.Postgres.Schema = setData(PostgresSchema)
	cfg.Postgres.SslMode = setData(PostgresSslMode)

	// การตั้งค่าสำหรับ Redis
	cfg.Redis.Host = setData(RedisHost)
	cfg.Redis.Port = setData(RedisPort)
	cfg.Redis.Password = setData(RedisPassword)

	// การตั้งค่าสำหรับ Keycloak
	cfg.KeyCloak.Host = setData(KeyCloakHost)
	cfg.KeyCloak.Port = setData(KeyCloakPort)
	cfg.KeyCloak.RealmName = setData(RealmName)
	cfg.KeyCloak.ClientID = setData(ClientID)
	cfg.KeyCloak.ClientSecret = setData(ClientSecret)
	cfg.KeyCloak.AdminUsername = setData(AdminUsername)
	cfg.KeyCloak.AdminPassword = setData(AdminPassword)

	// การตั้งค่าสำหรับ RabbitMQ
	cfg.RabbitMQ.Host = setData(RabbitMqHost)
	cfg.RabbitMQ.Port = setData(RabbitMqPort)
	cfg.RabbitMQ.Username = setData(RabbitMqUsername)
	cfg.RabbitMQ.Password = setData(RabbitMqPassword)
	cfg.RabbitMQ.VHost = setData(RabbitMqVHost)

	// การตั้งค่าสำหรับ PostgreSQL (DW Source)
	cfg.Postgres2.Host = setData(Postgres2Host)
	cfg.Postgres2.Port = setData(Postgres2Port)
	cfg.Postgres2.Username = setData(Postgres2Username)
	cfg.Postgres2.Password = setData(Postgres2Password)
	cfg.Postgres2.DatabaseName = setData(Postgres2Database)
	cfg.Postgres2.Schema = setData(Postgres2Schema)
	cfg.Postgres2.SslMode = setData(Postgres2SslMode)

	printLog(cfg)
}

func printLog(cfg *Config) {
	fields := map[CfgKey]interface{}{
		FiberPort:        cfg.App.Port,
		FiberMode:        cfg.App.Mode,
		RedisHost:        cfg.Redis.Host,
		RedisPort:        cfg.Redis.Port,
		InternalSecret:   cfg.App.InternalSecret,
		PostgresHost:     cfg.Postgres.Host,
		PostgresPort:     cfg.Postgres.Port,
		PostgresUsername: cfg.Postgres.Username,
		PostgresDatabase: cfg.Postgres.DatabaseName,
		KeyCloakHost:     cfg.KeyCloak.Host,
		KeyCloakPort:     cfg.KeyCloak.Port,
		RealmName:        cfg.KeyCloak.RealmName,
		RabbitMqHost:     cfg.RabbitMQ.Host,
		RabbitMqPort:     cfg.RabbitMQ.Port,
		RabbitMqUsername: cfg.RabbitMQ.Username,
		RabbitMqVHost:    cfg.RabbitMQ.VHost,

		// Source DW
		Postgres2Host:     cfg.Postgres2.Host,
		Postgres2Database: cfg.Postgres2.DatabaseName,
	}

	for key, value := range fields {
		logs.Debugf("%s: %v", key, value)
	}
}

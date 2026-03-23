package utils

import (
	"errors"
	"fmt"

	"p2p-back-end/configs"
)

func UrlBuilder(urlType string, cfg *configs.Config) (string, error) {

	var url string

	switch urlType {
	case "fiber":
		url = fmt.Sprintf(":%s", cfg.App.Port)
	case "postgres":
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d TimeZone=Asia/Bangkok",
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.Username,
			cfg.Postgres.Password,
			cfg.Postgres.DatabaseName,
			cfg.Postgres.SslMode,
			10, // 10 seconds timeout
		)
	case "central_postgres":
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d TimeZone=Asia/Bangkok",
			cfg.Postgres2.Host,
			cfg.Postgres2.Port,
			cfg.Postgres2.Username,
			cfg.Postgres2.Password,
			cfg.Postgres2.DatabaseName,
			cfg.Postgres2.SslMode,
			20, // Higher timeout for DW
		)

	case "redis":
		url = fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	case "keycloak":
		url = fmt.Sprintf("http://%s:%s", cfg.KeyCloak.Host, cfg.KeyCloak.Port)
	case "rabbitmq":
		vhost := cfg.RabbitMQ.VHost
		if vhost == "" {
			vhost = "/" // fallback เป็น default ถ้าไม่ได้ตั้ง
		}
		url = fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
			cfg.RabbitMQ.Username,
			cfg.RabbitMQ.Password,
			cfg.RabbitMQ.Host,
			cfg.RabbitMQ.Port,
			vhost,
		)
	default:
		err := fmt.Sprintf("error,url builder Unknown url type: %s", urlType)
		return "", errors.New(err)
	}
	return url, nil
}

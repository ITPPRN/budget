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
		// keepalives_*: kernel probes after 30s idle, every 10s, fails after 3 misses (~1 min)
		//   so a half-open TCP connection (network blip, DW server crash, NAT timeout) is
		//   detected in ≤1 minute instead of waiting for the kernel default of ~2 hours.
		// options statement_timeout: server-side cap on a single SQL statement (10 min for
		//   local — covers heavy aggregation/migration queries).
		// idle_in_transaction_session_timeout: prevents a stuck transaction from holding
		//   locks forever (10 min).
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d TimeZone=Asia/Bangkok "+
			"keepalives=1 keepalives_idle=30 keepalives_interval=10 keepalives_count=3 "+
			"options='-c statement_timeout=600000 -c idle_in_transaction_session_timeout=600000'",
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.Username,
			cfg.Postgres.Password,
			cfg.Postgres.DatabaseName,
			cfg.Postgres.SslMode,
			10, // 10 seconds timeout
		)
	case "central_postgres":
		// DW connection: tighter timeouts than local because DW reads are paginated
		// (LIMIT 5000 batches) — any single statement taking >5 min means the DW server
		// is unhealthy, fail fast instead of hanging the worker.
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d TimeZone=Asia/Bangkok "+
			"keepalives=1 keepalives_idle=30 keepalives_interval=10 keepalives_count=3 "+
			"options='-c statement_timeout=300000 -c idle_in_transaction_session_timeout=600000'",
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

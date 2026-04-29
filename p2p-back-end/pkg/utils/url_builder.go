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
		// GUC names go in DSN directly — pgx driver forwards unrecognized DSN keys
		// as RuntimeParams during the startup packet, which Postgres applies as
		// `SET name = value` for the new session. (libpq's `keepalives*` and the
		// `options='-c ...'` wrapper are NOT supported here — pgx forwards them to
		// the server which rejects them as unknown GUCs.)
		//
		// statement_timeout: server-side cap on one SQL statement (10 min for local).
		//   Heavy aggregation/migration queries can legitimately run longer than DW
		//   reads, so the local cap is more generous.
		// idle_in_transaction_session_timeout: kills sessions that hold a transaction
		//   open without progress for 10 min — prevents leaked transactions from
		//   blocking VACUUM and holding row locks forever.
		//
		// Note: TCP keepalive is handled by pgx's default *net.Dialer (KeepAlive=5m)
		// plus our SetConnMaxLifetime/IdleTime in postgres.go — combined with the
		// per-batch context timeouts in repository.go this is sufficient to avoid
		// silent hangs on half-open connections.
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d "+
			"TimeZone=Asia/Bangkok statement_timeout=600000 idle_in_transaction_session_timeout=600000",
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.Username,
			cfg.Postgres.Password,
			cfg.Postgres.DatabaseName,
			cfg.Postgres.SslMode,
			10, // 10 seconds timeout
		)
	case "central_postgres":
		// DW connection: tighter statement_timeout than local because DW reads are
		// paginated (LIMIT 5000 batches) — any single statement taking >5 min means
		// the DW server is unhealthy, fail fast instead of hanging the worker.
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d "+
			"TimeZone=Asia/Bangkok statement_timeout=300000 idle_in_transaction_session_timeout=600000",
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

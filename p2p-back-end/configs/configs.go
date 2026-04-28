package configs

type CfgKey string

const (
	GatewaySecret    CfgKey = "GATEWAY_SECRET"
	InternalSecret   CfgKey = "INTERNAL_SECRET"
	FiberPort        CfgKey = "APP_PORT"
	FiberMode        CfgKey = "APP_MODE"
	RedisHost        CfgKey = "REDIS_HOST"
	RedisPort        CfgKey = "REDIS_PORT"
	RedisPassword    CfgKey = "REDIS_PASSWORD"
	PostgresHost     CfgKey = "DB_HOST"
	PostgresPort     CfgKey = "DB_PORT"
	PostgresUsername CfgKey = "DB_USER"
	PostgresPassword CfgKey = "DB_PASSWORD"
	PostgresDatabase CfgKey = "DB_NAME"
	PostgresSchema   CfgKey = "DB_SCHEMA"
	PostgresSslMode  CfgKey = "DB_SSLMODE"
	KeyCloakHost     CfgKey = "KC_HOST"
	KeyCloakPort     CfgKey = "KC_PORT"
	ClientID         CfgKey = "KC_CLIENT_ID"
	ClientSecret     CfgKey = "KC_CLIENT_SECRET"
	RealmName        CfgKey = "KC_REALM_NAME"
	AdminUsername    CfgKey = "KC_ADMIN_USER"
	AdminPassword    CfgKey = "KC_ADMIN_PASS"
	RabbitMqHost     CfgKey = "RQ_HOST"
	RabbitMqPort     CfgKey = "RQ_PORT"
	RabbitMqUsername CfgKey = "RQ_USER"
	RabbitMqPassword CfgKey = "RQ_PASS"
	RabbitMqVHost    CfgKey = "RQ_VHOST"

	// Secondary DB (Data Warehouse)
	Postgres2Host     CfgKey = "DB2_HOST"
	Postgres2Port     CfgKey = "DB2_PORT"
	Postgres2Username CfgKey = "DB2_USER"
	Postgres2Password CfgKey = "DB2_PASSWORD"
	Postgres2Database CfgKey = "DB2_NAME"
	Postgres2Schema   CfgKey = "DB2_SCHEMA"
	Postgres2SslMode  CfgKey = "DB2_SSLMODE"

	// Sync tunables (DW pull)
	DwPerMonthTimeoutMinutes CfgKey = "DW_PER_MONTH_TIMEOUT_MINUTES"
)

type Config struct {
	App       Fiber
	Postgres  PostgresSql
	Postgres2 PostgresSql
	Redis     Redis
	KeyCloak  KeyCloak
	RabbitMQ  RabbitMQ
	Sync      Sync
}

// Sync holds runtime tunables for the DW + actual-fact sync pipeline.
type Sync struct {
	// DWPerMonthTimeoutMinutes — hard wall-clock cap for one (year, month) DW pull.
	// Heavy months (~14M rows) sometimes overrun the previous 30-min default.
	// Defaults to 90 if unset/invalid.
	DWPerMonthTimeoutMinutes int
}

type Fiber struct {
	Port           string
	Mode           string
	InternalSecret string
	GatewaySecret  string
}

type RabbitMQ struct {
	Host     string
	Port     string
	Username string
	Password string `json:"-"`
	VHost    string
	URL      string // Keep for compatibility if needed, though senior uses granular
}

type PostgresSql struct {
	Host         string
	Port         string
	Username     string
	Password     string `json:"-"`
	DatabaseName string
	SslMode      string
	Schema       string
}

type Redis struct {
	Host     string
	Port     string
	Password string `json:"-"`
}

type KeyCloak struct {
	Host          string
	Port          string
	RealmName     string
	ClientID      string
	ClientSecret  string // #nosec G117 -- internal config struct, not exposed via API
	AdminUsername string
	AdminPassword string
	PublicKey     string // Used for manual key management if needed
}



package configs

type CfgKey string

const (
	JWTKey           CfgKey = "JWT_SECRET_KEY"
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
	RabbitMQURL      CfgKey = "RABBITMQ_URL"
)

type Config struct {
	App       Fiber
	Postgres  PostgresSql
	Postgres2 PostgresSql
	Redis     Redis
	KeyCloak  KeyCloak
	RabbitMQ  RabbitMQ
}

type Fiber struct {
	Port string
	Mode string
}

type PostgresSql struct {
	Host         string
	Port         string
	Username     string
	Password     string
	DatabaseName string
	SslMode      string
	Schema       string
}

type Redis struct {
	Host     string
	Port     string
	Password string
}

type KeyCloak struct {
	Host          string
	Port          string
	ClientID      string
	ClientSecret  string
	RealmName     string
	AdminUsername string
	AdminPassword string
	// PublicKey     string
}

type RabbitMQ struct {
	URL string
}

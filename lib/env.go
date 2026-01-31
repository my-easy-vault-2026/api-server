package lib

import (
	"time"

	"github.com/spf13/viper"
)

type Env struct {
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	ServerPort  string `mapstructure:"SERVER_PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`
	GoNode      string `mapstructure:"GO_NODE"`

	DBUsername string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASS"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBName     string `mapstructure:"DB_NAME"`
	DBType     string `mapstructure:"DB_TYPE"`

	// 從庫配置 (支持多個)
	DBSlaveHosts     []string `mapstructure:"DB_SLAVE_HOSTS"`     // 逗號分隔
	DBSlavePorts     []string `mapstructure:"DB_SLAVE_PORTS"`     // 逗號分隔 (可選)
	DBSlaveUsernames []string `mapstructure:"DB_SLAVE_USERNAMES"` // 逗號分隔 (可選)
	DBSlavePasswords []string `mapstructure:"DB_SLAVE_PASSWORDS"` // 逗號分隔 (可選)

	// 連接池配置
	DBMaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	DBMaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	DBConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
	DBConnMaxIdleTime time.Duration `mapstructure:"DB_CONN_MAX_IDLE_TIME"`

	RedisAddr string `mapstructure:"REDIS_ADDR"`
	RedisPwd  string `mapstructure:"REDIS_PWD"`
	RedisDb   int    `mapstructure:"REDIS_DB"`
	RedisPool int    `mapstructure:"REDIS_POOL"`
	RedisTls  string `mapstructure:"REDIS_TLS"`

	MqList   []string `mapstructure:"MQ_LIST"`
	MqPubsub []string `mapstructure:"MQ_PUBSUB"`

	L2CacheExpire time.Duration `mapstructure:"L2_CACHE_EXPIRE"`

	WorkerPoolSize  int           `mapstructure:"WORKER_POOL_SIZE"`
	BucketSize      int           `mapstructure:"BUCKET_SIZE"`
	WriteWait       time.Duration `mapstructure:"WRITE_WAIT"`
	PongWait        time.Duration `mapstructure:"PONG_WAIT"`
	PingPeriod      time.Duration `mapstructure:"PING_PERIOD"`
	MaxMessageSize  int64         `mapstructure:"MAX_MESSAGE_SIZE"`
	ReadBufferSize  int           `mapstructure:"READ_BUFFER_SIZE"`
	WriteBufferSize int           `mapstructure:"WRITE_BUFFER_SIZE"`
	BroadcastSize   int64         `mapstructure:"BROADCAST_SIZE"`

	SaltLength          int           `mapstructure:"SALT_LENGTH"`
	TokenExpireTime     time.Duration `mapstructure:"TOKEN_EXPIRE_TIME"`
	LoginDataExpireTime time.Duration `mapstructure:"LOGIN_DATA_EXPIRE_TIME"`

	MailClientID     string `mapstructure:"MAIL_CLIENT_ID"`
	MailClientSecret string `mapstructure:"MAIL_CLIENT_SECRET"`
	MailTokenType    string `mapstructure:"MAIL_TOKEN_TYPE"`

	SentryDSN          string `mapstructure:"SENTRY_DSN"`
	MaxMultipartMemory int64  `mapstructure:"MAX_MULTIPART_MEMORY"`
	StorageBucketName  string `mapstructure:"STORAGE_BUCKET_NAME"`

	TimeZone      string `mapstructure:"TIMEZONE"`
	AdminEmail    string `mapstructure:"ADMIN_EMAIL"`
	AdminPassword string `mapstructure:"ADMIN_PASSWORD"`
}

var globalEnv = Env{
	MaxMultipartMemory: 10 << 20, // 10 MB
}

func GetEnv() Env {
	return globalEnv
}

func NewEnv(logger Logger) *Env {
	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal("cannot read cofiguration", err)
	}

	viper.SetDefault("TIMEZONE", "UTC")

	err = viper.Unmarshal(&globalEnv)
	if err != nil {
		logger.Fatal("environment cant be loaded: ", err)
	}

	return &globalEnv
}

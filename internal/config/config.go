package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"time"
)

const EnvPath = "./config/.env"

type Config struct {
	LogLevel    string `envconfig:"LOG_LEVEL" required:"true"`
	GrpcCfg     GrpcCfg
	PostgresCfg PostgresCfg
	Secret      string `envconfig:"SECRET" required:"true"`
}

type GrpcCfg struct {
	Port    int           `envconfig:"GRPC_PORT" required:"true"`
	Timeout time.Duration `envconfig:"GRPC_TIMEOUT" required:"true"`
}

type PostgresCfg struct {
	Host     string `envconfig:"POSTGRES_HOST" required:"true"`
	Port     int    `envconfig:"POSTGRES_PORT" required:"true"`
	Name     string `envconfig:"POSTGRES_NAME" required:"true"`
	User     string `envconfig:"POSTGRES_USER" required:"true"`
	Password string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	SSLMode  string `envconfig:"POSTGRES_SSL_MODE" required:"true"`

	PoolMaxConns        int           `envconfig:"POSTGRES_POOL_MAX_CONNS" default:"5"`
	PoolMaxConnLifetime time.Duration `envconfig:"POSTGRES_POOL_MAX_CONN_LIFETIME" default:"180s"`
	PoolMaxConnIdleTime time.Duration `envconfig:"POSTGRES_POOL_MAX_CONN_IDLE_TIME" default:"100s"`

	AutoMigrate bool `envconfig:"POSTGRES_AUTO_MIGRATE" default:"false"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(EnvPath); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}
	return &cfg, nil
}

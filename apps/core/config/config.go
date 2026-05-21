package config

import (
	"time"
)

var (
	AppEnvLocal       = "local"
	AppEnvDevelopment = "dev"
	AppEnvProduction  = "prod"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"dev"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Database   Database   `yaml:"db"`
	Nats       Nats       `yaml:"nats"`
	Temporal   Temporal   `yaml:"temporal"`
	JWT        JWT        `yaml:"jwt"`
}

type JWT struct {
	Secret     string        `yaml:"secret" env:"JWT_SECRET" env-required:"true"`
	AccessTTL  time.Duration `yaml:"access_ttl" env-default:"15m"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env-default:"720h"`
}

type HTTPServer struct {
	Host        string        `yaml:"host" env-default:"localhost"`
	Port        string        `yaml:"port" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"30s"`
}

type Database struct {
	URL string `yaml:"url" env-default:"postgres://root:root@localhost:5432/postgres?sslmode=disable"`
}

type Nats struct {
	URL                      string `yaml:"url" env-default:"tls://localhost:4222"`
	CAFile                   string `yaml:"ca_file"`
	CredentialsFile          string `yaml:"credentials_file"`
	AccountPublicKey         string `yaml:"account_public_key"`
	AccountJWTFile           string `yaml:"account_jwt_file"`
	AccountSeedEnv           string `yaml:"account_seed_env" env-default:"NATS_ACCOUNT_SEED"`
	AccountSeedFile          string `yaml:"account_seed_file"`
	OperatorSeedFile         string `yaml:"operator_seed_file"`
	SystemCredentialsFile    string `yaml:"system_credentials_file"`
	WorkerCredentialsEnabled bool   `yaml:"worker_credentials_enabled" env-default:"true"`
}

type Temporal struct {
	HostPort  string `yaml:"host_port" env-default:"localhost:7233"`
	Namespace string `yaml:"namespace" env-default:"default"`
	TaskQueue string `yaml:"task_queue" env-default:"cactus-core"`
}

// NatsURL реализует интерфейс bus.NatsConfig.
func (c *Config) NatsURL() string {
	return c.Nats.URL
}

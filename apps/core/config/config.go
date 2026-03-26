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
	URL string `yaml:"url" env-default:"nats://localhost:4222"`
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


package config

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
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
}

// NatsURL реализует интерфейс bus.NatsConfig.
func (c *Config) NatsURL() string {
	return c.Nats.URL
}

func MustLoad(configFileName string) *Config {

	configDir := os.Getenv("CONFIG_DIR")

	if configDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get working directory: %v", err)
		}
		configDir = wd
	}

	fullPath := filepath.Join(configDir, configFileName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", fullPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(fullPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

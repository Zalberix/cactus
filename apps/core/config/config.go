package config

import (
	"log"
	"os"
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

func MustLoad(cnfPath *string) *Config {
	configPath := "./configs/apps/core.yaml"
	if cnfPath != nil {
		configPath = *cnfPath
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

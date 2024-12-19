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
	Minio      Minio      `yaml:"minio"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"30s"`
}

type Database struct {
	Host string `yaml:"host" env-default:"db"`
	Port string `yaml:"port" env-default:"5432"`
	Name string `yaml:"name" env-default:"cactus"`
	User string `yaml:"user" env-default:"root"`
	Pass string `yaml:"pass" env-default:"root"`
}

type Minio struct {
	Endpoint   string `yaml:"endpoint" env-default:"local.work.ru:9000"`
	PublicKey  string `yaml:"publicKey"`
	PrivateKey string `yaml:"privateKey"`
	UseSSL     bool   `yaml:"useSSL" env-default:"false"`
}

func MustLoad() *Config {

	// TODO переделать на переменную среды так как нужно будет менять его при переезде на продакшен
	configPath := "./config/dev.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

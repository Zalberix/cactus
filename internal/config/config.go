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
	Redis      Redis      `yaml:"redis"`
}

type HTTPServer struct {
	Host        string        `yaml:"host" env-default:"localhost"`
	Port        string        `yaml:"port" env-default:"8080"`
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

type Redis struct {
	Address     string        `yaml:"address" env-default:"localhost:6379"`
	Password    string        `yaml:"password" env-default:""`
	User        string        `yaml:"user" env-default:""`
	DB          int           `yaml:"db" env-default:"0"`
	MaxRetries  int           `yaml:"max_retries" env-default:"1"`
	DialTimeout time.Duration `yaml:"dial_timeout" env-default:"10s"`
	Timeout     time.Duration `yaml:"timeout" env-default:"10s"`
}

func MustLoad() *Config {
	// TODO переделать на переменную среды так как нужно будет менять его при переезде на продакшен.
	configPath := "./config/local/cactus.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

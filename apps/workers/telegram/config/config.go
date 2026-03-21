package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"dev"`
	WorkerUUID string `yaml:"worker_uuid"`
	Nats       Nats   `yaml:"nats"`
	Token      string `yaml:"token"`
}

type Nats struct {
	URL string `yaml:"url" env-default:"nats://localhost:4222"`
}

func MustLoad() *Config {
	configPath := "./configs/apps/workers/telegram.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

package main

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"dev"`
	WorkerUUID string `yaml:"worker_uuid"`
	Redis      Redis  `yaml:"redis"`
	Token      string `yaml:"token"`
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
	// TODO переделать на переменную среды так как нужно будет менять его при переезде на продакшен
	configPath := "./config/telegram.worker.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

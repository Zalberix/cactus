package config

import cfgloader "github.com/zalberix/cactus/libs/config"

type Config struct {
	cfgloader.Worker `yaml:",inline"`

	Env       string `yaml:"env" env-default:"dev"`
	NatsURL   string `yaml:"nats_url" env-default:"nats://localhost:4222"`
	Token     string `yaml:"token"`
	ServerURL string `yaml:"server_url" env-default:"http://localhost:8080"`
}

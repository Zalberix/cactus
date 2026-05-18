package config

import cfgloader "github.com/zalberix/cactus/libs/config"

type Config struct {
	cfgloader.Worker `yaml:",inline"`

	Env       string `yaml:"env" env-default:"dev"`
	Token     string `yaml:"token"`
	ServerURL string `yaml:"server_url" env-default:"http://localhost:8080"`
}

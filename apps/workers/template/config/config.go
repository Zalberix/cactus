package config

import cfgloader "github.com/zalberix/cactus/libs/config"

type Config struct {
	cfgloader.Worker `yaml:",inline"`

	Env string `yaml:"env" env-default:"local"`
}

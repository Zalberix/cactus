package config

import cfgloader "github.com/zalberix/cactus/libs/config"

// Config describes SMTP worker configuration.
type Config struct {
	cfgloader.Worker `yaml:",inline"`

	Env     string `yaml:"env" env-default:"local"`
	NatsURL string `yaml:"nats_url" env-default:"nats://localhost:4222"`
	SMTP    SMTP   `yaml:"smtp"`
}

// SMTP contains SMTP server settings.
type SMTP struct {
	Host string `yaml:"host" env-default:"localhost"`
	Port int    `yaml:"port" env-default:"1025"`
	From string `yaml:"from" env-default:"noreply@cactus.local"`
	Auth string `yaml:"auth" env-default:"none"` // "none", "plain", "login"
	TLS  string `yaml:"tls" env-default:"none"`  // "none", "tls", "starttls"
}

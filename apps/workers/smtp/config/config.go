package config

type Config struct {
	Env        string `yaml:"env" env-default:"dev"`
	WorkerUUID string `yaml:"worker_uuid"`
	Nats       Nats   `yaml:"nats"`
	Token      string `yaml:"token"`
}

type Nats struct {
	URL string `yaml:"url" env-default:"nats://localhost:4222"`
}

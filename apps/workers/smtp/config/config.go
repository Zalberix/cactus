package config

// Config --- конфигурация SMTP воркера.
type Config struct {
	NatsURL        string `yaml:"nats_url" env-default:"nats://localhost:4222"`
	ManagerURL     string `yaml:"manager_url" env-default:"http://localhost:8080"`
	BootstrapToken string `yaml:"bootstrap_token" env-required:"true"`
	WorkTypeID     int32  `yaml:"work_type_id" env-required:"true"`
	RevisionID     int32  `yaml:"revision_id" env-required:"true"`
	WorkerIDFile   string `yaml:"worker_id_file" env-default:".worker_id"`
	SMTP           SMTP   `yaml:"smtp"`
}

// SMTP --- параметры SMTP-сервера.
type SMTP struct {
	Host string `yaml:"host" env-default:"localhost"`
	Port int    `yaml:"port" env-default:"1025"`
	From string `yaml:"from" env-default:"noreply@cactus.local"`
	Auth string `yaml:"auth" env-default:"none"` // "none", "plain", "login"
	TLS  string `yaml:"tls" env-default:"none"`  // "none", "tls", "starttls"
}

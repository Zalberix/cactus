package config

// Worker contains shared configuration required by worker services.
type Worker struct {
	ManagerURL     string `yaml:"manager_url" env-default:"http://localhost:3009"`
	BootstrapToken string `yaml:"bootstrap_token"`
	NatsCAFile     string `yaml:"nats_ca_file"`
}

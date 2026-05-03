package config

// Worker contains shared configuration required by worker services.
type Worker struct {
	ManagerURL     string `yaml:"manager_url" env-default:"http://localhost:3009"`
	BootstrapToken string `yaml:"bootstrap_token"`
	WorkTypeID     int32  `yaml:"work_type_id" env-default:"1"`
	RevisionID     int32  `yaml:"revision_id" env-default:"1"`
}

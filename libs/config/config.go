package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

// MustLoad читает YAML-конфиг в произвольную структуру T.
// Путь к директории конфигов берётся из CONFIG_DIR, иначе — cwd.
func MustLoad[T any](configFileName string) *T {
	configDir := os.Getenv("CONFIG_DIR")

	if configDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get working directory: %v", err)
		}
		configDir = wd
	}
	fullPath := filepath.Clean(filepath.Join(configDir, configFileName))

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %q", fullPath) // #nosec G706 -- fatal startup diagnostic.
	}

	var cfg T

	if err := cleanenv.ReadConfig(fullPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

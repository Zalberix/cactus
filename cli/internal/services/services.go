package services

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type WorkerDef struct {
	Name    string `yaml:"name"`
	App     string `yaml:"app"`
	Variant string `yaml:"variant"`
	Count   int    `yaml:"count"`
}

type CoreWorkersDef struct {
	Count int `yaml:"count"`
}

type Config struct {
	CoreWorkers CoreWorkersDef `yaml:"core_workers"`
	Workers     []WorkerDef    `yaml:"workers"`
}

type Lock struct {
	Workers map[string][]string `yaml:"workers"`
}

type WorkerInstance struct {
	Name    string
	App     string
	Variant string
	UUID    string
	IDPath  string
}

func LoadServices(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read services config %s: %w", path, err)
	}

	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse services config: %w", err)
	}

	if cfg.Workers == nil {
		cfg.Workers = make([]WorkerDef, 0)
	}
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateConfig(cfg Config) error {
	if cfg.CoreWorkers.Count < 0 {
		return fmt.Errorf("core_workers.count must be >= 0")
	}

	seen := make(map[string]struct{}, len(cfg.Workers))
	for i, worker := range cfg.Workers {
		if worker.Name == "" {
			return fmt.Errorf("workers[%d].name is required", i)
		}
		if worker.App == "" {
			return fmt.Errorf("workers[%d].app is required", i)
		}
		if worker.Variant == "" {
			return fmt.Errorf("workers[%d].variant is required", i)
		}
		if worker.Count < 0 {
			return fmt.Errorf("workers[%d].count must be >= 0", i)
		}
		if _, exists := seen[worker.Name]; exists {
			return fmt.Errorf("duplicate worker name %q", worker.Name)
		}
		seen[worker.Name] = struct{}{}
	}
	return nil
}

func Reconcile(cfgPath, lockPath, workerIDDir string) ([]WorkerInstance, error) {
	cfg, err := LoadServices(cfgPath)
	if err != nil {
		return nil, err
	}

	lock, err := loadLock(lockPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read lock %s: %w", lockPath, err)
	}
	if lock == nil {
		lock = &Lock{Workers: make(map[string][]string)}
	}

	changed := false
	configured := make(map[string]WorkerDef, len(cfg.Workers))

	for _, def := range cfg.Workers {
		configured[def.Name] = def

		current := lock.Workers[def.Name]
		desired := def.Count

		if len(current) < desired {
			for len(current) < desired {
				id, err := generateUUID()
				if err != nil {
					return nil, fmt.Errorf("generate UUID: %w", err)
				}
				current = append(current, id)
			}
			lock.Workers[def.Name] = current
			changed = true
		} else if len(current) > desired {
			toRemove := current[desired:]
			for _, uuid := range toRemove {
				removeWorkerIDFile(workerIDDir, def.Name, uuid)
			}
			lock.Workers[def.Name] = current[:desired]
			changed = true
		}
	}

	for workerName, uuids := range lock.Workers {
		if _, exists := configured[workerName]; !exists {
			for _, uuid := range uuids {
				removeWorkerIDFile(workerIDDir, workerName, uuid)
			}
			delete(lock.Workers, workerName)
			changed = true
		}
	}

	if changed {
		if err := saveLock(lockPath, lock); err != nil {
			return nil, err
		}
	}

	instances := make([]WorkerInstance, 0)
	for _, def := range cfg.Workers {
		for _, uuid := range lock.Workers[def.Name] {
			instances = append(instances, WorkerInstance{
				Name:    def.Name,
				App:     def.App,
				Variant: def.Variant,
				UUID:    uuid,
				IDPath:  filepath.Join(workerIDDir, def.Name, uuid),
			})
		}
	}

	return instances, nil
}

func loadLock(path string) (*Lock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var lock Lock
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parse lock file: %w", err)
	}

	if lock.Workers == nil {
		lock.Workers = make(map[string][]string)
	}

	return &lock, nil
}

func saveLock(path string, lock *Lock) error {
	data, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create lock dir: %w", err)
	}

	return os.WriteFile(path, data, 0o600)
}

func removeWorkerIDFile(workerIDDir, workerName, uuid string) {
	path := filepath.Join(workerIDDir, workerName, uuid)
	_ = os.Remove(path)

	groupDir := filepath.Join(workerIDDir, workerName)
	entries, err := os.ReadDir(groupDir)
	if err == nil && len(entries) == 0 {
		_ = os.Remove(groupDir)
	}
}

func generateUUID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16]), nil
}

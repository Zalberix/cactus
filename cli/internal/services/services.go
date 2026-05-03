package services

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type WorkerDef struct {
	Count int `yaml:"count"`
}

type Config struct {
	Workers map[string]WorkerDef `yaml:"workers"`
}

type Lock struct {
	Workers map[string][]string `yaml:"workers"`
}

// WorkerInstance — один инстанс воркера после reconciliation.
type WorkerInstance struct {
	Type string
	UUID string
	// IDPath — полный путь к runtime-файлу (.worker_id/{type}/{uuid}).
	IDPath string
}

// LoadServices читает cactus-services.yaml.
func LoadServices(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read services config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse services config: %w", err)
	}

	if cfg.Workers == nil {
		cfg.Workers = make(map[string]WorkerDef)
	}

	return &cfg, nil
}

// Reconcile загружает (или создаёт) lock-файл, сверяет с services config,
// добавляет/удаляет UUID, чистит runtime-файлы в workerIDDir.
// Возвращает итоговый список WorkerInstance.
func Reconcile(cfgPath, lockPath, workerIDDir string) ([]WorkerInstance, error) { //nolint:gocognit // Reconciliation keeps create/remove decisions in one pass over config and lock state.
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

	// Добавляем / убираем UUID по каждому типу из services config
	for workerType, def := range cfg.Workers {
		current := lock.Workers[workerType]
		desired := def.Count
		if desired < 0 {
			desired = 0
		}

		if len(current) < desired {
			// Нужно добавить инстансы
			for len(current) < desired {
				id, err := generateUUID()
				if err != nil {
					return nil, fmt.Errorf("generate UUID: %w", err)
				}
				current = append(current, id)
			}
			lock.Workers[workerType] = current
			changed = true
		} else if len(current) > desired {
			// Нужно удалить лишние инстансы (с конца)
			toRemove := current[desired:]
			for _, uuid := range toRemove {
				removeWorkerIDFile(workerIDDir, workerType, uuid)
			}
			lock.Workers[workerType] = current[:desired]
			changed = true
		}
	}

	// Удаляем типы, которых больше нет в services config
	for workerType, uuids := range lock.Workers {
		if _, exists := cfg.Workers[workerType]; !exists {
			for _, uuid := range uuids {
				removeWorkerIDFile(workerIDDir, workerType, uuid)
			}
			delete(lock.Workers, workerType)
			changed = true
		}
	}

	if changed {
		if err := saveLock(lockPath, lock); err != nil {
			return nil, err
		}
	}

	// Собираем итоговый список инстансов
	var instances []WorkerInstance
	for workerType, uuids := range lock.Workers {
		for _, uuid := range uuids {
			instances = append(instances, WorkerInstance{
				Type:   workerType,
				UUID:   uuid,
				IDPath: filepath.Join(workerIDDir, workerType, uuid),
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

func removeWorkerIDFile(workerIDDir, workerType, uuid string) {
	path := filepath.Join(workerIDDir, workerType, uuid)
	_ = os.Remove(path)

	// Попробуем удалить директорию типа если она пустая
	typeDir := filepath.Join(workerIDDir, workerType)
	entries, err := os.ReadDir(typeDir)
	if err == nil && len(entries) == 0 {
		_ = os.Remove(typeDir)
	}
}

func generateUUID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	// UUID v4
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16]), nil
}

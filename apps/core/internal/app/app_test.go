package app

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/config"
	temporalworker "github.com/zalberix/cactus/apps/core/internal/temporal/worker"
)

func minimalConfig() *config.Config {
	return &config.Config{
		Env: "local",
		HTTPServer: config.HTTPServer{
			Host: "127.0.0.1",
			Port: "0",
		},
		Nats: config.Nats{
			URL: "nats://127.0.0.1:4222",
		},
		Temporal: config.Temporal{
			HostPort:  "127.0.0.1:7233",
			Namespace: "default",
			TaskQueue: "cactus-core",
		},
		JWT: config.JWT{
			Secret: "dev-secret-change-in-production-32chars!",
		},
	}
}

func TestWorkerOptionsValidate(t *testing.T) {
	if err := fx.ValidateApp(WorkerOptions(minimalConfig()), fx.NopLogger); err != nil {
		t.Fatalf("worker fx graph: %v", err)
	}
}

func TestWorkerInvokesRegisterTemporalWorker(t *testing.T) {
	workerName := functionName(temporalworker.RegisterTemporalWorker)
	for _, invoke := range workerInvokes() {
		if functionName(invoke) == workerName {
			return
		}
	}
	t.Fatalf("worker app must register Temporal worker")
}

func functionName(fn any) string {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i+1:]
	}
	return name
}

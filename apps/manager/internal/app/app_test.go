package app

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/manager/config"
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

func TestManagerOptionsValidate(t *testing.T) {
	if err := fx.ValidateApp(ManagerOptions(minimalConfig()), fx.NopLogger); err != nil {
		t.Fatalf("manager fx graph: %v", err)
	}
}

func TestManagerInvokesDoNotRegisterTemporalWorker(t *testing.T) {
	for _, invoke := range managerInvokes() {
		if strings.Contains(functionName(invoke), "RegisterTemporalWorker") {
			t.Fatalf("manager must not register Temporal worker")
		}
	}
}

func functionName(fn any) string {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i+1:]
	}
	return name
}

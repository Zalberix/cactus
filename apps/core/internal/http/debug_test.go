package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zalberix/cactus/apps/core/config"
)

func TestNewDebugServerEnablesPprofOutsideProduction(t *testing.T) {
	srv := NewDebugServer(&config.Config{Env: config.AppEnvLocal})
	if !srv.Enabled {
		t.Fatal("expected debug server to be enabled in local env")
	}
	if srv.Server == nil {
		t.Fatal("expected debug server HTTP server")
	}
	if srv.Server.Addr != "127.0.0.1:6060" {
		t.Fatalf("expected debug server addr 127.0.0.1:6060, got %q", srv.Server.Addr)
	}

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	rec := httptest.NewRecorder()
	srv.Server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected /debug/pprof/ status 200, got %d", rec.Code)
	}
}

func TestNewDebugServerDisablesPprofInProduction(t *testing.T) {
	srv := NewDebugServer(&config.Config{Env: config.AppEnvProduction})
	if srv.Enabled {
		t.Fatal("expected debug server to be disabled in production")
	}
	if srv.Server != nil {
		t.Fatal("expected no debug HTTP server in production")
	}
}

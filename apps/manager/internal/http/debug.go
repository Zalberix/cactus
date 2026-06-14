package http

import (
	"net/http"
	"net/http/pprof"

	"github.com/zalberix/cactus/apps/manager/config"
)

const defaultDebugAddr = "127.0.0.1:6060"

type DebugServer struct {
	Enabled bool
	Server  *http.Server
}

func NewDebugServer(cfg *config.Config) *DebugServer {
	if cfg.Env == config.AppEnvProduction {
		return &DebugServer{}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &DebugServer{
		Enabled: true,
		Server: &http.Server{
			Addr:    defaultDebugAddr,
			Handler: mux,
		},
	}
}

package meta

import (
	"fmt"
	"os"

	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/config"
)

type ServerMeta struct {
	HostName string
	Port     string
}

type Opts struct {
	fx.In
	Config *config.Config
}

func NewFx(opts Opts) (*ServerMeta, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить hostname: %w", err)
	}
	return &ServerMeta{
		HostName: host,
		Port:     opts.Config.HTTPServer.Port,
	}, nil
}

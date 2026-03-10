package temporal

import (
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/config"
	"go.temporal.io/sdk/client"
)

type ClientOpts struct {
	fx.In
	Config *config.Config
}

func NewClientFx(opts ClientOpts) (client.Client, error) {
	return client.Dial(client.Options{
		HostPort:  opts.Config.Temporal.HostPort,
		Namespace: opts.Config.Temporal.Namespace,
	})
}

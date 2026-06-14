package targets

import (
	"github.com/zalberix/cactus/cli/internal/cmds/dev/frontend"
	devgolang "github.com/zalberix/cactus/cli/internal/cmds/dev/golang"
	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
	"github.com/zalberix/cactus/cli/internal/cmds/proxy"
)

func BuildTargets(opts devruntime.Options) ([]devruntime.TargetSpec, error) {
	targets := []devruntime.TargetSpec{
		devruntime.CleanPortsTarget(),
		proxy.TargetSpec(),
		devruntime.DependenciesTarget(opts.InstallDeps),
		devruntime.MigrationsTarget(opts.RunMigrations),
	}

	goApps, err := devgolang.New(opts.Debug)
	if err != nil {
		return nil, err
	}
	targets = append(targets, goApps.TargetSpecs()...)

	frontendTargets, err := frontend.TargetSpecs()
	if err != nil {
		return nil, err
	}
	targets = append(targets, frontendTargets...)

	return targets, nil
}

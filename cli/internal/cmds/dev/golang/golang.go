package golang

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/zalberix/cactus/cli/internal/goapp"
	"github.com/zalberix/cactus/cli/internal/services"
)

type GoApps struct {
	apps         []*goapp.GoApp
	appsByName   map[string]*goapp.GoApp
	debugEnabled bool
}

func New(enableDebug bool) (*GoApps, error) {
	apps := expandWorkers(goapp.Apps)

	ga := &GoApps{
		debugEnabled: enableDebug,
		appsByName:   make(map[string]*goapp.GoApp, len(apps)),
	}

	for _, app := range apps {
		if _, exists := ga.appsByName[app.Name]; exists {
			return nil, fmt.Errorf("duplicate app name: %q", app.Name)
		}
		ga.appsByName[app.Name] = nil
	}

	for _, app := range apps {
		for _, dep := range app.DependsOn {
			if _, exists := ga.appsByName[dep]; !exists {
				return nil, fmt.Errorf("app %q depends on unknown app %q", app.Name, dep)
			}
		}
	}

	if err := detectCycles(apps); err != nil {
		return nil, err
	}

	sorted := topoSort(apps)
	for _, app := range sorted {
		application, err := goapp.NewApplication(
			app.Name,
			enableDebug,
			app.AppDir,
			app.SourceDir,
			app.CommandDir,
			app.IsCoreWorker,
			app.Port,
			app.DebugPort,
			app.OnPortReady,
			app.DependsOn,
		)
		if err != nil {
			return nil, err
		}
		application.SourceDir = app.SourceDir
		application.CommandDir = app.CommandDir
		application.IsWorker = app.IsWorker
		application.IsCoreWorker = app.IsCoreWorker
		application.WorkerUUID = app.WorkerUUID
		application.WorkerGroupName = app.WorkerGroupName
		application.WorkerVariant = app.WorkerVariant
		application.BaseName = app.BaseName
		application.ExtraArgs = app.ExtraArgs
		ga.apps = append(ga.apps, application)
		ga.appsByName[app.Name] = application
	}

	return ga, nil
}

const (
	servicesConfigPath = "configs/cactus-services.yaml"
	servicesLockPath   = "configs/cactus-services-lock.yaml"
	workerIDDir        = ".worker_id"
)

func expandWorkers(templates []goapp.GoApp) []goapp.GoApp {
	cfg, cfgErr := services.LoadServices(servicesConfigPath)
	instances, err := services.Reconcile(servicesConfigPath, servicesLockPath, workerIDDir)
	if err != nil {
		pterm.Warning.Printfln("services reconcile: %v (workers will run with default count)", err)
		return templates
	}
	if cfgErr != nil {
		pterm.Warning.Printfln("services config: %v (core workers will not be expanded)", cfgErr)
		return expandServiceTemplates(templates, instances, 0)
	}

	return expandServiceTemplates(templates, instances, cfg.CoreWorkers.Count)
}

func expandServiceTemplates(templates []goapp.GoApp, instances []services.WorkerInstance, coreWorkerCount int) []goapp.GoApp {
	byApp := make(map[string][]services.WorkerInstance)
	for _, inst := range instances {
		byApp[inst.App] = append(byApp[inst.App], inst)
	}

	var result []goapp.GoApp
	for _, tmpl := range templates {
		switch {
		case tmpl.IsCoreWorker:
			for i := 1; i <= coreWorkerCount; i++ {
				expanded := tmpl
				expanded.BaseName = tmpl.Name
				expanded.Name = fmt.Sprintf("%s-%d", tmpl.Name, i)
				expanded.DebugPort = tmpl.DebugPort + i - 1
				expanded.IsWorker = false
				expanded.IsCoreWorker = true
				expanded.WorkerUUID = ""
				expanded.WorkerGroupName = ""
				expanded.WorkerVariant = ""
				expanded.ExtraArgs = nil
				result = append(result, expanded)
			}

		case !tmpl.IsWorker:
			result = append(result, tmpl)

		default:
			workerInstances, ok := byApp[tmpl.Name]
			if !ok || len(workerInstances) == 0 {
				pterm.Info.Printfln("Worker app %q not in cactus-services.yaml, skipping", tmpl.Name)
				continue
			}

			groupCounts := make(map[string]int)
			for i, inst := range workerInstances {
				expanded := tmpl
				expanded.BaseName = tmpl.Name
				expanded.WorkerUUID = inst.UUID
				expanded.WorkerGroupName = inst.Name
				expanded.WorkerVariant = inst.Variant
				expanded.IsWorker = true
				expanded.DebugPort = tmpl.DebugPort + i

				groupCounts[inst.Name]++
				if countInstancesForGroup(workerInstances, inst.Name) > 1 {
					expanded.Name = fmt.Sprintf("%s-%d", inst.Name, groupCounts[inst.Name])
				} else {
					expanded.Name = inst.Name
				}

				absPath, _ := filepath.Abs(inst.IDPath)
				expanded.ExtraArgs = []string{
					"--worker-id-path", absPath,
					"--worker-variant", inst.Variant,
					"--worker-name", expanded.Name,
				}

				result = append(result, expanded)
			}
		}
	}

	return result
}

func expandWorkerTemplates(templates []goapp.GoApp, instances []services.WorkerInstance) []goapp.GoApp {
	return expandServiceTemplates(templates, instances, 0)
}

func countInstancesForGroup(instances []services.WorkerInstance, name string) int {
	count := 0
	for _, inst := range instances {
		if inst.Name == name {
			count++
		}
	}
	return count
}

func (c *GoApps) Start(ctx context.Context) error {
	for _, app := range c.apps {
		app := app

		for _, depName := range app.DependsOn {
			dep := c.appsByName[depName]
			pterm.Info.Printfln("[%s] Waiting for %s...", app.Name, depName)
			select {
			case <-dep.Ready:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		pterm.Info.Printfln("Starting %s", app.Name)
		if err := app.Start(); err != nil {
			return err
		}

		go func() {
			chann, err := app.Watcher.Start(app.GetAppPath())
			if err != nil {
				pterm.Fatal.Println(err)
			}

			for range chann {
				pterm.Info.Println("ReStarting " + app.Name)
				if err := app.Start(); err != nil {
					pterm.Error.Println("Error Watch of: ", err)
				}
			}
		}()
	}

	return nil
}

func (c *GoApps) Stop() {
	for _, app := range c.apps {
		app.Watcher.Stop()
		app.Stop()
	}
}

func detectCycles(apps []goapp.GoApp) error {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int, len(apps))

	deps := make(map[string][]string, len(apps))
	for _, app := range apps {
		deps[app.Name] = app.DependsOn
	}

	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		color[name] = gray
		path = append(path, name)

		for _, dep := range deps[name] {
			switch color[dep] {
			case gray:
				cycle := append([]string{}, path...)
				cycle = append(cycle, dep)
				return fmt.Errorf("circular dependency: %s", formatCycle(cycle))
			case white:
				if err := visit(dep, path); err != nil {
					return err
				}
			}
		}

		color[name] = black
		return nil
	}

	for _, app := range apps {
		if color[app.Name] == white {
			if err := visit(app.Name, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func formatCycle(path []string) string {
	result := path[0]
	for i := 1; i < len(path); i++ {
		result += " -> " + path[i]
	}
	return result
}

func topoSort(apps []goapp.GoApp) []goapp.GoApp {
	byName := make(map[string]goapp.GoApp, len(apps))
	for _, app := range apps {
		byName[app.Name] = app
	}

	visited := make(map[string]bool, len(apps))
	var result []goapp.GoApp

	var visit func(name string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		app := byName[name]
		for _, dep := range app.DependsOn {
			visit(dep)
		}
		result = append(result, app)
	}

	for _, app := range apps {
		visit(app.Name)
	}

	return result
}

package golang

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/zalberix/cactus/cli/internal/goapp"
)

type GoApps struct {
	apps         []*goapp.GoApp
	appsByName   map[string]*goapp.GoApp
	debugEnabled bool
}

func New(enableDebug bool) (*GoApps, error) {
	ga := &GoApps{
		debugEnabled: enableDebug,
		appsByName:   make(map[string]*goapp.GoApp, len(goapp.Apps)),
	}

	// Валидация уникальности имён
	for _, app := range goapp.Apps {
		if _, exists := ga.appsByName[app.Name]; exists {
			return nil, fmt.Errorf("duplicate app name: %q", app.Name)
		}
		ga.appsByName[app.Name] = nil // placeholder, заполним после создания
	}

	// Валидация DependsOn — все имена должны существовать
	for _, app := range goapp.Apps {
		for _, dep := range app.DependsOn {
			if _, exists := ga.appsByName[dep]; !exists {
				return nil, fmt.Errorf("app %q depends on unknown app %q", app.Name, dep)
			}
		}
	}

	// Проверка циклических зависимостей
	if err := detectCycles(goapp.Apps); err != nil {
		return nil, err
	}

	// Топологическая сортировка
	sorted, err := topoSort(goapp.Apps)
	if err != nil {
		return nil, err
	}

	// Создание экземпляров в отсортированном порядке
	for _, app := range sorted {
		application, err := goapp.NewApplication(
			app.Name,
			enableDebug,
			app.AppDir,
			app.Port,
			app.DebugPort,
			app.OnPortReady,
			app.DependsOn,
		)
		if err != nil {
			return nil, err
		}
		ga.apps = append(ga.apps, application)
		ga.appsByName[app.Name] = application
	}

	return ga, nil
}

func (c *GoApps) Start(ctx context.Context) error {
	for _, app := range c.apps {
		app := app

		// Ждём готовности всех зависимостей
		for _, depName := range app.DependsOn {
			dep := c.appsByName[depName]
			pterm.Info.Printfln("[%s] Ожидание %s...", app.Name, depName)
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

// detectCycles проверяет наличие циклических зависимостей через DFS.
func detectCycles(apps []goapp.GoApp) error {
	const (
		white = 0 // не посещён
		gray  = 1 // в обработке (на стеке)
		black = 2 // завершён
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
				// Нашли цикл — формируем путь
				cycle := append(path, dep)
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

// topoSort выполняет топологическую сортировку приложений по зависимостям.
// Зависимости идут первыми.
func topoSort(apps []goapp.GoApp) ([]goapp.GoApp, error) {
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

	return result, nil
}

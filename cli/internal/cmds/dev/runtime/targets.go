package runtime

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/zalberix/cactus/cli/internal/cmds/dev/helpers"
	"github.com/zalberix/cactus/cli/internal/shell"
)

type Options struct {
	InstallDeps   bool
	RunMigrations bool
	Debug         bool
	LogCapacity   int
}

type writerReporter struct {
	w io.Writer
}

func (r writerReporter) Infof(format string, args ...any) {
	fmt.Fprintf(r.w, format+"\n", args...)
}

func (r writerReporter) Warnf(format string, args ...any) {
	fmt.Fprintf(r.w, "warning: "+format+"\n", args...)
}

func (r writerReporter) Successf(format string, args ...any) {
	fmt.Fprintf(r.w, format+"\n", args...)
}

func CleanPortsTarget() TargetSpec {
	return TargetSpec{
		ID:   "clean-ports",
		Name: "clean-ports",
		Kind: TargetTask,
		Build: func(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
			return helpers.CleanPorts(ctx, writerReporter{w: stdout})
		},
		Ready: ImmediateReady(),
	}
}

func DependenciesTarget(enabled bool) TargetSpec {
	if !enabled {
		return skippedTask("deps", "deps")
	}
	return TargetSpec{
		ID:   "deps",
		Name: "deps",
		Kind: TargetTask,
		Build: func(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
			if err := runShellTask(ctx, stdout, stderr, "go mod download"); err != nil {
				return err
			}
			return runShellTask(ctx, stdout, stderr, "npm install --frozen-lockfile")
		},
		Ready: ImmediateReady(),
	}
}

func MigrationsTarget(enabled bool) TargetSpec {
	if !enabled {
		return skippedTask("migrations", "migrations")
	}
	return TargetSpec{
		ID:   "migrations",
		Name: "migrations",
		Kind: TargetTask,
		Build: func(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
			if err := runShellTask(ctx, stdout, stderr, "go run ./cli/main.go migration up"); err != nil {
				fmt.Fprintf(stderr, "warning: migrations failed: %v\n", err)
			}
			return nil
		},
		Ready: ImmediateReady(),
	}
}

func skippedTask(id string, name string) TargetSpec {
	return TargetSpec{
		ID:    id,
		Name:  name,
		Kind:  TargetTask,
		Skip:  true,
		Ready: ImmediateReady(),
	}
}

func runShellTask(ctx context.Context, stdout io.Writer, stderr io.Writer, command string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return shell.ExecCommand(shell.ExecCommandOpts{
		Context: ctx,
		Command: command,
		Pwd:     wd,
		Stdout:  stdout,
		Stderr:  stderr,
	})
}

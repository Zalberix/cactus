package frontend

import (
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/zalberix/cactus/cli/internal/shell"
)

type app struct {
	name string
	cmd  *exec.Cmd
	path string
}

func newApplication(name, path string) (*app, error) {
	app := app{
		name: name,
		cmd:  nil,
		path: path,
	}

	cmd, err := app.createAppCommand()
	if err != nil {
		return nil, err
	}
	app.cmd = cmd

	return &app, nil
}

func (c *app) stop() error {
	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			return err
		}
	}

	return nil
}

func (c *app) createAppCommand() (*exec.Cmd, error) {
	return c.createAppCommandWith(context.Background(), os.Stdout, os.Stderr)
}

func (c *app) createAppCommandWith(ctx context.Context, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
	cmd, err := shell.CreateCommand(
		shell.ExecCommandOpts{
			Context: ctx,
			Command: "npm run dev",
			Pwd:     c.path,
			Stdout:  stdout,
			Stderr:  stderr,
		},
	)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (c *app) start() error {
	return c.cmd.Start()
}

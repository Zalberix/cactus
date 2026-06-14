package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
	devtargets "github.com/zalberix/cactus/cli/internal/cmds/dev/targets"
)

func Run(ctx context.Context, opts devruntime.Options) error {
	targets, err := devtargets.BuildTargets(opts)
	if err != nil {
		return err
	}

	events := make(chan devruntime.Event, 256)
	runner := devruntime.NewRunner(targets, events)
	model := NewModel(targets, runner, events)
	if opts.LogCapacity > 0 {
		model.logs = devruntime.NewLogStore(opts.LogCapacity)
	}

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)
	_, err = program.Run()
	return err
}

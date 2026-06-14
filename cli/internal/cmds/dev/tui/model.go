package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

type Mode string

const (
	ModeStartup  Mode = "startup"
	ModeMenu     Mode = "menu"
	ModeLogs     Mode = "logs"
	ModeShutdown Mode = "shutdown"
	ModeFailed   Mode = "failed"
)

type Model struct {
	mode         Mode
	specs        []devruntime.TargetSpec
	targets      map[string]devruntime.TargetState
	order        []string
	selected     int
	logs         *devruntime.LogStore
	events       <-chan devruntime.Event
	runner       RunnerPort
	spinner      spinner.Model
	viewport     viewport.Model
	width        int
	height       int
	shuttingDown bool
	err          error
	returnMode   Mode
}

type RunnerPort interface {
	Start(context.Context) error
	Restart(context.Context, string) error
	Shutdown(context.Context) error
}

func NewModel(specs []devruntime.TargetSpec, runner RunnerPort, events <-chan devruntime.Event) Model {
	targets := make(map[string]devruntime.TargetState, len(specs))
	order := make([]string, 0, len(specs))
	for _, spec := range specs {
		targets[spec.ID] = devruntime.TargetState{
			Spec:   spec,
			Status: devruntime.StatusPending,
		}
		order = append(order, spec.ID)
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return Model{
		mode:       ModeStartup,
		specs:      specs,
		targets:    targets,
		order:      order,
		logs:       devruntime.NewLogStore(2000),
		events:     events,
		runner:     runner,
		spinner:    sp,
		selected:   0,
		returnMode: ModeMenu,
	}
}

func (m *Model) applyRuntimeEvent(ev devruntime.Event) {
	state, ok := m.targets[ev.TargetID]
	if !ok {
		return
	}

	switch ev.Type {
	case devruntime.EventStatus:
		state.Status = ev.Status
		state.LastError = ""
		if ev.Err != nil {
			state.LastError = ev.Err.Error()
		}
		if ev.Status == devruntime.StatusFailed && ev.Message != "" {
			state.LastError = ev.Message
		}
		m.targets[ev.TargetID] = state
	case devruntime.EventLog:
		if ev.Line != nil {
			m.logs.Append(*ev.Line)
		}
	case devruntime.EventExit:
		if ev.Err != nil {
			state.Status = devruntime.StatusFailed
			state.LastError = ev.Err.Error()
			m.targets[ev.TargetID] = state
		}
	}

	if state.Status == devruntime.StatusFailed && m.mode != ModeShutdown {
		m.mode = ModeFailed
		m.returnMode = ModeFailed
	}
	if m.allReady() && m.mode == ModeStartup {
		m.mode = ModeMenu
		m.returnMode = ModeMenu
	}
}

func (m Model) allReady() bool {
	if len(m.targets) == 0 {
		return false
	}
	for _, id := range m.order {
		switch m.targets[id].Status {
		case devruntime.StatusReady, devruntime.StatusRunning, devruntime.StatusSkipped, devruntime.StatusStopped:
		default:
			return false
		}
	}
	return true
}

func (m Model) selectedTargetID() string {
	if len(m.order) == 0 {
		return ""
	}
	if m.selected < 0 {
		return m.order[0]
	}
	if m.selected >= len(m.order) {
		return m.order[len(m.order)-1]
	}
	return m.order[m.selected]
}

func (m *Model) refreshLogViewport() {
	targetID := m.selectedTargetID()
	entries := m.logs.Entries(targetID)
	var lines []string
	for _, entry := range entries {
		lines = append(lines, formatLogEntry(entry))
	}
	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.GotoBottom()
}

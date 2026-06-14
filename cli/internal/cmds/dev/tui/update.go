package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

type runtimeEventMsg devruntime.Event
type startupFinishedMsg struct{ err error }
type shutdownFinishedMsg struct{ err error }
type restartFinishedMsg struct {
	targetID string
	err      error
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startRunnerCmd(),
		m.waitRuntimeEventCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = max(1, msg.Height-6)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case runtimeEventMsg:
		m.applyRuntimeEvent(devruntime.Event(msg))
		if m.mode == ModeLogs {
			m.refreshLogViewport()
		}
		return m, m.waitRuntimeEventCmd()
	case startupFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.mode = ModeFailed
			m.returnMode = ModeFailed
		}
		return m, nil
	case restartFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			if state, ok := m.targets[msg.targetID]; ok {
				state.LastError = msg.err.Error()
				m.targets[msg.targetID] = state
			}
		}
		return m, nil
	case shutdownFinishedMsg:
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		return m.handleKey(msg)
	default:
		return m, nil
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyCtrlC:
		if m.shuttingDown {
			return m, nil
		}
		m.mode = ModeShutdown
		m.shuttingDown = true
		return m, m.shutdownCmd()
	case keyUp:
		if m.mode == ModeLogs {
			m.viewport.LineUp(1)
			return m, nil
		}
		if m.canMoveSelection() && m.selected > 0 {
			m.selected--
		}
		return m, nil
	case keyDown:
		if m.mode == ModeLogs {
			m.viewport.LineDown(1)
			return m, nil
		}
		if m.canMoveSelection() && m.selected < len(m.order)-1 {
			m.selected++
		}
		return m, nil
	case keyEnter:
		if m.mode == ModeMenu || m.mode == ModeFailed {
			m.returnMode = m.mode
			m.mode = ModeLogs
			m.refreshLogViewport()
		}
		return m, nil
	case keyEsc, keyQ, keyBackspace:
		if m.mode == ModeLogs {
			m.mode = m.returnMode
			if m.mode == "" {
				m.mode = ModeMenu
			}
		}
		return m, nil
	case keyRestart:
		if m.mode == ModeMenu || m.mode == ModeFailed || m.mode == ModeLogs {
			targetID := m.selectedTargetID()
			if targetID == "" {
				return m, nil
			}
			return m, m.restartCmd(targetID)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) startRunnerCmd() tea.Cmd {
	if m.runner == nil {
		return func() tea.Msg {
			return startupFinishedMsg{}
		}
	}
	return func() tea.Msg {
		return startupFinishedMsg{err: m.runner.Start(context.Background())}
	}
}

func (m Model) waitRuntimeEventCmd() tea.Cmd {
	if m.events == nil {
		return nil
	}
	return func() tea.Msg {
		ev, ok := <-m.events
		if !ok {
			return nil
		}
		return runtimeEventMsg(ev)
	}
}

func (m Model) shutdownCmd() tea.Cmd {
	return func() tea.Msg {
		if m.runner == nil {
			return shutdownFinishedMsg{}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return shutdownFinishedMsg{err: m.runner.Shutdown(ctx)}
	}
}

func (m Model) restartCmd(targetID string) tea.Cmd {
	return func() tea.Msg {
		if m.runner == nil {
			return restartFinishedMsg{targetID: targetID}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		return restartFinishedMsg{targetID: targetID, err: m.runner.Restart(ctx, targetID)}
	}
}

func (m Model) canMoveSelection() bool {
	return m.mode == ModeStartup || m.mode == ModeMenu || m.mode == ModeFailed
}

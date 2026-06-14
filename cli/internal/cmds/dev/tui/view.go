package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

const mainFooter = "up/down select  enter logs  r restart  Ctrl+C shutdown"

func (m Model) View() string {
	switch m.mode {
	case ModeLogs:
		return m.renderLogView()
	case ModeShutdown:
		return m.renderMainView("Shutdown", "shutdown in progress")
	default:
		return m.renderMainView("Cactus dev", mainFooter)
	}
}

func (m Model) renderMainView(title string, footer string) string {
	rows := make([]string, 0, len(m.order))
	for i, targetID := range m.order {
		rows = append(rows, m.renderTargetRow(i, targetID))
	}
	left := strings.Join(rows, "\n")
	if left == "" {
		left = dimStyle.Render("no targets")
	}

	body := left
	if m.width >= 80 && len(m.order) > 0 {
		leftWidth := max(38, m.width/2)
		panel := m.renderSelectedInfo()
		body = lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(leftWidth).Render(left), panel)
	}

	parts := []string{
		titleStyle.Render(title),
		"",
		body,
		"",
		footerStyle.Render(footer),
	}
	if m.err != nil {
		parts = append(parts[:len(parts)-1], errorStyle.Render(m.err.Error()), "", parts[len(parts)-1])
	}
	return strings.Join(parts, "\n")
}

func (m Model) renderTargetRow(index int, targetID string) string {
	state := m.targets[targetID]
	name := state.Spec.Name
	if name == "" {
		name = targetID
	}

	rowWidth := m.width
	if rowWidth <= 0 {
		rowWidth = 80
	}
	nameWidth := max(18, min(40, rowWidth-18))
	row := fmt.Sprintf("%-*s %s", nameWidth, name, m.renderStatus(state))
	if state.Status == devruntime.StatusFailed && state.LastError != "" {
		row += " " + errorStyle.Render(state.LastError)
	}
	if index == m.selected && m.canMoveSelection() {
		return selectedStyle.Render(row)
	}
	if state.Status == devruntime.StatusPending || state.Status == devruntime.StatusSkipped || state.Status == devruntime.StatusStopped {
		return dimStyle.Render(row)
	}
	return row
}

func (m Model) renderStatus(state devruntime.TargetState) string {
	switch state.Status {
	case devruntime.StatusCleaning, devruntime.StatusBuilding, devruntime.StatusStarting, devruntime.StatusRestarting, devruntime.StatusStopping:
		return m.spinner.View()
	case devruntime.StatusReady, devruntime.StatusRunning:
		return okStyle.Render("OK")
	case devruntime.StatusFailed:
		return errorStyle.Render("FAIL")
	case devruntime.StatusSkipped:
		return dimStyle.Render("skipped")
	case devruntime.StatusStopped:
		return dimStyle.Render("stopped")
	case devruntime.StatusPending:
		return dimStyle.Render("pending")
	default:
		return string(state.Status)
	}
}

func (m Model) renderSelectedInfo() string {
	targetID := m.selectedTargetID()
	state := m.targets[targetID]
	lines := []string{
		okStyle.Render(state.Spec.Name),
		"state: " + string(state.Status),
		"kind: " + string(state.Spec.Kind),
	}
	if state.PID != 0 {
		lines = append(lines, fmt.Sprintf("pid: %d", state.PID))
	}
	if len(state.Spec.DependsOn) > 0 {
		lines = append(lines, "deps: "+strings.Join(state.Spec.DependsOn, ", "))
	}
	if state.LastError != "" {
		lines = append(lines, errorStyle.Render(state.LastError))
	}
	return strings.Join(lines, "\n")
}

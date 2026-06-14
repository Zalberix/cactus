package tui

import (
	"fmt"
	"strings"

	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

const logFooter = "up/down scroll  r restart  Esc back  Ctrl+C shutdown"

func (m Model) renderLogView() string {
	targetID := m.selectedTargetID()
	state := m.targets[targetID]
	title := titleStyle.Render(targetID + " logs")

	meta := []string{
		"state: " + string(state.Status),
		"kind: " + string(state.Spec.Kind),
	}
	if state.PID != 0 {
		meta = append(meta, fmt.Sprintf("pid: %d", state.PID))
	}
	if len(state.Spec.DependsOn) > 0 {
		meta = append(meta, "deps: "+strings.Join(state.Spec.DependsOn, ", "))
	}
	if state.LastError != "" {
		meta = append(meta, errorStyle.Render(state.LastError))
	}

	content := m.viewport.View()
	if strings.TrimSpace(content) == "" {
		content = dimStyle.Render("no logs")
	}

	return strings.Join([]string{
		title,
		strings.Join(meta, "  "),
		"",
		content,
		"",
		footerStyle.Render(logFooter),
	}, "\n")
}

func formatLogEntry(entry devruntime.LogEntry) string {
	stream := string(entry.Stream)
	switch entry.Stream {
	case devruntime.StreamStderr:
		stream = warnStyle.Render(stream)
	case devruntime.StreamSystem:
		stream = dimStyle.Render(stream)
	default:
		stream = okStyle.Render(stream)
	}
	return stream + ": " + entry.Line
}

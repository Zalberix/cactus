package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

func TestModelAppliesStatusEvents(t *testing.T) {
	model := NewModel([]devruntime.TargetSpec{
		{ID: "manager", Name: "manager", Kind: devruntime.TargetProcess},
	}, nil, nil)

	model.applyRuntimeEvent(devruntime.StatusEvent("manager", devruntime.StatusStarting, "starting"))

	state := model.targets["manager"]
	if state.Status != devruntime.StatusStarting {
		t.Fatalf("status = %q, want %q", state.Status, devruntime.StatusStarting)
	}
}

func TestModelStoresLogEvents(t *testing.T) {
	model := NewModel([]devruntime.TargetSpec{
		{ID: "manager", Name: "manager", Kind: devruntime.TargetProcess},
	}, nil, nil)

	model.applyRuntimeEvent(devruntime.LogEvent("manager", devruntime.StreamStdout, "hello"))

	logs := model.logs.Entries("manager")
	if len(logs) != 1 || logs[0].Line != "hello" {
		t.Fatalf("logs = %#v", logs)
	}
}

func TestModelMovesSelectionWithArrowKeys(t *testing.T) {
	model := NewModel([]devruntime.TargetSpec{
		{ID: "manager", Name: "manager", Kind: devruntime.TargetProcess},
		{ID: "worker", Name: "worker", Kind: devruntime.TargetProcess},
	}, nil, nil)
	model.mode = ModeMenu

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	if model.selected != 1 {
		t.Fatalf("selected = %d, want 1", model.selected)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)
	if model.selected != 0 {
		t.Fatalf("selected = %d, want 0", model.selected)
	}
}

func TestModelOpensAndClosesLogsWithKeys(t *testing.T) {
	model := NewModel([]devruntime.TargetSpec{
		{ID: "manager", Name: "manager", Kind: devruntime.TargetProcess},
	}, nil, nil)
	model.mode = ModeMenu

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.mode != ModeLogs {
		t.Fatalf("mode = %q, want %q", model.mode, ModeLogs)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.mode != ModeMenu {
		t.Fatalf("mode = %q, want %q", model.mode, ModeMenu)
	}
}

package tui

import (
	"strings"
	"testing"

	devruntime "github.com/zalberix/cactus/cli/internal/cmds/dev/runtime"
)

func TestRenderStatusShowsTerminalMarkers(t *testing.T) {
	model := NewModel(nil, nil, nil)

	cases := []struct {
		status devruntime.Status
		want   string
	}{
		{status: devruntime.StatusReady, want: "OK"},
		{status: devruntime.StatusFailed, want: "FAIL"},
		{status: devruntime.StatusSkipped, want: "skipped"},
	}
	for _, tc := range cases {
		got := model.renderStatus(devruntime.TargetState{Status: tc.status})
		if !strings.Contains(got, tc.want) {
			t.Fatalf("status %q rendered as %q, want marker %q", tc.status, got, tc.want)
		}
	}
}

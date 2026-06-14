package runtime

import (
	"bytes"
	"strings"
	"testing"
)

func TestCleanPortsTargetIsStartupTask(t *testing.T) {
	spec := CleanPortsTarget()
	if spec.ID != "clean-ports" || spec.Name != "clean-ports" {
		t.Fatalf("unexpected clean ports identity: %#v", spec)
	}
	if spec.Kind != TargetTask {
		t.Fatalf("kind = %q, want %q", spec.Kind, TargetTask)
	}
	if spec.Build == nil || spec.Ready == nil {
		t.Fatalf("clean ports target must include build and readiness: %#v", spec)
	}
}

func TestOptionalStartupTargetsCanBeSkipped(t *testing.T) {
	for _, spec := range []TargetSpec{
		DependenciesTarget(false),
		MigrationsTarget(false),
	} {
		if spec.Kind != TargetTask {
			t.Fatalf("%s kind = %q, want %q", spec.ID, spec.Kind, TargetTask)
		}
		if !spec.Skip {
			t.Fatalf("%s should be marked skipped", spec.ID)
		}
		if spec.Build != nil {
			t.Fatalf("%s should not define build when skipped", spec.ID)
		}
	}
}

func TestWriterReporterWritesLogLines(t *testing.T) {
	var buf bytes.Buffer
	reporter := writerReporter{w: &buf}

	reporter.Infof("hello %s", "info")
	reporter.Warnf("hello %s", "warn")
	reporter.Successf("hello %s", "success")

	got := buf.String()
	for _, want := range []string{"hello info\n", "warning: hello warn\n", "hello success\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

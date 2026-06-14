package helpers

import (
	"context"
	"errors"
	"testing"
)

func TestCleanPortsHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := CleanPorts(ctx, noopPortReporter{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

type noopPortReporter struct{}

func (noopPortReporter) Infof(string, ...any) {}

func (noopPortReporter) Warnf(string, ...any) {}

func (noopPortReporter) Successf(string, ...any) {}

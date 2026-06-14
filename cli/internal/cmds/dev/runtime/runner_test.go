package runtime

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRunnerStartsTargetsAfterDependencies(t *testing.T) {
	var started []string
	events := make(chan Event, 50)

	target := func(id string, deps ...string) TargetSpec {
		return TargetSpec{
			ID:              id,
			Name:            id,
			Kind:            TargetProcess,
			DependsOn:       deps,
			Ready:           ImmediateReady(),
			GracefulTimeout: 50 * time.Millisecond,
		}
	}

	runner := NewRunner([]TargetSpec{
		target("manager"),
		target("worker", "manager"),
	}, events)
	runner.processFactory = func(spec TargetSpec, events chan<- Event) managedProcess {
		return &fakeManagedProcess{
			start: func(context.Context) error {
				started = append(started, spec.ID)
				events <- StatusEvent(spec.ID, StatusReady, "ready")
				return nil
			},
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := runner.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runner.Shutdown(context.Background())

	want := []string{"manager", "worker"}
	if !reflect.DeepEqual(started, want) {
		t.Fatalf("started order = %#v, want %#v", started, want)
	}
}

func TestRunnerShutdownStopsProcessesInReverseOrder(t *testing.T) {
	events := make(chan Event, 100)
	target := func(id string) TargetSpec {
		return TargetSpec{
			ID:              id,
			Name:            id,
			Kind:            TargetProcess,
			Ready:           ImmediateReady(),
			GracefulTimeout: 50 * time.Millisecond,
		}
	}
	runner := NewRunner([]TargetSpec{target("manager"), target("worker")}, events)
	runner.processFactory = func(spec TargetSpec, events chan<- Event) managedProcess {
		return &fakeManagedProcess{
			start: func(context.Context) error {
				events <- StatusEvent(spec.ID, StatusReady, "ready")
				return nil
			},
			stop: func(context.Context) error {
				events <- StatusEvent(spec.ID, StatusStopping, "stopping")
				events <- StatusEvent(spec.ID, StatusStopped, "stopped")
				return nil
			},
		}
	}

	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	drainEvents(events)

	if err := runner.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}

	var stopping []string
	deadline := time.After(time.Second)
	for len(stopping) < 2 {
		select {
		case ev := <-events:
			if ev.Type == EventStatus && ev.Status == StatusStopping {
				stopping = append(stopping, ev.TargetID)
			}
		case <-deadline:
			t.Fatalf("stopping order = %#v, want worker then manager", stopping)
		}
	}

	want := []string{"worker", "manager"}
	if !reflect.DeepEqual(stopping, want) {
		t.Fatalf("stopping order = %#v, want %#v", stopping, want)
	}
}

func TestRunnerRestartsWatchedTargetWithDebounce(t *testing.T) {
	events := make(chan Event, 100)
	changes := make(chan string, 10)
	var restarts int
	var stopped []string
	fakeWatcher := &fakeTargetWatcher{changes: changes}

	spec := TargetSpec{
		ID:                  "manager",
		Name:                "manager",
		Kind:                TargetProcess,
		Ready:               ImmediateReady(),
		RestartOnFileChange: true,
		WatchPaths:          []string{"apps/manager"},
	}
	runner := NewRunner([]TargetSpec{spec}, events)
	runner.watcherFactory = func() targetWatcher {
		return fakeWatcher
	}
	runner.processFactory = func(spec TargetSpec, events chan<- Event) managedProcess {
		return &fakeManagedProcess{
			start: func(context.Context) error {
				events <- StatusEvent(spec.ID, StatusReady, "ready")
				return nil
			},
			restart: func(context.Context) error {
				restarts++
				return nil
			},
			stop: func(context.Context) error {
				stopped = append(stopped, spec.ID)
				return nil
			},
		}
	}

	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	changes <- "apps/manager/main.go"
	changes <- "apps/manager/other.go"

	deadline := time.After(time.Second)
	for restarts == 0 {
		select {
		case <-deadline:
			t.Fatal("expected watched file change to restart target")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	time.Sleep(100 * time.Millisecond)
	if restarts != 1 {
		t.Fatalf("restarts = %d, want 1", restarts)
	}

	var sawLog bool
	drain := true
	for drain {
		select {
		case ev := <-events:
			if ev.Type == EventLog && ev.TargetID == "manager" && ev.Line != nil && ev.Line.Line == "file changed: apps/manager/main.go" {
				sawLog = true
			}
		default:
			drain = false
		}
	}
	if !sawLog {
		t.Fatal("expected file changed system log")
	}

	if err := runner.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fakeWatcher.stops != 1 {
		t.Fatalf("watcher stops = %d, want 1", fakeWatcher.stops)
	}
	if !reflect.DeepEqual(stopped, []string{"manager"}) {
		t.Fatalf("stopped = %#v, want manager", stopped)
	}
}

func drainEvents(events <-chan Event) {
	for {
		select {
		case <-events:
		default:
			return
		}
	}
}

type fakeManagedProcess struct {
	start   func(context.Context) error
	restart func(context.Context) error
	stop    func(context.Context) error
}

func (p *fakeManagedProcess) Start(ctx context.Context) error {
	if p.start == nil {
		return nil
	}
	return p.start(ctx)
}

func (p *fakeManagedProcess) Restart(ctx context.Context) error {
	if p.restart == nil {
		return nil
	}
	return p.restart(ctx)
}

func (p *fakeManagedProcess) Stop(ctx context.Context) error {
	if p.stop == nil {
		return nil
	}
	return p.stop(ctx)
}

func (p *fakeManagedProcess) PID() int {
	return 0
}

type fakeTargetWatcher struct {
	changes chan string
	stops   int
}

func (w *fakeTargetWatcher) StartEvents(string) (chan string, error) {
	return w.changes, nil
}

func (w *fakeTargetWatcher) Stop() {
	w.stops++
}

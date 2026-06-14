package runtime

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zalberix/cactus/cli/internal/watcher"
)

type Runner struct {
	specs          []TargetSpec
	events         chan<- Event
	processes      map[string]managedProcess
	processFactory func(TargetSpec, chan<- Event) managedProcess
	watchers       map[string][]targetWatcher
	watcherFactory func() targetWatcher
	watchDebounce  time.Duration
	order          []string
	mu             sync.Mutex
}

type managedProcess interface {
	Start(context.Context) error
	Restart(context.Context) error
	Stop(context.Context) error
	PID() int
}

type targetWatcher interface {
	StartEvents(path string) (chan string, error)
	Stop()
}

func NewRunner(specs []TargetSpec, events chan<- Event) *Runner {
	copied := make([]TargetSpec, len(specs))
	copy(copied, specs)
	return &Runner{
		specs:     copied,
		events:    events,
		processes: make(map[string]managedProcess),
		watchers:  make(map[string][]targetWatcher),
		processFactory: func(spec TargetSpec, events chan<- Event) managedProcess {
			return NewManagedProcess(spec, events)
		},
		watcherFactory: func() targetWatcher {
			return watcher.New()
		},
		watchDebounce: time.Second,
	}
}

func (r *Runner) Start(ctx context.Context) error {
	sorted, err := sortTargetSpecs(r.specs)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.order = targetIDs(sorted)
	r.mu.Unlock()

	for _, spec := range sorted {
		r.emit(StatusEvent(spec.ID, StatusPending, "pending"))
	}

	for _, spec := range sorted {
		if err := ctx.Err(); err != nil {
			return err
		}
		if spec.Skip {
			r.emit(StatusEvent(spec.ID, StatusSkipped, "skipped"))
			continue
		}

		switch spec.Kind {
		case TargetTask:
			if err := r.runTask(ctx, spec); err != nil {
				return err
			}
		case TargetProcess:
			process := r.processFactory(spec, r.events)
			if err := process.Start(ctx); err != nil {
				return err
			}
			r.mu.Lock()
			r.processes[spec.ID] = process
			r.mu.Unlock()
			if err := r.startWatchers(ctx, spec); err != nil {
				return err
			}
		default:
			return fmt.Errorf("target %q has unknown kind %q", spec.ID, spec.Kind)
		}
	}
	return nil
}

func (r *Runner) Restart(ctx context.Context, targetID string) error {
	spec, ok := r.specByID(targetID)
	if !ok {
		return fmt.Errorf("target %q not found", targetID)
	}
	if spec.Kind != TargetProcess {
		err := fmt.Errorf("target %s is not restartable", targetID)
		r.emit(LogEvent(targetID, StreamSystem, err.Error()))
		return err
	}

	r.mu.Lock()
	process := r.processes[targetID]
	r.mu.Unlock()
	if process == nil {
		return fmt.Errorf("target %q is not running", targetID)
	}

	if dependents := r.dependentsOf(targetID); len(dependents) > 0 {
		r.emit(LogEvent(targetID, StreamSystem, "warning: dependent targets stay running: "+strings.Join(dependents, ", ")))
	}
	return process.Restart(ctx)
}

func (r *Runner) Shutdown(ctx context.Context) error {
	r.stopWatchers()

	r.mu.Lock()
	order := append([]string(nil), r.order...)
	processes := make(map[string]managedProcess, len(r.processes))
	for id, process := range r.processes {
		processes[id] = process
	}
	r.mu.Unlock()

	var firstErr error
	for i := len(order) - 1; i >= 0; i-- {
		process := processes[order[i]]
		if process == nil {
			continue
		}
		if err := process.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (r *Runner) Specs() []TargetSpec {
	copied := make([]TargetSpec, len(r.specs))
	copy(copied, r.specs)
	return copied
}

func (r *Runner) runTask(ctx context.Context, spec TargetSpec) error {
	status := StatusBuilding
	if spec.ID == "clean-ports" {
		status = StatusCleaning
	}
	r.emit(StatusEvent(spec.ID, status, string(status)))

	stdout := newEventWriter(spec.ID, StreamStdout, r.events)
	stderr := newEventWriter(spec.ID, StreamStderr, r.events)
	if spec.Build != nil {
		if err := spec.Build(ctx, stdout, stderr); err != nil {
			stdout.Flush()
			stderr.Flush()
			r.emitFailed(spec.ID, err)
			return err
		}
	}
	stdout.Flush()
	stderr.Flush()
	r.emit(StatusEvent(spec.ID, StatusReady, "ready"))
	return nil
}

func (r *Runner) startWatchers(ctx context.Context, spec TargetSpec) error {
	if !spec.RestartOnFileChange || len(spec.WatchPaths) == 0 {
		return nil
	}

	for _, path := range spec.WatchPaths {
		w := r.watcherFactory()
		changes, err := w.StartEvents(path)
		if err != nil {
			return fmt.Errorf("watch %s: %w", path, err)
		}

		r.mu.Lock()
		r.watchers[spec.ID] = append(r.watchers[spec.ID], w)
		r.mu.Unlock()

		go r.watchTarget(ctx, spec.ID, changes)
	}
	return nil
}

func (r *Runner) watchTarget(ctx context.Context, targetID string, changes <-chan string) {
	var lastRestart time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case path, ok := <-changes:
			if !ok {
				return
			}
			now := time.Now()
			if !lastRestart.IsZero() && now.Sub(lastRestart) < r.watchDebounce {
				continue
			}
			lastRestart = now

			r.emit(LogEvent(targetID, StreamSystem, "file changed: "+path))
			restartCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			if err := r.Restart(restartCtx, targetID); err != nil {
				r.emit(LogEvent(targetID, StreamSystem, "restart failed: "+err.Error()))
			}
			cancel()
		}
	}
}

func (r *Runner) stopWatchers() {
	r.mu.Lock()
	watchersByTarget := r.watchers
	r.watchers = make(map[string][]targetWatcher)
	r.mu.Unlock()

	for _, watchers := range watchersByTarget {
		for _, w := range watchers {
			w.Stop()
		}
	}
}

func (r *Runner) specByID(targetID string) (TargetSpec, bool) {
	for _, spec := range r.specs {
		if spec.ID == targetID {
			return spec, true
		}
	}
	return TargetSpec{}, false
}

func (r *Runner) dependentsOf(targetID string) []string {
	var dependents []string
	for _, spec := range r.specs {
		for _, dep := range spec.DependsOn {
			if dep == targetID {
				dependents = append(dependents, spec.ID)
				break
			}
		}
	}
	sort.Strings(dependents)
	return dependents
}

func (r *Runner) emitFailed(targetID string, err error) {
	ev := StatusEvent(targetID, StatusFailed, err.Error())
	ev.Err = err
	r.emit(ev)
	r.emit(ExitEvent(targetID, err))
}

func (r *Runner) emit(ev Event) {
	if r.events == nil {
		return
	}
	r.events <- ev
}

func sortTargetSpecs(specs []TargetSpec) ([]TargetSpec, error) {
	byID := make(map[string]TargetSpec, len(specs))
	for _, spec := range specs {
		if spec.ID == "" {
			return nil, fmt.Errorf("target id is required")
		}
		if _, exists := byID[spec.ID]; exists {
			return nil, fmt.Errorf("duplicate target id: %q", spec.ID)
		}
		byID[spec.ID] = spec
	}

	for _, spec := range specs {
		for _, dep := range spec.DependsOn {
			if _, exists := byID[dep]; !exists {
				return nil, fmt.Errorf("target %q depends on unknown target %q", spec.ID, dep)
			}
		}
	}

	state := make(map[string]int, len(specs))
	var result []TargetSpec
	var visit func(string, []string) error
	visit = func(id string, path []string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("circular target dependency: %s -> %s", strings.Join(path, " -> "), id)
		case 2:
			return nil
		}

		state[id] = 1
		spec := byID[id]
		for _, dep := range spec.DependsOn {
			if err := visit(dep, append(path, id)); err != nil {
				return err
			}
		}
		state[id] = 2
		result = append(result, spec)
		return nil
	}

	for _, spec := range specs {
		if err := visit(spec.ID, nil); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func targetIDs(specs []TargetSpec) []string {
	ids := make([]string, 0, len(specs))
	for _, spec := range specs {
		ids = append(ids, spec.ID)
	}
	return ids
}

var _ io.Writer = (*eventWriter)(nil)

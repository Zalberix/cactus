package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	goruntime "runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const defaultGracefulTimeout = 5 * time.Second
const waitAfterKillTimeout = 2 * time.Second

type ManagedProcess struct {
	spec     TargetSpec
	events   chan<- Event
	mu       sync.Mutex
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	done     chan error
	stopping bool
}

func NewManagedProcess(spec TargetSpec, events chan<- Event) *ManagedProcess {
	return &ManagedProcess{spec: spec, events: events}
}

func (p *ManagedProcess) Start(parent context.Context) error {
	stdout := newEventWriter(p.spec.ID, StreamStdout, p.events)
	stderr := newEventWriter(p.spec.ID, StreamStderr, p.events)
	ctx, cancel := context.WithCancel(parent)

	if p.spec.Build != nil {
		p.emit(StatusEvent(p.spec.ID, StatusBuilding, "building"))
		if err := p.spec.Build(ctx, stdout, stderr); err != nil {
			cancel()
			p.emitFailed(err)
			return err
		}
		stdout.Flush()
		stderr.Flush()
	}

	if p.spec.Command == nil {
		cancel()
		err := fmt.Errorf("target %q has no command", p.spec.ID)
		p.emitFailed(err)
		return err
	}

	p.emit(StatusEvent(p.spec.ID, StatusStarting, "starting"))
	cmd, err := p.spec.Command(ctx, stdout, stderr)
	if err != nil {
		cancel()
		p.emitFailed(err)
		return err
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		cancel()
		p.emitFailed(err)
		return err
	}

	done := make(chan error, 1)
	p.mu.Lock()
	p.cmd = cmd
	p.cancel = cancel
	p.done = done
	p.stopping = false
	p.mu.Unlock()

	go p.wait(cmd, done, stdout, stderr)

	ready := p.spec.Ready
	if ready == nil {
		ready = ImmediateReady()
	}
	if err := ready(ctx); err != nil {
		p.emitFailed(err)
		return err
	}

	p.emit(StatusEvent(p.spec.ID, StatusReady, "ready"))
	return nil
}

func (p *ManagedProcess) Restart(parent context.Context) error {
	p.emit(StatusEvent(p.spec.ID, StatusRestarting, "restarting"))
	if err := p.Stop(parent); err != nil {
		return err
	}
	return p.Start(parent)
}

func (p *ManagedProcess) Stop(parent context.Context) error {
	p.mu.Lock()
	cmd := p.cmd
	cancel := p.cancel
	done := p.done
	if cmd == nil || cmd.Process == nil {
		p.mu.Unlock()
		p.emit(StatusEvent(p.spec.ID, StatusStopped, "stopped"))
		return nil
	}
	p.stopping = true
	p.mu.Unlock()

	p.emit(StatusEvent(p.spec.ID, StatusStopping, "stopping"))

	_ = terminateProcess(cmd)

	timeout := p.spec.GracefulTimeout
	if timeout <= 0 {
		timeout = defaultGracefulTimeout
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-parent.Done():
		killErr := killProcessTree(parent, cmd)
		if cancel != nil {
			cancel()
		}
		if killErr != nil {
			return killErr
		}
		p.emit(StatusEvent(p.spec.ID, StatusStopped, "stopped"))
		return parent.Err()
	case <-timer.C:
		p.emit(LogEvent(p.spec.ID, StreamSystem, "graceful stop timed out; killing process tree"))
		if err := killProcessTree(parent, cmd); err != nil {
			return err
		}
		if cancel != nil {
			cancel()
		}
		select {
		case <-done:
		case <-time.After(waitAfterKillTimeout):
			return fmt.Errorf("target %q did not exit after kill", p.spec.ID)
		}
	case <-done:
		if cancel != nil {
			cancel()
		}
	}

	p.emit(StatusEvent(p.spec.ID, StatusStopped, "stopped"))
	return nil
}

func (p *ManagedProcess) PID() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p *ManagedProcess) wait(cmd *exec.Cmd, done chan<- error, stdout *eventWriter, stderr *eventWriter) {
	err := cmd.Wait()
	stdout.Flush()
	stderr.Flush()

	p.mu.Lock()
	stopping := p.stopping
	if p.cmd == cmd {
		p.cmd = nil
		p.cancel = nil
		p.done = nil
		p.stopping = false
	}
	p.mu.Unlock()

	done <- err
	close(done)

	if stopping {
		return
	}
	if err != nil {
		p.emitFailed(err)
		return
	}
	p.emit(ExitEvent(p.spec.ID, nil))
}

func (p *ManagedProcess) emitFailed(err error) {
	ev := StatusEvent(p.spec.ID, StatusFailed, err.Error())
	ev.Err = err
	p.emit(ev)
	p.emit(ExitEvent(p.spec.ID, err))
}

func (p *ManagedProcess) emit(ev Event) {
	if p.events == nil {
		return
	}
	p.events <- ev
}

func terminateProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	if goruntime.GOOS == "windows" {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	return cmd.Process.Signal(syscall.SIGTERM)
}

func killProcessTree(ctx context.Context, cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	if goruntime.GOOS == "windows" {
		kill := exec.CommandContext(ctx, "taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid))
		if err := kill.Run(); err != nil {
			killErr := cmd.Process.Kill()
			if killErr == nil || errors.Is(killErr, os.ErrProcessDone) {
				return nil
			}
			return fmt.Errorf("%w; process kill failed: %w", err, killErr)
		}
	}
	err := cmd.Process.Kill()
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	return err
}

type eventWriter struct {
	targetID string
	stream   Stream
	events   chan<- Event
	mu       sync.Mutex
	buffer   bytes.Buffer
}

func newEventWriter(targetID string, stream Stream, events chan<- Event) *eventWriter {
	return &eventWriter{
		targetID: targetID,
		stream:   stream,
		events:   events,
	}
}

func (w *eventWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer.Write(p)
	for {
		line, err := w.buffer.ReadString('\n')
		if err != nil {
			w.buffer.WriteString(line)
			return len(p), nil
		}
		w.emitLine(line)
	}
}

func (w *eventWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buffer.Len() == 0 {
		return
	}
	w.emitLine(w.buffer.String())
	w.buffer.Reset()
}

func (w *eventWriter) emitLine(line string) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return
	}
	if w.events != nil {
		w.events <- LogEvent(w.targetID, w.stream, line)
	}
}

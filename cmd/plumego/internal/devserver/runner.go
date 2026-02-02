package devserver

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spcent/plumego/pubsub"
)

// AppRunner manages the user application lifecycle
type AppRunner struct {
	dir        string
	binaryPath string
	cmd        string
	args       []string
	env        []string

	process *os.Process
	cancel  context.CancelFunc
	mu      sync.Mutex

	pubsub  *pubsub.InProcPubSub
	running bool
}

// NewAppRunner creates a new application runner
func NewAppRunner(dir string, ps *pubsub.InProcPubSub) *AppRunner {
	return &AppRunner{
		dir:        dir,
		binaryPath: filepath.Join(dir, ".dev-server"),
		pubsub:     ps,
		env:        os.Environ(),
	}
}

// SetCustomCommand sets a custom run command
func (r *AppRunner) SetCustomCommand(cmd string, args []string) {
	r.cmd = cmd
	r.args = args
}

// SetEnv adds environment variables
func (r *AppRunner) SetEnv(key, value string) {
	r.env = append(r.env, fmt.Sprintf("%s=%s", key, value))
}

// IsRunning returns whether the app is currently running
func (r *AppRunner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// Start starts the application
func (r *AppRunner) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("application already running")
	}
	r.mu.Unlock()

	// Publish start event
	r.pubsub.Publish(EventAppStart, pubsub.Message{
		Topic: EventAppStart,
		Data: AppLifecycleEvent{
			State: "starting",
		},
	})

	// Create command
	var cmd *exec.Cmd
	if r.cmd != "" {
		cmd = exec.CommandContext(ctx, r.cmd, r.args...)
	} else {
		cmd = exec.CommandContext(ctx, r.binaryPath)
	}

	cmd.Dir = r.dir
	cmd.Env = r.env

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		r.pubsub.Publish(EventAppStart, pubsub.Message{
			Topic: EventAppStart,
			Data: AppLifecycleEvent{
				State: "crashed",
				Error: err.Error(),
			},
		})
		return fmt.Errorf("failed to start: %w", err)
	}

	r.mu.Lock()
	r.process = cmd.Process
	r.running = true
	r.mu.Unlock()

	// Create cancel context
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	// Stream output
	go r.streamOutput(ctx, stdout, "stdout")
	go r.streamOutput(ctx, stderr, "stderr")

	// Wait for process in background
	go func() {
		err := cmd.Wait()

		r.mu.Lock()
		r.running = false
		r.process = nil
		r.mu.Unlock()

		if err != nil {
			r.pubsub.Publish(EventAppStop, pubsub.Message{
				Topic: EventAppStop,
				Data: AppLifecycleEvent{
					State: "crashed",
					Error: err.Error(),
				},
			})
		} else {
			r.pubsub.Publish(EventAppStop, pubsub.Message{
				Topic: EventAppStop,
				Data: AppLifecycleEvent{
					State: "stopped",
				},
			})
		}
	}()

	// Publish running event
	r.pubsub.Publish(EventAppStart, pubsub.Message{
		Topic: EventAppStart,
		Data: AppLifecycleEvent{
			State: "running",
			PID:   cmd.Process.Pid,
		},
	})

	return nil
}

// Stop stops the application gracefully
func (r *AppRunner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running || r.process == nil {
		return nil
	}

	// Cancel context to stop log streaming
	if r.cancel != nil {
		r.cancel()
	}

	// Try graceful shutdown first (SIGTERM)
	if err := r.process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead
		return nil
	}

	// Wait for graceful shutdown with timeout
	done := make(chan bool, 1)
	go func() {
		r.process.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Process exited gracefully
		return nil
	case <-time.After(5 * time.Second):
		// Force kill after timeout
		if err := r.process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		r.process.Wait()
		return nil
	}
}

// Restart restarts the application
func (r *AppRunner) Restart(ctx context.Context) error {
	r.pubsub.Publish(EventAppRestart, pubsub.Message{
		Topic: EventAppRestart,
		Data: AppLifecycleEvent{
			State: "restarting",
		},
	})

	if err := r.Stop(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}

	// Small delay to ensure clean shutdown
	time.Sleep(500 * time.Millisecond)

	return r.Start(ctx)
}

// streamOutput captures and publishes application output
func (r *AppRunner) streamOutput(ctx context.Context, reader io.Reader, source string) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()

			// Determine log level from content
			level := "info"
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "fatal") {
				level = "error"
			} else if strings.Contains(lowerLine, "warn") {
				level = "warn"
			} else if strings.Contains(lowerLine, "debug") {
				level = "debug"
			}

			// Publish log event
			r.pubsub.Publish(EventAppLog, pubsub.Message{
				Topic: EventAppLog,
				Data: LogEvent{
					Level:   level,
					Message: line,
					Source:  source,
				},
			})

			// Also print to console for backward compatibility
			if source == "stderr" {
				fmt.Fprintln(os.Stderr, line)
			} else {
				fmt.Println(line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		r.pubsub.Publish(EventAppError, pubsub.Message{
			Topic: EventAppError,
			Data: LogEvent{
				Level:   "error",
				Message: fmt.Sprintf("Error reading %s: %v", source, err),
				Source:  source,
			},
		})
	}
}

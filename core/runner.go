package core

import (
	"context"
)

// Runner defines a background task lifecycle.
type Runner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Register registers a runner to participate in the app lifecycle.
// Runners start before the HTTP server and stop after it shuts down.
func (a *App) Register(runner Runner) error {
	if runner == nil {
		return nil
	}
	if err := a.ensureMutable("register_runner", "register runner"); err != nil {
		return err
	}
	a.mu.Lock()
	a.runners = append(a.runners, runner)
	a.mu.Unlock()
	return nil
}

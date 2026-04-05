package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/log"
)

// Prepare freezes routing/middleware state and constructs the HTTP server.
func (a *App) Prepare() error {
	if err := a.ensureServerPrepared(); err != nil {
		return err
	}

	return nil
}

// Server returns the prepared HTTP server instance.
func (a *App) Server() (*http.Server, error) {
	a.mu.RLock()
	server := a.httpServer
	a.mu.RUnlock()

	if server == nil {
		return nil, wrapCoreError(fmt.Errorf("server not prepared"), "get_server", nil)
	}
	return server, nil
}

// Start starts logger lifecycle hooks and marks the runtime started.
func (a *App) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.mu.RLock()
	if a.started {
		a.mu.RUnlock()
		return wrapCoreError(fmt.Errorf("app already started"), "start_app", nil)
	}
	prepared := a.httpServer != nil
	logger := a.logger
	a.mu.RUnlock()

	if !prepared {
		return wrapCoreError(fmt.Errorf("app not prepared"), "start_app", nil)
	}

	if lifecycle, ok := logger.(log.Lifecycle); ok {
		if err := lifecycle.Start(ctx); err != nil {
			return wrapCoreError(err, "start_app", nil)
		}
		a.mu.Lock()
		a.loggerStarted = true
		a.mu.Unlock()
	}

	a.mu.Lock()
	a.started = true
	a.mu.Unlock()

	return nil
}

// Shutdown gracefully stops the HTTP server and all runtime hooks.
func (a *App) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.mu.RLock()
	httpServer := a.httpServer
	connTracker := a.connTracker
	a.mu.RUnlock()

	if connTracker != nil {
		go connTracker.drain(ctx)
	}

	var shutdownErr error
	if httpServer != nil {
		if err := httpServer.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Server shutdown error", log.Fields{"error": err})
			shutdownErr = wrapCoreError(err, "shutdown_app", nil)
		}
	}

	a.stopLoggerLifecycle(ctx)

	a.mu.Lock()
	a.started = false
	a.mu.Unlock()

	return shutdownErr
}

func (a *App) stopLoggerLifecycle(ctx context.Context) {
	a.mu.RLock()
	started := a.loggerStarted
	logger := a.logger
	a.mu.RUnlock()
	if !started {
		return
	}
	if lifecycle, ok := logger.(log.Lifecycle); ok {
		_ = lifecycle.Stop(ctx)
	}
	a.mu.Lock()
	a.loggerStarted = false
	a.mu.Unlock()
}

type connectionTracker struct {
	active   atomic.Int64         // Number of active connections
	logger   log.StructuredLogger // Logger for connection tracking
	interval time.Duration        // Interval at which to log active connections
}

func newConnectionTracker(logger log.StructuredLogger, interval time.Duration) *connectionTracker {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &connectionTracker{logger: logger, interval: interval}
}

func (t *connectionTracker) track(_ net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		t.active.Add(1)
	case http.StateHijacked, http.StateClosed:
		t.active.Add(-1)
	}
}

func (t *connectionTracker) drain(ctx context.Context) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		if t.active.Load() <= 0 {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if t.logger != nil {
				t.logger.Info("draining active connections", log.Fields{"active_connections": t.active.Load()})
			}
		}
	}
}

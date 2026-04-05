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

	a.mu.RLock()
	started := a.started
	logger := a.logger
	a.mu.RUnlock()

	if started {
		return nil
	}

	if lifecycle, ok := logger.(log.Lifecycle); ok {
		if err := lifecycle.Start(context.Background()); err != nil {
			return wrapCoreError(err, "prepare_app", nil)
		}
	}

	a.mu.Lock()
	a.started = true
	a.mu.Unlock()

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

// Shutdown gracefully stops the HTTP server and all runtime hooks.
func (a *App) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.mu.RLock()
	httpServer := a.httpServer
	connTracker := a.connTracker
	started := a.started
	logger := a.logger
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

	if started {
		if lifecycle, ok := logger.(log.Lifecycle); ok {
			_ = lifecycle.Stop(ctx)
		}
	}

	a.mu.Lock()
	a.started = false
	a.mu.Unlock()

	return shutdownErr
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

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
	return a.ensureServerPrepared()
}

// Server returns the prepared HTTP server instance.
func (a *App) Server() (*http.Server, error) {
	if a == nil {
		return nil, nilAppError("get_server", nil)
	}
	a.mu.RLock()
	server := a.httpServer
	_, initialized := a.stateAndInitializedLocked()
	a.mu.RUnlock()

	if !initialized {
		return nil, uninitializedAppError("get_server", nil)
	}
	if server == nil {
		return nil, wrapCoreError(fmt.Errorf("server not prepared"), "get_server", nil)
	}
	return server, nil
}

// Shutdown gracefully stops the prepared HTTP server.
func (a *App) Shutdown(ctx context.Context) error {
	if a == nil {
		return nilAppError("shutdown_app", nil)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	a.mu.RLock()
	httpServer := a.httpServer
	connTracker := a.connTracker
	_, initialized := a.stateAndInitializedLocked()
	a.mu.RUnlock()

	if !initialized {
		return uninitializedAppError("shutdown_app", nil)
	}
	if connTracker != nil {
		go connTracker.drain(ctx)
	}

	var shutdownErr error
	if httpServer != nil {
		if err := httpServer.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			a.Logger().Error("Server shutdown error", log.Fields{"error": err})
			shutdownErr = wrapCoreError(err, "shutdown_app", nil)
		}
	}

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
		t.decrementActive()
	}
}

func (t *connectionTracker) decrementActive() {
	for {
		current := t.active.Load()
		if current <= 0 {
			return
		}
		if t.active.CompareAndSwap(current, current-1) {
			return
		}
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

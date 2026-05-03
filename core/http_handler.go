package core

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/spcent/plumego/contract"
)

// ensureHandlerPrepared performs the one-time transition required for using the
// app as an http.Handler. It freezes config/router state and builds the handler.
func (a *App) ensureHandlerPrepared() {
	if a == nil {
		return
	}
	a.handlerOnce.Do(func() {
		a.freezeConfig()
		r := a.ensureRouter()
		if r != nil {
			r.Freeze()
		}
		a.buildHandler()
	})
}

// ensureServerPrepared constructs the backing http.Server for the explicit
// Prepare/Server lifecycle path.
func (a *App) ensureServerPrepared() error {
	if a == nil {
		return nilAppError("prepare_server", nil)
	}
	a.mu.RLock()
	if a.httpServer != nil {
		a.mu.RUnlock()
		a.mu.Lock()
		a.preparationState = PreparationStateServerPrepared
		a.mu.Unlock()
		return nil
	}
	cfg, initialized := a.config, a.config != nil && a.router != nil && a.middlewareChain != nil
	if cfg == nil {
		a.mu.RUnlock()
		return uninitializedAppError("prepare_server", nil)
	}
	config := *cfg
	a.mu.RUnlock()

	if !initialized {
		return uninitializedAppError("prepare_server", nil)
	}

	tlsConfig, err := prepareTLSConfig(config.TLS)
	if err != nil {
		return wrapCoreError(err, "prepare_server", nil)
	}

	a.ensureHandlerPrepared()

	a.mu.RLock()
	if a.httpServer != nil {
		a.mu.RUnlock()
		a.mu.Lock()
		a.preparationState = PreparationStateServerPrepared
		a.mu.Unlock()
		return nil
	}
	handler := a.handler
	a.mu.RUnlock()

	if handler == nil {
		return wrapCoreError(fmt.Errorf("handler not configured"), "prepare_server", nil)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.httpServer != nil {
		a.preparationState = PreparationStateServerPrepared
		return nil
	}

	a.httpServer = &http.Server{
		Addr:              config.Addr,
		Handler:           handler,
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
		TLSConfig:         tlsConfig,
	}
	a.connTracker = newConnectionTracker(a.Logger(), config.DrainInterval)
	a.httpServer.ConnState = a.connTracker.track
	a.preparationState = PreparationStateServerPrepared

	if !config.HTTP2Enabled {
		a.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}

	return nil
}

func prepareTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("TLS enabled but certificate or key file not provided")
	}
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load tls certificate: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

// ServeHTTP allows App to be used directly with net/http servers.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a == nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeUnavailable).Message("app not configured").Build())
		return
	}
	a.mu.RLock()
	_, initialized := a.stateAndInitializedLocked()
	a.mu.RUnlock()
	if !initialized {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeUnavailable).Message("app not initialized").Build())
		return
	}
	a.ensureHandlerPrepared()

	a.mu.RLock()
	handler := a.handler
	a.mu.RUnlock()

	if handler == nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeUnavailable).Message("handler not configured").Build())
		return
	}

	handler.ServeHTTP(w, r)
}

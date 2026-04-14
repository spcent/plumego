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
	cfg := *a.config
	a.mu.RUnlock()

	if handler == nil {
		return wrapCoreError(fmt.Errorf("handler not configured"), "prepare_server", nil)
	}

	var tlsConfig *tls.Config
	if cfg.TLS.Enabled {
		if cfg.TLS.CertFile == "" || cfg.TLS.KeyFile == "" {
			return wrapCoreError(fmt.Errorf("TLS enabled but certificate or key file not provided"), "prepare_server", nil)
		}
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return wrapCoreError(fmt.Errorf("load tls certificate: %w", err), "prepare_server", nil)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.httpServer != nil {
		a.preparationState = PreparationStateServerPrepared
		return nil
	}

	a.httpServer = &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		TLSConfig:         tlsConfig,
	}
	a.connTracker = newConnectionTracker(a.logger, cfg.DrainInterval)
	a.httpServer.ConnState = a.connTracker.track
	a.preparationState = PreparationStateServerPrepared

	if !cfg.HTTP2Enabled {
		a.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}

	return nil
}

// ServeHTTP allows App to be used directly with net/http servers.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

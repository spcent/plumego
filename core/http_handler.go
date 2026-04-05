package core

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/spcent/plumego/contract"
)

// ensurePrepared performs the canonical one-time preparation transition shared
// by Prepare and ServeHTTP. It freezes config/router state, builds the handler,
// and constructs the backing http.Server.
func (a *App) ensurePrepared() error {
	a.handlerOnce.Do(func() {
		a.freezeConfig()
		r := a.ensureRouter()
		if r != nil {
			r.Freeze()
		}
		a.buildHandler()
	})

	a.mu.RLock()
	if a.httpServer != nil {
		a.mu.RUnlock()
		return nil
	}
	handler := a.handler
	cfg := a.serverRuntimeConfigLocked()
	a.mu.RUnlock()

	if handler == nil {
		return fmt.Errorf("handler not configured")
	}

	var tlsConfig *tls.Config
	if cfg.tls.Enabled {
		if cfg.tls.CertFile == "" || cfg.tls.KeyFile == "" {
			return fmt.Errorf("TLS enabled but certificate or key file not provided")
		}
		cert, err := tls.LoadX509KeyPair(cfg.tls.CertFile, cfg.tls.KeyFile)
		if err != nil {
			return fmt.Errorf("load tls certificate: %w", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.httpServer != nil {
		return nil
	}

	a.httpServer = &http.Server{
		Addr:              cfg.addr,
		Handler:           handler,
		ReadTimeout:       cfg.readTimeout,
		ReadHeaderTimeout: cfg.readHeaderTimeout,
		WriteTimeout:      cfg.writeTimeout,
		IdleTimeout:       cfg.idleTimeout,
		MaxHeaderBytes:    cfg.maxHeaderBytes,
		TLSConfig:         tlsConfig,
	}
	a.connTracker = newConnectionTracker(a.logger, cfg.drainInterval)
	a.httpServer.ConnState = a.connTracker.track

	if !cfg.enableHTTP2 {
		a.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}

	return nil
}

// ServeHTTP allows App to be used directly with net/http servers.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := a.ensurePrepared(); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Status(http.StatusServiceUnavailable).Code("handler_not_configured").Message(err.Error()).Build())
		return
	}

	a.mu.RLock()
	handler := a.handler
	a.mu.RUnlock()

	if handler == nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Status(http.StatusServiceUnavailable).Code("handler_not_configured").Message("handler not configured").Build())
		return
	}

	handler.ServeHTTP(w, r)
}

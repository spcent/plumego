package core

import "net/http"

// ensureHandler lazily builds the application's handler chain so App can satisfy http.Handler.
func (a *App) ensureHandler() {
	a.handlerOnce.Do(func() {
		a.freezeConfig()
		r := a.ensureRouter()
		a.ensureComponents()
		a.applyGuardrails()
		if r != nil {
			r.Freeze()
		}
		a.buildHandler()
	})
}

func (a *App) ensureComponents() {
	a.mu.RLock()
	mounted := a.componentsMounted
	a.mu.RUnlock()

	if mounted {
		return
	}

	a.mountComponents()
}

// ServeHTTP allows App to be used directly with net/http servers.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.ensureHandler()

	a.mu.RLock()
	handler := a.handler
	a.mu.RUnlock()

	if handler == nil {
		writeHTTPError(w, r, http.StatusServiceUnavailable, "handler_not_configured", "handler not configured")
		return
	}

	handler.ServeHTTP(w, r)
}

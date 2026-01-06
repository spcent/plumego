package core

import (
	"net/http"

	"github.com/spcent/plumego/router"
)

// ensureHandler lazily builds the application's handler chain so App can satisfy http.Handler.
func (a *App) ensureHandler() {
	a.handlerOnce.Do(func() {
		if a.router == nil {
			a.router = router.NewRouter()
			a.router.SetLogger(a.logger)
		}

		a.ensureComponents()
		a.applyGuardrails()
		a.router.Freeze()
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

	if a.handler == nil {
		http.Error(w, "handler not configured", http.StatusServiceUnavailable)
		return
	}

	a.handler.ServeHTTP(w, r)
}

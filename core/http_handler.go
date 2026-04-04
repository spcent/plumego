package core

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// ensureHandler lazily builds the application's handler chain so App can satisfy http.Handler.
func (a *App) ensureHandler() {
	a.handlerOnce.Do(func() {
		a.freezeConfig()
		r := a.ensureRouter()
		if r != nil {
			r.Freeze()
		}
		a.buildHandler()
	})
}

// ServeHTTP allows App to be used directly with net/http servers.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.ensureHandler()

	a.mu.RLock()
	handler := a.handler
	a.mu.RUnlock()

	if handler == nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Status(http.StatusServiceUnavailable).Code("handler_not_configured").Message("handler not configured").Build())
		return
	}

	handler.ServeHTTP(w, r)
}

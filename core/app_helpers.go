package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spcent/plumego/router"
)

const (
	operationPrepareServer = "prepare_server"
	operationGetServer     = "get_server"
	operationShutdownApp   = "shutdown_app"
	operationAddRoute      = "add_route"
	operationUseMiddleware = "use_middleware"
)

func (a *App) stateAndInitializedLocked() (PreparationState, bool) {
	return a.preparationState, a.config != nil && a.router != nil && a.middlewareChain != nil
}

func uninitializedAppError(operation string, params map[string]any) error {
	return wrapCoreError(fmt.Errorf("app not initialized"), operation, params)
}

func immutableAppError(operation, action string, params map[string]any) error {
	return wrapCoreError(fmt.Errorf("cannot %s after app has been prepared", action), operation, params)
}

func wrapCoreError(err error, operation string, params map[string]any) error {
	if err == nil {
		return nil
	}
	if len(params) == 0 {
		return fmt.Errorf("core %s: %w", operation, err)
	}
	return fmt.Errorf("core %s %s: %w", operation, formatErrorParams(params), err)
}

func formatErrorParams(params map[string]any) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, params[key]))
	}
	return strings.Join(parts, " ")
}

func (a *App) freezeConfig() {
	a.mu.Lock()
	if a.preparationState == PreparationStateMutable {
		a.preparationState = PreparationStateHandlerPrepared
	}
	a.mu.Unlock()
}

func (a *App) ensureRouter() *router.Router {
	a.mu.RLock()
	r := a.router
	a.mu.RUnlock()
	return r
}

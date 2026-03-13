package rest

import (
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

// RouteOptions controls the standard REST route surface.
type RouteOptions struct {
	EnableBatch   bool
	EnableHead    bool
	EnableOptions bool
}

// DefaultRouteOptions returns the canonical REST route surface.
func DefaultRouteOptions() RouteOptions {
	return RouteOptions{
		EnableBatch:   true,
		EnableHead:    true,
		EnableOptions: true,
	}
}

// RegisterResourceRoutes binds the canonical REST routes for a resource controller.
func RegisterResourceRoutes(r *router.Router, prefix string, controller ResourceController, opts RouteOptions) {
	if r == nil || controller == nil {
		return
	}
	prefix = normalizePrefix(prefix)

	r.Get(prefix, http.HandlerFunc(controller.Index))
	r.Get(prefix+"/:id", http.HandlerFunc(controller.Show))
	r.Post(prefix, http.HandlerFunc(controller.Create))
	r.Put(prefix+"/:id", http.HandlerFunc(controller.Update))
	r.Delete(prefix+"/:id", http.HandlerFunc(controller.Delete))
	r.Patch(prefix+"/:id", http.HandlerFunc(controller.Patch))

	if opts.EnableOptions {
		r.Options(prefix, http.HandlerFunc(controller.Options))
		r.Options(prefix+"/:id", http.HandlerFunc(controller.Options))
	}
	if opts.EnableHead {
		r.Head(prefix, http.HandlerFunc(controller.Head))
		r.Head(prefix+"/:id", http.HandlerFunc(controller.Head))
	}
	if opts.EnableBatch {
		r.Post(prefix+"/batch", http.HandlerFunc(controller.BatchCreate))
		r.Delete(prefix+"/batch", http.HandlerFunc(controller.BatchDelete))
	}
}

// RegisterContextResourceRoutes binds the canonical REST routes for a context-aware resource controller.
func RegisterContextResourceRoutes(r *router.Router, prefix string, controller ContextResourceController) {
	if r == nil || controller == nil {
		return
	}
	prefix = normalizePrefix(prefix)
	logger := r.Logger()

	r.Get(prefix, contract.AdaptCtxHandler(controller.IndexCtx, logger))
	r.Get(prefix+"/:id", contract.AdaptCtxHandler(controller.ShowCtx, logger))
	r.Post(prefix, contract.AdaptCtxHandler(controller.CreateCtx, logger))
	r.Put(prefix+"/:id", contract.AdaptCtxHandler(controller.UpdateCtx, logger))
	r.Delete(prefix+"/:id", contract.AdaptCtxHandler(controller.DeleteCtx, logger))
	r.Patch(prefix+"/:id", contract.AdaptCtxHandler(controller.PatchCtx, logger))
	r.Post(prefix+"/batch", contract.AdaptCtxHandler(controller.BatchCreateCtx, logger))
	r.Delete(prefix+"/batch", contract.AdaptCtxHandler(controller.BatchDeleteCtx, logger))
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "/"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return strings.TrimRight(prefix, "/")
}

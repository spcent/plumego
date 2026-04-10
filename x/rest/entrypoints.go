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
func RegisterResourceRoutes(r *router.Router, prefix string, controller ResourceController, opts RouteOptions) error {
	if r == nil || controller == nil {
		return nil
	}
	prefix = normalizePrefix(prefix)

	if err := r.AddRoute(http.MethodGet, prefix, http.HandlerFunc(controller.Index)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, prefix+"/:id", http.HandlerFunc(controller.Show)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPost, prefix, http.HandlerFunc(controller.Create)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPut, prefix+"/:id", http.HandlerFunc(controller.Update)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodDelete, prefix+"/:id", http.HandlerFunc(controller.Delete)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPatch, prefix+"/:id", http.HandlerFunc(controller.Patch)); err != nil {
		return err
	}

	if opts.EnableOptions {
		if err := r.AddRoute(http.MethodOptions, prefix, http.HandlerFunc(controller.Options)); err != nil {
			return err
		}
		if err := r.AddRoute(http.MethodOptions, prefix+"/:id", http.HandlerFunc(controller.Options)); err != nil {
			return err
		}
	}
	if opts.EnableHead {
		if err := r.AddRoute(http.MethodHead, prefix, http.HandlerFunc(controller.Head)); err != nil {
			return err
		}
		if err := r.AddRoute(http.MethodHead, prefix+"/:id", http.HandlerFunc(controller.Head)); err != nil {
			return err
		}
	}
	if opts.EnableBatch {
		if err := r.AddRoute(http.MethodPost, prefix+"/batch", http.HandlerFunc(controller.BatchCreate)); err != nil {
			return err
		}
		if err := r.AddRoute(http.MethodDelete, prefix+"/batch", http.HandlerFunc(controller.BatchDelete)); err != nil {
			return err
		}
	}
	return nil
}

// RegisterContextResourceRoutes binds the canonical REST routes for a context-aware resource controller.
func RegisterContextResourceRoutes(r *router.Router, prefix string, controller ContextResourceController) error {
	if r == nil || controller == nil {
		return nil
	}
	prefix = normalizePrefix(prefix)

	if err := r.AddRoute(http.MethodGet, prefix, contract.AdaptCtxHandler(controller.IndexCtx)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, prefix+"/:id", contract.AdaptCtxHandler(controller.ShowCtx)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPost, prefix, contract.AdaptCtxHandler(controller.CreateCtx)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPut, prefix+"/:id", contract.AdaptCtxHandler(controller.UpdateCtx)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodDelete, prefix+"/:id", contract.AdaptCtxHandler(controller.DeleteCtx)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPatch, prefix+"/:id", contract.AdaptCtxHandler(controller.PatchCtx)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPost, prefix+"/batch", contract.AdaptCtxHandler(controller.BatchCreateCtx)); err != nil {
		return err
	}
	return r.AddRoute(http.MethodDelete, prefix+"/batch", contract.AdaptCtxHandler(controller.BatchDeleteCtx))
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

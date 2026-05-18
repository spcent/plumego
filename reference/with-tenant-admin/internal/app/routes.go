package app

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/reference/with-tenant-admin/internal/auth"
)

func (a *App) RegisterRoutes() error {
	admin := auth.RequireAdminToken(a.Cfg.AdminToken)
	if err := a.registerTenantAdminRoutes(admin); err != nil {
		return err
	}
	if err := a.registerQuotaAdminRoutes(admin); err != nil {
		return err
	}
	return a.registerUsageAdminRoutes(admin)
}

func (a *App) registerTenantAdminRoutes(admin func(http.Handler) http.Handler) error {
	return a.Core.Get("/admin/tenants", admin(http.HandlerFunc(notImplemented("tenant admin handlers are added in card 1581"))))
}

func (a *App) registerQuotaAdminRoutes(admin func(http.Handler) http.Handler) error {
	return a.Core.Get("/admin/quota", admin(http.HandlerFunc(notImplemented("quota admin handlers are added in card 1582"))))
}

func (a *App) registerUsageAdminRoutes(admin func(http.Handler) http.Handler) error {
	return a.Core.Post("/admin/usage", admin(http.HandlerFunc(notImplemented("usage recording handlers are added in card 1583"))))
}

func notImplemented(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotImplemented).
			Message(message).
			Build())
	}
}

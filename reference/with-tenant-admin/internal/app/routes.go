package app

import (
	"net/http"

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
	if err := a.Core.Post("/admin/tenants", admin(http.HandlerFunc(a.Tenants.CreateTenant))); err != nil {
		return err
	}
	if err := a.Core.Get("/admin/tenants/:id", admin(http.HandlerFunc(a.Tenants.GetTenant))); err != nil {
		return err
	}
	if err := a.Core.Post("/admin/tenants/:id/suspend", admin(http.HandlerFunc(a.Tenants.SuspendTenant))); err != nil {
		return err
	}
	return a.Core.Delete("/admin/tenants/:id", admin(http.HandlerFunc(a.Tenants.DeleteTenant)))
}

func (a *App) registerQuotaAdminRoutes(admin func(http.Handler) http.Handler) error {
	if err := a.Core.Get("/admin/quota/:tenantID", admin(http.HandlerFunc(a.Quotas.GetQuota))); err != nil {
		return err
	}
	if err := a.Core.Put("/admin/quota/:tenantID", admin(http.HandlerFunc(a.Quotas.SetQuota))); err != nil {
		return err
	}
	return a.Core.Post("/admin/quota/:tenantID/reset", admin(http.HandlerFunc(a.Quotas.ResetQuota)))
}

func (a *App) registerUsageAdminRoutes(admin func(http.Handler) http.Handler) error {
	if err := a.Core.Post("/admin/usage/:tenantID", admin(http.HandlerFunc(a.Usage.RecordUsage))); err != nil {
		return err
	}
	return a.Core.Get("/admin/usage/:tenantID", admin(http.HandlerFunc(a.Usage.GetUsageReport)))
}

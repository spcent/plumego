package app

import (
	"net/http"

	"with-tenant-admin/internal/auth"
)

// RegisterRoutes wires all HTTP routes for the with-tenant-admin demo.
// One method, one path, one handler per line — full route table visible at a glance.
func (a *App) RegisterRoutes() error {
	admin := auth.RequireAdminToken(a.Cfg.AdminToken)
	reg := newRouteReg(a.Core)

	// Tenant admin
	reg.post("/admin/tenants", admin(http.HandlerFunc(a.Tenants.CreateTenant)))
	reg.get("/admin/tenants/:id", admin(http.HandlerFunc(a.Tenants.GetTenant)))
	reg.post("/admin/tenants/:id/suspend", admin(http.HandlerFunc(a.Tenants.SuspendTenant)))
	reg.delete("/admin/tenants/:id", admin(http.HandlerFunc(a.Tenants.DeleteTenant)))

	// Quota admin
	reg.get("/admin/quota/:tenantID", admin(http.HandlerFunc(a.Quotas.GetQuota)))
	reg.put("/admin/quota/:tenantID", admin(http.HandlerFunc(a.Quotas.SetQuota)))
	reg.post("/admin/quota/:tenantID/reset", admin(http.HandlerFunc(a.Quotas.ResetQuota)))

	// Usage reporting
	reg.post("/admin/usage/:tenantID", admin(http.HandlerFunc(a.Usage.RecordUsage)))
	reg.get("/admin/usage/:tenantID", admin(http.HandlerFunc(a.Usage.GetUsageReport)))

	return reg.err
}

// routeAdder is the minimal interface shared by *core.App and *core.RouteGroup.
type routeAdder interface {
	Get(path string, h http.Handler) error
	Post(path string, h http.Handler) error
	Put(path string, h http.Handler) error
	Delete(path string, h http.Handler) error
}

// routeReg wraps a routeAdder and records the first registration error.
// This lets the route table be written one route per line without per-call
// error checks; inspect reg.err once after all registrations.
type routeReg struct {
	adder routeAdder
	err   error
}

func newRouteReg(adder routeAdder) *routeReg { return &routeReg{adder: adder} }

func (r *routeReg) get(path string, h http.Handler)    { r.record(r.adder.Get(path, h)) }
func (r *routeReg) post(path string, h http.Handler)   { r.record(r.adder.Post(path, h)) }
func (r *routeReg) put(path string, h http.Handler)    { r.record(r.adder.Put(path, h)) }
func (r *routeReg) delete(path string, h http.Handler) { r.record(r.adder.Delete(path, h)) }
func (r *routeReg) record(err error) {
	if r.err == nil {
		r.err = err
	}
}

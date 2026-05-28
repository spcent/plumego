package app

import (
	"net/http"

	midauth "github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/x/frontend"

	dbauthn "dbadmin/internal/domain/authn"
	"dbadmin/internal/handler"
)

// RegisterRoutes wires all HTTP routes for dbadmin.
// Routes are explicit — one method, one path, one handler per line.
func (a *App) RegisterRoutes() error {
	healthH := handler.HealthHandler{
		ServiceName: a.Cfg.App.ServiceName,
		Logger:      a.Core.Logger(),
	}
	authH := handler.AuthHandler{
		AdminUser:     a.Cfg.App.AdminUser,
		AdminPassword: a.Cfg.App.AdminPassword,
		Sessions:      a.SessionStore,
		Logger:        a.Core.Logger(),
	}

	// Session authenticator wraps the session store.
	sessionAuthn := &dbauthn.SessionAuthenticator{Sessions: a.SessionStore}
	authMw, err := midauth.Authenticate(sessionAuthn)
	if err != nil {
		return err
	}

	// Public routes — no auth required.
	root := newRouteReg(a.Core)
	root.get("/healthz", http.HandlerFunc(healthH.Live))
	root.get("/readyz", http.HandlerFunc(healthH.Ready))
	root.post("/api/auth/login", http.HandlerFunc(authH.Login))
	if root.err != nil {
		return root.err
	}

	exportH := handler.ExportHandler{
		Connections: a.ConnectionStore,
		Manager:     a.DBManager,
		Logger:      a.Core.Logger(),
	}
	importH := handler.ImportHandler{
		Connections: a.ConnectionStore,
		Manager:     a.DBManager,
		Logger:      a.Core.Logger(),
	}
	ddlH := handler.DDLHandler{
		Connections: a.ConnectionStore,
		Manager:     a.DBManager,
		Logger:      a.Core.Logger(),
	}
	queryH := handler.QueryHandler{
		Connections: a.ConnectionStore,
		Manager:     a.DBManager,
		History:     a.HistoryStore,
		Logger:      a.Core.Logger(),
	}
	rowH := handler.RowHandler{
		Connections: a.ConnectionStore,
		Manager:     a.DBManager,
		Logger:      a.Core.Logger(),
	}
	inspectH := handler.InspectHandler{
		Connections: a.ConnectionStore,
		Manager:     a.DBManager,
		Logger:      a.Core.Logger(),
	}
	connH := handler.ConnectionHandler{
		Connections: a.ConnectionStore,
		Manager:     a.DBManager,
		Logger:      a.Core.Logger(),
	}
	sqliteH := handler.SQLiteHandler{
		Connections:    a.ConnectionStore,
		Manager:        a.DBManager,
		UploadDir:      a.UploadDir,
		MaxUploadBytes: a.Cfg.App.MaxUploadBytes,
		Logger:         a.Core.Logger(),
	}

	// Protected routes — all require a valid session cookie.
	guard := func(h http.Handler) http.Handler { return authMw(h) }

	protected := newRouteReg(a.Core)
	protected.post("/api/auth/logout", guard(http.HandlerFunc(authH.Logout)))
	protected.get("/api/auth/me", guard(http.HandlerFunc(authH.Me)))

	// Connection management.
	protected.get("/api/connections", guard(http.HandlerFunc(connH.List)))
	protected.post("/api/connections", guard(http.HandlerFunc(connH.Create)))
	protected.get("/api/connections/:id", guard(http.HandlerFunc(connH.Get)))
	protected.put("/api/connections/:id", guard(http.HandlerFunc(connH.Update)))
	protected.delete("/api/connections/:id", guard(http.HandlerFunc(connH.Delete)))
	protected.post("/api/connections/:id/test", guard(http.HandlerFunc(connH.Test)))

	// Database introspection.
	protected.get("/api/conn/:id/databases", guard(http.HandlerFunc(inspectH.Databases)))
	protected.get("/api/conn/:id/db/:db/tables", guard(http.HandlerFunc(inspectH.Tables)))
	protected.get("/api/conn/:id/db/:db/tables/:table/structure", guard(http.HandlerFunc(inspectH.TableStructure)))
	protected.get("/api/conn/:id/db/:db/tables/:table/columns", guard(http.HandlerFunc(inspectH.Columns)))
	protected.get("/api/conn/:id/db/:db/tables/:table/indexes", guard(http.HandlerFunc(inspectH.Indexes)))
	protected.get("/api/conn/:id/db/:db/tables/:table/foreign-keys", guard(http.HandlerFunc(inspectH.ForeignKeys)))

	// Row CRUD.
	protected.get("/api/conn/:id/db/:db/tables/:table/rows", guard(http.HandlerFunc(rowH.List)))
	protected.post("/api/conn/:id/db/:db/tables/:table/rows", guard(http.HandlerFunc(rowH.Create)))
	protected.patch("/api/conn/:id/db/:db/tables/:table/rows", guard(http.HandlerFunc(rowH.Update)))
	protected.delete("/api/conn/:id/db/:db/tables/:table/rows", guard(http.HandlerFunc(rowH.Delete)))

	// DDL operations.
	protected.post("/api/conn/:id/db/:db/tables", guard(http.HandlerFunc(ddlH.CreateTable)))
	protected.put("/api/conn/:id/db/:db/tables/:table", guard(http.HandlerFunc(ddlH.AlterTable)))
	protected.delete("/api/conn/:id/db/:db/tables/:table", guard(http.HandlerFunc(ddlH.DropTable)))

	// Export and Import.
	protected.get("/api/conn/:id/db/:db/tables/:table/export", guard(http.HandlerFunc(exportH.Export)))
	protected.post("/api/conn/:id/db/:db/import", guard(http.HandlerFunc(importH.Import)))

	// SQL console.
	protected.post("/api/conn/:id/db/:db/query", guard(http.HandlerFunc(queryH.Execute)))
	protected.get("/api/conn/:id/history", guard(http.HandlerFunc(queryH.ListHistory)))

	// SQLite file upload and download.
	protected.post("/api/sqlite/upload", guard(http.HandlerFunc(sqliteH.Upload)))
	protected.get("/api/conn/:id/sqlite/download", guard(http.HandlerFunc(sqliteH.Download)))
	if protected.err != nil {
		return protected.err
	}

	// SPA: must be registered last so API routes take precedence over the catch-all.
	// In development (no dist/), skip gracefully.
	return frontend.RegisterFromDir(a.Core, "./web/dist",
		frontend.WithFallback(true),
		frontend.WithCacheControl("public, max-age=31536000, immutable"),
		frontend.WithIndexCacheControl("no-cache, must-revalidate"),
	)
}

// routeAdder is the minimal interface shared by *core.App and *core.RouteGroup.
type routeAdder interface {
	Get(path string, h http.Handler) error
	Post(path string, h http.Handler) error
	Put(path string, h http.Handler) error
	Patch(path string, h http.Handler) error
	Delete(path string, h http.Handler) error
}

type routeReg struct {
	adder routeAdder
	err   error
}

func newRouteReg(adder routeAdder) *routeReg { return &routeReg{adder: adder} }

func (r *routeReg) get(path string, h http.Handler)    { r.record(r.adder.Get(path, h)) }
func (r *routeReg) post(path string, h http.Handler)   { r.record(r.adder.Post(path, h)) }
func (r *routeReg) put(path string, h http.Handler)    { r.record(r.adder.Put(path, h)) }
func (r *routeReg) patch(path string, h http.Handler)  { r.record(r.adder.Patch(path, h)) }
func (r *routeReg) delete(path string, h http.Handler) { r.record(r.adder.Delete(path, h)) }

func (r *routeReg) record(err error) {
	if r.err == nil {
		r.err = err
	}
}

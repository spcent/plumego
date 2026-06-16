package app

import (
	"io/fs"
	"net/http"
	"time"

	midauth "github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/x/frontend"

	dbauthn "dbadmin/internal/domain/authn"
	"dbadmin/internal/handler"
	"dbadmin/web"
)

// RegisterRoutes wires all HTTP routes for dbadmin.
// Routes are explicit — one method, one path, one handler per line.
func (a *App) RegisterRoutes() error {
	// Build health checkers for all connections
	checkers, err := handler.BuildHealthCheckers(
		a.ConnectionStore,
		a.DBManager,
		a.RedisManager,
		a.MongoManager,
		a.ESManager,
	)
	if err != nil {
		return err
	}

	healthH := handler.HealthHandler{
		ServiceName: a.Cfg.App.ServiceName,
		Logger:      a.Core.Logger(),
		Checkers:    checkers,
	}
	poolStatsH := handler.PoolStatsHandler{
		DBManager:    a.DBManager,
		RedisManager: a.RedisManager,
		MongoManager: a.MongoManager,
		ESManager:    a.ESManager,
		Logger:       a.Core.Logger(),
	}
	runtimeStatsH := handler.RuntimeStatsHandler{
		Logger:    a.Core.Logger(),
		StartTime: a.StartTime,
	}
	operationsH := handler.OperationsHandler{
		Registry: a.OperationRegistry,
		Logger:   a.Core.Logger(),
	}
	auditH := handler.AuditHandler{
		Store:  a.AuditStore,
		Logger: a.Core.Logger(),
	}
	authH := handler.AuthHandler{
		AdminUser:     a.Cfg.App.AdminUser,
		AdminPassword: a.Cfg.App.AdminPassword,
		AdminRole:     a.Cfg.App.AdminRole,
		Users:         a.Users,
		SessionTTL:    a.Cfg.App.SessionTTL,
		Sessions:      a.SessionStore,
		MFA:           a.MFAStore,
		LoginLimiter:  handler.NewLoginLimiter(5, 15*time.Minute),
		Logger:        a.Core.Logger(),
	}

	// Session authenticator wraps the session store.
	sessionAuthn := &dbauthn.SessionAuthenticator{Sessions: a.SessionStore}
	authMw, err := midauth.Authenticate(sessionAuthn)
	if err != nil {
		return err
	}

	docsH := handler.DocsHandler{}

	// Public routes — no auth required.
	root := newRouteReg(a.Core)
	root.get("/healthz", http.HandlerFunc(healthH.Live))
	root.get("/readyz", http.HandlerFunc(healthH.Ready))
	root.get("/pool-stats", http.HandlerFunc(poolStatsH.GetAllStats))
	root.get("/pool-stats/sql", http.HandlerFunc(poolStatsH.GetSQLPoolStats))
	root.get("/runtime-stats", http.HandlerFunc(runtimeStatsH.GetStats))
	root.get("/metrics", a.Metrics.Handler())
	root.get("/openapi.json", http.HandlerFunc(docsH.SpecJSON))
	root.get("/openapi.yaml", http.HandlerFunc(docsH.SpecYAML))
	root.get("/docs", http.HandlerFunc(docsH.UI))
	sameOriginMw := handler.SameOriginMiddleware(a.Core.Logger())
	ipMw := handler.IPAllowlistMiddleware(a.Cfg.App.AllowedIPs, a.Core.Logger())
	root.post("/api/auth/login", ipMw(sameOriginMw(http.HandlerFunc(authH.Login))))
	root.post("/api/auth/mfa/verify", ipMw(sameOriginMw(http.HandlerFunc(authH.MFAVerify))))
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
		Connections:         a.ConnectionStore,
		Manager:             a.DBManager,
		History:             a.HistoryStore,
		Logger:              a.Core.Logger(),
		QueryTimeoutSeconds: a.Cfg.App.QueryTimeoutSeconds,
		Registry:            nil,
	}
	if a.Cfg.App.QueryCancelEnabled {
		queryH.Registry = a.QueryRegistry
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
		Connections:  a.ConnectionStore,
		Manager:      a.DBManager,
		RedisManager: a.RedisManager,
		MongoManager: a.MongoManager,
		ESManager:    a.ESManager,
		Logger:       a.Core.Logger(),
	}
	resourceH := handler.ResourceHandler{
		Connections:    a.ConnectionStore,
		SQLAdapter:     a.SQLAdapter,
		RedisAdapter:   a.RedisAdapter,
		MongoAdapter:   a.MongoAdapter,
		ESAdapter:      a.ESAdapter,
		Logger:         a.Core.Logger(),
		TimeoutSeconds: a.Cfg.App.ResourceListTimeoutSeconds,
	}
	redisH := handler.RedisHandler{
		Connections:           a.ConnectionStore,
		RedisManager:          a.RedisManager,
		History:               a.RedisHistoryStore,
		Registry:              a.OperationRegistry,
		Logger:                a.Core.Logger(),
		CommandTimeoutSeconds: a.Cfg.App.RedisCommandTimeoutSeconds,
	}
	sqliteH := handler.SQLiteHandler{
		Connections:    a.ConnectionStore,
		Manager:        a.DBManager,
		UploadDir:      a.UploadDir,
		MaxUploadBytes: a.Cfg.App.MaxUploadBytes,
		Logger:         a.Core.Logger(),
	}
	mongoH := handler.MongoDBHandler{
		Connections:         a.ConnectionStore,
		MongoManager:        a.MongoManager,
		History:             a.MongoHistoryStore,
		Registry:            a.OperationRegistry,
		Logger:              a.Core.Logger(),
		QueryTimeoutSeconds: a.Cfg.App.MongoQueryTimeoutSeconds,
	}
	esH := handler.ElasticsearchHandler{
		Connections:         a.ConnectionStore,
		ESManager:           a.ESManager,
		History:             a.ESHistoryStore,
		Registry:            a.OperationRegistry,
		Logger:              a.Core.Logger(),
		QueryTimeoutSeconds: a.Cfg.App.ESQueryTimeoutSeconds,
	}

	// Protected routes — all require a valid session cookie.
	roleMw := handler.RoleMiddleware(a.Cfg.App.AdminRole, a.Core.Logger())
	auditMw := handler.AuditMiddleware(a.AuditStore, a.Cfg.App.AdminRole, a.Core.Logger())
	guard := func(h http.Handler) http.Handler {
		return ipMw(sameOriginMw(authMw(auditMw(roleMw(h)))))
	}

	protected := newRouteReg(a.Core)
	protected.post("/api/auth/logout", guard(http.HandlerFunc(authH.Logout)))
	protected.get("/api/auth/me", guard(http.HandlerFunc(authH.Me)))
	protected.post("/api/auth/sessions/revoke", guard(http.HandlerFunc(authH.RevokeAllSessions)))
	protected.get("/api/auth/users", guard(http.HandlerFunc(authH.ListUsers)))
	protected.get("/api/auth/mfa", guard(http.HandlerFunc(authH.MFAStatus)))
	protected.post("/api/auth/mfa/enroll", guard(http.HandlerFunc(authH.MFAEnroll)))
	protected.post("/api/auth/mfa/confirm", guard(http.HandlerFunc(authH.MFAConfirm)))
	protected.post("/api/auth/mfa/disable", guard(http.HandlerFunc(authH.MFADisable)))
	protected.get("/api/audit/events", guard(http.HandlerFunc(auditH.List)))
	protected.get("/api/audit/export", guard(http.HandlerFunc(auditH.Export)))

	// Connection management.
	protected.get("/api/connections", guard(http.HandlerFunc(connH.List)))
	protected.post("/api/connections", guard(http.HandlerFunc(connH.Create)))
	protected.get("/api/connections/:id", guard(http.HandlerFunc(connH.Get)))
	protected.put("/api/connections/:id", guard(http.HandlerFunc(connH.Update)))
	protected.delete("/api/connections/:id", guard(http.HandlerFunc(connH.Delete)))
	protected.post("/api/connections/:id/test", guard(http.HandlerFunc(connH.Test)))
	protected.delete("/api/connections/:id/runtime", guard(http.HandlerFunc(connH.CloseRuntime)))
	protected.get("/api/connections/export", guard(http.HandlerFunc(connH.Export)))
	protected.post("/api/connections/import", guard(http.HandlerFunc(connH.Import)))
	protected.get("/api/pool-stats", guard(http.HandlerFunc(poolStatsH.GetAllStats)))
	protected.get("/api/pool-stats/sql", guard(http.HandlerFunc(poolStatsH.GetSQLPoolStats)))

	// Unified resource tree — driver-agnostic.
	// GET /api/connections/:id/resources?parentId=<path>
	// Returns ResourceNode[] for SQL (sql_database, sql_table, sql_view).
	// Future: returns redis_db, mongo_collection, es_index nodes for those drivers.
	protected.get("/api/connections/:id/resources", guard(http.HandlerFunc(resourceH.ListResources)))

	// Database introspection — table-level and below.
	// Database and table listing is served by the unified /resources endpoint above.
	protected.get("/api/conn/:id/db/:db/schema-doc", guard(http.HandlerFunc(inspectH.SchemaDoc)))
	protected.get("/api/conn/:id/db/:db/tables/:table/structure", guard(http.HandlerFunc(inspectH.TableStructure)))
	protected.get("/api/conn/:id/db/:db/tables/:table/columns", guard(http.HandlerFunc(inspectH.Columns)))
	protected.get("/api/conn/:id/db/:db/tables/:table/indexes", guard(http.HandlerFunc(inspectH.Indexes)))
	protected.get("/api/conn/:id/db/:db/tables/:table/foreign-keys", guard(http.HandlerFunc(inspectH.ForeignKeys)))

	// Row CRUD.
	protected.get("/api/conn/:id/db/:db/tables/:table/rows", guard(http.HandlerFunc(rowH.List)))
	protected.post("/api/conn/:id/db/:db/tables/:table/rows", guard(http.HandlerFunc(rowH.Create)))
	protected.patch("/api/conn/:id/db/:db/tables/:table/rows", guard(http.HandlerFunc(rowH.Update)))
	protected.delete("/api/conn/:id/db/:db/tables/:table/rows", guard(http.HandlerFunc(rowH.Delete)))
	protected.post("/api/connections/:id/db/:db/tables/:table/rows/bulk-delete", guard(http.HandlerFunc(rowH.BulkDelete)))
	protected.post("/api/connections/:id/db/:db/tables/:table/rows/bulk-update", guard(http.HandlerFunc(rowH.BulkUpdate)))

	// DDL operations.
	protected.post("/api/conn/:id/db/:db/tables", guard(http.HandlerFunc(ddlH.CreateTable)))
	protected.put("/api/conn/:id/db/:db/tables/:table", guard(http.HandlerFunc(ddlH.AlterTable)))
	protected.delete("/api/conn/:id/db/:db/tables/:table", guard(http.HandlerFunc(ddlH.DropTable)))
	protected.post("/api/conn/:id/db/:db/views", guard(http.HandlerFunc(ddlH.CreateView)))
	protected.delete("/api/conn/:id/db/:db/views/:view", guard(http.HandlerFunc(ddlH.DropView)))

	// Export and Import.
	protected.get("/api/conn/:id/db/:db/tables/:table/export", guard(http.HandlerFunc(exportH.Export)))
	protected.post("/api/conn/:id/db/:db/import", guard(http.HandlerFunc(importH.Import)))

	// SQL console — SQL-specific routes (MySQL / SQLite only).
	// Future non-SQL data sources (Redis, MongoDB, Elasticsearch) should use
	// their own route groups, e.g. /api/conn/:id/redis/... or /api/conn/:id/mongo/...
	protected.post("/api/conn/:id/db/:db/query", guard(http.HandlerFunc(queryH.Execute)))
	protected.get("/api/conn/:id/history", guard(http.HandlerFunc(queryH.ListHistory)))
	protected.delete("/api/conn/:id/history", guard(http.HandlerFunc(queryH.ClearHistory)))
	protected.delete("/api/conn/:id/history/:entryId", guard(http.HandlerFunc(queryH.DeleteHistory)))

	// Query cancellation — cancel active queries and list running queries.
	if a.Cfg.App.QueryCancelEnabled {
		protected.post("/api/queries/cancel", guard(http.HandlerFunc(queryH.Cancel)))
		protected.get("/api/queries/active", guard(http.HandlerFunc(queryH.ListActive)))
		protected.post("/api/operations/cancel", guard(http.HandlerFunc(operationsH.Cancel)))
		protected.get("/api/operations/active", guard(http.HandlerFunc(operationsH.ListActive)))
	}

	// Redis operations — Redis-specific route group.
	protected.get("/api/conn/:id/redis/databases", guard(http.HandlerFunc(redisH.ListDBs)))
	protected.get("/api/conn/:id/redis/:dbIndex/keys", guard(http.HandlerFunc(redisH.ListKeys)))
	protected.get("/api/conn/:id/redis/:dbIndex/key", guard(http.HandlerFunc(redisH.GetKey)))
	protected.patch("/api/conn/:id/redis/:dbIndex/key/ttl", guard(http.HandlerFunc(redisH.SetTTL)))
	protected.delete("/api/conn/:id/redis/:dbIndex/key", guard(http.HandlerFunc(redisH.DeleteKey)))
	protected.post("/api/conn/:id/redis/:dbIndex/command", guard(http.HandlerFunc(redisH.Command)))
	protected.post("/api/conn/:id/redis/:dbIndex/batch-preview", guard(http.HandlerFunc(redisH.BatchPreview)))
	protected.post("/api/conn/:id/redis/:dbIndex/batch-delete", guard(http.HandlerFunc(redisH.BatchDelete)))
	protected.get("/api/conn/:id/redis/:dbIndex/export", guard(http.HandlerFunc(redisH.ExportKeys)))
	protected.post("/api/conn/:id/redis/:dbIndex/import", guard(http.HandlerFunc(redisH.ImportKeys)))
	protected.get("/api/conn/:id/redis/history", guard(http.HandlerFunc(redisH.ListHistory)))
	protected.delete("/api/conn/:id/redis/history", guard(http.HandlerFunc(redisH.ClearHistory)))
	protected.delete("/api/conn/:id/redis/history/:entryId", guard(http.HandlerFunc(redisH.DeleteHistoryEntry)))

	// SQLite file upload and download.
	protected.post("/api/sqlite/upload", guard(http.HandlerFunc(sqliteH.Upload)))
	protected.get("/api/conn/:id/sqlite/download", guard(http.HandlerFunc(sqliteH.Download)))

	// MongoDB operations — MongoDB-specific route group.
	protected.get("/api/connections/:id/mongo/databases", guard(http.HandlerFunc(mongoH.ListDatabases)))
	protected.get("/api/connections/:id/mongo/collections", guard(http.HandlerFunc(mongoH.ListCollections)))
	protected.post("/api/connections/:id/mongo/documents/query", guard(http.HandlerFunc(mongoH.QueryDocuments)))
	protected.get("/api/connections/:id/mongo/indexes", guard(http.HandlerFunc(mongoH.ListIndexes)))
	protected.post("/api/connections/:id/mongo/documents", guard(http.HandlerFunc(mongoH.InsertDocument)))
	protected.patch("/api/connections/:id/mongo/documents", guard(http.HandlerFunc(mongoH.UpdateDocument)))
	protected.delete("/api/connections/:id/mongo/documents", guard(http.HandlerFunc(mongoH.DeleteDocument)))

	// MongoDB collection and index management (DDL-equivalent operations).
	protected.post("/api/connections/:id/mongo/collections", guard(http.HandlerFunc(mongoH.CreateCollection)))
	protected.delete("/api/connections/:id/mongo/collections", guard(http.HandlerFunc(mongoH.DropCollection)))
	protected.post("/api/connections/:id/mongo/indexes/create", guard(http.HandlerFunc(mongoH.CreateIndex)))
	protected.delete("/api/connections/:id/mongo/indexes", guard(http.HandlerFunc(mongoH.DropIndex)))

	// MongoDB P1 - Advanced features
	protected.post("/api/connections/:id/mongo/aggregate", guard(http.HandlerFunc(mongoH.Aggregate)))
	protected.post("/api/connections/:id/mongo/explain", guard(http.HandlerFunc(mongoH.ExplainQuery)))
	protected.get("/api/connections/:id/mongo/schema", guard(http.HandlerFunc(mongoH.SchemaSample)))
	protected.get("/api/connections/:id/mongo/stats", guard(http.HandlerFunc(mongoH.CollectionStats)))
	protected.get("/api/connections/:id/mongo/export", guard(http.HandlerFunc(mongoH.ExportDocuments)))
	protected.post("/api/connections/:id/mongo/import", guard(http.HandlerFunc(mongoH.ImportDocuments)))
	protected.get("/api/mongo/objectid/:id/parse", guard(http.HandlerFunc(mongoH.ParseObjectId)))

	// MongoDB P1 - Pipeline history
	protected.get("/api/connections/:id/mongo/history", guard(http.HandlerFunc(mongoH.ListHistory)))
	protected.delete("/api/connections/:id/mongo/history/:entryId", guard(http.HandlerFunc(mongoH.DeleteHistoryEntry)))
	protected.delete("/api/connections/:id/mongo/history", guard(http.HandlerFunc(mongoH.ClearHistory)))

	// Elasticsearch operations — ES-specific route group.
	protected.get("/api/connections/:id/es/info", guard(http.HandlerFunc(esH.ClusterInfo)))
	protected.get("/api/connections/:id/es/indices", guard(http.HandlerFunc(esH.ListIndices)))
	protected.post("/api/connections/:id/es/indices", guard(http.HandlerFunc(esH.CreateIndex)))
	protected.delete("/api/connections/:id/es/indices", guard(http.HandlerFunc(esH.DeleteIndex)))
	protected.get("/api/connections/:id/es/templates", guard(http.HandlerFunc(esH.ListTemplates)))
	protected.post("/api/connections/:id/es/templates", guard(http.HandlerFunc(esH.CreateTemplate)))
	protected.delete("/api/connections/:id/es/templates", guard(http.HandlerFunc(esH.DeleteTemplate)))
	protected.get("/api/connections/:id/es/index/mapping", guard(http.HandlerFunc(esH.GetMapping)))
	protected.get("/api/connections/:id/es/index/settings", guard(http.HandlerFunc(esH.GetSettings)))
	protected.post("/api/connections/:id/es/search", guard(http.HandlerFunc(esH.Search)))
	protected.get("/api/connections/:id/es/document", guard(http.HandlerFunc(esH.GetDocument)))
	protected.delete("/api/connections/:id/es/document", guard(http.HandlerFunc(esH.DeleteDocument)))
	protected.get("/api/connections/:id/es/export", guard(http.HandlerFunc(esH.ExportDocuments)))
	protected.post("/api/connections/:id/es/import", guard(http.HandlerFunc(esH.ImportDocuments)))
	protected.get("/api/connections/:id/es/history", guard(http.HandlerFunc(esH.ListHistory)))
	protected.delete("/api/connections/:id/es/history/:entryId", guard(http.HandlerFunc(esH.DeleteHistoryEntry)))
	protected.delete("/api/connections/:id/es/history", guard(http.HandlerFunc(esH.ClearHistory)))
	if protected.err != nil {
		return protected.err
	}

	// SPA: must be registered last so API routes take precedence over the catch-all.
	// Use embedded assets in production, fall back to disk in development.
	opts := []frontend.Option{
		frontend.WithFallback(true),
		frontend.WithCacheControl("public, max-age=31536000, immutable"),
		frontend.WithIndexCacheControl("no-cache, must-revalidate"),
	}

	if web.Available {
		// Production: serve from embedded filesystem
		distFS, err := fs.Sub(web.Assets, "dist")
		if err != nil {
			return err
		}
		return frontend.RegisterFS(a.Core, http.FS(distFS), opts...)
	}

	// Development: serve from disk, skip if not built
	return frontend.RegisterFromDir(a.Core, "./web/dist", opts...)
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

package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	midsecurity "github.com/spcent/plumego/middleware/securityheaders"
	"github.com/spcent/plumego/middleware/timeout"
	kvstore "github.com/spcent/plumego/store/kv"
	"github.com/spcent/plumego/x/observability"

	"dbadmin/internal/config"
	"dbadmin/internal/datasource"
	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/audit"
	"dbadmin/internal/domain/connection"
	"dbadmin/internal/domain/eshistory"
	"dbadmin/internal/domain/history"
	"dbadmin/internal/domain/mfa"
	"dbadmin/internal/domain/mongohistory"
	"dbadmin/internal/domain/redishistory"
	"dbadmin/internal/domain/session"
	"dbadmin/internal/esmanager"
	"dbadmin/internal/handler"
	"dbadmin/internal/mongomanager"
	"dbadmin/internal/redismanager"
)

// App holds application-wide dependencies.
type App struct {
	Core              *core.App
	Cfg               config.Config
	SessionStore      *session.Store
	MFAStore          *mfa.Store
	ConnectionStore   *connection.Store
	HistoryStore      *history.Store
	MongoHistoryStore *mongohistory.Store
	ESHistoryStore    *eshistory.Store
	RedisHistoryStore *redishistory.Store
	AuditStore        *audit.Store
	DBManager         *dbmanager.Manager
	RedisManager      *redismanager.Manager
	MongoManager      *mongomanager.Manager
	ESManager         *esmanager.Manager
	SQLAdapter        *datasource.SQLAdapter
	RedisAdapter      *datasource.RedisAdapter
	MongoAdapter      *datasource.MongoAdapter
	ESAdapter         *datasource.ESAdapter
	UploadDir         string
	QueryRegistry     *handler.QueryRegistry
	OperationRegistry *handler.OperationRegistry
	StartTime         time.Time           // Track app start time for uptime monitoring
	Users             []config.UserConfig // resolved user list (from cfg.App.ResolveUsers())
	Metrics           *observability.PrometheusCollector
}

// New constructs the App with all middleware and shared dependencies wired.
func New(cfg config.Config) (*App, error) {
	coreApp := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})

	sessKV, err := kvstore.NewKVStore(kvstore.Options{
		DataDir: cfg.App.DataDir + "/sessions",
	})
	if err != nil {
		return nil, fmt.Errorf("create session KV store: %w", err)
	}

	mfaKV, err := kvstore.NewKVStore(kvstore.Options{
		DataDir: cfg.App.DataDir + "/mfa",
	})
	if err != nil {
		return nil, fmt.Errorf("create mfa KV store: %w", err)
	}

	connKV, err := kvstore.NewKVStore(kvstore.Options{
		DataDir: cfg.App.DataDir + "/connections",
	})
	if err != nil {
		return nil, fmt.Errorf("create connection KV store: %w", err)
	}

	connStore, err := connection.NewStore(connKV, cfg.App.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create connection store: %w", err)
	}

	histKV, err := kvstore.NewKVStore(kvstore.Options{
		DataDir: cfg.App.DataDir + "/history",
	})
	if err != nil {
		return nil, fmt.Errorf("create history KV store: %w", err)
	}

	mongoHistKV, err := kvstore.NewKVStore(kvstore.Options{
		DataDir: cfg.App.DataDir + "/mongo-history",
	})
	if err != nil {
		return nil, fmt.Errorf("create mongo history KV store: %w", err)
	}

	esHistKV, err := kvstore.NewKVStore(kvstore.Options{
		DataDir: cfg.App.DataDir + "/es-history",
	})
	if err != nil {
		return nil, fmt.Errorf("create ES history KV store: %w", err)
	}

	redisHistKV, err := kvstore.NewKVStore(kvstore.Options{
		DataDir: cfg.App.DataDir + "/redis-history",
	})
	if err != nil {
		return nil, fmt.Errorf("create Redis history KV store: %w", err)
	}

	auditKV, err := kvstore.NewKVStore(kvstore.Options{
		DataDir: cfg.App.DataDir + "/audit",
	})
	if err != nil {
		return nil, fmt.Errorf("create audit KV store: %w", err)
	}

	securityMw, err := midsecurity.Middleware(midsecurity.Config{})
	if err != nil {
		return nil, fmt.Errorf("configure security headers middleware: %w", err)
	}
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: coreApp.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: coreApp.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	timeoutMw := timeout.Middleware(timeout.Config{Timeout: 60 * time.Second})

	metricsCollector := observability.NewPrometheusCollector("dbadmin")

	if err := coreApp.Use(
		requestid.Middleware(),
		securityMw,
		cors.Middleware(cors.CORSOptions{}),
		recoveryMw,
		accesslogMw,
		bodylimit.Middleware(bodylimit.Config{
			MaxBytes: cfg.App.MaxBodyBytes,
			Logger:   coreApp.Logger(),
		}),
		httpmetrics.Middleware(metricsCollector),
		timeoutMw,
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	uploadDir := filepath.Join(cfg.App.DataDir, "uploads")
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	mgr := dbmanager.NewManager()
	redisMgr := redismanager.NewManager()
	mongoMgr := mongomanager.NewManager()
	esMgr := esmanager.NewManager()
	return &App{
		Core:              coreApp,
		Cfg:               cfg,
		SessionStore:      session.NewStore(sessKV, cfg.App.SessionTTL),
		MFAStore:          mfa.NewStore(mfaKV),
		ConnectionStore:   connStore,
		HistoryStore:      history.NewStore(histKV),
		MongoHistoryStore: mongohistory.NewStore(mongoHistKV),
		ESHistoryStore:    eshistory.NewStore(esHistKV),
		RedisHistoryStore: redishistory.NewStore(redisHistKV),
		AuditStore: audit.NewStoreWithOptions(auditKV, audit.Options{
			MaxEvents: cfg.App.AuditMaxEvents,
			Retention: time.Duration(cfg.App.AuditRetentionDays) *
				24 * time.Hour,
		}),
		DBManager:         mgr,
		RedisManager:      redisMgr,
		MongoManager:      mongoMgr,
		ESManager:         esMgr,
		SQLAdapter:        datasource.NewSQLAdapter(mgr),
		RedisAdapter:      datasource.NewRedisAdapter(redisMgr, cfg.App.ResourceListTimeoutSeconds),
		MongoAdapter:      datasource.NewMongoAdapter(mongoMgr),
		ESAdapter:         datasource.NewESAdapter(esMgr),
		UploadDir:         uploadDir,
		QueryRegistry:     handler.NewQueryRegistry(),
		OperationRegistry: handler.NewOperationRegistry(),
		StartTime:         time.Now(),
		Users:             cfg.App.ResolveUsers(),
		Metrics:           metricsCollector,
	}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
func (a *App) Start(ctx context.Context) error {
	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	a.Core.Logger().Info("starting dbadmin", plumelog.Fields{
		"addr":    a.Cfg.Core.Addr,
		"dataDir": a.Cfg.App.DataDir,
	})

	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		shutdownErr <- a.Core.Shutdown(shutdownCtx)
	}()

	var serveErr error
	if a.Cfg.Core.TLS.Enabled {
		serveErr = srv.ListenAndServeTLS("", "")
	} else {
		serveErr = srv.ListenAndServe()
	}
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("server stopped: %w", serveErr)
	}
	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}

package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	midsecurity "github.com/spcent/plumego/middleware/security"
	"github.com/spcent/plumego/middleware/timeout"

	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
	"cloud-vault/internal/document"
	"cloud-vault/internal/importer"
	"cloud-vault/internal/search"
	"cloud-vault/internal/storage"
	"cloud-vault/internal/tag"
)

// App holds application-wide dependencies.
type App struct {
	Core     *core.App
	Cfg      config.Config
	DB       *database.DB
	Docs     *document.Handler
	Tags     *tag.Handler
	Importer *importer.Handler
	Search   *search.Handler
	indexer  *search.Indexer
}

// New constructs the App: opens the database, initialises storage, and wires middleware.
func New(cfg config.Config) (*App, error) {
	db, err := database.Open(cfg.DB.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Migrate(); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	store, err := buildStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("build storage: %w", err)
	}

	app := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})

	securityMw, err := midsecurity.Middleware(midsecurity.Config{})
	if err != nil {
		return nil, fmt.Errorf("configure security middleware: %w", err)
	}
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	timeoutMw := timeout.Middleware(timeout.Config{Timeout: 30 * time.Second})

	maxBodyBytes := cfg.App.MaxUploadSizeMB * 1024 * 1024

	if err := app.Use(
		requestid.Middleware(),
		securityMw,
		cors.Middleware(cors.CORSOptions{}),
		recoveryMw,
		accesslogMw,
		bodylimit.Middleware(bodylimit.Config{
			MaxBytes: maxBodyBytes,
			Logger:   app.Logger(),
		}),
		httpmetrics.Middleware(metrics.NewNoopCollector()),
		timeoutMw,
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	// Document handler.
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)
	docs := document.NewHandler(docSvc, app.Logger())

	// Tag handler.
	tagRepo := tag.NewSQLiteRepository(db)
	tagSvc := tag.NewService(tagRepo)
	tags := tag.NewHandler(tagSvc, app.Logger())

	// Importer handler.
	importRepo := importer.NewRepository(db)
	importSvc := importer.NewService(importRepo, docSvc, importer.Config{
		MaxFileSizeMB: cfg.Import.MaxFileSizeMB,
	})
	imp := importer.NewHandler(importSvc, app.Logger())

	// Search handler + indexer.
	searchRepo := search.NewRepository(db)
	searchEngine := search.NewFTSEngine(db, cfg.Search.SnippetTokens)
	searchSvc := search.NewService(searchEngine, searchRepo, store, cfg.Search)
	searchHandler := search.NewHandler(searchSvc, app.Logger())
	indexer := search.NewIndexer(searchEngine, searchRepo, store, cfg.Search, app.Logger())

	// Wire index hook: document mutations trigger search indexing.
	docSvc.SetIndexHook(func(ctx context.Context, ev document.IndexEvent) {
		searchSvc.HandleIndexEvent(ctx, search.IndexEvent{
			DocID:    ev.DocID,
			Content:  ev.Content,
			Version:  ev.Version,
			Hash:     ev.Hash,
			Deleted:  ev.Deleted,
			IsImport: ev.IsImport,
		})
	})

	return &App{
		Core:     app,
		Cfg:      cfg,
		DB:       db,
		Docs:     docs,
		Tags:     tags,
		Importer: imp,
		Search:   searchHandler,
		indexer:  indexer,
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

	a.Core.Logger().Info("starting server", plumelog.Fields{
		"addr":             a.Cfg.Core.Addr,
		"storage_provider": a.Cfg.Storage.Provider,
		"db_path":          a.Cfg.DB.Path,
		"search_enabled":   a.Cfg.Search.Enabled,
	})

	// Start background indexer (cancelled when ctx is done).
	go a.indexer.Run(ctx)

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
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func buildStorage(cfg config.Config) (storage.ObjectStorage, error) {
	switch cfg.Storage.Provider {
	case "local":
		return storage.NewLocalStorage(cfg.Local.Root), nil
	case "qiniu":
		return storage.NewQiniuStorage(storage.QiniuConfig{
			AccessKey: cfg.Qiniu.AccessKey,
			SecretKey: cfg.Qiniu.SecretKey,
			Bucket:    cfg.Qiniu.Bucket,
			Domain:    cfg.Qiniu.Domain,
			Region:    cfg.Qiniu.Region,
			UseHTTPS:  cfg.Qiniu.UseHTTPS,
		})
	default:
		return nil, fmt.Errorf("unknown storage provider: %q", cfg.Storage.Provider)
	}
}

// routeReg is a helper that collects route registration errors.
type routeReg struct {
	core *core.App
	err  error
}

func newRouteReg(c *core.App) *routeReg {
	return &routeReg{core: c}
}

func (rr *routeReg) get(path string, h http.Handler) {
	if rr.err != nil {
		return
	}
	rr.err = rr.core.Get(path, h)
}

func (rr *routeReg) post(path string, h http.Handler) {
	if rr.err != nil {
		return
	}
	rr.err = rr.core.Post(path, h)
}

func (rr *routeReg) put(path string, h http.Handler) {
	if rr.err != nil {
		return
	}
	rr.err = rr.core.Put(path, h)
}

func (rr *routeReg) delete(path string, h http.Handler) {
	if rr.err != nil {
		return
	}
	rr.err = rr.core.Delete(path, h)
}

type groupRouteReg struct {
	group *core.RouteGroup
	err   error
}

func newGroupRouteReg(g *core.RouteGroup) *groupRouteReg {
	return &groupRouteReg{group: g}
}

func (rr *groupRouteReg) get(path string, h http.Handler) {
	if rr.err != nil {
		return
	}
	rr.err = rr.group.Get(path, h)
}

func (rr *groupRouteReg) post(path string, h http.Handler) {
	if rr.err != nil {
		return
	}
	rr.err = rr.group.Post(path, h)
}

func (rr *groupRouteReg) put(path string, h http.Handler) {
	if rr.err != nil {
		return
	}
	rr.err = rr.group.Put(path, h)
}

func (rr *groupRouteReg) delete(path string, h http.Handler) {
	if rr.err != nil {
		return
	}
	rr.err = rr.group.Delete(path, h)
}

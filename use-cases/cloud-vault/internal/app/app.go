package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
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

	"cloud-vault/internal/ai"
	"cloud-vault/internal/auth"
	"cloud-vault/internal/backup"
	"cloud-vault/internal/collection"
	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
	"cloud-vault/internal/document"
	"cloud-vault/internal/importer"
	"cloud-vault/internal/organize"
	"cloud-vault/internal/search"
	"cloud-vault/internal/storage"
	"cloud-vault/internal/system"
	"cloud-vault/internal/tag"
)

// App holds application-wide dependencies.
type App struct {
	Core           *core.App
	Cfg            config.Config
	DB             *database.DB
	Docs           *document.Handler
	Tags           *tag.Handler
	Importer       *importer.Handler
	Search         *search.Handler
	Organize       *organize.Handler
	Collection     *collection.Handler
	AI             *ai.Handler
	System         *system.Handler
	Auth           *auth.Handler
	Backup         *backup.Handler
	authService    *auth.Service
	authMiddleware func(http.Handler) http.Handler
	indexer        *search.Indexer
	aiWorkers      []*ai.Worker
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

	// Organize handler (V0.4).
	organizeRepo := organize.NewRepository(db)
	organizeSvc := organize.NewService(organizeRepo, store, cfg.Organize)
	organizeHandler := organize.NewHandler(organizeSvc, app.Logger())

	// Collection handler (V0.4).
	collectionRepo := collection.NewRepository(db)
	collectionSvc := collection.NewService(collectionRepo)
	collectionHandler := collection.NewHandler(collectionSvc, app.Logger())

	// AI handler (V0.5).
	aiRepo := ai.NewRepository(db.DB)
	aiProvider := buildAIProvider(cfg.AI)
	aiDocSaver := &aiDocumentSaver{docSvc: docSvc, aiRepo: aiRepo}
	aiSvc := ai.NewService(aiRepo, store, cfg.AI, aiProvider, aiDocSaver)
	aiHandler := ai.NewHandler(aiSvc, cfg.AI.Enabled, app.Logger())

	var aiWorkers []*ai.Worker
	if cfg.AI.Enabled {
		n := cfg.AI.TaskWorkers
		if n <= 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			workerLogger := app.Logger().WithFields(plumelog.Fields{"worker": i})
			aiWorkers = append(aiWorkers, ai.NewWorker(aiSvc, workerLogger, cfg.AI.MaxRetries))
		}
	}

	// System handler (V0.6).
	systemSvc := system.NewService(db.DB, store, cfg.AI)
	systemHandler := system.NewHandler(systemSvc, app.Logger())

	// Auth handler (V0.7).
	authRepo := auth.NewRepository(db)
	authConfig := auth.Config{
		Enabled:                   cfg.Auth.Enabled,
		SessionTTLHours:           cfg.Auth.SessionTTLHours,
		CookieName:                cfg.Auth.CookieName,
		SecureCookie:              cfg.Auth.SecureCookie,
		MaxLoginFailures:          cfg.Auth.MaxLoginFailures,
		LoginFailureWindowMinutes: cfg.Auth.LoginFailureWindowMinutes,
		LockoutMinutes:            cfg.Auth.LockoutMinutes,
		PasswordMinLength:         cfg.Auth.PasswordMinLength,
		BootstrapAdminEnabled:     cfg.Auth.BootstrapAdminEnabled,
		BootstrapAdminUsername:    cfg.Auth.BootstrapAdminUsername,
		BootstrapAdminEmail:       cfg.Auth.BootstrapAdminEmail,
		BootstrapAdminPassword:    cfg.Auth.BootstrapAdminPassword,
	}
	authService := auth.NewService(authRepo, authConfig)
	authHandler := auth.NewHandler(authService, authConfig, app.Logger())

	// Bootstrap admin if enabled.
	if cfg.Auth.BootstrapAdminEnabled {
		if err := authService.BootstrapAdmin(context.Background()); err != nil {
			app.Logger().Error("failed to bootstrap admin", plumelog.Fields{"error": err.Error()})
		}
	}

	// Auth middleware.
	var authMiddleware func(http.Handler) http.Handler
	if cfg.Auth.Enabled {
		authMiddleware = auth.RequireAuth(authService, cfg.Auth.CookieName)
	}

	// Backup handler (V0.8).
	backupDir := filepath.Join(filepath.Dir(cfg.DB.Path), "backups")
	backupRepo := backup.NewRepository(backupDir)
	backupSvc := backup.NewService(backupRepo, cfg.DB.Path, cfg.Storage.Provider, cfg.Local.Root, "", cfg.App.Version)
	backupHandler := backup.NewHandler(backupSvc, filepath.Dir(cfg.DB.Path), app.Logger())

	return &App{
		Core:           app,
		Cfg:            cfg,
		DB:             db,
		Docs:           docs,
		Tags:           tags,
		Importer:       imp,
		Search:         searchHandler,
		Organize:       organizeHandler,
		Collection:     collectionHandler,
		AI:             aiHandler,
		System:         systemHandler,
		Auth:           authHandler,
		Backup:         backupHandler,
		authService:    authService,
		authMiddleware: authMiddleware,
		indexer:        indexer,
		aiWorkers:      aiWorkers,
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

	// Start AI task workers.
	for _, w := range a.aiWorkers {
		go w.Run(ctx)
	}

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

func buildAIProvider(cfg config.AIConfig) ai.LLMProvider {
	if cfg.Provider == "openai_compatible" && cfg.BaseURL != "" {
		return ai.NewOpenAIProvider(cfg.BaseURL, cfg.APIKey, cfg.Model)
	}
	return ai.NewMockProvider()
}

// aiDocumentSaver saves AI-generated content as a vault document and records source links.
type aiDocumentSaver struct {
	docSvc *document.Service
	aiRepo *ai.Repository
}

func (s *aiDocumentSaver) SaveAIDocument(ctx context.Context, title, content, _ string, sourceIDs []string) (string, error) {
	result, err := s.docSvc.Create(ctx, document.CreateRequest{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return "", err
	}
	for _, srcID := range sourceIDs {
		_ = s.aiRepo.SaveSourceLink(result.ID, srcID, "ai_generated")
	}
	return result.ID, nil
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

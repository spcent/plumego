package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GetTemplateFiles returns the files that would be created for a template.
func GetTemplateFiles(template string) []string {
	base := []string{
		"cmd/app/main.go",
		"go.mod",
		"env.example",
		".gitignore",
		"README.md",
	}

	switch template {
	case "canonical":
		// Mirrors reference/standard-service exactly: explicit bootstrap,
		// stable-root-only imports, constructor injection, explicit routes.
		return []string{
			"cmd/app/main.go",
			"internal/app/app.go",
			"internal/app/routes.go",
			"internal/handler/api.go",
			"internal/handler/health.go",
			"internal/config/config.go",
			"go.mod",
			"env.example",
			".gitignore",
			"README.md",
		}
	case "minimal":
		return base
	case "api":
		return append(base, []string{
			"internal/httpapp/app.go",
			"internal/httpapp/routes.go",
			"internal/httpapp/handlers/health.go",
			"internal/httpapp/handlers/user.go",
			"internal/domain/user/service.go",
			"internal/domain/user/repository.go",
		}...)
	case "fullstack":
		return append(base, []string{
			"internal/httpapp/app.go",
			"internal/httpapp/routes.go",
			"internal/httpapp/handlers/health.go",
			"internal/httpapp/handlers/api.go",
			"frontend/index.html",
			"frontend/app.js",
			"frontend/styles.css",
		}...)
	case "microservice":
		return append(base, []string{
			"internal/httpapp/app.go",
			"internal/httpapp/routes.go",
			"internal/httpapp/handlers/health.go",
			"internal/httpapp/handlers/metrics.go",
			"internal/domain/user/service.go",
			"internal/domain/user/repository.go",
			"internal/platform/httpjson/response.go",
			"internal/platform/httperr/error.go",
			"Dockerfile",
			"docker-compose.yml",
		}...)
	default:
		return base
	}
}

// CreateProject creates a new project from a template.
func CreateProject(dir, name, module, template string, initGit bool) ([]string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	files := GetTemplateFiles(template)
	created := []string{}

	for _, file := range files {
		path := filepath.Join(dir, file)

		parent := filepath.Dir(path)
		if err := os.MkdirAll(parent, 0755); err != nil {
			return created, fmt.Errorf("failed to create directory %s: %w", parent, err)
		}

		content := getTemplateContent(file, name, module, template)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return created, fmt.Errorf("failed to write file %s: %w", path, err)
		}

		created = append(created, file)
	}

	if initGit {
		cmd := exec.Command("git", "init")
		cmd.Dir = dir
		if err := cmd.Run(); err == nil {
			created = append(created, ".git/")
		}
	}

	cmd := exec.Command("go", "mod", "init", module)
	cmd.Dir = dir
	if err := cmd.Run(); err == nil {
		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Dir = dir
		_ = tidyCmd.Run()
	}

	return created, nil
}

func getTemplateContent(file, name, module, template string) string {
	switch file {
	case "cmd/app/main.go":
		if template == "canonical" {
			return getCanonicalMainGoContent(module, name)
		}
		return getMainGoContent(module, template)
	case "internal/app/app.go":
		return getCanonicalAppGoContent(module)
	case "internal/app/routes.go":
		return getCanonicalRoutesGoContent(module, name)
	case "internal/handler/api.go":
		return getCanonicalAPIHandlerContent(name)
	case "internal/handler/health.go":
		return getCanonicalHealthHandlerContent()
	case "internal/config/config.go":
		return getCanonicalConfigGoContent(module)
	case "go.mod":
		return fmt.Sprintf("module %s\n\ngo 1.24\n\nrequire github.com/spcent/plumego v0.0.0\n", module)
	case "env.example":
		return getEnvExampleContent()
	case ".gitignore":
		return getGitignoreContent()
	case "README.md":
		return getReadmeContent(name, template)
	default:
		return getDefaultFileContent(file, name, module)
	}
}

func getDefaultFileContent(file, name, module string) string {
	if module == "" {
		module = "example.com/app"
	}
	switch filepath.Base(file) {
	case "app.go":
		return `package httpapp

import (
	"log"
	"net/http"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/recovery"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
)

// New constructs the HTTP application with middleware and routes.
func New() *core.App {
	cfg := core.DefaultConfig()
	cfg.Addr = ":8080"
	app := core.New(cfg, core.AppDependencies{Logger: plumelog.NewLogger()})
	if err := app.Use(
		requestid.Middleware(),
		mwtracing.Middleware(nil),
		httpmetrics.Middleware(nil),
		accesslog.Middleware(app.Logger(), nil, nil),
		recovery.Recovery(app.Logger()),
	); err != nil {
		log.Fatal(err)
	}
	if err := registerRoutes(app); err != nil {
		log.Fatal(err)
	}
	return app
}

// ListenAndServe starts the server.
func ListenAndServe(app *core.App) error {
	return http.ListenAndServe(":8080", app)
}
`
	case "routes.go":
		return fmt.Sprintf(`package httpapp

import (
	"github.com/spcent/plumego/core"

	userdom "%s/internal/domain/user"
	"%s/internal/httpapp/handlers"
)

// registerRoutes wires all HTTP routes with explicit handler construction.
func registerRoutes(app *core.App) error {
	if err := app.Get("/healthz", handlers.Health); err != nil {
		return err
	}

	// Wire user domain: repository → service → handler.
	userRepo := userdom.NewMemRepository()
	userSvc := userdom.NewService(userRepo)
	userH := handlers.UserHandler{Service: userSvc}

	if err := app.Get("/api/v1/users", userH.List); err != nil {
		return err
	}
	if err := app.Get("/api/v1/users/:id", userH.Get); err != nil {
		return err
	}
	if err := app.Post("/api/v1/users", userH.Create); err != nil {
		return err
	}
	return nil
}
`, module, module)
	case "health.go":
		return `package handlers

import (
	"net/http"
)

// Health responds with a simple JSON liveness check.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(` + "`" + `{"status":"ok"}` + "`" + `))
}
`
	case "api.go":
		return fmt.Sprintf(`package handlers

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// APIHandler handles general API endpoints.
type APIHandler struct{}

// Hello responds with service metadata.
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"message":   "hello from %s",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if err := contract.WriteResponse(w, r, http.StatusOK, resp, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
`, name)
	case "user.go":
		return fmt.Sprintf(`package handlers

import (
	"encoding/json"
	"net/http"

	userdom %q
)

// UserHandler handles HTTP requests for the user domain.
type UserHandler struct {
	Service userdom.Service
}

type createUserRequest struct {
	Name  string `+"`json:\"name\"`"+`
	Email string `+"`json:\"email\"`"+`
}

// List handles GET /api/v1/users
func (h UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.Service.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

// Get handles GET /api/v1/users/:id
func (h UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	u, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

// Create handles POST /api/v1/users
func (h UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	u, err := h.Service.Create(r.Context(), req.Name, req.Email)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(u)
}
`, module+"/internal/domain/user")
	case "service.go":
		return `package user

import "context"

// Service defines the domain operations for users.
type Service interface {
	Create(ctx context.Context, name, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]*User, error)
}

type service struct {
	repo Repository
}

// NewService constructs a Service backed by repo.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, name, email string) (*User, error) {
	return s.repo.Insert(ctx, name, email)
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) List(ctx context.Context) ([]*User, error) {
	return s.repo.FindAll(ctx)
}
`
	case "repository.go":
		return `package user

import (
	"context"
	"fmt"
	"sync"
)

// User represents a user entity.
type User struct {
	ID    string ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}

// Repository defines persistence operations for users.
type Repository interface {
	Insert(ctx context.Context, name, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
}

type memRepository struct {
	mu    sync.Mutex
	items map[string]*User
	seq   int
}

// NewMemRepository returns an in-memory Repository implementation.
func NewMemRepository() Repository {
	return &memRepository{items: make(map[string]*User)}
}

func (r *memRepository) Insert(_ context.Context, name, email string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	u := &User{ID: fmt.Sprintf("%d", r.seq), Name: name, Email: email}
	r.items[u.ID] = u
	return u, nil
}

func (r *memRepository) FindByID(_ context.Context, id string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.items[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return u, nil
}

func (r *memRepository) FindAll(_ context.Context) ([]*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*User, 0, len(r.items))
	for _, u := range r.items {
		out = append(out, u)
	}
	return out, nil
}
`
	case "metrics.go":
		return `package handlers

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// MetricsHandler handles the service metrics summary endpoint.
type MetricsHandler struct{}

// Summary responds with a basic runtime metrics snapshot.
func (h MetricsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"uptime":    time.Since(startTime).String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

var startTime = time.Now()
`
	case "response.go":
		return `package httpjson

import (
	"encoding/json"
	"net/http"
)

// Write encodes v as JSON and writes it with the given status code.
func Write(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
`
	case "error.go":
		if filepath.Dir(file) == "internal/platform/httperr" || filepath.Base(filepath.Dir(file)) == "httperr" {
			return `package httperr

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the canonical error shape for this service.
type ErrorResponse struct {
	Code    string ` + "`json:\"code\"`" + `
	Message string ` + "`json:\"message\"`" + `
}

// Write writes a structured JSON error response.
func Write(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Code: code, Message: message})
}

// BadRequest writes a 400 error response.
func BadRequest(w http.ResponseWriter, code, message string) {
	Write(w, http.StatusBadRequest, code, message)
}

// Internal writes a 500 error response.
func Internal(w http.ResponseWriter, code, message string) {
	Write(w, http.StatusInternalServerError, code, message)
}
`
		}
		// Derive package name from directory for other error.go files.
		pkg := filepath.Base(filepath.Dir(file))
		if pkg == "." || pkg == "" {
			pkg = "main"
		}
		return fmt.Sprintf("package %s\n", pkg)
	case "Dockerfile":
		return `FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -o /bin/app ./cmd/app

FROM alpine:3.19
RUN adduser -D -u 1000 appuser
USER appuser
COPY --from=builder /bin/app /bin/app
EXPOSE 8080
ENTRYPOINT ["/bin/app"]
`
	case "docker-compose.yml":
		return `version: "3.9"
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ADDR=:8080
`
	case "index.html":
		return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>%s</title>
  <link rel="stylesheet" href="/styles.css" />
</head>
<body>
  <h1>%s</h1>
  <div id="app"></div>
  <script src="/app.js"></script>
</body>
</html>
`, name, name)
	case "app.js":
		return `// Entry point for the frontend application.
document.addEventListener('DOMContentLoaded', function () {
  fetch('/healthz')
    .then(function (r) { return r.json(); })
    .then(function (data) {
      document.getElementById('app').textContent = 'Status: ' + data.status;
    })
    .catch(function (err) {
      document.getElementById('app').textContent = 'Error: ' + err.message;
    });
});
`
	case "styles.css":
		return `/* Application styles */
*, *::before, *::after {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: system-ui, -apple-system, sans-serif;
  padding: 2rem;
  color: #1a1a1a;
}

h1 {
  margin-bottom: 1rem;
  font-size: 1.5rem;
}
`
	default:
		// Derive a compilable Go file with the correct package name.
		ext := filepath.Ext(file)
		if ext != ".go" {
			return ""
		}
		pkg := filepath.Base(filepath.Dir(file))
		if pkg == "." || pkg == "" {
			pkg = "main"
		}
		return fmt.Sprintf("package %s\n", pkg)
	}
}

func getCanonicalMainGoContent(module, name string) string {
	return fmt.Sprintf(`package main

import (
	"log"

	"%s/internal/app"
	"%s/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %%v", err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %%v", err)
	}

	if err := a.RegisterRoutes(); err != nil {
		log.Fatalf("failed to register routes: %%v", err)
	}

	log.Printf("Starting %s on %%s", cfg.Core.Addr)
	if err := a.Start(); err != nil {
		log.Fatalf("server stopped: %%v", err)
	}
}
`, module, module, name)
}

func getCanonicalAppGoContent(module string) string {
	return fmt.Sprintf(`// Package app wires together the application dependencies and manages the
// server lifecycle.
package app

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"%s/internal/config"
)

// App holds application-wide dependencies.
type App struct {
	Core *core.App
	Cfg  config.Config
}

// New constructs the App with explicit stable-root wiring only.
func New(cfg config.Config) (*App, error) {
	a := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})
	a.Use(requestid.Middleware())
	a.Use(recovery.Recovery(a.Logger()))
	a.Use(accesslog.Middleware(a.Logger(), nil, nil))

	return &App{Core: a, Cfg: cfg}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
func (a *App) Start() error {
	ctx := context.Background()

	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %%w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %%w", err)
	}
	defer a.Core.Shutdown(ctx)

	if a.Cfg.Core.TLS.Enabled {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			return fmt.Errorf("server stopped: %%w", err)
		}
		return nil
	}
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("server stopped: %%w", err)
	}
	return nil
}
`, module)
}

func getCanonicalRoutesGoContent(module, name string) string {
	return fmt.Sprintf(`package app

import (
	"%s/internal/handler"
)

// RegisterRoutes wires all HTTP routes.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "%s"}

	if err := a.Core.Get("/", api.Hello); err != nil {
		return err
	}
	if err := a.Core.Get("/healthz", health.Live); err != nil {
		return err
	}
	if err := a.Core.Get("/readyz", health.Ready); err != nil {
		return err
	}
	if err := a.Core.Get("/api/hello", api.Hello); err != nil {
		return err
	}
	return nil
}
`, module, name)
}

func getCanonicalAPIHandlerContent(name string) string {
	return fmt.Sprintf(`// Package handler contains the HTTP handlers.
package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// APIHandler handles the core JSON API endpoints.
type APIHandler struct{}

// Hello responds with service metadata.
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"message":   "hello from %s",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	if err := contract.WriteResponse(w, r, http.StatusOK, resp, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
`, name)
}

func getCanonicalHealthHandlerContent() string {
	return `package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// HealthHandler serves liveness and readiness endpoints.
type HealthHandler struct {
	ServiceName string
}

// Live reports that the process is serving HTTP traffic.
func (h HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"status":    "ok",
		"service":   h.ServiceName,
		"check":     "liveness",
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// Ready reports that the service is ready to accept requests.
func (h HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"status":    "ready",
		"service":   h.ServiceName,
		"check":     "readiness",
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
`
}

func getCanonicalConfigGoContent(_ string) string {
	return `// Package config loads and validates the application configuration.
package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spcent/plumego/core"
	plumecfg "github.com/spcent/plumego/internal/config"
	plumelog "github.com/spcent/plumego/log"
)

// Config holds all application configuration.
type Config struct {
	Core core.AppConfig
	App  AppConfig
}

// AppConfig holds app-local, non-kernel configuration.
type AppConfig struct {
	EnvFile string
	Debug   bool
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Addr = ":8080"
	return Config{
		Core: coreCfg,
		App: AppConfig{
			EnvFile: ".env",
		},
	}
}

// Load reads configuration from environment variables and flags.
func Load() (Config, error) {
	cfg := Defaults()

	cfg.App.EnvFile = resolveEnvFile(os.Args, cfg.App.EnvFile)
	if err := loadEnvFile(cfg.App.EnvFile); err != nil {
		return cfg, err
	}

	if err := applyEnv(&cfg); err != nil {
		return cfg, err
	}
	applyFlags(&cfg)

	return cfg, Validate(cfg)
}

// Validate returns an error if cfg is unusable.
func Validate(cfg Config) error {
	if cfg.Core.Addr == "" {
		return fmt.Errorf("addr is required")
	}
	return nil
}

func applyEnv(cfg *Config) error {
	manager := plumecfg.NewManager(plumelog.NewLogger())
	if err := manager.AddSource(plumecfg.NewEnvSource("")); err != nil {
		return err
	}
	if err := manager.Load(context.Background()); err != nil {
		return err
	}
	cfg.Core.Addr = manager.GetString("app_addr", cfg.Core.Addr)
	cfg.App.EnvFile = manager.GetString("app_env_file", cfg.App.EnvFile)
	cfg.App.Debug = manager.GetBool("app_debug", cfg.App.Debug)
	return nil
}

func applyFlags(cfg *Config) {
	flag.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	flag.StringVar(&cfg.App.EnvFile, "env-file", cfg.App.EnvFile, "path to .env file")
	flag.BoolVar(&cfg.App.Debug, "debug", cfg.App.Debug, "enable debug mode")
	flag.Parse()
}

func resolveEnvFile(args []string, defaultPath string) string {
	if envPath := strings.TrimSpace(os.Getenv("APP_ENV_FILE")); envPath != "" {
		defaultPath = envPath
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--env-file" || arg == "-env-file" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--env-file=") {
			return strings.TrimPrefix(arg, "--env-file=")
		}
		if strings.HasPrefix(arg, "-env-file=") {
			return strings.TrimPrefix(arg, "-env-file=")
		}
	}
	return defaultPath
}

func loadEnvFile(path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	return plumecfg.LoadEnv(path, true)
}
`
}

func getMainGoContent(module, template string) string {
	_ = module // module path is in go.mod, not needed inline
	_ = template
	return `package main

import (
	"log"
	"net/http"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/recovery"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
)

func main() {
	cfg := core.DefaultConfig()
	cfg.Addr = ":8080"
	app := core.New(cfg, core.AppDependencies{Logger: plumelog.NewLogger()})
	if err := app.Use(
		requestid.Middleware(),
		mwtracing.Middleware(nil),
		httpmetrics.Middleware(nil),
		accesslog.Middleware(app.Logger(), nil, nil),
		recovery.Recovery(app.Logger()),
	); err != nil {
		log.Fatal(err)
	}

	if err := app.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(` + "`" + `{"status":"ok"}` + "`" + `))
	}); err != nil {
		log.Fatal(err)
	}

	if err := http.ListenAndServe(":8080", app); err != nil {
		log.Fatal(err)
	}
}
`
}

func getEnvExampleContent() string {
	return `APP_ADDR=:8080
APP_DEBUG=false
APP_SHUTDOWN_TIMEOUT_MS=5000
APP_MAX_BODY_BYTES=10485760
`
}

func getGitignoreContent() string {
	return `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test coverage
*.out
coverage.html

# Environment
.env
.env.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
}

func getReadmeContent(name, template string) string {
	return fmt.Sprintf(`# %s

A plumego application built with the **%s** template.

## Getting Started

### Install dependencies
`+"```bash"+`
go mod tidy
`+"```"+`

### Run development server
`+"```bash"+`
plumego dev
# or
go run ./cmd/app
`+"```"+`

### Run tests
`+"```bash"+`
plumego test
# or
go test ./...
`+"```"+`

### Build
`+"```bash"+`
plumego build
# or
go build -o bin/app ./cmd/app
`+"```"+`

## Documentation

See [Plumego documentation](https://github.com/spcent/plumego) for more information.
`, name, template)
}

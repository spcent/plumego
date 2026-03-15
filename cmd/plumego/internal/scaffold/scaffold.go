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
		return getDefaultFileContent(file)
	}
}

func getDefaultFileContent(file string) string {
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
	app := core.New(
		core.WithAddr(":8080"),
		core.WithLogger(plumelog.NewGLogger()),
	)
	if err := app.Use(
		requestid.Middleware(),
		mwtracing.Middleware(nil),
		httpmetrics.Middleware(nil),
		accesslog.Middleware(app.Logger()),
		recovery.Recovery(app.Logger()),
	); err != nil {
		log.Fatal(err)
	}
	registerRoutes(app)
	return app
}

// ListenAndServe starts the server.
func ListenAndServe(app *core.App) error {
	return http.ListenAndServe(":8080", app)
}
`
	case "routes.go":
		return `package httpapp

import (
	"github.com/spcent/plumego/core"

	"internal/httpapp/handlers"
)

func registerRoutes(app *core.App) {
	app.Get("/healthz", handlers.Health)
}
`
	case "health.go":
		return `package handlers

import (
	"net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(` + "`" + `{"status":"ok"}` + "`" + `))
}
`
	case "user.go":
		return `package handlers

import (
	"encoding/json"
	"net/http"
)

type UserHandler struct {
	Service UserService
}

type UserService interface {
	// TODO: define service methods
}

type CreateUserRequest struct {
	Name  string ` + "`" + `json:"name"` + "`" + `
	Email string ` + "`" + `json:"email"` + "`" + `
}

type CreateUserResponse struct {
	ID string ` + "`" + `json:"id"` + "`" + `
}

func (h UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: call h.Service
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateUserResponse{ID: "TODO"})
}
`
	case "service.go":
		return `package user

// Service defines the user domain operations.
type Service interface {
	// TODO: define service methods
}

type service struct{}

// NewService constructs a new user Service.
func NewService() Service {
	return &service{}
}
`
	case "repository.go":
		return `package user

// Repository defines the user persistence operations.
type Repository interface {
	// TODO: define repository methods
}
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
	Code    string ` + "`" + `json:"code"` + "`" + `
	Message string ` + "`" + `json:"message"` + "`" + `
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
		return fmt.Sprintf("// TODO: Implement %s\npackage main\n", file)
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
	default:
		return fmt.Sprintf("// TODO: Implement %s\npackage main\n", file)
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
	opts := []core.Option{
		core.WithAddr(cfg.Core.Addr),
		core.WithEnvPath(cfg.Core.EnvFile),
		core.WithLogger(plumelog.NewGLogger()),
	}
	if cfg.Core.Debug {
		opts = append(opts, core.WithDebug())
	}

	a := core.New(opts...)
	a.Use(requestid.Middleware())
	a.Use(recovery.Recovery(a.Logger()))
	a.Use(accesslog.Middleware(a.Logger()))

	return &App{Core: a, Cfg: cfg}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
func (a *App) Start() error {
	ctx := context.Background()

	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %%w", err)
	}
	if err := a.Core.Start(ctx); err != nil {
		return fmt.Errorf("start runtime: %%w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %%w", err)
	}
	defer a.Core.Shutdown(ctx)

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

	a.Core.Get("/", api.Hello)
	a.Core.Get("/healthz", health.Live)
	a.Core.Get("/readyz", health.Ready)
	a.Core.Get("/api/hello", api.Hello)
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

func getCanonicalConfigGoContent(module string) string {
	return fmt.Sprintf(`// Package config loads and validates the application configuration.
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
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	return Config{
		Core: core.AppConfig{
			Addr:    ":8080",
			EnvFile: ".env",
			Debug:   false,
		},
	}
}

// Load reads configuration from environment variables and flags.
func Load() (Config, error) {
	cfg := Defaults()

	cfg.Core.EnvFile = resolveEnvFile(os.Args, cfg.Core.EnvFile)
	if err := loadEnvFile(cfg.Core.EnvFile); err != nil {
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
	manager := plumecfg.NewManager(plumelog.NewGLogger())
	if err := manager.AddSource(plumecfg.NewEnvSource("")); err != nil {
		return err
	}
	if err := manager.Load(context.Background()); err != nil {
		return err
	}
	cfg.Core.Addr = manager.GetString("app_addr", cfg.Core.Addr)
	cfg.Core.EnvFile = manager.GetString("app_env_file", cfg.Core.EnvFile)
	cfg.Core.Debug = manager.GetBool("app_debug", cfg.Core.Debug)
	return nil
}

func applyFlags(cfg *Config) {
	flag.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	flag.StringVar(&cfg.Core.EnvFile, "env-file", cfg.Core.EnvFile, "path to .env file")
	flag.BoolVar(&cfg.Core.Debug, "debug", cfg.Core.Debug, "enable debug mode")
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
`, module)
}

func getMainGoContent(module, template string) string {
	_ = module // module path is in go.mod, not needed inline
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
	app := core.New(
		core.WithAddr(":8080"),
		core.WithLogger(plumelog.NewGLogger()),
	)
	if err := app.Use(
		requestid.Middleware(),
		mwtracing.Middleware(nil),
		httpmetrics.Middleware(nil),
		accesslog.Middleware(app.Logger()),
		recovery.Recovery(app.Logger()),
	); err != nil {
		log.Fatal(err)
	}

	app.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(` + "`" + `{"status":"ok"}` + "`" + `))
	})

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

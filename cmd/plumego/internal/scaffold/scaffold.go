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
		return []string{
			"cmd/app/main.go",
			"internal/app/app.go",
			"internal/app/routes.go",
			"internal/handler/api.go",
			"internal/handler/health.go",
			"internal/config/config.go",
			"internal/resource/users.go",
			"go.mod",
			"env.example",
			".gitignore",
			"README.md",
		}
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
		if template == "canonical" || template == "api" {
			return getCanonicalMainGoContent(module, name)
		}
		return getMainGoContent(module, template)
	case "internal/app/app.go":
		return getCanonicalAppGoContent(module)
	case "internal/app/routes.go":
		if template == "api" {
			return getAPIRoutesGoContent(module, name)
		}
		return getCanonicalRoutesGoContent(module, name)
	case "internal/handler/api.go":
		return getCanonicalAPIHandlerContent(name)
	case "internal/handler/health.go":
		return getCanonicalHealthHandlerContent()
	case "internal/config/config.go":
		return getCanonicalConfigGoContent(module)
	case "internal/resource/users.go":
		return getAPIUsersResourceContent()
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

	"github.com/spcent/plumego/contract"
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

	"github.com/spcent/plumego/contract"
)

type healthResponse struct {
	Status string ` + "`json:\"status\"`" + `
}

// Health responds with a simple JSON liveness check.
func Health(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{Status: "ok"}, nil)
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

type helloResponse struct {
	Message   string `+"`json:\"message\"`"+`
	Timestamp string `+"`json:\"timestamp\"`"+`
}

// Hello responds with service metadata.
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	resp := helloResponse{
		Message:   "hello from %s",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}
`, name)
	case "user.go":
		return fmt.Sprintf(`package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/contract"
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
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to list users").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, users, nil)
}

// Get handles GET /api/v1/users/:id
func (h UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := contract.RequestContextFromContext(r.Context()).Params["id"]
	u, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Message("user not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, u, nil)
}

// Create handles POST /api/v1/users
func (h UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Category(contract.CategoryValidation).
			Build())
		return
	}
	u, err := h.Service.Create(r.Context(), req.Name, req.Email)
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to create user").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusCreated, u, nil)
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

type metricsSummaryResponse struct {
	Uptime    string ` + "`json:\"uptime\"`" + `
	Timestamp string ` + "`json:\"timestamp\"`" + `
}

// Summary responds with a basic runtime metrics snapshot.
func (h MetricsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, metricsSummaryResponse{
		Uptime:    time.Since(startTime).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil)
}

var startTime = time.Now()
`
	case "error.go":
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
	"net/http"

	"%s/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the application.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "%s"}

	if err := a.Core.Get("/", http.HandlerFunc(api.Hello)); err != nil {
		return err
	}
	if err := a.Core.Get("/healthz", http.HandlerFunc(health.Live)); err != nil {
		return err
	}
	if err := a.Core.Get("/readyz", http.HandlerFunc(health.Ready)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/hello", http.HandlerFunc(api.Hello)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/status", http.HandlerFunc(api.Status)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/v1/greet", http.HandlerFunc(api.Greet)); err != nil {
		return err
	}
	return nil
}
`, module, name)
}

func getAPIRoutesGoContent(module, name string) string {
	return fmt.Sprintf(`package app

import (
	"net/http"

	"%s/internal/handler"
	"%s/internal/resource"

	"github.com/spcent/plumego/x/rest"
)

// RegisterRoutes wires all HTTP routes for the application.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "%s"}

	if err := a.Core.Get("/", http.HandlerFunc(api.Hello)); err != nil {
		return err
	}
	if err := a.Core.Get("/healthz", http.HandlerFunc(health.Live)); err != nil {
		return err
	}
	if err := a.Core.Get("/readyz", http.HandlerFunc(health.Ready)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/hello", http.HandlerFunc(api.Hello)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/status", http.HandlerFunc(api.Status)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/v1/greet", http.HandlerFunc(api.Greet)); err != nil {
		return err
	}

	spec := rest.DefaultResourceSpec("users").WithPrefix("/api/users")
	users := rest.NewDBResource[resource.User](spec, resource.NewUserRepository())
	if err := a.Core.Get(spec.Prefix, http.HandlerFunc(users.Index)); err != nil {
		return err
	}
	if err := a.Core.Get(spec.Prefix+"/:id", http.HandlerFunc(users.Show)); err != nil {
		return err
	}
	if err := a.Core.Post(spec.Prefix, http.HandlerFunc(users.Create)); err != nil {
		return err
	}
	return nil
}
`, module, module, name)
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

type helloResponse struct {
	Message   string            `+"`json:\"message\"`"+`
	Service   string            `+"`json:\"service\"`"+`
	Mode      string            `+"`json:\"mode\"`"+`
	Timestamp string            `+"`json:\"timestamp\"`"+`
	Version   string            `+"`json:\"version\"`"+`
	Features  []string          `+"`json:\"features\"`"+`
	Endpoints map[string]string `+"`json:\"endpoints\"`"+`
}

type greetResponse struct {
	Message string `+"`json:\"message\"`"+`
}

type statusResponse struct {
	Status    string          `+"`json:\"status\"`"+`
	Service   string          `+"`json:\"service\"`"+`
	Version   string          `+"`json:\"version\"`"+`
	Timestamp string          `+"`json:\"timestamp\"`"+`
	Structure statusStructure `+"`json:\"structure\"`"+`
	Modules   []string        `+"`json:\"modules\"`"+`
}

type statusStructure struct {
	Bootstrap  string `+"`json:\"bootstrap\"`"+`
	Extensions string `+"`json:\"extensions\"`"+`
	Handlers   string `+"`json:\"handlers\"`"+`
	Routes     string `+"`json:\"routes\"`"+`
}

// Hello responds with service metadata and available endpoints.
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	resp := helloResponse{
		Message:   "hello from %s standard-service",
		Service:   "%s",
		Mode:      "canonical",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Features: []string{
			"stable_root_only",
			"explicit_routes",
			"stdlib_handlers",
			"minimal_bootstrap",
		},
		Endpoints: map[string]string{
			"root":       "/",
			"healthz":    "/healthz",
			"readyz":     "/readyz",
			"api_hello":  "/api/hello",
			"api_status": "/api/status",
		},
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}

// Greet demonstrates the canonical error response pattern.
func (h APIHandler) Greet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Detail("field", "name").
			Message("name is required").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, greetResponse{Message: "hello, " + name}, nil)
}

// Status responds with a summary of system health and component state.
func (h APIHandler) Status(w http.ResponseWriter, r *http.Request) {
	resp := statusResponse{
		Status:    "healthy",
		Service:   "%s",
		Version:   "1.0.0",
		Timestamp: time.Now().Format(time.RFC3339),
		Structure: statusStructure{
			Bootstrap:  "explicit",
			Extensions: "excluded_from_canonical_path",
			Handlers:   "net/http",
			Routes:     "one_method_one_path_one_handler",
		},
		Modules: []string{
			"core",
			"router",
			"contract",
			"middleware",
		},
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}
`, name, name, name)
}

func getAPIUsersResourceContent() string {
	return `package resource

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/spcent/plumego/x/rest"
)

// User is the sample resource exposed by the API template.
type User struct {
	ID    string ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}

// UserRepository is an in-memory rest.Repository implementation.
type UserRepository struct {
	mu    sync.Mutex
	seq   int
	items map[string]User
}

// NewUserRepository returns a seeded in-memory repository.
func NewUserRepository() *UserRepository {
	repo := &UserRepository{items: make(map[string]User)}
	_ = repo.Create(context.Background(), &User{Name: "Ada", Email: "ada@example.com"})
	return repo
}

// FindAll returns all users.
func (r *UserRepository) FindAll(_ context.Context, _ *rest.QueryParams) ([]User, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	users := make([]User, 0, len(r.items))
	for _, user := range r.items {
		users = append(users, user)
	}
	return users, int64(len(users)), nil
}

// FindByID returns one user by ID.
func (r *UserRepository) FindByID(_ context.Context, id string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.items[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return &user, nil
}

// Create inserts a user.
func (r *UserRepository) Create(_ context.Context, data *User) error {
	if data == nil {
		return fmt.Errorf("user is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if data.ID == "" {
		r.seq++
		data.ID = fmt.Sprintf("%d", r.seq)
	}
	r.items[data.ID] = *data
	return nil
}

// Update replaces a user.
func (r *UserRepository) Update(_ context.Context, id string, data *User) error {
	if data == nil {
		return fmt.Errorf("user is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return sql.ErrNoRows
	}
	data.ID = id
	r.items[id] = *data
	return nil
}

// Delete removes a user.
func (r *UserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.items, id)
	return nil
}

// Count returns the number of users.
func (r *UserRepository) Count(_ context.Context, _ *rest.QueryParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.items)), nil
}

// Exists reports whether a user exists.
func (r *UserRepository) Exists(_ context.Context, id string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.items[id]
	return ok, nil
}
`
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

type healthResponse struct {
	Status    string ` + "`json:\"status\"`" + `
	Service   string ` + "`json:\"service\"`" + `
	Check     string ` + "`json:\"check\"`" + `
	Timestamp string ` + "`json:\"timestamp\"`" + `
}

// Live reports that the process is serving HTTP traffic.
func (h HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	service := h.ServiceName
	if service == "" {
		service = "plumego-reference"
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   service,
		Check:     "liveness",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil)
}

// Ready reports that the application is ready to accept requests.
func (h HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	service := h.ServiceName
	if service == "" {
		service = "plumego-reference"
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ready",
		Service:   service,
		Check:     "readiness",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil)
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

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/recovery"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
)

type healthResponse struct {
	Status string ` + "`json:\"status\"`" + `
}

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
		_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{Status: "ok"}, nil)
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

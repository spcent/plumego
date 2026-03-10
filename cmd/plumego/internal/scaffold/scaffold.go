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
		return getMainGoContent(module, template)
	case "go.mod":
		return fmt.Sprintf("module %s\n\ngo 1.24\n", module)
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
	"github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/middleware/recovery"
)

// New constructs the HTTP application with middleware and routes.
func New() *core.App {
	app := core.New(core.WithAddr(":8080"))
	if err := app.Use(
		observability.RequestID(),
		observability.Logging(app.Logger(), nil, nil),
		recovery.RecoveryMiddleware,
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

func getMainGoContent(module, template string) string {
	_ = module // module path is in go.mod, not needed inline
	return `package main

import (
	"log"
	"net/http"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/middleware/recovery"
)

func main() {
	app := core.New(core.WithAddr(":8080"))
	if err := app.Use(
		observability.RequestID(),
		observability.Logging(app.Logger(), nil, nil),
		recovery.RecoveryMiddleware,
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

package scaffold

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/spcent/plumego/cmd/plumego/internal/executil"
)

var scenarioProfiles = map[string]struct{}{
	"rest-api":    {},
	"tenant-api":  {},
	"gateway":     {},
	"realtime":    {},
	"ai-service":  {},
	"ops-service": {},
}

const defaultPlumegoVersion = "v0.0.0"

// ProjectOptions configures project-level scaffold output.
type ProjectOptions struct {
	PlumegoVersion string
	PlumegoReplace string
	CleanExisting  bool
}

// GetTemplateFiles returns the files that would be created for a template.
func GetTemplateFiles(template string) []string {
	switch template {
	case "canonical":
		// Mirrors reference/standard-service exactly: explicit bootstrap,
		// stable-root-only imports, constructor injection, explicit routes.
		return canonicalTemplateFiles()
	case "minimal", "fullstack", "microservice":
		return canonicalTemplateFiles()
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
	case "rest-api":
		return []string{
			"cmd/app/main.go",
			"internal/app/app.go",
			"internal/app/routes.go",
			"internal/handler/api.go",
			"internal/handler/health.go",
			"internal/config/config.go",
			"internal/resource/users.go",
			"internal/scenario/profile.go",
			"go.mod",
			"env.example",
			".gitignore",
			"README.md",
		}
	case "tenant-api", "gateway", "realtime", "ai-service", "ops-service":
		return []string{
			"cmd/app/main.go",
			"internal/app/app.go",
			"internal/app/routes.go",
			"internal/handler/api.go",
			"internal/handler/health.go",
			"internal/config/config.go",
			"internal/scenario/profile.go",
			"go.mod",
			"env.example",
			".gitignore",
			"README.md",
		}
	default:
		return nil
	}
}

func canonicalTemplateFiles() []string {
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
}

// CreateProject creates a new project from a template.
func CreateProject(dir, name, module, template string, initGit bool, options ...ProjectOptions) ([]string, error) {
	if err := validateProjectName(name); err != nil {
		return nil, err
	}
	if err := validateModulePath(module); err != nil {
		return nil, err
	}

	files := GetTemplateFiles(template)
	if len(files) == 0 {
		return nil, fmt.Errorf("unknown project template: %s", template)
	}
	projectOptions := resolveProjectOptions(options)
	if projectOptions.CleanExisting {
		if err := cleanKnownTemplateFiles(dir); err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	created := []string{}

	for _, file := range files {
		path := filepath.Join(dir, file)

		parent := filepath.Dir(path)
		if err := os.MkdirAll(parent, 0755); err != nil {
			return created, fmt.Errorf("failed to create directory %s: %w", parent, err)
		}

		content := getTemplateContent(file, name, module, template)
		if file == "go.mod" {
			content = getGoModContent(module, projectOptions)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return created, fmt.Errorf("failed to write file %s: %w", path, err)
		}

		created = append(created, file)
	}

	if initGit {
		result, err := executil.Run(context.Background(), executil.Options{
			Name:        "git",
			Args:        []string{"init"},
			Dir:         dir,
			Timeout:     30 * time.Second,
			OutputLimit: 16 * 1024,
		})
		if err == nil {
			created = append(created, ".git/")
		} else {
			output := strings.TrimSpace(result.CombinedOutput())
			if output == "" {
				return created, fmt.Errorf("failed to initialize git repository: %w", err)
			}
			return created, fmt.Errorf("failed to initialize git repository: %w: %s", err, output)
		}
	}

	return created, nil
}

func cleanKnownTemplateFiles(dir string) error {
	for _, file := range allTemplateFiles() {
		path := filepath.Join(dir, file)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to inspect stale template file %s: %w", path, err)
		}
		if info.IsDir() {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove stale template file %s: %w", path, err)
		}
	}
	return nil
}

func allTemplateFiles() []string {
	seen := make(map[string]struct{})
	var files []string
	for _, template := range []string{
		"canonical",
		"minimal",
		"api",
		"fullstack",
		"microservice",
		"rest-api",
		"tenant-api",
		"gateway",
		"realtime",
		"ai-service",
		"ops-service",
	} {
		for _, file := range GetTemplateFiles(template) {
			if _, ok := seen[file]; ok {
				continue
			}
			seen[file] = struct{}{}
			files = append(files, file)
		}
	}
	return files
}

func validateProjectName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	if name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid project name: %s", name)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("invalid project name: %s", name)
	}
	for _, r := range name {
		if r == '-' || r == '_' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return fmt.Errorf("invalid project name: %s", name)
	}
	return nil
}

func validateModulePath(module string) error {
	module = strings.TrimSpace(module)
	if module == "" {
		return fmt.Errorf("module path is required")
	}
	if strings.ContainsAny(module, "\\ \t\r\n") || strings.Contains(module, "://") {
		return fmt.Errorf("invalid module path: %s", module)
	}
	if strings.HasPrefix(module, "/") || strings.HasSuffix(module, "/") || strings.Contains(module, "//") {
		return fmt.Errorf("invalid module path: %s", module)
	}
	for _, segment := range strings.Split(module, "/") {
		if segment == "" || segment == "." || segment == ".." || strings.HasPrefix(segment, "-") {
			return fmt.Errorf("invalid module path: %s", module)
		}
		for _, r := range segment {
			if r == '-' || r == '_' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r) {
				continue
			}
			return fmt.Errorf("invalid module path: %s", module)
		}
	}
	return nil
}

func resolveProjectOptions(options []ProjectOptions) ProjectOptions {
	resolved := ProjectOptions{PlumegoVersion: defaultPlumegoVersion}
	if len(options) > 0 {
		resolved = options[0]
	}
	if resolved.PlumegoVersion == "" {
		resolved.PlumegoVersion = defaultPlumegoVersion
	}
	return resolved
}

func getGoModContent(module string, opts ProjectOptions) string {
	opts = resolveProjectOptions([]ProjectOptions{opts})
	content := fmt.Sprintf("module %s\n\ngo 1.24\n\nrequire github.com/spcent/plumego %s\n", module, opts.PlumegoVersion)
	if opts.PlumegoReplace != "" {
		content += fmt.Sprintf("\nreplace github.com/spcent/plumego => %s\n", filepath.ToSlash(opts.PlumegoReplace))
	}
	return content
}

func getTemplateContent(file, name, module, template string) string {
	switch file {
	case "cmd/app/main.go":
		return getCanonicalMainGoContent(module, name)
	case "internal/app/app.go":
		return getCanonicalAppGoContent(module)
	case "internal/app/routes.go":
		if template == "tenant-api" {
			return getTenantAPIRoutesGoContent(module, name)
		}
		if template == "gateway" {
			return getGatewayRoutesGoContent(module, name)
		}
		if template == "realtime" {
			return getRealtimeRoutesGoContent(module, name)
		}
		if template == "ai-service" {
			return getAIServiceRoutesGoContent(module, name)
		}
		if template == "ops-service" {
			return getOpsServiceRoutesGoContent(module, name)
		}
		if template == "api" || template == "rest-api" {
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
	case "internal/scenario/profile.go":
		return getScenarioProfileContent(template)
	case "go.mod":
		return getGoModContent(module, ProjectOptions{})
	case "env.example":
		return getEnvExampleContent()
	case ".gitignore":
		return getGitignoreContent()
	case "README.md":
		return getReadmeContent(name, template)
	default:
		return ""
	}
}

func getCanonicalMainGoContent(module, name string) string {
	return fmt.Sprintf(`package main

import (
	"log"
	"os"

	"%s/internal/app"
	"%s/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Printf("server stopped: %%v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	a, err := app.New(cfg)
	if err != nil {
		return err
	}

	if err := a.RegisterRoutes(); err != nil {
		return err
	}

	log.Printf("Starting %s on %%s", cfg.Core.Addr)
	return a.Start()
}
`, module, module, name)
}

func getCanonicalAppGoContent(module string) string {
	return fmt.Sprintf(`// Package app wires together the application dependencies and manages the
// server lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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
	if err := a.Use(
		requestid.Middleware(),
		recovery.Recovery(a.Logger()),
		accesslog.Middleware(a.Logger(), nil, nil),
	); err != nil {
		return nil, fmt.Errorf("register middleware: %%w", err)
	}

	return &App{Core: a, Cfg: cfg}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
func (a *App) Start() (err error) {
	ctx := context.Background()

	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %%w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %%w", err)
	}
	defer func() {
		if shutdownErr := a.Core.Shutdown(ctx); shutdownErr != nil && err == nil {
			err = fmt.Errorf("shutdown server: %%w", shutdownErr)
		}
	}()

	if a.Cfg.Core.TLS.Enabled {
		if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server stopped: %%w", err)
		}
		return nil
	}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func getTenantAPIRoutesGoContent(module, name string) string {
	return fmt.Sprintf(`package app

import (
	"net/http"

	"%s/internal/handler"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	"github.com/spcent/plumego/x/tenant/policy"
	"github.com/spcent/plumego/x/tenant/quota"
	"github.com/spcent/plumego/x/tenant/ratelimit"
	"github.com/spcent/plumego/x/tenant/resolve"
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

	tenantConfig := tenantcore.NewInMemoryConfigManager()
	tenantConfig.SetTenantConfig(tenantcore.Config{
		TenantID: "tenant-a",
		Policy: tenantcore.PolicyConfig{
			AllowedModels: []string{"gpt-4o"},
		},
		Quota: tenantcore.QuotaConfig{
			Limits: []tenantcore.QuotaLimit{
				{Window: tenantcore.QuotaWindowMinute, Requests: 100},
			},
		},
	})
	rateLimits := tenantcore.NewInMemoryRateLimitManager()
	rateLimits.SetRateLimit("tenant-a", tenantcore.RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             10,
	})
	tenantChain := middleware.NewChain(
		resolve.Middleware(resolve.Options{DisablePrincipal: true}),
		policy.Middleware(policy.Options{Evaluator: tenantcore.NewConfigPolicyEvaluator(tenantConfig)}),
		quota.Middleware(quota.Options{Manager: tenantcore.NewFixedWindowQuotaManager(tenantConfig)}),
		ratelimit.Middleware(ratelimit.Options{Limiter: tenantcore.NewTokenBucketRateLimiter(rateLimits)}),
	)
	if err := a.Core.Get("/api/models", tenantChain.Build(http.HandlerFunc(models))); err != nil {
		return err
	}
	return nil
}

type modelsResponse struct {
	TenantID string   `+"`json:\"tenant_id\"`"+`
	Models   []string `+"`json:\"models\"`"+`
}

func models(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantcore.TenantIDFromContext(r.Context())
	_ = contract.WriteResponse(w, r, http.StatusOK, modelsResponse{
		TenantID: tenantID,
		Models:   []string{"gpt-4o"},
	}, nil)
}
`, module, name)
}

func getGatewayRoutesGoContent(module, name string) string {
	return fmt.Sprintf(`package app

import (
	"net/http"
	"strings"

	"%s/internal/handler"

	"github.com/spcent/plumego/x/gateway"
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

	proxy, err := gateway.NewGatewayE(gateway.GatewayConfig{
		Targets:     []string{gatewayLoopbackTarget(a.Cfg.Core.Addr)},
		PathRewrite: gateway.ReplacePrefix("/edge", "/api/status"),
		RetryCount:  0,
	})
	if err != nil {
		return err
	}
	if err := a.Core.Get("/edge", proxy); err != nil {
		return err
	}
	return nil
}

func gatewayLoopbackTarget(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr
	}
	return "http://" + addr
}
`, module, name)
}

func getRealtimeRoutesGoContent(module, name string) string {
	return fmt.Sprintf(`package app

import (
	"net/http"

	"%s/internal/handler"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/x/websocket"
)

// RegisterRoutes wires all HTTP routes for the application.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "%s"}
	hub := websocket.NewHub(4, 1024)

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
	if err := a.Core.Get("/realtime/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteResponse(w, r, http.StatusOK, hub.Metrics(), nil)
	})); err != nil {
		return err
	}
	return nil
}
`, module, name)
}

func getAIServiceRoutesGoContent(module, name string) string {
	return fmt.Sprintf(`package app

import (
	"net/http"

	"%s/internal/handler"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/x/ai/provider"
	"github.com/spcent/plumego/x/ai/session"
	"github.com/spcent/plumego/x/ai/tool"
)

// RegisterRoutes wires all HTTP routes for the application.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "%s"}
	offline := provider.NewMockProvider("offline")
	offline.QueueResponse(&provider.CompletionResponse{
		ID:         "offline-demo",
		Model:      "mock-model",
		Role:       provider.RoleAssistant,
		Content:    []provider.ContentBlock{{Type: provider.ContentTypeText, Text: "offline ai response"}},
		StopReason: provider.StopReasonEndTurn,
	})
	sessions := session.NewManager(session.NewMemoryStorage())
	tools := tool.NewRegistry()
	_ = tools.Register(tool.NewEchoTool())

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
	if err := a.Core.Get("/ai/demo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		created, err := sessions.Create(r.Context(), session.CreateOptions{
			TenantID: "demo-tenant",
			UserID:   "demo-user",
			AgentID:  "offline-agent",
			Model:    "mock-model",
		})
		if err != nil {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).
				Message("failed to create session").
				Build())
			return
		}
		resp, err := offline.Complete(r.Context(), &provider.CompletionRequest{
			Model: "mock-model",
			Messages: []provider.Message{{
				Role:    provider.RoleUser,
				Content: "hello",
			}},
			Tools: tools.ToProviderTools(r.Context()),
		})
		if err != nil {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).
				Message("offline provider failed").
				Build())
			return
		}
		_ = contract.WriteResponse(w, r, http.StatusOK, aiDemoResponse{
			SessionID: created.ID,
			Provider:  offline.Name(),
			Model:     resp.Model,
			ToolCount: len(tools.ListForContext(r.Context())),
			Content:   resp.GetText(),
		}, nil)
	})); err != nil {
		return err
	}
	return nil
}

type aiDemoResponse struct {
	SessionID string `+"`json:\"session_id\"`"+`
	Provider  string `+"`json:\"provider\"`"+`
	Model     string `+"`json:\"model\"`"+`
	ToolCount int    `+"`json:\"tool_count\"`"+`
	Content   string `+"`json:\"content\"`"+`
}
`, module, name)
}

func getOpsServiceRoutesGoContent(module, name string) string {
	return fmt.Sprintf(`package app

import (
	"net/http"
	"os"
	"time"

	"%s/internal/handler"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/security/authn"
	"github.com/spcent/plumego/x/observability"
	"github.com/spcent/plumego/x/ops"
)

// RegisterRoutes wires all HTTP routes for the application.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "%s"}
	collector := observability.NewPrometheusCollector("app")
	collector.ObserveHTTP(nil, "GET", "/healthz", http.StatusOK, 0, time.Millisecond)
	metrics := observability.NewPrometheusExporter(collector).Handler()
	opsAuth := auth.Authenticate(authn.StaticToken(os.Getenv("OPS_TOKEN")), auth.WithAuthRealm("ops-service"))

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
	if err := a.Core.Get("/ops/metrics", opsAuth(metrics)); err != nil {
		return err
	}
	if err := a.Core.Get("/ops/admin", opsAuth(http.HandlerFunc(opsAdmin))); err != nil {
		return err
	}
	return nil
}

type opsAdminResponse struct {
	AdminRoutes string         `+"`json:\"admin_routes\"`"+`
	DebugRoutes string         `+"`json:\"debug_routes\"`"+`
	Metrics     string         `+"`json:\"metrics\"`"+`
	Queue       ops.QueueStats `+"`json:\"queue\"`"+`
}

func opsAdmin(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, opsAdminResponse{
		AdminRoutes: "protected",
		DebugRoutes: "not_mounted_by_default",
		Metrics:     "/ops/metrics",
		Queue: ops.QueueStats{
			Queue:     "demo",
			Queued:    0,
			Leased:    0,
			Dead:      0,
			Expired:   0,
			UpdatedAt: time.Now().UTC(),
		},
	}, nil)
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
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spcent/plumego/core"
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
	if value := strings.TrimSpace(os.Getenv("APP_ADDR")); value != "" {
		cfg.Core.Addr = value
	}
	if value := strings.TrimSpace(os.Getenv("APP_ENV_FILE")); value != "" {
		cfg.App.EnvFile = value
	}
	if value := strings.TrimSpace(os.Getenv("APP_DEBUG")); value != "" {
		cfg.App.Debug = value == "true" || value == "1"
	}
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
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("invalid env line: %s", line)
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), ` + "`\"'`" + `)
		if key == "" {
			return fmt.Errorf("invalid empty env key")
		}
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
`
}

func getScenarioProfileContent(template string) string {
	switch template {
	case "rest-api":
		return `// Package scenario documents the selected scaffold profile.
package scenario

import "github.com/spcent/plumego/x/rest"

// Name identifies the scaffold profile.
const Name = "rest-api"

// Capabilities returns the optional Plumego capability families this profile uses.
func Capabilities() []string {
	spec := rest.DefaultResourceSpec("users").WithPrefix("/api/users")
	return []string{"x/rest", spec.Prefix}
}
`
	case "tenant-api":
		return `// Package scenario documents the selected scaffold profile.
package scenario

import (
	"github.com/spcent/plumego/x/tenant/policy"
	"github.com/spcent/plumego/x/tenant/quota"
	"github.com/spcent/plumego/x/tenant/ratelimit"
	"github.com/spcent/plumego/x/tenant/resolve"
)

// Name identifies the scaffold profile.
const Name = "tenant-api"

// Capabilities returns the optional Plumego capability families this profile uses.
func Capabilities() []string {
	_ = resolve.Middleware
	_ = policy.Middleware
	_ = quota.Middleware
	_ = ratelimit.Middleware
	return []string{"x/tenant/resolve", "x/tenant/policy", "x/tenant/quota", "x/tenant/ratelimit"}
}
`
	case "gateway":
		return `// Package scenario documents the selected scaffold profile.
package scenario

import "github.com/spcent/plumego/x/gateway"

// Name identifies the scaffold profile.
const Name = "gateway"

// Capabilities returns the optional Plumego capability families this profile uses.
func Capabilities() []string {
	_ = gateway.RegisterProxy
	return []string{"x/gateway"}
}
`
	case "realtime":
		return `// Package scenario documents the selected scaffold profile.
package scenario

import (
	"github.com/spcent/plumego/x/messaging"
	"github.com/spcent/plumego/x/websocket"
)

// Name identifies the scaffold profile.
const Name = "realtime"

// Capabilities returns the optional Plumego capability families this profile uses.
func Capabilities() []string {
	_ = websocket.NewHub
	_ = messaging.New
	return []string{"x/websocket", "x/messaging"}
}
`
	case "ai-service":
		return `// Package scenario documents the selected scaffold profile.
package scenario

import (
	"github.com/spcent/plumego/x/ai/provider"
	"github.com/spcent/plumego/x/ai/session"
	"github.com/spcent/plumego/x/ai/streaming"
	"github.com/spcent/plumego/x/ai/tool"
)

// Name identifies the scaffold profile.
const Name = "ai-service"

// Capabilities returns the optional Plumego capability families this profile uses.
func Capabilities() []string {
	_ = provider.NewMockProvider
	_ = session.NewManager
	_ = streaming.NewStreamManager
	_ = tool.NewRegistry
	return []string{"x/ai/provider", "x/ai/session", "x/ai/streaming", "x/ai/tool"}
}
`
	case "ops-service":
		return `// Package scenario documents the selected scaffold profile.
package scenario

import (
	"github.com/spcent/plumego/x/observability"
	"github.com/spcent/plumego/x/ops"
)

// Name identifies the scaffold profile.
const Name = "ops-service"

// Capabilities returns the optional Plumego capability families this profile uses.
func Capabilities() []string {
	_ = observability.Configure
	_ = ops.New
	return []string{"x/observability", "x/ops"}
}
`
	default:
		return `// Package scenario documents the selected scaffold profile.
package scenario

// Name identifies the scaffold profile.
const Name = "custom"

// Capabilities returns optional Plumego capability families this profile uses.
func Capabilities() []string {
	return nil
}
`
	}
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

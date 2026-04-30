package scaffold

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spcent/plumego/cmd/plumego/internal/testassert"
)

var allTemplates = []string{
	"minimal",
	"api",
	"rest-api",
	"tenant-api",
	"gateway",
	"realtime",
	"ai-service",
	"ops-service",
	"fullstack",
	"microservice",
	"canonical",
}

func assertParseableGo(t *testing.T, filename string, content string) {
	t.Helper()

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, filename, content, parser.AllErrors); err != nil {
		t.Errorf("%s parse error: %v\ncontent:\n%s", filename, err, content)
	}
}

func requiresMainPackage(file string) bool {
	return strings.HasPrefix(file, "cmd/") && filepath.Ext(file) == ".go"
}

func assertFileOmitted(t *testing.T, files []string, file string, reason string) {
	t.Helper()

	if slices.Contains(files, file) {
		t.Fatalf("%s should not emit %s", reason, file)
	}
}

func assertContainsAll(t *testing.T, content string, patterns []string) {
	t.Helper()

	for _, pattern := range patterns {
		if !strings.Contains(content, pattern) {
			t.Fatalf("content missing required pattern %q:\n%s", pattern, content)
		}
	}
}

func assertContainsNone(t *testing.T, content string, patterns []string) {
	t.Helper()

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			t.Fatalf("content contains disallowed pattern %q:\n%s", pattern, content)
		}
	}
}

// TestGetTemplateFiles_NoEmpty verifies every template returns at least one file.
func TestGetTemplateFiles_NoEmpty(t *testing.T) {
	for _, tmpl := range allTemplates {
		t.Run(tmpl, func(t *testing.T) {
			files := GetTemplateFiles(tmpl)
			if len(files) == 0 {
				t.Errorf("template %q returned no files", tmpl)
			}
		})
	}
}

// TestTemplateContent_NoTODO verifies no generated file contains a bare // TODO comment.
func TestTemplateContent_NoTODO(t *testing.T) {
	const (
		testName   = "myapp"
		testModule = "example.com/myapp"
	)
	for _, tmpl := range allTemplates {
		t.Run(tmpl, func(t *testing.T) {
			files := GetTemplateFiles(tmpl)
			for _, file := range files {
				content := getTemplateContent(file, testName, testModule, tmpl)
				testassert.NoBareTODO(t, "template="+tmpl+" file="+file, content)
			}
		})
	}
}

// TestTemplateContent_GoFilesParseable verifies all generated .go files are
// syntactically valid Go source.
func TestTemplateContent_GoFilesParseable(t *testing.T) {
	const (
		testName   = "myapp"
		testModule = "example.com/myapp"
	)
	for _, tmpl := range allTemplates {
		t.Run(tmpl, func(t *testing.T) {
			files := GetTemplateFiles(tmpl)
			for _, file := range files {
				if filepath.Ext(file) != ".go" {
					continue
				}
				content := getTemplateContent(file, testName, testModule, tmpl)
				if content == "" {
					continue
				}
				assertParseableGo(t, file, content)
			}
		})
	}
}

// TestTemplateContent_CorrectPackageNames verifies Go files declare a package
// name that matches their directory name (or "main" for cmd/ paths).
func TestTemplateContent_CorrectPackageNames(t *testing.T) {
	const (
		testName   = "myapp"
		testModule = "example.com/myapp"
	)
	fset := token.NewFileSet()
	for _, tmpl := range allTemplates {
		files := GetTemplateFiles(tmpl)
		for _, file := range files {
			if filepath.Ext(file) != ".go" {
				continue
			}
			content := getTemplateContent(file, testName, testModule, tmpl)
			if content == "" {
				continue
			}
			f, err := parser.ParseFile(fset, file, content, 0)
			if err != nil {
				continue // parse errors caught elsewhere
			}
			pkg := f.Name.Name
			if pkg == "" {
				t.Errorf("template=%q file=%q has empty package name", tmpl, file)
				continue
			}
			if requiresMainPackage(file) {
				if pkg != "main" {
					t.Errorf("template=%q file=%q: expected package main for cmd path, got %q",
						tmpl, file, pkg)
				}
			}
		}
	}
}

// TestDefaultFileContent_NoTODO verifies getDefaultFileContent never emits // TODO.
func TestDefaultFileContent_NoTODO(t *testing.T) {
	cases := []struct {
		file   string
		name   string
		module string
	}{
		{"internal/httpapp/routes.go", "myapp", "example.com/myapp"},
		{"internal/httpapp/handlers/user.go", "myapp", "example.com/myapp"},
		{"internal/domain/user/service.go", "myapp", "example.com/myapp"},
		{"internal/domain/user/repository.go", "myapp", "example.com/myapp"},
		{"internal/httpapp/handlers/metrics.go", "myapp", "example.com/myapp"},
		{"internal/httpapp/handlers/health.go", "myapp", "example.com/myapp"},
		{"frontend/index.html", "myapp", "example.com/myapp"},
		{"frontend/app.js", "myapp", "example.com/myapp"},
		{"frontend/styles.css", "myapp", "example.com/myapp"},
		{"Dockerfile", "myapp", "example.com/myapp"},
		{"docker-compose.yml", "myapp", "example.com/myapp"},
		{"internal/unknown/foo.go", "myapp", "example.com/myapp"},
	}

	for _, tc := range cases {
		content := getDefaultFileContent(tc.file, tc.name, tc.module)
		testassert.NoBareTODO(t, "file="+tc.file, content)
	}
}

func TestGetTemplateFiles_MicroserviceDoesNotEmitLegacyHTTPHelpers(t *testing.T) {
	files := GetTemplateFiles("microservice")
	assertFileOmitted(t, files, "internal/platform/httpjson/response.go", "microservice template")
	assertFileOmitted(t, files, "internal/platform/httperr/error.go", "microservice template")
}

func TestTemplateContent_UsesCanonicalHTTPContract(t *testing.T) {
	const (
		testName   = "myapp"
		testModule = "example.com/myapp"
	)

	disallowed := []string{
		"internal/platform/httpjson",
		"internal/platform/httperr",
		"PathValue(",
		"http.Error(",
		"json.NewEncoder(w).Encode",
		`w.Header().Set("Content-Type", "application/json")`,
		`"encoding error"`,
	}

	for _, tmpl := range allTemplates {
		files := GetTemplateFiles(tmpl)
		for _, file := range files {
			if filepath.Ext(file) != ".go" {
				continue
			}
			content := getTemplateContent(file, testName, testModule, tmpl)
			assertContainsNone(t, content, disallowed)
		}
	}
}

func TestAPITemplate_UsesCanonicalBootstrapWithRestProfile(t *testing.T) {
	files := GetTemplateFiles("api")
	assertContainsAll(t, strings.Join(files, "\n"), []string{
		"internal/app/app.go",
		"internal/app/routes.go",
		"internal/handler/api.go",
		"internal/handler/health.go",
		"internal/config/config.go",
		"internal/resource/users.go",
	})
	assertContainsNone(t, strings.Join(files, "\n"), []string{
		"internal/httpapp/",
		"internal/domain/user/",
	})

	routes := getTemplateContent("internal/app/routes.go", "myapp", "example.com/myapp", "api")
	assertContainsAll(t, routes, []string{
		`"github.com/spcent/plumego/x/rest"`,
		`rest.DefaultResourceSpec("users").WithPrefix("/api/users")`,
		`rest.NewDBResource[resource.User](spec, resource.NewUserRepository())`,
		`a.Core.Get(spec.Prefix, http.HandlerFunc(users.Index))`,
		`a.Core.Get(spec.Prefix+"/:id", http.HandlerFunc(users.Show))`,
		`a.Core.Post(spec.Prefix, http.HandlerFunc(users.Create))`,
	})
}

func TestScenarioProfiles_GenerateRunnableRoutes(t *testing.T) {
	restRoutes := getTemplateContent("internal/app/routes.go", "myapp", "example.com/myapp", "rest-api")
	assertContainsAll(t, restRoutes, []string{
		`"github.com/spcent/plumego/x/rest"`,
		`rest.NewDBResource[resource.User](spec, resource.NewUserRepository())`,
		`a.Core.Get(spec.Prefix, http.HandlerFunc(users.Index))`,
		`a.Core.Post(spec.Prefix, http.HandlerFunc(users.Create))`,
	})

	tenantRoutes := getTemplateContent("internal/app/routes.go", "myapp", "example.com/myapp", "tenant-api")
	assertContainsAll(t, tenantRoutes, []string{
		`"github.com/spcent/plumego/middleware"`,
		`"github.com/spcent/plumego/x/tenant/resolve"`,
		`"github.com/spcent/plumego/x/tenant/policy"`,
		`"github.com/spcent/plumego/x/tenant/quota"`,
		`"github.com/spcent/plumego/x/tenant/ratelimit"`,
		`tenantChain := middleware.NewChain(`,
		`a.Core.Get("/api/models", tenantChain.Build(http.HandlerFunc(models)))`,
		`tenantcore.TenantIDFromContext(r.Context())`,
	})

	gatewayRoutes := getTemplateContent("internal/app/routes.go", "myapp", "example.com/myapp", "gateway")
	assertContainsAll(t, gatewayRoutes, []string{
		`"github.com/spcent/plumego/x/gateway"`,
		`gateway.NewGatewayE(gateway.GatewayConfig{`,
		`PathRewrite: gateway.ReplacePrefix("/edge", "/api/status")`,
		`a.Core.Get("/edge", proxy)`,
	})

	realtimeRoutes := getTemplateContent("internal/app/routes.go", "myapp", "example.com/myapp", "realtime")
	assertContainsAll(t, realtimeRoutes, []string{
		`"github.com/spcent/plumego/x/websocket"`,
		`hub := websocket.NewHub(4, 1024)`,
		`a.Core.Get("/realtime/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)`,
		`hub.Metrics()`,
	})

	aiRoutes := getTemplateContent("internal/app/routes.go", "myapp", "example.com/myapp", "ai-service")
	assertContainsAll(t, aiRoutes, []string{
		`"github.com/spcent/plumego/x/ai/provider"`,
		`"github.com/spcent/plumego/x/ai/session"`,
		`"github.com/spcent/plumego/x/ai/tool"`,
		`provider.NewMockProvider("offline")`,
		`session.NewManager(session.NewMemoryStorage())`,
		`tools.Register(tool.NewEchoTool())`,
		`a.Core.Get("/ai/demo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)`,
	})

	opsRoutes := getTemplateContent("internal/app/routes.go", "myapp", "example.com/myapp", "ops-service")
	assertContainsAll(t, opsRoutes, []string{
		`"github.com/spcent/plumego/x/observability"`,
		`"github.com/spcent/plumego/x/ops"`,
		`observability.NewPrometheusCollector("app")`,
		`auth.Authenticate(authn.StaticToken(os.Getenv("OPS_TOKEN"))`,
		`a.Core.Get("/ops/metrics", opsAuth(metrics))`,
		`a.Core.Get("/ops/admin", opsAuth(http.HandlerFunc(opsAdmin)))`,
		`DebugRoutes: "not_mounted_by_default"`,
	})
}

func TestScenarioProfiles_UseCanonicalScaffoldWithExplicitCapabilityProfile(t *testing.T) {
	tests := []struct {
		template string
		want     []string
	}{
		{template: "rest-api", want: []string{`"github.com/spcent/plumego/x/rest"`}},
		{template: "tenant-api", want: []string{
			`"github.com/spcent/plumego/x/tenant/resolve"`,
			`"github.com/spcent/plumego/x/tenant/policy"`,
			`"github.com/spcent/plumego/x/tenant/quota"`,
			`"github.com/spcent/plumego/x/tenant/ratelimit"`,
		}},
		{template: "gateway", want: []string{`"github.com/spcent/plumego/x/gateway"`}},
		{template: "realtime", want: []string{
			`"github.com/spcent/plumego/x/websocket"`,
			`"github.com/spcent/plumego/x/messaging"`,
		}},
		{template: "ai-service", want: []string{
			`"github.com/spcent/plumego/x/ai/provider"`,
			`"github.com/spcent/plumego/x/ai/session"`,
			`"github.com/spcent/plumego/x/ai/streaming"`,
			`"github.com/spcent/plumego/x/ai/tool"`,
		}},
		{template: "ops-service", want: []string{
			`"github.com/spcent/plumego/x/observability"`,
			`"github.com/spcent/plumego/x/ops"`,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			files := GetTemplateFiles(tt.template)
			assertContainsAll(t, strings.Join(files, "\n"), []string{
				"cmd/app/main.go",
				"internal/app/app.go",
				"internal/app/routes.go",
				"internal/handler/api.go",
				"internal/handler/health.go",
				"internal/config/config.go",
				"internal/scenario/profile.go",
			})

			mainContent := getTemplateContent("cmd/app/main.go", "myapp", "example.com/myapp", tt.template)
			assertContainsAll(t, mainContent, []string{
				`"example.com/myapp/internal/app"`,
				`"example.com/myapp/internal/config"`,
			})

			profile := getTemplateContent("internal/scenario/profile.go", "myapp", "example.com/myapp", tt.template)
			assertContainsAll(t, profile, tt.want)
			assertContainsNone(t, profile, []string{
				"func init(",
				"DefaultProvider",
				"Global",
			})
		})
	}
}

func TestTemplateContent_UsesLocalResponseDTOs(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		template string
		want     []string
	}{
		{
			name:     "minimal main health",
			file:     "cmd/app/main.go",
			template: "minimal",
			want: []string{
				`"github.com/spcent/plumego/contract"`,
				"type healthResponse struct",
				"contract.WriteResponse(",
			},
		},
		{
			name:     "api health handler",
			file:     "internal/handler/health.go",
			template: "api",
			want: []string{
				"type healthResponse struct",
				"contract.WriteResponse(",
			},
		},
		{
			name:     "fullstack hello handler",
			file:     "internal/httpapp/handlers/api.go",
			template: "fullstack",
			want: []string{
				"type helloResponse struct",
				"contract.WriteResponse(",
			},
		},
		{
			name:     "canonical health handler",
			file:     "internal/handler/health.go",
			template: "canonical",
			want: []string{
				"type healthResponse struct",
				"contract.WriteResponse(",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := getTemplateContent(tt.file, "myapp", "example.com/myapp", tt.template)
			assertContainsAll(t, content, tt.want)
		})
	}
}

func TestCanonicalTemplate_MatchesReferenceRouteShape(t *testing.T) {
	content := getTemplateContent("internal/app/routes.go", "myapp", "example.com/myapp", "canonical")

	assertContainsAll(t, content, []string{
		`"net/http"`,
		`a.Core.Get("/", http.HandlerFunc(api.Hello))`,
		`a.Core.Get("/healthz", http.HandlerFunc(health.Live))`,
		`a.Core.Get("/readyz", http.HandlerFunc(health.Ready))`,
		`a.Core.Get("/api/hello", http.HandlerFunc(api.Hello))`,
		`a.Core.Get("/api/status", http.HandlerFunc(api.Status))`,
		`a.Core.Get("/api/v1/greet", http.HandlerFunc(api.Greet))`,
	})
}

func TestCanonicalTemplate_FileSetMatchesReferenceContract(t *testing.T) {
	files := GetTemplateFiles("canonical")
	want := []string{
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

	if !slices.Equal(files, want) {
		t.Fatalf("canonical file set drifted from reference contract:\n got: %#v\nwant: %#v", files, want)
	}
}

func TestCanonicalTemplate_APIHandlerMatchesReferenceSurface(t *testing.T) {
	content := getTemplateContent("internal/handler/api.go", "myapp", "example.com/myapp", "canonical")

	assertContainsAll(t, content, []string{
		"type helloResponse struct",
		"type greetResponse struct",
		"type statusResponse struct",
		"func (h APIHandler) Hello(",
		"func (h APIHandler) Greet(",
		"func (h APIHandler) Status(",
		`Extensions: "excluded_from_canonical_path"`,
		`contract.WriteError(`,
		`contract.WriteResponse(`,
	})
}

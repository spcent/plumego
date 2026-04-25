package scaffold

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var allTemplates = []string{"minimal", "api", "fullstack", "microservice", "canonical"}

func assertNoBareTODO(t *testing.T, label string, content string) {
	t.Helper()

	if strings.Contains(content, "// TODO") {
		t.Errorf("%s contains '// TODO':\n%s", label, content)
	}
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
				assertNoBareTODO(t, "template="+tmpl+" file="+file, content)
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
		assertNoBareTODO(t, "file="+tc.file, content)
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

func TestTemplateContent_UsesCanonicalRouteParams(t *testing.T) {
	content := getTemplateContent("internal/httpapp/handlers/user.go", "myapp", "example.com/myapp", "api")
	assertContainsAll(t, content, []string{`contract.RequestContextFromContext(r.Context()).Params["id"]`})
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
			file:     "internal/httpapp/handlers/health.go",
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

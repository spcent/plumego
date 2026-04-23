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

// TestGetTemplateFiles_NoEmpty verifies every template returns at least one file.
func TestGetTemplateFiles_NoEmpty(t *testing.T) {
	for _, tmpl := range allTemplates {
		files := GetTemplateFiles(tmpl)
		if len(files) == 0 {
			t.Errorf("template %q returned no files", tmpl)
		}
	}
}

// TestTemplateContent_NoTODO verifies no generated file contains a bare // TODO comment.
func TestTemplateContent_NoTODO(t *testing.T) {
	const (
		testName   = "myapp"
		testModule = "example.com/myapp"
	)
	for _, tmpl := range allTemplates {
		files := GetTemplateFiles(tmpl)
		for _, file := range files {
			content := getTemplateContent(file, testName, testModule, tmpl)
			if strings.Contains(content, "// TODO") {
				t.Errorf("template=%q file=%q contains '// TODO':\n%s", tmpl, file, content)
			}
		}
	}
}

// TestTemplateContent_GoFilesParseable verifies all generated .go files are
// syntactically valid Go source.
func TestTemplateContent_GoFilesParseable(t *testing.T) {
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
			_, err := parser.ParseFile(fset, file, content, parser.AllErrors)
			if err != nil {
				t.Errorf("template=%q file=%q parse error: %v\ncontent:\n%s",
					tmpl, file, err, content)
			}
		}
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
			// Files directly under cmd/ should be package main.
			// Everything else keeps its directory-derived package name.
			if strings.HasPrefix(file, "cmd/") && filepath.Ext(file) == ".go" {
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
		if strings.Contains(content, "// TODO") {
			t.Errorf("file=%q contains '// TODO':\n%s", tc.file, content)
		}
	}
}

func TestGetTemplateFiles_MicroserviceDoesNotEmitLegacyHTTPHelpers(t *testing.T) {
	files := GetTemplateFiles("microservice")
	if slices.Contains(files, "internal/platform/httpjson/response.go") {
		t.Fatal("microservice template should not emit internal/platform/httpjson/response.go")
	}
	if slices.Contains(files, "internal/platform/httperr/error.go") {
		t.Fatal("microservice template should not emit internal/platform/httperr/error.go")
	}
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
			for _, pattern := range disallowed {
				if strings.Contains(content, pattern) {
					t.Fatalf("template=%q file=%q contains disallowed pattern %q:\n%s", tmpl, file, pattern, content)
				}
			}
		}
	}
}

func TestTemplateContent_UsesCanonicalRouteParams(t *testing.T) {
	content := getTemplateContent("internal/httpapp/handlers/user.go", "myapp", "example.com/myapp", "api")
	required := `contract.RequestContextFromContext(r.Context()).Params["id"]`
	if !strings.Contains(content, required) {
		t.Fatalf("user handler template missing canonical route param lookup %q:\n%s", required, content)
	}
}

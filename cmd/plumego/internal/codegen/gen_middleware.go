package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateMiddleware(dir string, opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Type:  "middleware",
		Name:  opts.Name,
		Files: make(map[string][]string),
		Imports: []string{
			"net/http",
		},
	}

	outputPath := opts.OutputPath
	if outputPath == "" {
		middlewareName := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "internal", "middleware", middlewareName+".go")
	}

	packageName := opts.PackageName
	if packageName == "" {
		packageName = "middleware"
	}
	if err := validateGoIdentifier("package", packageName); err != nil {
		return nil, err
	}

	content := generateMiddlewareCode(opts.Name, packageName)
	outputPaths := []string{outputPath}
	if opts.WithTests {
		outputPaths = append(outputPaths, strings.TrimSuffix(outputPath, ".go")+"_test.go")
	}
	if err := validateOutputPaths(outputPaths, opts.Force); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	if opts.WithTests {
		testPath := outputPaths[1]
		testContent := generateMiddlewareTestCode(opts.Name, packageName)
		if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write test file: %w", err)
		}
		result.Files["created"] = append(result.Files["created"], testPath)
	}

	return result, nil
}

// generateMiddlewareCode generates canonical middleware: func(http.Handler) http.Handler.
func generateMiddlewareCode(name, pkg string) string {
	return fmt.Sprintf(`package %s

import (
	"net/http"
)

// %s returns an HTTP middleware that applies transport-layer behaviour for %s.
// Add request/response inspection, header injection, or early-exit logic here.
// Do not add business logic or service calls inside middleware.
func %s() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Handled-By", "%s")
			next.ServeHTTP(w, r)
		})
	}
}
`, pkg, name, strings.ToLower(name), name, name)
}

func generateMiddlewareTestCode(name, pkg string) string {
	return fmt.Sprintf(`package %s

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test%s(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := %s()(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %%d", rec.Code)
	}
	if rec.Header().Get("X-Handled-By") != "%s" {
		t.Fatalf("expected X-Handled-By header to be set")
	}
}
`, pkg, name, name, name)
}

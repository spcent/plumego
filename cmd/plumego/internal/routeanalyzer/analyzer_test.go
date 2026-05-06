package routeanalyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeRoutesRecognizesCanonicalHandlerFuncWrapper(t *testing.T) {
	tmp := t.TempDir()
	appDir := filepath.Join(tmp, "internal", "app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("mkdir app dir: %v", err)
	}

	source := `package app

import "net/http"

func registerRoutes(app interface{ Get(string, http.Handler) error }) error {
	api := APIHandler{}
	if err := app.Get("/api/hello", http.HandlerFunc(api.Hello)); err != nil {
		return err
	}
	if err := app.Get("/healthz", healthHandler); err != nil {
		return err
	}
	return nil
}

type APIHandler struct{}

func (APIHandler) Hello(http.ResponseWriter, *http.Request) {}

func healthHandler(http.ResponseWriter, *http.Request) {}
`
	if err := os.WriteFile(filepath.Join(appDir, "routes.go"), []byte(source), 0644); err != nil {
		t.Fatalf("write routes.go: %v", err)
	}

	result, err := AnalyzeRoutes(tmp, AnalyzeOptions{SortBy: "path"})
	if err != nil {
		t.Fatalf("analyze routes: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("total = %d, want 2; routes: %#v", result.Total, result.Routes)
	}

	var hello Route
	for _, route := range result.Routes {
		if route.Path == "/api/hello" {
			hello = route
			break
		}
	}
	if hello.Handler != "api.Hello" {
		t.Fatalf("handler = %q, want api.Hello; route: %#v", hello.Handler, hello)
	}
	if hello.File != filepath.Join("internal", "app", "routes.go") || hello.Line == 0 {
		t.Fatalf("unexpected location: %#v", hello)
	}
}

func TestAnalyzeRoutesReturnsParseErrors(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "broken.go"), []byte("package app\nfunc broken("), 0644); err != nil {
		t.Fatalf("write broken.go: %v", err)
	}

	_, err := AnalyzeRoutes(tmp, AnalyzeOptions{})
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "broken.go") {
		t.Fatalf("expected parse error to name file, got: %v", err)
	}
}

func TestAnalyzeRoutesRejectsUnsupportedSortField(t *testing.T) {
	tmp := t.TempDir()
	if _, err := AnalyzeRoutes(tmp, AnalyzeOptions{SortBy: "group"}); err == nil {
		t.Fatal("expected unsupported sort field error")
	}
}

func TestAnalyzeRoutesOutputOmitsUnsupportedGroupAndMiddlewareFields(t *testing.T) {
	result := AnalyzeResult{
		Routes: []Route{{Method: "GET", Path: "/healthz", Handler: "health"}},
		Total:  1,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	text := string(data)
	if strings.Contains(text, "group") || strings.Contains(text, "middleware") {
		t.Fatalf("route analyzer output should not expose unsupported fields: %s", text)
	}
}

func TestRouteAnalyzerDocsStateStableBoundary(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("read CLI README: %v", err)
	}
	readme := strings.Join(strings.Fields(string(data)), " ")
	for _, want := range []string{
		"static best-effort analyzer",
		"direct literal route registrations",
		"does not infer route groups",
		"middleware chains",
		"variable-built paths",
		"wrapper registration helpers",
	} {
		if !strings.Contains(readme, want) {
			t.Fatalf("route analyzer docs missing %q", want)
		}
	}
}

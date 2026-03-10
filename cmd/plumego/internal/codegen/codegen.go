package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateOptions represents code generation options
type GenerateOptions struct {
	Type           string
	Name           string
	OutputPath     string
	PackageName    string
	Methods        string
	WithTests      bool
	WithValidation bool
	Force          bool
}

// GenerateResult represents the result of code generation
type GenerateResult struct {
	Type    string              `json:"type" yaml:"type"`
	Name    string              `json:"name" yaml:"name"`
	Files   map[string][]string `json:"files" yaml:"files"`
	Imports []string            `json:"imports,omitempty" yaml:"imports,omitempty"`
}

// Generate generates code based on options
func Generate(dir string, opts GenerateOptions) (*GenerateResult, error) {
	switch opts.Type {
	case "component":
		return generateComponent(dir, opts)
	case "middleware":
		return generateMiddleware(dir, opts)
	case "handler":
		return generateHandler(dir, opts)
	case "model":
		return generateModel(dir, opts)
	default:
		return nil, fmt.Errorf("unknown generation type: %s", opts.Type)
	}
}

func generateComponent(dir string, opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Type:  "component",
		Name:  opts.Name,
		Files: make(map[string][]string),
		Imports: []string{
			"net/http",
			"github.com/spcent/plumego/core",
		},
	}

	outputPath := opts.OutputPath
	if outputPath == "" {
		componentDir := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "internal", "httpapp", "handlers", componentDir+".go")
	}

	packageName := opts.PackageName
	if packageName == "" {
		packageName = "handlers"
	}

	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("file %s already exists (use --force to overwrite)", outputPath)
	}

	content := generateComponentCode(opts.Name, packageName)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	if opts.WithTests {
		testPath := strings.TrimSuffix(outputPath, ".go") + "_test.go"
		testContent := generateComponentTestCode(opts.Name, packageName)
		if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write test file: %w", err)
		}
		result.Files["created"] = append(result.Files["created"], testPath)
	}

	return result, nil
}

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
		outputPath = filepath.Join(dir, "internal", "httpapp", "middleware", middlewareName+".go")
	}

	packageName := opts.PackageName
	if packageName == "" {
		packageName = "middleware"
	}

	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("file %s already exists (use --force to overwrite)", outputPath)
	}

	content := generateMiddlewareCode(opts.Name, packageName)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	if opts.WithTests {
		testPath := strings.TrimSuffix(outputPath, ".go") + "_test.go"
		testContent := generateMiddlewareTestCode(opts.Name, packageName)
		if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write test file: %w", err)
		}
		result.Files["created"] = append(result.Files["created"], testPath)
	}

	return result, nil
}

func generateHandler(dir string, opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Type:  "handler",
		Name:  opts.Name,
		Files: make(map[string][]string),
		Imports: []string{
			"encoding/json",
			"net/http",
		},
	}

	outputPath := opts.OutputPath
	if outputPath == "" {
		handlerName := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "internal", "httpapp", "handlers", handlerName+".go")
	}

	packageName := opts.PackageName
	if packageName == "" {
		packageName = "handlers"
	}

	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("file %s already exists (use --force to overwrite)", outputPath)
	}

	methods := strings.Split(opts.Methods, ",")
	content := generateHandlerCode(opts.Name, packageName, methods)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	if opts.WithTests {
		testPath := strings.TrimSuffix(outputPath, ".go") + "_test.go"
		testContent := generateHandlerTestCode(opts.Name, packageName, methods)
		if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write test file: %w", err)
		}
		result.Files["created"] = append(result.Files["created"], testPath)
	}

	return result, nil
}

func generateModel(dir string, opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Type:    "model",
		Name:    opts.Name,
		Files:   make(map[string][]string),
		Imports: []string{},
	}

	outputPath := opts.OutputPath
	if outputPath == "" {
		modelName := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "internal", "domain", modelName, modelName+".go")
	}

	packageName := opts.PackageName
	if packageName == "" {
		packageName = strings.ToLower(opts.Name)
	}

	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("file %s already exists (use --force to overwrite)", outputPath)
	}

	content := generateModelCode(opts.Name, packageName, opts.WithValidation)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	return result, nil
}

// --- Code templates ---

// generateComponentCode generates a canonical handler struct with health endpoint.
func generateComponentCode(name, pkg string) string {
	lower := strings.ToLower(name)
	return fmt.Sprintf(`package %s

import (
	"net/http"
)

// %sHandler handles HTTP requests for the %s domain.
type %sHandler struct {
	Service %sService
}

// %sService defines the operations required by %sHandler.
type %sService interface {
	// TODO: define service methods
}

func (h %sHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`+"`"+`{"status":"ok"}`+"`"+`))
}

func (h %sHandler) RegisterRoutes(mux interface {
	Get(string, http.HandlerFunc)
}) {
	mux.Get("/%s/health", h.Health)
}
`, pkg, name, lower, name, name, name, lower, name, name, name, lower)
}

func generateComponentTestCode(name, pkg string) string {
	return fmt.Sprintf(`package %s

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test%sHandler_Health(t *testing.T) {
	h := %sHandler{}

	req := httptest.NewRequest(http.MethodGet, "/%s/health", nil)
	rec := httptest.NewRecorder()
	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %%d", rec.Code)
	}
}
`, pkg, name, name, strings.ToLower(name))
}

// generateMiddlewareCode generates canonical middleware: func(http.Handler) http.Handler.
func generateMiddlewareCode(name, pkg string) string {
	return fmt.Sprintf(`package %s

import (
	"net/http"
)

// %s is an HTTP middleware that adds transport-layer behaviour for %s.
func %s() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: add transport-only logic here
			next.ServeHTTP(w, r)
		})
	}
}
`, pkg, name, strings.ToLower(name), name)
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
}
`, pkg, name, name)
}

// generateHandlerCode generates a canonical handler struct with methods.
func generateHandlerCode(name, pkg string, methods []string) string {
	serviceIface := fmt.Sprintf(`
// %sService defines the operations required by %sHandler.
type %sService interface {
	// TODO: define service methods
}

// %sHandler handles HTTP requests for the %s domain.
type %sHandler struct {
	Service %sService
}
`, name, name, name, name, strings.ToLower(name), name, name)

	handlers := ""
	for _, method := range methods {
		method = strings.TrimSpace(strings.ToUpper(method))
		handlers += generateHandlerMethodCode(name, method)
	}

	return fmt.Sprintf(`package %s

import (
	"encoding/json"
	"net/http"
)
%s%s`, pkg, serviceIface, handlers)
}

func generateHandlerMethodCode(name, method string) string {
	lower := strings.ToLower(name)
	switch method {
	case "GET":
		return fmt.Sprintf(`
// Get handles GET /%s
func (h %sHandler) Get(w http.ResponseWriter, r *http.Request) {
	// TODO: read params, call h.Service, write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "%s retrieved"})
}
`, lower, name, lower)
	case "POST":
		return fmt.Sprintf(`
type Create%sRequest struct {
	Name string `+"`json:\"name\"`"+`
}

type Create%sResponse struct {
	ID string `+"`json:\"id\"`"+`
}

// Create handles POST /%s
func (h %sHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req Create%sRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// TODO: call h.Service.Create(req.Name)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(Create%sResponse{ID: "TODO"})
}
`, name, name, lower, name, name, name)
	case "PUT":
		return fmt.Sprintf(`
// Update handles PUT /%s/:id
func (h %sHandler) Update(w http.ResponseWriter, r *http.Request) {
	// TODO: read id param, decode body, call h.Service.Update
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "%s updated"})
}
`, lower, name, lower)
	case "DELETE":
		return fmt.Sprintf(`
// Delete handles DELETE /%s/:id
func (h %sHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// TODO: read id param, call h.Service.Delete
	w.WriteHeader(http.StatusNoContent)
}
`, lower, name)
	default:
		return ""
	}
}

func generateHandlerTestCode(name, pkg string, methods []string) string {
	tests := ""
	for _, method := range methods {
		method = strings.TrimSpace(strings.ToUpper(method))
		tests += generateHandlerTestMethodCode(name, method)
	}

	return fmt.Sprintf(`package %s

import (
	"net/http"
	"net/http/httptest"
	"testing"
)
%s`, pkg, tests)
}

func generateHandlerTestMethodCode(name, method string) string {
	lower := strings.ToLower(name)
	switch method {
	case "GET":
		return fmt.Sprintf(`
func TestGet%s(t *testing.T) {
	h := %sHandler{}
	req := httptest.NewRequest(http.MethodGet, "/%s", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %%d", rec.Code)
	}
}
`, name, name, lower)
	case "POST":
		return fmt.Sprintf(`
func TestCreate%s(t *testing.T) {
	h := %sHandler{}
	req := httptest.NewRequest(http.MethodPost, "/%s", nil)
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusCreated {
		t.Fatalf("unexpected status %%d", rec.Code)
	}
}
`, name, name, lower)
	case "PUT":
		return fmt.Sprintf(`
func TestUpdate%s(t *testing.T) {
	h := %sHandler{}
	req := httptest.NewRequest(http.MethodPut, "/%s/1", nil)
	rec := httptest.NewRecorder()
	h.Update(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %%d", rec.Code)
	}
}
`, name, name, lower)
	case "DELETE":
		return fmt.Sprintf(`
func TestDelete%s(t *testing.T) {
	h := %sHandler{}
	req := httptest.NewRequest(http.MethodDelete, "/%s/1", nil)
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %%d", rec.Code)
	}
}
`, name, name, lower)
	default:
		return ""
	}
}

func generateModelCode(name, pkg string, withValidation bool) string {
	validation := ""
	if withValidation {
		validation = fmt.Sprintf(`
// Validate checks required fields.
func (m *%s) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}`, name)
	}

	imports := ""
	if withValidation {
		imports = `
import (
	"fmt"
	"strings"
)
`
	}

	return fmt.Sprintf(`package %s
%s
// %s represents a %s entity.
type %s struct {
	ID        int64  `+"`json:\"id\"`"+`
	CreatedAt string `+"`json:\"created_at\"`"+`
	UpdatedAt string `+"`json:\"updated_at\"`"+`
	Name      string `+"`json:\"name\"`"+`
}
%s
`, pkg, imports, name, strings.ToLower(name), name, validation)
}

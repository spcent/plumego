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
			"context",
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

// generateHandlerCode generates a handler struct, an entity type, and a service
// interface whose methods correspond to the requested HTTP methods.
func generateHandlerCode(name, pkg string, methods []string) string {
	lower := strings.ToLower(name)
	needsJSON := false

	// Build service interface methods from requested HTTP methods.
	svcMethods := ""
	for _, m := range methods {
		m = strings.TrimSpace(strings.ToUpper(m))
		switch m {
		case "GET":
			svcMethods += fmt.Sprintf("\tGet(ctx context.Context, id string) (*%s, error)\n", name)
		case "POST":
			needsJSON = true
			svcMethods += fmt.Sprintf("\tCreate(ctx context.Context, name string) (*%s, error)\n", name)
		case "PUT":
			needsJSON = true
			svcMethods += fmt.Sprintf("\tUpdate(ctx context.Context, id, name string) (*%s, error)\n", name)
		case "DELETE":
			svcMethods += "\tDelete(ctx context.Context, id string) error\n"
		}
	}

	jsonImport := ""
	if needsJSON {
		jsonImport = "\n\t\"encoding/json\""
	}

	header := fmt.Sprintf(`package %s

import (
	"context"
	"net/http"%s

	"github.com/spcent/plumego/contract"
)

// %s represents a %s entity.
type %s struct {
	ID   string `+"`json:\"id\"`"+`
	Name string `+"`json:\"name\"`"+`
}

// %sService defines the operations required by %sHandler.
type %sService interface {
%s}

// %sHandler handles HTTP requests for the %s domain.
type %sHandler struct {
	Service %sService
}
`, pkg, jsonImport, name, lower, name, name, name, name, svcMethods, name, lower, name, name)

	handlers := ""
	for _, m := range methods {
		m = strings.TrimSpace(strings.ToUpper(m))
		handlers += generateHandlerMethodCode(name, m)
	}

	return header + handlers
}

func generateHandlerMethodCode(name, method string) string {
	lower := strings.ToLower(name)
	switch method {
	case "GET":
		return fmt.Sprintf(`
// Get handles GET /%s/:id
func (h %sHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, err := h.Service.Get(r.Context(), id)
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Message("%s not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, item, nil)
}
`, lower, name, lower)
	case "POST":
		return fmt.Sprintf(`
// Create%sRequest carries the fields for creating a new %s.
type Create%sRequest struct {
	Name string `+"`json:\"name\"`"+`
}

// Create handles POST /%s
func (h %sHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req Create%sRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Category(contract.CategoryValidation).
			Build())
		return
	}
	item, err := h.Service.Create(r.Context(), req.Name)
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to create %s").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusCreated, item, nil)
}
`, name, lower, name, lower, name, name, lower)
	case "PUT":
		return fmt.Sprintf(`
// Update%sRequest carries the fields for updating a %s.
type Update%sRequest struct {
	Name string `+"`json:\"name\"`"+`
}

// Update handles PUT /%s/:id
func (h %sHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req Update%sRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Category(contract.CategoryValidation).
			Build())
		return
	}
	item, err := h.Service.Update(r.Context(), id, req.Name)
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to update %s").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, item, nil)
}
`, name, lower, name, lower, name, name, lower)
	case "DELETE":
		return fmt.Sprintf(`
// Delete handles DELETE /%s/:id
func (h %sHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.Service.Delete(r.Context(), id); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to delete %s").
			Build())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
`, lower, name, lower)
	default:
		return ""
	}
}

func generateHandlerTestCode(name, pkg string, methods []string) string {
	// Determine if any POST/PUT method is present (needs strings import).
	needsStrings := false
	for _, m := range methods {
		m = strings.TrimSpace(strings.ToUpper(m))
		if m == "POST" || m == "PUT" {
			needsStrings = true
		}
	}

	stringsImport := ""
	if needsStrings {
		stringsImport = "\n\t\"strings\""
	}

	// Build mock service implementing all requested methods.
	mockMethods := ""
	for _, m := range methods {
		m = strings.TrimSpace(strings.ToUpper(m))
		switch m {
		case "GET":
			mockMethods += fmt.Sprintf(`
func (m *mock%sService) Get(_ context.Context, id string) (*%s, error) {
	return &%s{ID: id, Name: "stub"}, nil
}
`, name, name, name)
		case "POST":
			mockMethods += fmt.Sprintf(`
func (m *mock%sService) Create(_ context.Context, name string) (*%s, error) {
	return &%s{ID: "1", Name: name}, nil
}
`, name, name, name)
		case "PUT":
			mockMethods += fmt.Sprintf(`
func (m *mock%sService) Update(_ context.Context, id, name string) (*%s, error) {
	return &%s{ID: id, Name: name}, nil
}
`, name, name, name)
		case "DELETE":
			mockMethods += fmt.Sprintf(`
func (m *mock%sService) Delete(_ context.Context, _ string) error {
	return nil
}
`, name)
		}
	}

	tests := ""
	for _, m := range methods {
		m = strings.TrimSpace(strings.ToUpper(m))
		tests += generateHandlerTestMethodCode(name, m)
	}

	return fmt.Sprintf(`package %s

import (
	"context"
	"net/http"
	"net/http/httptest"%s
	"testing"
)

type mock%sService struct{}
%s%s`, pkg, stringsImport, name, mockMethods, tests)
}

func generateHandlerTestMethodCode(name, method string) string {
	lower := strings.ToLower(name)
	switch method {
	case "GET":
		return fmt.Sprintf(`
func TestGet%s(t *testing.T) {
	h := %sHandler{Service: &mock%sService{}}
	req := httptest.NewRequest(http.MethodGet, "/%s/1", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %%d", rec.Code)
	}
}
`, name, name, name, lower)
	case "POST":
		return fmt.Sprintf(`
func TestCreate%s(t *testing.T) {
	h := %sHandler{Service: &mock%sService{}}
	body := strings.NewReader(`+"`"+`{"name":"alice"}`+"`"+`)
	req := httptest.NewRequest(http.MethodPost, "/%s", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %%d", rec.Code)
	}
}
`, name, name, name, lower)
	case "PUT":
		return fmt.Sprintf(`
func TestUpdate%s(t *testing.T) {
	h := %sHandler{Service: &mock%sService{}}
	body := strings.NewReader(`+"`"+`{"name":"bob"}`+"`"+`)
	req := httptest.NewRequest(http.MethodPut, "/%s/1", body)
	rec := httptest.NewRecorder()
	h.Update(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %%d", rec.Code)
	}
}
`, name, name, name, lower)
	case "DELETE":
		return fmt.Sprintf(`
func TestDelete%s(t *testing.T) {
	h := %sHandler{Service: &mock%sService{}}
	req := httptest.NewRequest(http.MethodDelete, "/%s/1", nil)
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %%d", rec.Code)
	}
}
`, name, name, name, lower)
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

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
			"context",
			"github.com/spcent/plumego/core",
			"github.com/spcent/plumego/router",
			"github.com/spcent/plumego/middleware",
			"github.com/spcent/plumego/health",
		},
	}

	// Determine output path
	outputPath := opts.OutputPath
	if outputPath == "" {
		componentDir := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "components", componentDir, componentDir+".go")
	}

	// Determine package name
	packageName := opts.PackageName
	if packageName == "" {
		packageName = strings.ToLower(opts.Name)
	}

	// Check if file exists
	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("file %s already exists (use --force to overwrite)", outputPath)
	}

	// Generate component code
	content := generateComponentCode(opts.Name, packageName)

	// Create directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	// Generate tests if requested
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

	// Determine output path
	outputPath := opts.OutputPath
	if outputPath == "" {
		middlewareName := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "middleware", middlewareName+".go")
	}

	// Determine package name
	packageName := opts.PackageName
	if packageName == "" {
		packageName = "middleware"
	}

	// Check if file exists
	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("file %s already exists (use --force to overwrite)", outputPath)
	}

	// Generate middleware code
	content := generateMiddlewareCode(opts.Name, packageName)

	// Create directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	// Generate tests if requested
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
			"net/http",
			"github.com/spcent/plumego/contract",
		},
	}

	// Determine output path
	outputPath := opts.OutputPath
	if outputPath == "" {
		handlerName := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "handlers", handlerName+".go")
	}

	// Determine package name
	packageName := opts.PackageName
	if packageName == "" {
		packageName = "handlers"
	}

	// Check if file exists
	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("file %s already exists (use --force to overwrite)", outputPath)
	}

	// Parse methods
	methods := strings.Split(opts.Methods, ",")

	// Generate handler code
	content := generateHandlerCode(opts.Name, packageName, methods)

	// Create directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	// Generate tests if requested
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

	if opts.WithValidation {
		result.Imports = append(result.Imports, "github.com/spcent/plumego/validator")
	}

	// Determine output path
	outputPath := opts.OutputPath
	if outputPath == "" {
		modelName := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "models", modelName+".go")
	}

	// Determine package name
	packageName := opts.PackageName
	if packageName == "" {
		packageName = "models"
	}

	// Check if file exists
	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("file %s already exists (use --force to overwrite)", outputPath)
	}

	// Generate model code
	content := generateModelCode(opts.Name, packageName, opts.WithValidation)

	// Create directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	return result, nil
}

func generateComponentCode(name, pkg string) string {
	return fmt.Sprintf(`package %s

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

// %sComponent implements the Component interface
type %sComponent struct {
	// Name identifies this component instance.
	Name string
}

// New%sComponent creates a new %s component
func New%sComponent() *%sComponent {
	return &%sComponent{Name: "%s"}
}

// RegisterRoutes registers routes for this component
func (c *%sComponent) RegisterRoutes(r *router.Router) {
	r.Get("/%s/health", http.HandlerFunc(c.handleHealth))
}

// RegisterMiddleware registers middleware for this component
func (c *%sComponent) RegisterMiddleware(m *middleware.Registry) {
	// No middleware registered by default.
}

// Start starts the component
func (c *%sComponent) Start(ctx context.Context) error {
	return nil
}

// Stop stops the component
func (c *%sComponent) Stop(ctx context.Context) error {
	return nil
}

// Health returns the component health status
func (c *%sComponent) Health() (string, health.HealthStatus) {
	if c.Name == "" {
		return "%s", health.Healthy()
	}
	return c.Name, health.Healthy()
}

func (c *%sComponent) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{\"status\":\"ok\"}"))
}
`, pkg, name, name, name, name, name, name, name, name, strings.ToLower(name), name, name, name, name, strings.ToLower(name), strings.ToLower(name), name)
}

func generateComponentTestCode(name, pkg string) string {
	return fmt.Sprintf(`package %s

import (
	"context"
	"testing"
)

func Test%sComponent_Start(t *testing.T) {
	comp := New%sComponent()
	ctx := context.Background()

	if err := comp.Start(ctx); err != nil {
		t.Errorf("Start failed: %%v", err)
	}
}

func Test%sComponent_Health(t *testing.T) {
	comp := New%sComponent()

	name, status := comp.Health()
	if name == "" {
		t.Error("Health name is empty")
	}
	if status.Status != "healthy" {
		t.Errorf("Expected healthy, got %%s", status.Status)
	}
}
`, pkg, name, name, name, name)
}

func generateMiddlewareCode(name, pkg string) string {
	return fmt.Sprintf(`package %s

import (
	"net/http"
)

// %s adds a header and passes the request through.
func %s(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-%s", "true")
		next.ServeHTTP(w, r)
	})
}
`, pkg, name, name, name)
}

func generateMiddlewareTestCode(name, pkg string) string {
	return fmt.Sprintf(`package %s

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test%s(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := %s(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %%d", w.Code)
	}
}
`, pkg, name, name)
}

func generateHandlerCode(name, pkg string, methods []string) string {
	handlers := ""
	for _, method := range methods {
		method = strings.TrimSpace(strings.ToUpper(method))
		handlers += generateHandlerMethodCode(name, method)
	}

	return fmt.Sprintf(`package %s

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

%s
`, pkg, handlers)
}

func generateHandlerMethodCode(name, method string) string {
	switch method {
	case "GET":
		return fmt.Sprintf(`
// Get%s handles GET requests for %s
func Get%s(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{
		"message": "%s retrieved",
	}, nil)
}
`, name, name, name, strings.ToLower(name))
	case "POST":
		return fmt.Sprintf(`
// Create%s handles POST requests for %s
func Create%s(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusCreated, map[string]string{
		"message": "%s created",
	}, nil)
}
`, name, name, name, strings.ToLower(name))
	case "PUT":
		return fmt.Sprintf(`
// Update%s handles PUT requests for %s
func Update%s(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{
		"message": "%s updated",
	}, nil)
}
`, name, name, name, strings.ToLower(name))
	case "DELETE":
		return fmt.Sprintf(`
// Delete%s handles DELETE requests for %s
func Delete%s(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusNoContent, nil, nil)
}
`, name, name, name)
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
	"net/http/httptest"
	"testing"
)

%s
`, pkg, tests)
}

func generateHandlerTestMethodCode(name, method string) string {
	switch method {
	case "GET":
		return fmt.Sprintf(`
func TestGet%s(t *testing.T) {
	req := httptest.NewRequest("GET", "/%s", nil)
	w := httptest.NewRecorder()

	Get%s(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %%d", w.Code)
	}
}
`, name, strings.ToLower(name), name)
	case "POST":
		return fmt.Sprintf(`
func TestCreate%s(t *testing.T) {
	req := httptest.NewRequest("POST", "/%s", nil)
	w := httptest.NewRecorder()

	Create%s(w, req)

	if w.Code != 201 {
		t.Errorf("Expected status 201, got %%d", w.Code)
	}
}
`, name, strings.ToLower(name), name)
	case "PUT":
		return fmt.Sprintf(`
func TestUpdate%s(t *testing.T) {
	req := httptest.NewRequest("PUT", "/%s/1", nil)
	w := httptest.NewRecorder()

	Update%s(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %%d", w.Code)
	}
}
`, name, strings.ToLower(name), name)
	case "DELETE":
		return fmt.Sprintf(`
func TestDelete%s(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/%s/1", nil)
	w := httptest.NewRecorder()

	Delete%s(w, req)

	if w.Code != 204 {
		t.Errorf("Expected status 204, got %%d", w.Code)
	}
}
`, name, strings.ToLower(name), name)
	default:
		return ""
	}
}

func generateModelCode(name, pkg string, withValidation bool) string {
	validation := ""
	if withValidation {
		validation = `
// Validate validates the model
func (m *` + name + `) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}`
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

// %s represents a %s model
type %s struct {
	ID        int64  `+"`json:\"id\"`"+`
	CreatedAt string `+"`json:\"created_at\"`"+`
	UpdatedAt string `+"`json:\"updated_at\"`"+`

	Name string `+"`json:\"name\"`"+`
}
%s
`, pkg, imports, name, strings.ToLower(name), name, validation)
}

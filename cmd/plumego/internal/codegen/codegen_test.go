package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func assertNoBareTODO(t *testing.T, label string, content string) {
	t.Helper()

	if strings.Contains(content, "// TODO") {
		t.Errorf("%s contains '// TODO':\n%s", label, content)
	}
}

// TestGenerateMiddlewareCode_NoTODO verifies the middleware template has no // TODO.
func TestGenerateMiddlewareCode_NoTODO(t *testing.T) {
	content := generateMiddlewareCode("Logging", "middleware")
	assertNoBareTODO(t, "middleware code", content)
}

// TestGenerateMiddlewareCode_Parseable verifies the middleware template is valid Go.
func TestGenerateMiddlewareCode_Parseable(t *testing.T) {
	content := generateMiddlewareCode("Logging", "middleware")
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "logging.go", content, parser.AllErrors); err != nil {
		t.Errorf("middleware code parse error: %v\ncontent:\n%s", err, content)
	}
}

// TestGenerateMiddlewareCode_CallsNext verifies generated middleware calls next.ServeHTTP.
func TestGenerateMiddlewareCode_CallsNext(t *testing.T) {
	content := generateMiddlewareCode("Tracing", "middleware")
	if !strings.Contains(content, "next.ServeHTTP") {
		t.Errorf("middleware code does not call next.ServeHTTP:\n%s", content)
	}
}

// TestGenerateHandlerCode_NoTODO verifies the handler template has no // TODO
// for any combination of HTTP methods.
func TestGenerateHandlerCode_NoTODO(t *testing.T) {
	cases := [][]string{
		{"GET"},
		{"POST"},
		{"PUT"},
		{"DELETE"},
		{"GET", "POST"},
		{"GET", "POST", "PUT", "DELETE"},
	}
	for _, methods := range cases {
		content := generateHandlerCode("Order", "handlers", methods)
		assertNoBareTODO(t, "handler code for "+strings.Join(methods, ","), content)
	}
}

// TestGenerateHandlerCode_Parseable verifies all handler method combinations produce
// valid Go source.
func TestGenerateHandlerCode_Parseable(t *testing.T) {
	fset := token.NewFileSet()
	cases := [][]string{
		{"GET"},
		{"POST"},
		{"PUT"},
		{"DELETE"},
		{"GET", "POST", "PUT", "DELETE"},
	}
	for _, methods := range cases {
		content := generateHandlerCode("Product", "handlers", methods)
		if _, err := parser.ParseFile(fset, "product.go", content, parser.AllErrors); err != nil {
			t.Errorf("handler code for %v parse error: %v\ncontent:\n%s", methods, err, content)
		}
	}
}

// TestGenerateHandlerCode_ServiceMethodsPresent verifies the service interface
// contains methods that match the requested HTTP operations.
func TestGenerateHandlerCode_ServiceMethodsPresent(t *testing.T) {
	tests := []struct {
		method string
		wantFn string
	}{
		{"GET", "Get(ctx context.Context"},
		{"POST", "Create(ctx context.Context"},
		{"PUT", "Update(ctx context.Context"},
		{"DELETE", "Delete(ctx context.Context"},
	}
	for _, tt := range tests {
		content := generateHandlerCode("Widget", "handlers", []string{tt.method})
		if !strings.Contains(content, tt.wantFn) {
			t.Errorf("handler %s: service interface missing %q:\n%s", tt.method, tt.wantFn, content)
		}
	}
}

// TestGenerateHandlerCode_HandlerCallsService verifies handler methods call the
// service rather than returning stub values.
func TestGenerateHandlerCode_HandlerCallsService(t *testing.T) {
	tests := []struct {
		method   string
		wantCall string
	}{
		{"GET", "h.Service.Get("},
		{"POST", "h.Service.Create("},
		{"PUT", "h.Service.Update("},
		{"DELETE", "h.Service.Delete("},
	}
	for _, tt := range tests {
		content := generateHandlerCode("Widget", "handlers", []string{tt.method})
		if !strings.Contains(content, tt.wantCall) {
			t.Errorf("handler %s: missing service call %q:\n%s", tt.method, tt.wantCall, content)
		}
	}
}

func TestGenerateHandlerCode_UsesCanonicalHTTPContract(t *testing.T) {
	content := generateHandlerCode("Widget", "handlers", []string{"GET", "POST", "PUT", "DELETE"})

	required := []string{
		`"github.com/spcent/plumego/contract"`,
		`contract.RequestContextFromContext(r.Context()).Params["id"]`,
		"contract.WriteResponse(",
		"contract.WriteError(",
		"json.NewDecoder(r.Body).Decode",
		"Code(contract.CodeInvalidJSON)",
		"Category(contract.CategoryValidation)",
	}
	for _, pattern := range required {
		if !strings.Contains(content, pattern) {
			t.Fatalf("handler code missing required pattern %q:\n%s", pattern, content)
		}
	}

	disallowed := []string{
		"PathValue(",
		"http.Error(",
		"json.NewEncoder(w).Encode",
		`w.Header().Set("Content-Type", "application/json")`,
	}
	for _, pattern := range disallowed {
		if strings.Contains(content, pattern) {
			t.Fatalf("handler code contains disallowed pattern %q:\n%s", pattern, content)
		}
	}
}

// TestGenerateHandlerTestCode_NoTODO verifies generated test files have no // TODO.
func TestGenerateHandlerTestCode_NoTODO(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	content := generateHandlerTestCode("Order", "handlers", methods)
	assertNoBareTODO(t, "handler test code", content)
}

// TestGenerateHandlerTestCode_Parseable verifies test files are valid Go.
func TestGenerateHandlerTestCode_Parseable(t *testing.T) {
	fset := token.NewFileSet()
	cases := [][]string{
		{"GET"},
		{"POST"},
		{"PUT"},
		{"DELETE"},
		{"GET", "POST", "PUT", "DELETE"},
	}
	for _, methods := range cases {
		content := generateHandlerTestCode("Order", "handlers", methods)
		if _, err := parser.ParseFile(fset, "order_test.go", content, parser.AllErrors); err != nil {
			t.Errorf("handler test code for %v parse error: %v\ncontent:\n%s",
				methods, err, content)
		}
	}
}

// TestGenerateHandlerTestCode_InjectsMock verifies the mock service is injected
// into each handler test.
func TestGenerateHandlerTestCode_InjectsMock(t *testing.T) {
	tests := []struct {
		method string
		wantFn string
	}{
		{"GET", "Service: &mockOrderService{}"},
		{"POST", "Service: &mockOrderService{}"},
		{"PUT", "Service: &mockOrderService{}"},
		{"DELETE", "Service: &mockOrderService{}"},
	}
	for _, tt := range tests {
		content := generateHandlerTestCode("Order", "handlers", []string{tt.method})
		if !strings.Contains(content, tt.wantFn) {
			t.Errorf("handler test %s: missing mock injection %q:\n%s",
				tt.method, tt.wantFn, content)
		}
	}
}

// TestGenerateModelCode_NoTODO verifies the model template has no // TODO.
func TestGenerateModelCode_NoTODO(t *testing.T) {
	for _, withVal := range []bool{false, true} {
		content := generateModelCode("Invoice", "invoice", withVal)
		label := "model code without validation"
		if withVal {
			label = "model code with validation"
		}
		assertNoBareTODO(t, label, content)
	}
}

// TestGenerateModelCode_Parseable verifies model code is valid Go.
func TestGenerateModelCode_Parseable(t *testing.T) {
	fset := token.NewFileSet()
	for _, withVal := range []bool{false, true} {
		content := generateModelCode("Invoice", "invoice", withVal)
		if _, err := parser.ParseFile(fset, "invoice.go", content, parser.AllErrors); err != nil {
			t.Errorf("model code (validation=%v) parse error: %v\ncontent:\n%s",
				withVal, err, content)
		}
	}
}

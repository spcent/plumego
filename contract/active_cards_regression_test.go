package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBindJSONAndValidateStructExplicitFlow(t *testing.T) {
	type payload struct {
		Name string `json:"name" validate:"required"`
	}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":""}`))
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	var dst payload
	if err := ctx.BindJSON(&dst); err != nil {
		t.Fatalf("expected bind to succeed, got %v", err)
	}

	if err := ValidateStruct(&dst); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestBindQuerySupportsPointersSlicesAndOmitempty(t *testing.T) {
	type query struct {
		Limit  *int     `query:"limit"`
		Filter *string  `query:"filter,omitempty"`
		IDs    []int    `query:"id"`
		Flags  []bool   `query:"flag"`
		Names  []string `query:"name,omitempty"`
		Miss   *string  `query:"missing,omitempty"`
	}

	req := httptest.NewRequest(http.MethodGet, "/?limit=10&filter=alpha&id=1&id=2&flag=true&flag=false&name=a&name=b", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	var got query
	if err := ctx.BindQuery(&got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Limit == nil || *got.Limit != 10 {
		t.Fatalf("expected Limit pointer to be set, got %+v", got.Limit)
	}
	if got.Filter == nil || *got.Filter != "alpha" {
		t.Fatalf("expected Filter pointer to be set, got %+v", got.Filter)
	}
	if got.Miss != nil {
		t.Fatalf("expected missing pointer field to stay nil, got %+v", got.Miss)
	}
	if want := []int{1, 2}; len(got.IDs) != len(want) || got.IDs[0] != want[0] || got.IDs[1] != want[1] {
		t.Fatalf("expected IDs %v, got %v", want, got.IDs)
	}
	if want := []bool{true, false}; len(got.Flags) != len(want) || got.Flags[0] != want[0] || got.Flags[1] != want[1] {
		t.Fatalf("expected Flags %v, got %v", want, got.Flags)
	}
	if want := []string{"a", "b"}; len(got.Names) != len(want) || got.Names[0] != want[0] || got.Names[1] != want[1] {
		t.Fatalf("expected Names %v, got %v", want, got.Names)
	}

	req = httptest.NewRequest(http.MethodGet, "/?id=bad", nil)
	ctx = NewCtx(httptest.NewRecorder(), req, nil)
	if err := ctx.BindQuery(&got); err == nil {
		t.Fatal("expected invalid slice element to fail")
	}
}

func TestValidateStructNestedUnknownRuleAndStringLength(t *testing.T) {
	type address struct {
		Street string `validate:"required"`
	}
	type user struct {
		Address address
	}

	err := ValidateStruct(&user{})
	fields := FieldErrorsFrom(err)
	if len(fields) == 0 || fields[0].Field != "Address.Street" {
		t.Fatalf("expected nested validation failure, got %v", fields)
	}

	type badRules struct {
		Name string `validate:"requried"`
	}
	err = ValidateStruct(&badRules{})
	if err == nil {
		t.Fatal("expected error for unknown rule, got nil")
	}
	// Unknown rules are programmer errors: not wrapped in ValidationErrors.
	if FieldErrorsFrom(err) != nil {
		t.Fatalf("expected unknown rule error to not be a ValidationErrors, got fields: %v", FieldErrorsFrom(err))
	}
	if !strings.Contains(err.Error(), "unknown validation rule") {
		t.Fatalf("expected error to mention unknown validation rule, got: %v", err)
	}

	type stringLengths struct {
		Code string `validate:"min=10"`
		Name string `validate:"max=10"`
	}
	err = ValidateStruct(&stringLengths{Code: "42", Name: "hello world"})
	fields = FieldErrorsFrom(err)
	if len(fields) != 2 {
		t.Fatalf("expected string min/max failures, got %v", fields)
	}
	for _, field := range fields {
		if field.Code != CodeOutOfRange {
			t.Fatalf("expected out of range validation codes, got %v", fields)
		}
	}
}

func TestValidateStructDepthLimitReturnsFieldError(t *testing.T) {
	type level11 struct {
		Value string `validate:"required"`
	}
	type level10 struct{ Child level11 }
	type level9 struct{ Child level10 }
	type level8 struct{ Child level9 }
	type level7 struct{ Child level8 }
	type level6 struct{ Child level7 }
	type level5 struct{ Child level6 }
	type level4 struct{ Child level5 }
	type level3 struct{ Child level4 }
	type level2 struct{ Child level3 }
	type level1 struct{ Child level2 }
	type root struct{ Child level1 }

	err := ValidateStruct(&root{})
	if err == nil {
		t.Fatal("expected depth-limit error, got nil")
	}

	fields := FieldErrorsFrom(err)
	var found bool
	for _, field := range fields {
		if field.Code == CodeOutOfRange {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected out of range field error, got %v", fields)
	}
}

func TestWriteJSONEncodeFailureWritesNothing(t *testing.T) {
	rec := httptest.NewRecorder()
	err := WriteJSON(rec, http.StatusCreated, map[string]any{"bad": func() {}})
	if err == nil {
		t.Fatal("expected encode failure")
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected no partial body write, got %q", rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "" {
		t.Fatalf("expected no content type on encode failure, got %q", rec.Header().Get("Content-Type"))
	}
}

func TestWriteErrorUsesTopLevelRequestIDAndTypedFields(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(WithRequestID(req.Context(), "req-123"))

	writeErr := NewErrorBuilder().
		Type(TypeValidation).
		Severity(SeverityWarning).
		Message("validation failed").
		Build()
	if err := WriteError(rec, req, writeErr); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	var payload struct {
		RequestID string `json:"request_id"`
		Error     struct {
			Type      ErrorType     `json:"type"`
			Severity  ErrorSeverity `json:"severity"`
			RequestID string        `json:"request_id,omitempty"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.RequestID != "req-123" {
		t.Fatalf("expected top-level request_id, got %v", payload.RequestID)
	}
	if payload.Error.Type != TypeValidation || payload.Error.Severity != SeverityWarning {
		t.Fatalf("expected type/severity in error body, got %+v", payload.Error)
	}
	if payload.Error.RequestID != "" {
		t.Fatalf("expected request_id to be promoted out of nested error body")
	}
}

func TestErrorBuilderStatusOnlyDerivesCategoryRegression(t *testing.T) {
	got := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Code(CodeBadRequest).
		Message("bad request").
		Build()
	if got.Category != CategoryClient {
		t.Fatalf("expected category %q, got %q", CategoryClient, got.Category)
	}
}

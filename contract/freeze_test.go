package contract

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFreezeWriteResponseNormalizesInvalidStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if err := WriteResponse(rec, req, 42, map[string]string{"ok": "true"}, nil); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var got Response
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Data == nil {
		t.Fatalf("expected response data to be preserved after status normalization")
	}
}

func TestFreezeWritersRejectNilResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	errs := []error{
		WriteJSON(nil, http.StatusOK, map[string]string{"ok": "true"}),
		WriteResponse(nil, req, http.StatusOK, nil, nil),
		WriteError(nil, req, NewErrorBuilder().Type(TypeInternal).Message("boom").Build()),
	}
	for i, err := range errs {
		if !errors.Is(err, ErrResponseWriterNil) {
			t.Fatalf("writer %d error = %v, want %v", i, err, ErrResponseWriterNil)
		}
	}
}

func TestFreezeErrorBuilderTypeOrdering(t *testing.T) {
	typeFirst := NewErrorBuilder().
		Type(TypeValidation).
		Status(http.StatusConflict).
		Code(CodeConflict).
		Category(CategoryClient).
		Message("conflict after validation type").
		Build()
	if typeFirst.Type != TypeValidation {
		t.Fatalf("type-first Type = %q, want %q", typeFirst.Type, TypeValidation)
	}
	if typeFirst.Status != http.StatusBadRequest || typeFirst.Code != CodeConflict || typeFirst.Category != CategoryValidation {
		t.Fatalf("typed errors should preserve custom code only, got status=%d code=%q category=%q", typeFirst.Status, typeFirst.Code, typeFirst.Category)
	}

	typeLast := NewErrorBuilder().
		Status(http.StatusConflict).
		Code(CodeConflict).
		Category(CategoryClient).
		Type(TypeValidation).
		Message("validation type after explicit fields").
		Build()
	if typeLast.Status != http.StatusBadRequest || typeLast.Code != CodeValidationError || typeLast.Category != CategoryValidation {
		t.Fatalf("Type should overwrite earlier status/code/category, got status=%d code=%q category=%q", typeLast.Status, typeLast.Code, typeLast.Category)
	}
}

func TestFreezeErrorBuilderDetailsAreCloned(t *testing.T) {
	details := map[string]any{"field": "name"}
	got := NewErrorBuilder().
		Type(TypeValidation).
		Message("validation failed").
		Details(details).
		Build()

	details["field"] = "email"
	details["later"] = "ignored"

	if got.Details["field"] != "name" {
		t.Fatalf("details field = %v, want %q", got.Details["field"], "name")
	}
	if _, ok := got.Details["later"]; ok {
		t.Fatalf("built details should not observe later mutation: %+v", got.Details)
	}
}

func TestFreezeDirectAPIErrorLiteralNormalization(t *testing.T) {
	got := normalizeAPIError(APIError{
		Status:   http.StatusConflict,
		Code:     "CUSTOM_REQUIRED",
		Message:  "custom required",
		Category: CategoryServer,
		Type:     TypeRequired,
		Severity: SeverityWarning,
	})

	if got.Status != http.StatusBadRequest {
		t.Fatalf("typed literal status = %d, want %d", got.Status, http.StatusBadRequest)
	}
	if got.Category != CategoryValidation {
		t.Fatalf("typed literal category = %q, want %q", got.Category, CategoryValidation)
	}
	if got.Code != "CUSTOM_REQUIRED" {
		t.Fatalf("typed literal code = %q, want custom code", got.Code)
	}
	if got.Type != TypeRequired {
		t.Fatalf("typed literal type = %q, want %q", got.Type, TypeRequired)
	}
	if got.Severity != SeverityWarning {
		t.Fatalf("typed literal severity = %q, want %q", got.Severity, SeverityWarning)
	}
}

func TestFreezeInvalidAPIErrorLiteralRepair(t *testing.T) {
	got := normalizeAPIError(APIError{
		Status:   http.StatusOK,
		Code:     "",
		Message:  "",
		Category: CategoryValidation,
		Type:     ErrorType("extension_unknown"),
		Severity: ErrorSeverity("urgent"),
		Details: map[string]any{
			"":      "ignored",
			"field": "name",
		},
	})

	if got.Status != http.StatusInternalServerError {
		t.Fatalf("invalid literal status = %d, want %d", got.Status, http.StatusInternalServerError)
	}
	if got.Code != CodeInternalError {
		t.Fatalf("invalid literal code = %q, want %q", got.Code, CodeInternalError)
	}
	if got.Category != CategoryServer {
		t.Fatalf("invalid literal category = %q, want %q", got.Category, CategoryServer)
	}
	if got.Type != "" {
		t.Fatalf("invalid literal type = %q, want empty", got.Type)
	}
	if got.Severity != "" {
		t.Fatalf("invalid literal severity = %q, want empty", got.Severity)
	}
	if got.Message != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("invalid literal message = %q, want status text", got.Message)
	}
	if _, ok := got.Details[""]; ok {
		t.Fatalf("empty detail key should be omitted: %+v", got.Details)
	}
	if got.Details["field"] != "name" {
		t.Fatalf("non-empty detail should remain: %+v", got.Details)
	}
}

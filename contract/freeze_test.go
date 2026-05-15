package contract

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFreezeWriteResponseRejectsInvalidStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if err := WriteResponse(rec, req, 42, map[string]string{"ok": "true"}, nil); !errors.Is(err, ErrInvalidResponseStatus) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidResponseStatus)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", rec.Body.String())
	}
	if got := rec.Header().Get(HeaderContentType); got != "" {
		t.Fatalf("content type = %q, want empty", got)
	}
}

func TestFreezeWriteResponseRejectsNonSuccessStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if err := WriteResponse(rec, req, http.StatusServiceUnavailable, map[string]string{"status": "degraded"}, nil); !errors.Is(err, ErrInvalidResponseStatus) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidResponseStatus)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", rec.Body.String())
	}
}

func TestFreezeWritersRejectNilResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	errs := []error{
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
		Code(CodeConflict).
		Message("conflict after validation type").
		Build()
	if typeFirst.errorType != TypeValidation {
		t.Fatalf("type-first Type = %q, want %q", typeFirst.errorType, TypeValidation)
	}
	if typeFirst.status != http.StatusBadRequest || typeFirst.code != CodeConflict || typeFirst.category != CategoryValidation {
		t.Fatalf("typed errors should preserve custom code only, got status=%d code=%q category=%q", typeFirst.status, typeFirst.code, typeFirst.category)
	}

	typeLast := NewErrorBuilder().
		Code(CodeConflict).
		Type(TypeValidation).
		Message("validation type after explicit fields").
		Build()
	if typeLast.status != http.StatusBadRequest || typeLast.code != CodeValidationError || typeLast.category != CategoryValidation {
		t.Fatalf("Type should overwrite earlier status/code/category, got status=%d code=%q category=%q", typeLast.status, typeLast.code, typeLast.category)
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

	if got.details["field"] != "name" {
		t.Fatalf("details field = %v, want %q", got.details["field"], "name")
	}
	if _, ok := got.details["later"]; ok {
		t.Fatalf("built details should not observe later mutation: %+v", got.details)
	}
}

func TestFreezeDirectAPIErrorLiteralNormalization(t *testing.T) {
	got := normalizeAPIError(APIError{
		status:    http.StatusConflict,
		code:      "CUSTOM_REQUIRED",
		message:   "custom required",
		category:  CategoryServer,
		errorType: TypeRequired,
	})

	if got.status != http.StatusBadRequest {
		t.Fatalf("typed literal status = %d, want %d", got.status, http.StatusBadRequest)
	}
	if got.category != CategoryValidation {
		t.Fatalf("typed literal category = %q, want %q", got.category, CategoryValidation)
	}
	if got.code != "CUSTOM_REQUIRED" {
		t.Fatalf("typed literal code = %q, want custom code", got.code)
	}
	if got.errorType != TypeRequired {
		t.Fatalf("typed literal type = %q, want %q", got.errorType, TypeRequired)
	}
}

func TestFreezeInvalidAPIErrorLiteralRepair(t *testing.T) {
	got := normalizeAPIError(APIError{
		status:    http.StatusOK,
		code:      "",
		message:   "",
		category:  CategoryValidation,
		errorType: ErrorType("extension_unknown"),
		details: map[string]any{
			"":      "ignored",
			"field": "name",
		},
	})

	if got.status != http.StatusInternalServerError {
		t.Fatalf("invalid literal status = %d, want %d", got.status, http.StatusInternalServerError)
	}
	if got.code != CodeInternalError {
		t.Fatalf("invalid literal code = %q, want %q", got.code, CodeInternalError)
	}
	if got.category != CategoryServer {
		t.Fatalf("invalid literal category = %q, want %q", got.category, CategoryServer)
	}
	if got.errorType != "" {
		t.Fatalf("invalid literal type = %q, want empty", got.errorType)
	}
	if got.message != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("invalid literal message = %q, want status text", got.message)
	}
	if _, ok := got.details[""]; ok {
		t.Fatalf("empty detail key should be omitted: %+v", got.details)
	}
	if got.details["field"] != "name" {
		t.Fatalf("non-empty detail should remain: %+v", got.details)
	}
}

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
	if typeFirst.Status != http.StatusConflict || typeFirst.Code != CodeConflict || typeFirst.Category != CategoryClient {
		t.Fatalf("explicit fields after Type should win, got status=%d code=%q category=%q", typeFirst.Status, typeFirst.Code, typeFirst.Category)
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

func TestFreezeErrorBuilderTypeOnlyPreservesExplicitFields(t *testing.T) {
	got := NewErrorBuilder().
		Status(http.StatusConflict).
		Code(CodeConflict).
		Category(CategoryClient).
		TypeOnly(TypeValidation).
		Message("typed without default overwrite").
		Build()

	if got.Type != TypeValidation {
		t.Fatalf("Type = %q, want %q", got.Type, TypeValidation)
	}
	if got.Status != http.StatusConflict || got.Code != CodeConflict || got.Category != CategoryClient {
		t.Fatalf("TypeOnly should preserve explicit fields, got status=%d code=%q category=%q", got.Status, got.Code, got.Category)
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

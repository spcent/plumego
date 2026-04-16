package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestWriteErrorWithBuilder(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	err := WriteError(w, r, NewErrorBuilder().
		Status(http.StatusBadRequest).
		Code("invalid_request").
		Message("bad request").
		Detail("field", "name").
		Build())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.Error.Code != "invalid_request" {
		t.Fatalf("expected code invalid_request, got %q", response.Error.Code)
	}
	if response.Error.Category != CategoryClient {
		t.Fatalf("expected category %q, got %q", CategoryClient, response.Error.Category)
	}
}

func TestCategoryForStatus(t *testing.T) {
	tests := []struct {
		name             string
		status           int
		expectedCategory ErrorCategory
	}{
		{"bad request", http.StatusBadRequest, CategoryClient},
		{"unauthorized", http.StatusUnauthorized, CategoryAuth},
		{"forbidden", http.StatusForbidden, CategoryAuth},
		{"not found", http.StatusNotFound, CategoryClient},
		{"too many requests", http.StatusTooManyRequests, CategoryRateLimit},
		{"request timeout", http.StatusRequestTimeout, CategoryTimeout},
		{"internal server error", http.StatusInternalServerError, CategoryServer},
		{"unprocessable entity fallback", http.StatusUnprocessableEntity, CategoryClient},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategoryForStatus(tt.status); got != tt.expectedCategory {
				t.Fatalf("status %d: expected category %q, got %q", tt.status, tt.expectedCategory, got)
			}
		})
	}
}

func TestParam(t *testing.T) {
	ctx := &Ctx{Params: map[string]string{"id": "123", "name": "test"}}

	val, ok := ctx.Param("id")
	if !ok || val != "123" {
		t.Fatalf("expected param id=123, got %q ok=%v", val, ok)
	}
	if val, ok = ctx.Param("missing"); ok || val != "" {
		t.Fatalf("expected missing param to be absent, got %q ok=%v", val, ok)
	}

	ctx.Params = nil
	if val, ok = ctx.Param("id"); ok || val != "" {
		t.Fatalf("expected nil params to be absent, got %q ok=%v", val, ok)
	}
}

func TestMustParam(t *testing.T) {
	ctx := &Ctx{Params: map[string]string{"id": "123", "empty": "", "spaces": "  "}}

	val, err := ctx.MustParam("id")
	if err != nil || val != "123" {
		t.Fatalf("expected id=123, got value=%q err=%v", val, err)
	}
	if _, err := ctx.MustParam("missing"); err == nil {
		t.Fatal("expected error for missing param")
	}
	if _, err := ctx.MustParam("empty"); err == nil {
		t.Fatal("expected error for empty param")
	}
	if val, err := ctx.MustParam("spaces"); err != nil || val != "  " {
		t.Fatalf("expected raw whitespace value to round-trip, got value=%q err=%v", val, err)
	}
}

func TestBindJSONBodyTooLarge(t *testing.T) {
	largeBody := strings.Repeat("a", 1024*1025)
	ctx := NewCtxWithConfig(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(largeBody)),
		nil,
		RequestConfig{MaxBodySize: 1024 * 1024},
	)

	var dst struct{}
	err := ctx.BindJSON(&dst)

	if err == nil {
		t.Fatal("expected error for body too large")
	}

	var bindErr *bindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected bindError, got %T", err)
	}
	if bindErr.Status != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, bindErr.Status)
	}
}

func TestRequestContextFromContext(t *testing.T) {
	result := RequestContextFromContext(t.Context())
	if result.Params != nil {
		t.Fatal("expected empty RequestContext")
	}

	rc := RequestContext{
		Params:       map[string]string{"id": "123"},
		RoutePattern: "/users/:id",
		RouteName:    "user_show",
	}
	ctx := WithRequestContext(context.Background(), rc)
	result = RequestContextFromContext(ctx)
	if result.Params == nil || result.Params["id"] != "123" {
		t.Fatal("expected RequestContext with params")
	}
	if result.RoutePattern != "/users/:id" || result.RouteName != "user_show" {
		t.Fatal("expected route fields from RequestContext")
	}
}

func TestWithRequestContextNilContext(t *testing.T) {
	rc := RequestContext{
		Params:       map[string]string{"id": "123"},
		RoutePattern: "/users/:id",
		RouteName:    "user_show",
	}

	ctx := WithRequestContext(nil, rc)
	result := RequestContextFromContext(ctx)
	if result.Params == nil || result.Params["id"] != "123" {
		t.Fatal("expected RequestContext to survive nil parent context")
	}
	if result.RoutePattern != "/users/:id" || result.RouteName != "user_show" {
		t.Fatal("expected route fields from nil-parent RequestContext")
	}
}

func TestCtxConfigDefaults(t *testing.T) {
	ctx := &Ctx{
		W:      httptest.NewRecorder(),
		R:      httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"ok":true}`)),
		Params: map[string]string{},
		config: nil,
	}

	if _, err := ctx.bodyBytes(); err != nil {
		t.Fatalf("unexpected error with nil config: %v", err)
	}

	ctx = &Ctx{
		W:      httptest.NewRecorder(),
		R:      httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"ok":true}`)),
		Params: map[string]string{},
		config: &RequestConfig{},
	}
	if _, err := ctx.bodyBytes(); err != nil {
		t.Fatalf("unexpected error with zero config: %v", err)
	}
}

func TestBindJSONAliasRemoval(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"demo","age":30}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	var payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := ctx.BindJSON(&payload); err != nil {
		t.Fatalf("expected successful bind, got %v", err)
	}
	if payload.Name != "demo" || payload.Age != 30 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestBindJSONError(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("")), nil)

	var payload struct{ Name string }
	err := ctx.BindJSON(&payload)
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	var bindErr *bindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected bindError, got %T", err)
	}
}

func TestBindQueryUint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?count=42", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct {
		Count uint `query:"count"`
	}

	var f filter
	if err := ctx.BindQuery(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Count != 42 {
		t.Fatalf("expected count=42, got %d", f.Count)
	}
}

func TestBindQueryAliasRemoval(t *testing.T) {
	type Q struct {
		Name string `query:"name"`
	}
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?name=alice", nil), nil)

	var q Q
	if err := ctx.BindQuery(&q); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Name != "alice" {
		t.Fatalf("expected alice, got %q", q.Name)
	}
}

func TestBindQueryThenValidateStruct(t *testing.T) {
	type Q struct {
		Name string `query:"name" validate:"required"`
	}

	t.Run("valid query", func(t *testing.T) {
		ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?name=bob", nil), nil)

		var q Q
		if err := ctx.BindQuery(&q); err != nil {
			t.Fatalf("unexpected bind error: %v", err)
		}
		if err := ValidateStruct(&q); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})

	t.Run("missing required field fails validation", func(t *testing.T) {
		ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)

		var q Q
		if err := ctx.BindQuery(&q); err != nil {
			t.Fatalf("unexpected bind error: %v", err)
		}
		if err := ValidateStruct(&q); err == nil {
			t.Fatal("expected validation error for missing required field")
		}
	})
}

func TestBindQueryInvalidUint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?count=bad", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct {
		Count uint `query:"count"`
	}

	var f filter
	err := ctx.BindQuery(&f)
	if err == nil {
		t.Fatal("expected invalid uint error")
	}
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Fatalf("expected strconv.NumError to remain reachable, got %v", err)
	}
}

package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	log "github.com/spcent/plumego/log"
)

func TestNewCtxPopulatesFields(t *testing.T) {
	deadline := time.Now().Add(time.Minute)
	baseCtx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/users/123?foo=bar", nil).WithContext(baseCtx)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	ctx := NewCtx(httptest.NewRecorder(), req, map[string]string{"id": "123"})

	if ctx.Params["id"] != "123" {
		t.Fatalf("expected param to be propagated")
	}

	if ctx.Query.Get("foo") != "bar" {
		t.Fatalf("expected query to be parsed")
	}

	if ctx.ClientIP != "1.2.3.4" {
		t.Fatalf("expected client ip from header, got %s", ctx.ClientIP)
	}

	if ctx.Deadline.IsZero() || !ctx.Deadline.Equal(deadline) {
		t.Fatalf("expected deadline to be copied")
	}
}

func TestCtxResponseHelpers(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := NewCtx(recorder, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	if err := ctx.JSON(http.StatusAccepted, map[string]string{"msg": "ok"}); err != nil {
		t.Fatalf("json response failed: %v", err)
	}

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status %d got %d", http.StatusAccepted, recorder.Code)
	}

	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected json content type, got %s", ct)
	}

	var payload map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload["msg"] != "ok" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestBindJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"demo"}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	var payload struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&payload); err != nil {
		t.Fatalf("expected successful bind, got %v", err)
	}

	if payload.Name != "demo" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestBindJSONMaxBodySize(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"too-long"}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)
	ctx.Config = &RequestConfig{
		MaxBodySize:     10,
		EnableBodyCache: true,
	}

	var payload map[string]any
	err := ctx.BindJSON(&payload)
	if err == nil {
		t.Fatalf("expected error for body too large")
	}

	var bindErr *BindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected BindError, got %T", err)
	}

	if bindErr.Status != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, bindErr.Status)
	}
}

func TestBindJSONBodyCacheToggle(t *testing.T) {
	payloadBody := `{"name":"demo"}`

	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(payloadBody)), nil)
	ctx.Config = &RequestConfig{
		EnableBodyCache: true,
	}

	var payload map[string]any
	if err := ctx.BindJSON(&payload); err != nil {
		t.Fatalf("expected bind to succeed, got %v", err)
	}

	cachedBody, err := io.ReadAll(ctx.R.Body)
	if err != nil {
		t.Fatalf("expected to read cached body, got %v", err)
	}
	if string(cachedBody) != payloadBody {
		t.Fatalf("expected cached body to match payload")
	}

	ctxNoCache := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(payloadBody)), nil)
	ctxNoCache.Config = &RequestConfig{
		EnableBodyCache: false,
	}

	if err := ctxNoCache.BindJSON(&payload); err != nil {
		t.Fatalf("expected bind to succeed, got %v", err)
	}

	uncachedBody, err := io.ReadAll(ctxNoCache.R.Body)
	if err != nil {
		t.Fatalf("expected to read body, got %v", err)
	}
	if len(uncachedBody) != 0 {
		t.Fatalf("expected request body to be consumed when cache is disabled")
	}
}

func TestBindJSONErrors(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{name: "empty", body: "", wantMsg: "request body is empty"},
		{name: "invalid", body: "{", wantMsg: "invalid JSON payload"},
		{name: "extra", body: `{"name":"demo"} {}`, wantMsg: "unexpected extra JSON data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body)), nil)
			var payload map[string]any
			err := ctx.BindJSON(&payload)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}

			var bindErr *BindError
			if !errors.As(err, &bindErr) {
				t.Fatalf("expected BindError, got %T", err)
			}

			if bindErr.Message != tt.wantMsg {
				t.Fatalf("unexpected message: %s", bindErr.Message)
			}
		})
	}
}

func TestBindAndValidateJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"demo@example.com","password":"superpass","username":"demo"}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	type payload struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		Username string `json:"username" validate:"required,min=3"`
	}

	var p payload
	if err := ctx.BindAndValidateJSON(&p); err != nil {
		t.Fatalf("expected successful bind+validate, got %v", err)
	}

	if p.Email == "" || p.Password == "" || p.Username == "" {
		t.Fatalf("expected fields to be populated: %+v", p)
	}
}

func TestBindAndValidateJSONErrors(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"invalid","password":"short","username":"de"}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	type payload struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		Username string `json:"username" validate:"required,min=3"`
	}

	var p payload
	err := ctx.BindAndValidateJSON(&p)
	if err == nil {
		t.Fatalf("expected validation error")
	}

	var bindErr *BindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected BindError, got %T", err)
	}

	if !strings.Contains(bindErr.Message, "Email") {
		t.Fatalf("unexpected bind error message: %s", bindErr.Message)
	}
}

func TestNewCtxAppliesRequestTimeout(t *testing.T) {
	originalConfig := DefaultRequestConfig
	DefaultRequestConfig = &RequestConfig{
		RequestTimeout: 50 * time.Millisecond,
	}
	t.Cleanup(func() {
		DefaultRequestConfig = originalConfig
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	if ctx.Deadline.IsZero() {
		t.Fatalf("expected deadline to be set from request timeout")
	}

	until := time.Until(ctx.Deadline)
	if until <= 0 || until > 200*time.Millisecond {
		t.Fatalf("expected deadline within timeout window, got %v", until)
	}
}

func TestParamsAndRequestContextHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(context.WithValue(context.Background(), ParamsContextKey{}, map[string]string{"id": "42"}))
	if val, ok := Param(req, "id"); !ok || val != "42" {
		t.Fatalf("unexpected Param lookup: %v %v", val, ok)
	}

	rc := RequestContextFrom(req.Context())
	if rc.Params["id"] != "42" {
		t.Fatalf("request context did not surface params")
	}

	if params := ParamsFromContext(context.Background()); params != nil {
		t.Fatalf("expected nil params for empty context")
	}
}

func TestBindErrorHelpers(t *testing.T) {
	err := &BindError{Message: "oops", Err: errors.New("root")}
	if err.Error() != "oops" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
	if !errors.Is(err, err.Err) {
		t.Fatalf("expected unwrap to work")
	}

	var nilErr *BindError
	if nilErr.Error() != "" {
		t.Fatalf("nil receiver should return empty string")
	}
	if nilErr.Unwrap() != nil {
		t.Fatalf("nil receiver unwrap should be nil")
	}
}

func TestCtxHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	ctx := NewCtx(httptest.NewRecorder(), req, map[string]string{"name": "demo"})

	if val, ok := ctx.Param("name"); !ok || val != "demo" {
		t.Fatalf("Param lookup failed")
	}
	if _, err := ctx.MustParam("missing"); err == nil {
		t.Fatalf("MustParam should error on missing key")
	}

	// response helpers
	if err := ctx.Text(http.StatusCreated, "hello"); err != nil {
		t.Fatalf("text write failed: %v", err)
	}
	if err := ctx.Bytes(http.StatusOK, []byte("bin")); err != nil {
		t.Fatalf("bytes write failed: %v", err)
	}
}

func TestAdaptCtxHandler(t *testing.T) {
	logger := log.NewGLogger()
	called := false
	h := AdaptCtxHandler(func(c *Ctx) {
		called = true
	}, logger)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("handler was not called")
	}

	if err := ValidateCtxHandler(nil); err == nil {
		t.Fatalf("expected validation error for nil handler")
	}
}

func TestStreamChunkSizeValidation(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)

	if err := ctx.StreamBinary(strings.NewReader("data"), 0); !errors.Is(err, ErrInvalidChunkSize) {
		t.Fatalf("expected invalid chunk size error, got %v", err)
	}

	if err := ctx.StreamJSONChunked([]any{map[string]string{"a": "b"}}, 0); !errors.Is(err, ErrInvalidChunkSize) {
		t.Fatalf("expected invalid chunk size error, got %v", err)
	}

	if err := ctx.StreamTextChunked([]string{"line"}, -1); !errors.Is(err, ErrInvalidChunkSize) {
		t.Fatalf("expected invalid chunk size error, got %v", err)
	}
}

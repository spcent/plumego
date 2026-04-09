package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestWriteErrorWithBuilder(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	err := WriteError(w, r, NewErrorBuilder().
		Status(http.StatusBadRequest).
		Code("invalid_request").
		Message("Bad request").
		Detail("field", "name").
		Build())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Error.Code != "invalid_request" {
		t.Errorf("expected code 'invalid_request', got '%s'", response.Error.Code)
	}
	if response.Error.Message != "Bad request" {
		t.Errorf("expected message 'Bad request', got '%s'", response.Error.Message)
	}
	if response.Error.Category != CategoryClient {
		t.Errorf("expected category %q, got %q", CategoryClient, response.Error.Category)
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
		{"unprocessable entity (fallback)", http.StatusUnprocessableEntity, CategoryClient},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategoryForStatus(tt.status); got != tt.expectedCategory {
				t.Errorf("status %d: expected category %q, got %q", tt.status, tt.expectedCategory, got)
			}
		})
	}
}

func TestUnsafeRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewCtx(w, r, nil)

	err := ctx.UnsafeRedirect(http.StatusFound, "https://external.example.com/redirect")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "https://external.example.com/redirect" {
		t.Errorf("expected external location, got '%s'", location)
	}
}

func TestRedirect(t *testing.T) {
	tests := []struct {
		name        string
		location    string
		host        string
		expectError bool
	}{
		{"relative path", "/dashboard", "example.com", false},
		{"relative path with query", "/search?q=test", "example.com", false},
		{"same-origin absolute", "http://example.com/path", "example.com", false},
		{"same-origin https", "https://example.com/path", "example.com", false},
		{"external URL", "https://evil.com/phish", "example.com", true},
		{"protocol-relative URL", "//evil.com/path", "example.com", true},
		{"external with path", "https://attacker.com/fake-login", "example.com", true},
		{"javascript scheme", "javascript:alert(1)", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.Host = tt.host
			ctx := NewCtx(w, r, nil)

			err := ctx.Redirect(http.StatusFound, tt.location)

			if tt.expectError && err == nil {
				t.Errorf("expected error for %q, got nil", tt.location)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.location, err)
			}
			if tt.expectError && err != nil && !errors.Is(err, ErrUnsafeRedirect) {
				t.Errorf("expected ErrUnsafeRedirect, got %v", err)
			}
		})
	}
}

func TestFile(t *testing.T) {
	// Create a temporary file
	content := []byte("test file content")
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewCtx(w, r, nil)

	err = ctx.File(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if !bytes.Equal(w.Body.Bytes(), content) {
		t.Errorf("expected file content '%s', got '%s'", content, w.Body.Bytes())
	}
}

func TestParam(t *testing.T) {
	ctx := &Ctx{
		Params: map[string]string{"id": "123", "name": "test"},
	}

	// Test existing param
	val, ok := ctx.Param("id")
	if !ok {
		t.Error("expected param 'id' to exist")
	}
	if val != "123" {
		t.Errorf("expected '123', got '%s'", val)
	}

	// Test non-existing param
	val, ok = ctx.Param("missing")
	if ok {
		t.Error("expected param 'missing' to not exist")
	}
	if val != "" {
		t.Errorf("expected empty string, got '%s'", val)
	}

	// Test with nil params
	ctx.Params = nil
	val, ok = ctx.Param("id")
	if ok {
		t.Error("expected param to not exist with nil params")
	}
}

func TestMustParam(t *testing.T) {
	ctx := &Ctx{
		Params: map[string]string{"id": "123", "empty": "", "spaces": "  "},
	}

	// Test existing param
	val, err := ctx.MustParam("id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "123" {
		t.Errorf("expected '123', got '%s'", val)
	}

	// Test missing param
	_, err = ctx.MustParam("missing")
	if err == nil {
		t.Error("expected error for missing param")
	}
	if !strings.Contains(err.Error(), "missing param") {
		t.Errorf("expected 'missing param' error, got '%v'", err)
	}

	// Test empty param
	_, err = ctx.MustParam("empty")
	if err == nil {
		t.Error("expected error for empty param")
	}
	if val, err := ctx.MustParam("spaces"); err != nil || val != "  " {
		t.Fatalf("expected raw whitespace value to round-trip, got value=%q err=%v", val, err)
	}
}

func TestBindJSONBodyTooLarge(t *testing.T) {
	// Create a body larger than max size
	largeBody := strings.Repeat("a", 1024*1025) // 1MB + 1 byte

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))

	// Use custom config with small max body size
	config := &RequestConfig{MaxBodySize: 1024 * 1024} // 1MB
	ctx := &Ctx{
		W:      w,
		R:      r,
		Params: map[string]string{},
		Config: config,
	}

	var dst struct{}
	err := ctx.BindJSON(&dst)

	if err == nil {
		t.Error("expected error for body too large")
	}

	var bindErr *bindError
	if errors.As(err, &bindErr) {
		if bindErr.Status != http.StatusRequestEntityTooLarge {
			t.Errorf("expected status %d, got %d", http.StatusRequestEntityTooLarge, bindErr.Status)
		}
	}
}

func TestRelease(t *testing.T) {
	// Test with cancel function
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Create context with timeout to get a cancel function
	timeoutCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	r = r.WithContext(timeoutCtx)

	ctx := NewCtx(w, r, nil)
	ctx.cancel = cancel // Set the cancel function

	// release should not panic
	ctx.release()

	// Second release should not panic
	ctx.release()
}

func TestClientIPFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remoteIP string
		expected string
	}{
		{
			name:     "X-Forwarded-For last hop wins",
			headers:  map[string]string{"X-Forwarded-For": "198.51.100.9, 203.0.113.1"},
			remoteIP: "192.168.1.1:8080",
			expected: "203.0.113.1",
		},
		{
			name:     "X-Real-IP",
			headers:  map[string]string{"X-Real-IP": "203.0.113.2"},
			remoteIP: "192.168.1.1:8080",
			expected: "203.0.113.2",
		},
		{
			name:     "RemoteAddr",
			headers:  map[string]string{},
			remoteIP: "192.168.1.1:8080",
			expected: "192.168.1.1",
		},
		{
			name:     "RemoteAddr without port",
			headers:  map[string]string{},
			remoteIP: "192.168.1.1",
			expected: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/test", nil)
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}
			r.RemoteAddr = tt.remoteIP

			// Use the internal function through a Ctx
			w := httptest.NewRecorder()
			ctx := NewCtx(w, r, nil)

			if ctx.ClientIP != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, ctx.ClientIP)
			}
		})
	}
}

func TestValidateCtxHandler(t *testing.T) {
	// Test with nil handler
	err := ValidateCtxHandler(nil)
	if err == nil {
		t.Error("expected error for nil handler")
	}
	if !errors.Is(err, ErrHandlerNil) {
		t.Errorf("expected ErrHandlerNil, got '%v'", err)
	}

	// Test with valid handler
	handler := func(ctx *Ctx) {}
	err = ValidateCtxHandler(handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequestContextFromContext(t *testing.T) {
	// Test with nil context
	result := RequestContextFromContext(t.Context())
	if result.Params != nil {
		t.Error("expected empty RequestContext")
	}

	// Test with RequestContext
	rc := RequestContext{
		Params:       map[string]string{"id": "123"},
		RoutePattern: "/users/:id",
		RouteName:    "user_show",
	}
	ctx := context.WithValue(context.Background(), requestContextKey{}, rc)
	result = RequestContextFromContext(ctx)
	if result.Params == nil || result.Params["id"] != "123" {
		t.Error("expected RequestContext with params")
	}
	if result.RoutePattern != "/users/:id" || result.RouteName != "user_show" {
		t.Error("expected route fields from RequestContext")
	}
}

func TestCtxResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r = r.WithContext(WithRequestID(r.Context(), "req-xyz"))
	ctx := NewCtx(w, r, nil)

	err := ctx.Response(http.StatusOK, map[string]string{"status": "ok"}, map[string]any{"page": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.RequestID != "req-xyz" {
		t.Fatalf("expected request id, got %q", response.RequestID)
	}
	if response.Meta == nil || response.Meta["page"] != float64(1) {
		t.Fatalf("expected meta to be included")
	}
}

func TestParamFromRequest(t *testing.T) {
	// Test with RequestContext
	rc := RequestContext{Params: map[string]string{"id": "123"}}
	ctx := context.WithValue(context.Background(), requestContextKey{}, rc)
	r := httptest.NewRequest("GET", "/test", nil)
	r = r.WithContext(ctx)

	params := RequestContextFromContext(r.Context()).Params
	val, ok := params["id"]
	if !ok {
		t.Error("expected param to exist")
	}
	if val != "123" {
		t.Errorf("expected '123', got '%s'", val)
	}

	// Test with missing param
	_, ok = params["missing"]
	if ok {
		t.Error("expected param to not exist")
	}

	// Test with no context
	r = httptest.NewRequest("GET", "/test", nil)
	params = RequestContextFromContext(r.Context()).Params
	if _, ok = params["id"]; ok {
		t.Error("expected param to not exist")
	}
}

func TestBindErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	bindErr := &bindError{
		Status:  http.StatusBadRequest,
		Message: "test error",
		Err:     innerErr,
	}

	if bindErr.Unwrap() != innerErr {
		t.Error("Unwrap should return inner error")
	}

	// Test nil case
	var nilErr *bindError
	if nilErr.Unwrap() != nil {
		t.Error("Unwrap on nil should return nil")
	}

	// Test with no inner error
	bindErr2 := &bindError{
		Status:  http.StatusBadRequest,
		Message: "test error",
	}
	if bindErr2.Unwrap() != nil {
		t.Error("Unwrap should return nil when no inner error")
	}
}

func TestCtxConfigDefaults(t *testing.T) {
	// Test with nil config
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := &Ctx{
		W:      w,
		R:      r,
		Params: map[string]string{},
		Config: nil,
	}

	// Should not panic with nil config
	_, err := ctx.bodyBytes()
	if err != nil {
		t.Fatalf("unexpected error with nil config: %v", err)
	}

	// Test with zero config
	ctx.Config = &RequestConfig{}
	_, err = ctx.bodyBytes()
	if err != nil {
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

func TestBindQuery(t *testing.T) {
	type filter struct {
		Name   string   `query:"name"`
		Page   int      `query:"page"`
		Limit  int64    `query:"limit"`
		Score  float64  `query:"score"`
		Active bool     `query:"active"`
		Tags   []string `query:"tags"`
		Ignore string   // no query tag
	}

	req := httptest.NewRequest(http.MethodGet, "/?name=alice&page=2&limit=50&score=9.5&active=true&tags=go&tags=http", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	var f filter
	if err := ctx.BindQuery(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f.Name != "alice" {
		t.Errorf("expected name=alice, got %s", f.Name)
	}
	if f.Page != 2 {
		t.Errorf("expected page=2, got %d", f.Page)
	}
	if f.Limit != 50 {
		t.Errorf("expected limit=50, got %d", f.Limit)
	}
	if f.Score != 9.5 {
		t.Errorf("expected score=9.5, got %f", f.Score)
	}
	if !f.Active {
		t.Error("expected active=true")
	}
	if len(f.Tags) != 2 || f.Tags[0] != "go" || f.Tags[1] != "http" {
		t.Errorf("expected tags=[go http], got %v", f.Tags)
	}
	if f.Ignore != "" {
		t.Error("field without query tag should not be set")
	}
}

func TestBindQueryMissingParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct {
		Name string `query:"name"`
		Page int    `query:"page"`
	}

	var f filter
	if err := ctx.BindQuery(&f); err != nil {
		t.Fatalf("missing params should not error, got %v", err)
	}
	if f.Name != "" || f.Page != 0 {
		t.Error("missing params should leave zero values")
	}
}

func TestBindQueryInvalidType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=notanumber", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct {
		Page int `query:"page"`
	}

	var f filter
	err := ctx.BindQuery(&f)
	if err == nil {
		t.Fatal("expected error for invalid int")
	}
	var bindErr *bindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected bindError, got %T", err)
	}
	if bindErr.Status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", bindErr.Status)
	}
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("expected errors.Is(err, ErrInvalidParam) to be true, got %v", err)
	}
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Fatalf("expected strconv.NumError to remain reachable, got %v", err)
	}
}

func TestBindQueryNonPointer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct{ Name string }
	var f filter
	err := ctx.BindQuery(f) // not a pointer
	if err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestBindQueryDashTag(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?skip=yes", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct {
		Skip string `query:"-"`
	}

	var f filter
	if err := ctx.BindQuery(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Skip != "" {
		t.Error("field with tag '-' should be skipped")
	}
}

func TestSetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := NewCtx(w, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	ctx.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    "abc123",
		MaxAge:   3600,
		Path:     "/app",
		Domain:   "example.com",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "session" || c.Value != "abc123" {
		t.Errorf("unexpected cookie: %v", c)
	}
	if c.MaxAge != 3600 {
		t.Errorf("expected MaxAge=3600, got %d", c.MaxAge)
	}
	if c.Path != "/app" {
		t.Errorf("expected Path=/app, got %s", c.Path)
	}
	if c.Domain != "example.com" {
		t.Errorf("expected Domain=example.com, got %s", c.Domain)
	}
	if !c.Secure || !c.HttpOnly {
		t.Error("expected Secure and HttpOnly to be true")
	}

	// Verify SameSite=Strict appears in the Set-Cookie header
	setCookie := w.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "SameSite=Strict") {
		t.Errorf("expected SameSite=Strict in Set-Cookie header, got: %s", setCookie)
	}
}

func TestSetCookieDefaultPath(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := NewCtx(w, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	ctx.SetCookie(&http.Cookie{Name: "token", Value: "xyz"})

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Path != "/" {
		t.Errorf("expected default path '/', got %s", cookies[0].Path)
	}
}

func TestCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	val, err := ctx.Cookie("session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "abc123" {
		t.Errorf("expected abc123, got %s", val)
	}
}

func TestCookieNotFound(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)

	_, err := ctx.Cookie("missing")
	if !errors.Is(err, http.ErrNoCookie) {
		t.Fatalf("expected ErrNoCookie, got %v", err)
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
		t.Errorf("expected count=42, got %d", f.Count)
	}
}

func TestCtxBodyReadOnce(t *testing.T) {
	body := "test body"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	ctx := NewCtx(w, r, nil)

	// Read body multiple times
	data1, err1 := ctx.bodyBytes()
	data2, err2 := ctx.bodyBytes()

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}

	if !bytes.Equal(data1, data2) {
		t.Error("body should be cached and return same data")
	}
}

func TestBindQueryAliasRemoval(t *testing.T) {
	type Q struct {
		Name string `query:"name"`
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?name=alice", nil)
	ctx := NewCtx(w, r, nil)

	var q Q
	if err := ctx.BindQuery(&q); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Name != "alice" {
		t.Errorf("expected 'alice', got %q", q.Name)
	}
}

func TestBindQueryThenValidateStruct(t *testing.T) {
	type Q struct {
		Name string `query:"name" validate:"required"`
	}

	t.Run("valid query", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/?name=bob", nil)
		ctx := NewCtx(w, r, nil)

		var q Q
		if err := ctx.BindQuery(&q); err != nil {
			t.Fatalf("unexpected bind error: %v", err)
		}
		if err := ValidateStruct(&q); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
		if q.Name != "bob" {
			t.Errorf("expected 'bob', got %q", q.Name)
		}
	})

	t.Run("missing required field fails validation", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		ctx := NewCtx(w, r, nil)

		var q Q
		if err := ctx.BindQuery(&q); err != nil {
			t.Fatalf("unexpected bind error: %v", err)
		}
		if err := ValidateStruct(&q); err == nil {
			t.Fatal("expected validation error for missing required field")
		}
	})
}

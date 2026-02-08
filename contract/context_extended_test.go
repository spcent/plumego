package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestErrorJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewCtx(w, r, nil)

	err := ctx.ErrorJSON(http.StatusBadRequest, "invalid_request", "Bad request", map[string]any{"field": "name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Code != "invalid_request" {
		t.Errorf("expected code 'invalid_request', got '%s'", response.Code)
	}
	if response.Message != "Bad request" {
		t.Errorf("expected message 'Bad request', got '%s'", response.Message)
	}
}

func TestErrorJSONCategoryFromStatus(t *testing.T) {
	tests := []struct {
		name             string
		status           int
		expectedCategory ErrorCategory
	}{
		{"bad request", http.StatusBadRequest, CategoryClient},
		{"unauthorized", http.StatusUnauthorized, CategoryAuthentication},
		{"forbidden", http.StatusForbidden, CategoryAuthentication},
		{"not found", http.StatusNotFound, CategoryClient},
		{"too many requests", http.StatusTooManyRequests, CategoryRateLimit},
		{"request timeout", http.StatusRequestTimeout, CategoryTimeout},
		{"internal server error", http.StatusInternalServerError, CategoryServer},
		{"unprocessable entity (fallback)", http.StatusUnprocessableEntity, CategoryClient},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			ctx := NewCtx(w, r, nil)

			err := ctx.ErrorJSON(tt.status, "TEST", "test message", nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var response APIError
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if response.Category != tt.expectedCategory {
				t.Errorf("status %d: expected category %q, got %q", tt.status, tt.expectedCategory, response.Category)
			}
		})
	}
}

func TestRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewCtx(w, r, nil)

	err := ctx.Redirect(http.StatusFound, "/redirect")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/redirect" {
		t.Errorf("expected location '/redirect', got '%s'", location)
	}
}

func TestSafeRedirect(t *testing.T) {
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

			err := ctx.SafeRedirect(http.StatusFound, tt.location)

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
		Params: map[string]string{"id": "123", "empty": "  "},
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
}

func TestBindJSONBodyTooLarge(t *testing.T) {
	// Create a body larger than max size
	largeBody := strings.Repeat("a", 1024*1025) // 1MB + 1 byte

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))

	// Use custom config with small max body size
	config := &RequestConfig{MaxBodySize: 1024 * 1024} // 1MB
	ctx := &Ctx{
		W:         w,
		R:         r,
		Params:    map[string]string{},
		Config:    config,
		startedAt: time.Now(),
	}

	var dst struct{}
	err := ctx.BindJSON(&dst)

	if err == nil {
		t.Error("expected error for body too large")
	}

	var bindErr *BindError
	if errors.As(err, &bindErr) {
		if bindErr.Status != http.StatusRequestEntityTooLarge {
			t.Errorf("expected status %d, got %d", http.StatusRequestEntityTooLarge, bindErr.Status)
		}
	}
}

func TestGetRequestDuration(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewCtx(w, r, nil)

	// Sleep a bit to ensure duration is measurable
	time.Sleep(10 * time.Millisecond)

	duration := ctx.GetRequestDuration()
	if duration < 10*time.Millisecond {
		t.Errorf("expected duration >= 10ms, got %v", duration)
	}
}

func TestIsCompressed(t *testing.T) {
	originalConfig := DefaultRequestConfig
	config := defaultRequestConfig()
	config.EnableCompression = true
	DefaultRequestConfig = config
	t.Cleanup(func() {
		DefaultRequestConfig = originalConfig
	})

	tests := []struct {
		name            string
		contentEncoding string
		expected        bool
	}{
		{"gzip", "gzip", true},
		{"deflate", "deflate", true},
		{"none", "", false},
		{"unknown", "br", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/test", nil)
			if tt.contentEncoding != "" {
				r.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			ctx := NewCtx(w, r, nil)

			if ctx.IsCompressed() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, ctx.IsCompressed())
			}
		})
	}
}

func TestGetBodySize(t *testing.T) {
	body := "test body content"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	ctx := NewCtx(w, r, nil)

	// Trigger body read
	_, err := ctx.bodyBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	size := ctx.GetBodySize()
	if size != int64(len(body)) {
		t.Errorf("expected size %d, got %d", len(body), size)
	}
}

func TestClose(t *testing.T) {
	// Test with cancel function
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Create context with timeout to get a cancel function
	timeoutCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	r = r.WithContext(timeoutCtx)

	ctx := NewCtx(w, r, nil)
	ctx.cancel = cancel // Set the cancel function

	// Close should not panic
	ctx.Close()

	// Second close should not panic
	ctx.Close()
}

func TestClientIPFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remoteIP string
		expected string
	}{
		{
			name:     "X-Forwarded-For",
			headers:  map[string]string{"X-Forwarded-For": "203.0.113.1"},
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
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("expected 'cannot be nil' error, got '%v'", err)
	}

	// Test with valid handler
	handler := func(ctx *Ctx) {}
	err = ValidateCtxHandler(handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParamsFromContext(t *testing.T) {
	// Test with nil context
	result := ParamsFromContext(t.Context())
	if result != nil {
		t.Error("expected nil for nil context")
	}

	// Test with RequestContext
	rc := RequestContext{Params: map[string]string{"id": "123"}}
	ctx := context.WithValue(context.Background(), RequestContextKey{}, rc)
	result = ParamsFromContext(ctx)
	if result == nil || result["id"] != "123" {
		t.Error("expected params from RequestContext")
	}

	// Test with ParamsContextKey (legacy)
	params := map[string]string{"name": "test"}
	ctx = context.WithValue(context.Background(), ParamsContextKey{}, params)
	result = ParamsFromContext(ctx)
	if result == nil || result["name"] != "test" {
		t.Error("expected params from ParamsContextKey")
	}

	// Test with empty context
	ctx = context.Background()
	result = ParamsFromContext(ctx)
	if result != nil {
		t.Error("expected nil for empty context")
	}
}

func TestRequestContextFrom(t *testing.T) {
	// Test with nil context
	result := RequestContextFrom(t.Context())
	if result.Params != nil {
		t.Error("expected empty RequestContext")
	}

	// Test with RequestContext
	rc := RequestContext{
		Params:       map[string]string{"id": "123"},
		RoutePattern: "/users/:id",
		RouteName:    "user_show",
	}
	ctx := context.WithValue(context.Background(), RequestContextKey{}, rc)
	result = RequestContextFrom(ctx)
	if result.Params == nil || result.Params["id"] != "123" {
		t.Error("expected RequestContext with params")
	}
	if result.RoutePattern != "/users/:id" || result.RouteName != "user_show" {
		t.Error("expected route fields from RequestContext")
	}

	// Test fallback to ParamsContextKey
	params := map[string]string{"name": "test"}
	ctx = context.WithValue(context.Background(), ParamsContextKey{}, params)
	result = RequestContextFrom(ctx)
	if result.Params == nil || result.Params["name"] != "test" {
		t.Error("expected fallback to ParamsContextKey")
	}
}

func TestCtxResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r = r.WithContext(context.WithValue(r.Context(), TraceIDKey{}, "trace-xyz"))
	ctx := NewCtx(w, r, nil)

	err := ctx.Response(http.StatusOK, map[string]string{"status": "ok"}, map[string]any{"page": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.TraceID != "trace-xyz" {
		t.Fatalf("expected trace id, got %q", response.TraceID)
	}
	if response.Meta == nil || response.Meta["page"] != float64(1) {
		t.Fatalf("expected meta to be included")
	}
}

func TestParamFromRequest(t *testing.T) {
	// Test with RequestContext
	rc := RequestContext{Params: map[string]string{"id": "123"}}
	ctx := context.WithValue(context.Background(), RequestContextKey{}, rc)
	r := httptest.NewRequest("GET", "/test", nil)
	r = r.WithContext(ctx)

	val, ok := Param(r, "id")
	if !ok {
		t.Error("expected param to exist")
	}
	if val != "123" {
		t.Errorf("expected '123', got '%s'", val)
	}

	// Test with missing param
	val, ok = Param(r, "missing")
	if ok {
		t.Error("expected param to not exist")
	}

	// Test with no context
	r = httptest.NewRequest("GET", "/test", nil)
	val, ok = Param(r, "id")
	if ok {
		t.Error("expected param to not exist")
	}
}

func TestBindErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	bindErr := &BindError{
		Status:  http.StatusBadRequest,
		Message: "test error",
		Err:     innerErr,
	}

	if bindErr.Unwrap() != innerErr {
		t.Error("Unwrap should return inner error")
	}

	// Test nil case
	var nilErr *BindError
	if nilErr.Unwrap() != nil {
		t.Error("Unwrap on nil should return nil")
	}

	// Test with no inner error
	bindErr2 := &BindError{
		Status:  http.StatusBadRequest,
		Message: "test error",
	}
	if bindErr2.Unwrap() != nil {
		t.Error("Unwrap should return nil when no inner error")
	}
}

func TestCtxCompression(t *testing.T) {
	originalConfig := DefaultRequestConfig
	config := defaultRequestConfig()
	config.EnableCompression = true
	DefaultRequestConfig = config
	t.Cleanup(func() {
		DefaultRequestConfig = originalConfig
	})

	// Test with gzip
	r := httptest.NewRequest("POST", "/test", nil)
	r.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()
	ctx := NewCtx(w, r, nil)

	if !ctx.IsCompressed() {
		t.Error("expected compression to be enabled")
	}

	// Test with deflate
	r = httptest.NewRequest("POST", "/test", nil)
	r.Header.Set("Content-Encoding", "deflate")
	w = httptest.NewRecorder()
	ctx = NewCtx(w, r, nil)

	if !ctx.IsCompressed() {
		t.Error("expected compression to be enabled")
	}
}

func TestCtxConfigDefaults(t *testing.T) {
	// Test with nil config
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := &Ctx{
		W:         w,
		R:         r,
		Params:    map[string]string{},
		Config:    nil,
		startedAt: time.Now(),
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

func TestShouldBindJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"demo","age":30}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	var payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		t.Fatalf("expected successful bind, got %v", err)
	}
	if payload.Name != "demo" || payload.Age != 30 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestShouldBindJSONError(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("")), nil)

	var payload struct{ Name string }
	err := ctx.ShouldBindJSON(&payload)
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	var bindErr *BindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected BindError, got %T", err)
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
	var bindErr *BindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected BindError, got %T", err)
	}
	if bindErr.Status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", bindErr.Status)
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

	ctx.SetCookie("session", "abc123", 3600, "/app", "example.com", true, true)

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
}

func TestSetCookieDefaultPath(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := NewCtx(w, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	ctx.SetCookie("token", "xyz", 0, "", "", false, false)

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

func TestFormFile(t *testing.T) {
	// Build a multipart form with a file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("upload", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("file content here"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	fh, err := ctx.FormFile("upload")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fh.Filename != "test.txt" {
		t.Errorf("expected filename test.txt, got %s", fh.Filename)
	}
	if fh.Size != int64(len("file content here")) {
		t.Errorf("expected size %d, got %d", len("file content here"), fh.Size)
	}
}

func TestFormFileMissing(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	_, err := ctx.FormFile("missing")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSaveUploadedFile(t *testing.T) {
	// Build a multipart form with a file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("upload", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	content := "saved file content"
	part.Write([]byte(content))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	fh, err := ctx.FormFile("upload")
	if err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir() + "/subdir/saved.txt"
	if err := ctx.SaveUploadedFile(fh, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestSetGet(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)

	// Get on empty store returns false
	val, ok := ctx.Get("missing")
	if ok || val != nil {
		t.Fatalf("expected no value for missing key, got %v", val)
	}

	// Set and Get
	ctx.Set("user", "alice")
	val, ok = ctx.Get("user")
	if !ok || val != "alice" {
		t.Fatalf("expected alice, got %v (ok=%v)", val, ok)
	}

	// Overwrite
	ctx.Set("user", "bob")
	val, ok = ctx.Get("user")
	if !ok || val != "bob" {
		t.Fatalf("expected bob after overwrite, got %v", val)
	}

	// Different types
	ctx.Set("count", 42)
	ctx.Set("active", true)
	ctx.Set("data", map[string]int{"x": 1})

	if v, _ := ctx.Get("count"); v != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
	if v, _ := ctx.Get("active"); v != true {
		t.Fatalf("expected true, got %v", v)
	}
	if v, _ := ctx.Get("data"); v == nil {
		t.Fatal("expected non-nil map")
	}

	// Nil value is a valid stored value
	ctx.Set("nilval", nil)
	val, ok = ctx.Get("nilval")
	if !ok {
		t.Fatal("expected key to exist even with nil value")
	}
	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}
}

func TestMustGetPanics(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)

	// MustGet on existing key
	ctx.Set("key", "val")
	if v := ctx.MustGet("key"); v != "val" {
		t.Fatalf("expected val, got %v", v)
	}

	// MustGet on missing key should panic
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for missing key")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "missing key") {
			t.Fatalf("unexpected panic value: %v", r)
		}
	}()
	ctx.MustGet("nonexistent")
}

func TestSetGetConcurrent(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)

	done := make(chan struct{})
	const n = 100

	// Concurrent writers
	for i := 0; i < n; i++ {
		go func(i int) {
			ctx.Set("key", i)
			done <- struct{}{}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < n; i++ {
		go func() {
			ctx.Get("key")
			done <- struct{}{}
		}()
	}

	for i := 0; i < 2*n; i++ {
		<-done
	}

	// Final value should exist
	_, ok := ctx.Get("key")
	if !ok {
		t.Fatal("expected key to be present after concurrent writes")
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

	if ctx.GetBodySize() != int64(len(body)) {
		t.Errorf("expected body size %d, got %d", len(body), ctx.GetBodySize())
	}
}

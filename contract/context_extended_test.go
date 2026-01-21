package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	rc := RequestContext{Params: map[string]string{"id": "123"}}
	ctx := context.WithValue(context.Background(), RequestContextKey{}, rc)
	result = RequestContextFrom(ctx)
	if result.Params == nil || result.Params["id"] != "123" {
		t.Error("expected RequestContext with params")
	}

	// Test fallback to ParamsContextKey
	params := map[string]string{"name": "test"}
	ctx = context.WithValue(context.Background(), ParamsContextKey{}, params)
	result = RequestContextFrom(ctx)
	if result.Params == nil || result.Params["name"] != "test" {
		t.Error("expected fallback to ParamsContextKey")
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

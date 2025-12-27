package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func TestCORSMiddleware(t *testing.T) {
	opts := CORSOptions{
		AllowedOrigins:   []string{"http://allowed.com"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-My-Custom-Header"},
		MaxAge:           10 * time.Minute,
	}

	handler := CORSWithOptions(opts, dummyHandler)

	t.Run("No Origin header (non-CORS)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		if resp.Header.Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("expected no CORS header, got %s", resp.Header.Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("Allowed Origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://allowed.com")
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://allowed.com" {
			t.Errorf("expected Allow-Origin http://allowed.com, got %s", got)
		}
		if resp.Header.Get("Access-Control-Allow-Credentials") != "true" {
			t.Errorf("expected Allow-Credentials true")
		}
		if !strings.Contains(resp.Header.Get("Access-Control-Expose-Headers"), "X-My-Custom-Header") {
			t.Errorf("expected Expose-Headers to contain X-My-Custom-Header, got %s", resp.Header.Get("Access-Control-Expose-Headers"))
		}
	})

	t.Run("Disallowed Origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://evil.com")
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		if resp.Header.Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("expected no Allow-Origin header, got %s", resp.Header.Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("Preflight Allowed", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://allowed.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204, got %d", resp.StatusCode)
		}
		if !strings.Contains(resp.Header.Get("Access-Control-Allow-Methods"), "POST") {
			t.Errorf("expected Allow-Methods to contain POST, got %s", resp.Header.Get("Access-Control-Allow-Methods"))
		}
		if !strings.Contains(resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type") {
			t.Errorf("expected Allow-Headers to contain Content-Type, got %s", resp.Header.Get("Access-Control-Allow-Headers"))
		}
		if resp.Header.Get("Access-Control-Max-Age") != "600" {
			t.Errorf("expected Max-Age=600, got %s", resp.Header.Get("Access-Control-Max-Age"))
		}
	})

	t.Run("AllowCredentials with * origins", func(t *testing.T) {
		opts2 := CORSOptions{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		}
		handler2 := CORSWithOptions(opts2, dummyHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://foo.com")
		w := httptest.NewRecorder()

		handler2(w, req)
		resp := w.Result()

		if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://foo.com" {
			t.Errorf("expected echoed origin http://foo.com, got %s", got)
		}
	})
}

func TestCORSMiddleware_ResponseBody(t *testing.T) {
	opts := CORSOptions{AllowedOrigins: []string{"*"}}
	handler := CORSWithOptions(opts, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://anything.com")
	w := httptest.NewRecorder()

	handler(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("expected body 'ok', got %s", string(body))
	}
}

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/security/abuse"
	"github.com/spcent/plumego/utils/httpx"
)

type abuseClock struct {
	now time.Time
}

func (c *abuseClock) Now() time.Time {
	return c.now
}

func (c *abuseClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
}

func TestAbuseGuardBlocks(t *testing.T) {
	clock := &abuseClock{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	limiter := abuse.NewLimiter(abuse.Config{
		Rate:            1,
		Capacity:        1,
		Now:             clock.Now,
		CleanupInterval: time.Hour,
		MaxIdle:         time.Hour,
	})
	defer limiter.Stop()

	include := true
	mw := AbuseGuard(AbuseGuardConfig{
		Limiter:        limiter,
		KeyFunc:        func(*http.Request) string { return "client" },
		IncludeHeaders: &include,
	})
	wrapped := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("expected remaining 0, got %q", got)
	}

	w = httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Retry-After"); got == "" {
		t.Fatalf("expected Retry-After header")
	}
}

func TestAbuseGuardSkip(t *testing.T) {
	clock := &abuseClock{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	limiter := abuse.NewLimiter(abuse.Config{
		Rate:            1,
		Capacity:        1,
		Now:             clock.Now,
		CleanupInterval: time.Hour,
		MaxIdle:         time.Hour,
	})
	defer limiter.Stop()

	include := true
	mw := AbuseGuard(AbuseGuardConfig{
		Limiter:        limiter,
		KeyFunc:        func(*http.Request) string { return "client" },
		IncludeHeaders: &include,
		Skip:           func(*http.Request) bool { return true },
	})
	wrapped := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	decision := limiter.Allow("client")
	if !decision.Allowed {
		t.Fatalf("expected limiter bucket to be untouched by Skip")
	}
}

func TestAbuseGuardHeaderToggle(t *testing.T) {
	clock := &abuseClock{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	limiter := abuse.NewLimiter(abuse.Config{
		Rate:            1,
		Capacity:        2,
		Now:             clock.Now,
		CleanupInterval: time.Hour,
		MaxIdle:         time.Hour,
	})
	defer limiter.Stop()

	include := false
	mw := AbuseGuard(AbuseGuardConfig{
		Limiter:        limiter,
		KeyFunc:        func(*http.Request) string { return "client" },
		IncludeHeaders: &include,
	})
	wrapped := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-RateLimit-Limit"); got != "" {
		t.Fatalf("expected rate limit headers to be omitted, got %q", got)
	}
}

func TestClientIP(t *testing.T) {
	if got := httpx.ClientIP(nil); got != "" {
		t.Fatalf("expected empty key for nil request, got %q", got)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", " 1.1.1.1 , 2.2.2.2 ")
	req.RemoteAddr = "9.9.9.9:1234"
	if got := httpx.ClientIP(req); got != "1.1.1.1" {
		t.Fatalf("expected X-Forwarded-For to win, got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("X-Real-IP", " 3.3.3.3 ")
	req.RemoteAddr = "9.9.9.9:1234"
	if got := httpx.ClientIP(req); got != "3.3.3.3" {
		t.Fatalf("expected X-Real-IP to win, got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.RemoteAddr = "5.5.5.5:8080"
	if got := httpx.ClientIP(req); got != "5.5.5.5" {
		t.Fatalf("expected host part from RemoteAddr, got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.RemoteAddr = "6.6.6.6"
	if got := httpx.ClientIP(req); got != "6.6.6.6" {
		t.Fatalf("expected raw RemoteAddr when no port, got %q", got)
	}
}

func BenchmarkAbuseGuard(b *testing.B) {
	limit := b.N + 1
	limiter := abuse.NewLimiter(abuse.Config{
		Rate:            float64(limit),
		Capacity:        limit,
		CleanupInterval: time.Hour,
		MaxIdle:         time.Hour,
	})
	defer limiter.Stop()

	include := true
	mw := AbuseGuard(AbuseGuardConfig{
		Limiter:        limiter,
		KeyFunc:        func(*http.Request) string { return "client" },
		IncludeHeaders: &include,
	})
	wrapped := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}
}

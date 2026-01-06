package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/security/abuse"
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

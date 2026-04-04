package httpmetrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
)

type stubMetrics struct {
	mu     sync.Mutex
	count  int
	path   string
	status int
	bytes  int
}

func (m *stubMetrics) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.count++
	m.path = path
	m.status = status
	m.bytes = bytes
}

func TestMiddlewareObservesRoutePattern(t *testing.T) {
	collector := &stubMetrics{}
	handler := Middleware(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/users/42", nil)
	req = req.WithContext(contract.WithRequestContext(req.Context(), contract.RequestContext{
		RoutePattern: "/users/:id",
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if collector.count != 1 {
		t.Fatalf("expected one metrics observation, got %d", collector.count)
	}
	if collector.path != "/users/:id" {
		t.Fatalf("expected route pattern to be observed, got %q", collector.path)
	}
	if collector.status != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", collector.status)
	}
	if collector.bytes != len("ok") {
		t.Fatalf("expected bytes %d, got %d", len("ok"), collector.bytes)
	}
}

package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/health"
	plumelog "github.com/spcent/plumego/log"
)

type fakeChecker struct {
	name string
	err  error
}

func (f fakeChecker) Name() string                  { return f.name }
func (f fakeChecker) Check(_ context.Context) error { return f.err }

var _ health.ComponentChecker = fakeChecker{}

func discardLogger() plumelog.StructuredLogger {
	return plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatDiscard})
}

// TestHealthLiveAlwaysOK verifies the liveness probe returns 200.
func TestHealthLiveAlwaysOK(t *testing.T) {
	h := HealthHandler{ServiceName: "test", Logger: discardLogger()}
	rec := httptest.NewRecorder()
	h.Live(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Live: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestHealthReadyNoCheckers verifies the readiness probe returns 200 with no checkers registered.
func TestHealthReadyNoCheckers(t *testing.T) {
	h := HealthHandler{ServiceName: "test", Logger: discardLogger()}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Ready (no checkers): status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestHealthReadyCheckerFails verifies the readiness probe returns 503 when any checker fails.
func TestHealthReadyCheckerFails(t *testing.T) {
	h := HealthHandler{
		ServiceName: "test",
		Logger:      discardLogger(),
		Checkers:    []health.ComponentChecker{fakeChecker{"db", errors.New("connection refused")}},
	}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("Ready (failing checker): status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestHealthReadyAllCheckersPass(t *testing.T) {
	h := HealthHandler{
		ServiceName: "test",
		Logger:      discardLogger(),
		Checkers:    []health.ComponentChecker{fakeChecker{"a", nil}, fakeChecker{"b", nil}},
	}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Ready (all pass): status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestHealthReadyPartialFailure verifies all checkers run even when one fails.
func TestHealthReadyPartialFailure(t *testing.T) {
	h := HealthHandler{
		ServiceName: "test",
		Logger:      discardLogger(),
		Checkers: []health.ComponentChecker{
			fakeChecker{"cache", nil},
			fakeChecker{"db", errors.New("timeout")},
		},
	}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("Ready (partial failure): status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

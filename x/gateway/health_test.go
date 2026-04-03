package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestHealthCheckerCheck performs an immediate check on a live backend.
func TestHealthCheckerCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	pool, _ := NewBackendPool([]string{srv.URL})
	checker := NewHealthChecker(pool, &HealthCheckConfig{
		Path:           "/",
		ExpectedStatus: http.StatusOK,
		Timeout:        2 * time.Second,
	})

	backend := pool.Backends()[0]
	healthy := checker.Check(backend)
	if !healthy {
		t.Error("backend should be healthy")
	}
}

// TestHealthCheckerCheckUnhealthy checks a backend returning non-200.
func TestHealthCheckerCheckUnhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	pool, _ := NewBackendPool([]string{srv.URL})
	checker := NewHealthChecker(pool, &HealthCheckConfig{
		Path:           "/",
		ExpectedStatus: http.StatusOK,
		Timeout:        2 * time.Second,
	})

	backend := pool.Backends()[0]
	healthy := checker.Check(backend)
	if healthy {
		t.Error("backend should be unhealthy when status != expected")
	}
}

// TestHealthCheckerCheckAll checks all backends at once.
func TestHealthCheckerCheckAll(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthy.Close()

	pool, _ := NewBackendPool([]string{healthy.URL, unhealthy.URL})
	checker := NewHealthChecker(pool, &HealthCheckConfig{
		Path:           "/",
		ExpectedStatus: http.StatusOK,
		Timeout:        2 * time.Second,
	})

	checker.CheckAll()

	backends := pool.Backends()
	if !backends[0].IsHealthy() {
		t.Error("first backend should be healthy")
	}
	if backends[1].IsHealthy() {
		t.Error("second backend should be unhealthy")
	}
}

// TestHealthCheckerOnHealthChange verifies the callback fires on status change.
func TestHealthCheckerOnHealthChange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	pool, _ := NewBackendPool([]string{srv.URL})

	var changed bool
	checker := NewHealthChecker(pool, &HealthCheckConfig{
		Path:           "/",
		ExpectedStatus: http.StatusOK,
		Timeout:        2 * time.Second,
		OnHealthChange: func(b *Backend, h bool) {
			changed = true
		},
	})

	// Backend starts healthy; check returns unhealthy → triggers callback
	checker.Check(pool.Backends()[0])
	if !changed {
		t.Error("OnHealthChange should have been called")
	}
}

// TestHealthCheckerStartStop verifies the background goroutine lifecycle.
func TestHealthCheckerStartStop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	pool, _ := NewBackendPool([]string{srv.URL})
	checker := NewHealthChecker(pool, &HealthCheckConfig{
		Path:           "/",
		ExpectedStatus: http.StatusOK,
		Interval:       50 * time.Millisecond,
		Timeout:        1 * time.Second,
	})

	checker.Start()

	// Let at least one tick happen
	time.Sleep(120 * time.Millisecond)

	// Stop should not hang
	done := make(chan struct{})
	go func() {
		checker.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() timed out")
	}
}

// TestHealthCheckerStopIdempotent verifies Stop can be called multiple times.
func TestHealthCheckerStopIdempotent(t *testing.T) {
	pool, _ := NewBackendPool([]string{"http://localhost:9999"})
	checker := NewHealthChecker(pool, &HealthCheckConfig{
		Path:     "/health",
		Timeout:  100 * time.Millisecond,
		Interval: 10 * time.Second,
	})
	checker.Start()
	checker.Stop()
	checker.Stop() // second call should not panic
}

// TestHealthCheckerUnreachableBackend verifies unreachable backends are marked unhealthy.
func TestHealthCheckerUnreachableBackend(t *testing.T) {
	pool, _ := NewBackendPool([]string{"http://127.0.0.1:19999"}) // nothing listening
	checker := NewHealthChecker(pool, &HealthCheckConfig{
		Path:           "/health",
		ExpectedStatus: http.StatusOK,
		Timeout:        200 * time.Millisecond,
	})

	backend := pool.Backends()[0]
	healthy := checker.Check(backend)
	if healthy {
		t.Error("unreachable backend should be unhealthy")
	}
}

// TestHealthCheckerDefaultConfig verifies defaults are applied.
func TestHealthCheckerDefaultConfig(t *testing.T) {
	pool, _ := NewBackendPool([]string{"http://localhost:9999"})
	checker := NewHealthChecker(pool, &HealthCheckConfig{})

	if checker.config.Path != "/health" {
		t.Errorf("default path = %q, want /health", checker.config.Path)
	}
	if checker.config.Interval != 10*time.Second {
		t.Errorf("default interval = %v, want 10s", checker.config.Interval)
	}
	if checker.config.Timeout != 5*time.Second {
		t.Errorf("default timeout = %v, want 5s", checker.config.Timeout)
	}
	if checker.config.ExpectedStatus != http.StatusOK {
		t.Errorf("default status = %d, want 200", checker.config.ExpectedStatus)
	}
}

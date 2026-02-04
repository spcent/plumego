package coalesce

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoalesce_SingleRequest(t *testing.T) {
	callCount := int32(0)

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	middleware := Middleware(Config{})
	handler := middleware(backend)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected backend to be called once, got %d", callCount)
	}

	if w.Body.String() != "response" {
		t.Errorf("Expected response 'response', got '%s'", w.Body.String())
	}
}

func TestCoalesce_ConcurrentIdenticalRequests(t *testing.T) {
	callCount := int32(0)

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(100 * time.Millisecond) // Simulate slow backend
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	middleware := Middleware(Config{})
	handler := middleware(backend)

	// Launch 10 concurrent identical requests
	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// All should receive the same response
			if w.Body.String() != "response" {
				t.Errorf("Expected response 'response', got '%s'", w.Body.String())
			}
		}()
	}

	wg.Wait()

	// Backend should be called only once
	count := atomic.LoadInt32(&callCount)
	if count != 1 {
		t.Errorf("Expected backend to be called once, got %d", count)
	}
}

func TestCoalesce_DifferentRequests(t *testing.T) {
	callCount := int32(0)

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.Path))
	})

	middleware := Middleware(Config{})
	handler := middleware(backend)

	var wg sync.WaitGroup
	wg.Add(2)

	// Request 1
	go func() {
		defer wg.Done()
		req := httptest.NewRequest("GET", "/test1", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}()

	// Request 2 (different URL)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest("GET", "/test2", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}()

	wg.Wait()

	// Backend should be called twice (different URLs)
	count := atomic.LoadInt32(&callCount)
	if count != 2 {
		t.Errorf("Expected backend to be called twice, got %d", count)
	}
}

func TestCoalesce_NonCacheableMethods(t *testing.T) {
	callCount := int32(0)

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(Config{})
	handler := middleware(backend)

	var wg sync.WaitGroup
	wg.Add(2)

	// POST requests should not be coalesced
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("POST", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}()
	}

	wg.Wait()

	// Backend should be called twice (POST not coalesced)
	count := atomic.LoadInt32(&callCount)
	if count != 2 {
		t.Errorf("Expected backend to be called twice for POST, got %d", count)
	}
}

func TestCoalesce_CoalescedHeader(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	middleware := Middleware(Config{})
	handler := middleware(backend)

	var wg sync.WaitGroup
	var firstResponse *httptest.ResponseRecorder
	var secondResponse *httptest.ResponseRecorder

	wg.Add(2)

	// First request
	go func() {
		defer wg.Done()
		req := httptest.NewRequest("GET", "/test", nil)
		firstResponse = httptest.NewRecorder()
		handler.ServeHTTP(firstResponse, req)
	}()

	// Wait a bit to ensure first request starts
	time.Sleep(10 * time.Millisecond)

	// Second request (should be coalesced)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest("GET", "/test", nil)
		secondResponse = httptest.NewRecorder()
		handler.ServeHTTP(secondResponse, req)
	}()

	wg.Wait()

	// Second response should have X-Coalesced header
	if secondResponse.Header().Get("X-Coalesced") != "true" {
		t.Error("Expected X-Coalesced header on second response")
	}
}

func TestCoalesce_HeaderAwareKeyFunc(t *testing.T) {
	callCount := int32(0)

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	// Use header-aware key function
	middleware := Middleware(Config{
		KeyFunc: HeaderAwareKeyFunc([]string{"Accept"}),
	})
	handler := middleware(backend)

	var wg sync.WaitGroup
	wg.Add(2)

	// Request 1 with Accept: application/json
	go func() {
		defer wg.Done()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}()

	// Request 2 with Accept: application/xml (different header)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept", "application/xml")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}()

	wg.Wait()

	// Backend should be called twice (different Accept headers)
	count := atomic.LoadInt32(&callCount)
	if count != 2 {
		t.Errorf("Expected backend to be called twice with different headers, got %d", count)
	}
}

func TestCoalesce_Stats(t *testing.T) {
	coalescer := New(Config{})

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := coalescer.Middleware()(backend)

	// Start a request in background
	go func() {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}()

	// Wait a bit for request to be in-flight
	time.Sleep(20 * time.Millisecond)

	// Check stats
	stats := coalescer.Stats()
	if stats.InFlight != 1 {
		t.Errorf("Expected 1 in-flight request, got %d", stats.InFlight)
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	// Check stats again
	stats = coalescer.Stats()
	if stats.InFlight != 0 {
		t.Errorf("Expected 0 in-flight requests after completion, got %d", stats.InFlight)
	}
}

func TestCoalesce_OnCoalescedCallback(t *testing.T) {
	coalescedCount := int32(0)
	var coalescedKey string

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(Config{
		OnCoalesced: func(key string, count int) {
			atomic.AddInt32(&coalescedCount, 1)
			coalescedKey = key
		},
	})
	handler := middleware(backend)

	var wg sync.WaitGroup
	wg.Add(3)

	// Launch 3 concurrent identical requests
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}()
	}

	wg.Wait()

	// Callback should be called for waiters (2 requests waited)
	count := atomic.LoadInt32(&coalescedCount)
	if count != 2 {
		t.Errorf("Expected OnCoalesced to be called twice, got %d", count)
	}

	if coalescedKey == "" {
		t.Error("Expected coalescedKey to be set")
	}
}

func TestDefaultKeyFunc(t *testing.T) {
	req1 := httptest.NewRequest("GET", "/test?a=1", nil)
	req2 := httptest.NewRequest("GET", "/test?a=1", nil)
	req3 := httptest.NewRequest("GET", "/test?a=2", nil)

	key1 := DefaultKeyFunc(req1)
	key2 := DefaultKeyFunc(req2)
	key3 := DefaultKeyFunc(req3)

	if key1 != key2 {
		t.Error("Expected identical requests to generate same key")
	}

	if key1 == key3 {
		t.Error("Expected different requests to generate different keys")
	}
}

func TestCoalesce_ResponseStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(50 * time.Millisecond)
				w.WriteHeader(tt.statusCode)
			})

			middleware := Middleware(Config{})
			handler := middleware(backend)

			var wg sync.WaitGroup
			var responses []*httptest.ResponseRecorder

			wg.Add(2)

			// Launch 2 concurrent requests
			for i := 0; i < 2; i++ {
				go func() {
					defer wg.Done()
					req := httptest.NewRequest("GET", "/test", nil)
					w := httptest.NewRecorder()
					handler.ServeHTTP(w, req)
					responses = append(responses, w)
				}()
			}

			wg.Wait()

			// Both should receive the same status code
			for _, resp := range responses {
				if resp.Code != tt.statusCode {
					t.Errorf("Expected status %d, got %d", tt.statusCode, resp.Code)
				}
			}
		})
	}
}

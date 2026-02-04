package recovery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		shouldPanic    bool
		expectedStatus int
	}{
		{
			name: "Normal handler execution - no panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}),
			shouldPanic:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Handler panics with string",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			}),
			shouldPanic:    true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Handler panics with error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(http.ErrMissingFile)
			}),
			shouldPanic:    true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Handler panics with integer",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(42)
			}),
			shouldPanic:    true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			recoveryHandler := RecoveryMiddleware(tt.handler)
			recoveryHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.shouldPanic {
				var response contract.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error.Code != "internal_error" {
					t.Errorf("expected error code internal_error, got %s", response.Error.Code)
				}
				if response.Error.Category != contract.CategoryServer {
					t.Errorf("expected server error category, got %s", response.Error.Category)
				}
			} else if body := w.Body.String(); body != "success" {
				t.Errorf("expected success body, got %q", body)
			}
		})
	}
}

// Test that recovery middleware doesn't interfere with normal responses
func TestRecoveryMiddleware_NormalFlow(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Custom-Header", "test-value")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("custom response"))
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	recoveryHandler := RecoveryMiddleware(handler)
	recoveryHandler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	if w.Header().Get("Custom-Header") != "test-value" {
		t.Errorf("custom header not preserved")
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("content-type header not preserved")
	}

	if body := w.Body.String(); body != "custom response" {
		t.Errorf("response body not preserved, got %q", body)
	}
}

func TestRecoveryMiddleware_Concurrent(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("concurrent panic")
	})
	recoveryHandler := RecoveryMiddleware(panicHandler)

	const numRequests = 10
	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/panic", nil)
			w := httptest.NewRecorder()
			recoveryHandler.ServeHTTP(w, req)
			if w.Code != http.StatusInternalServerError {
				t.Errorf("expected status 500, got %d", w.Code)
			}
		}()
	}

	wg.Wait()
}

// Test that middleware properly handles different HTTP methods
func TestRecoveryMiddleware_DifferentMethods(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("method panic")
			})

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			recoveryHandler := RecoveryMiddleware(panicHandler)
			recoveryHandler.ServeHTTP(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("method %s: expected status 500, got %d", method, w.Code)
			}
		})
	}
}

// Benchmarks remain to ensure middleware overhead stays bounded.
func BenchmarkRecoveryMiddleware_NoPanic(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	recoveryHandler := RecoveryMiddleware(handler)
	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		recoveryHandler.ServeHTTP(w, req)
	}
}

func BenchmarkRecoveryMiddleware_WithPanic(b *testing.B) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("benchmark panic")
	})

	recoveryHandler := RecoveryMiddleware(panicHandler)
	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		recoveryHandler.ServeHTTP(w, req)
	}
}

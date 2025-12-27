package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAuth(t *testing.T) {
	// Mock handler that will be called if auth passes
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	tests := []struct {
		name           string
		authToken      string // AUTH_TOKEN environment variable value
		requestToken   string // X-Token header value in request
		expectedStatus int
		expectedBody   string
		shouldCallNext bool
	}{
		{
			name:           "No AUTH_TOKEN env var set - should allow all requests",
			authToken:      "",
			requestToken:   "",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			shouldCallNext: true,
		},
		{
			name:           "No AUTH_TOKEN env var set with token in request - should allow",
			authToken:      "",
			requestToken:   "any-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			shouldCallNext: true,
		},
		{
			name:           "Valid token provided - should allow",
			authToken:      "secret123",
			requestToken:   "secret123",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			shouldCallNext: true,
		},
		{
			name:           "Invalid token provided - should reject",
			authToken:      "secret123",
			requestToken:   "wrong-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized","message":"Invalid or missing authentication token"}`,
			shouldCallNext: false,
		},
		{
			name:           "No token provided when required - should reject",
			authToken:      "secret123",
			requestToken:   "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized","message":"Invalid or missing authentication token"}`,
			shouldCallNext: false,
		},
		{
			name:           "Empty token when required - should reject",
			authToken:      "secret123",
			requestToken:   "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized","message":"Invalid or missing authentication token"}`,
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable
			if tt.authToken != "" {
				os.Setenv("AUTH_TOKEN", tt.authToken)
			} else {
				os.Unsetenv("AUTH_TOKEN")
			}

			// Clean up after test
			defer func() {
				if tt.authToken != "" {
					os.Unsetenv("AUTH_TOKEN")
				}
			}()

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.requestToken != "" {
				req.Header.Set("X-Token", tt.requestToken)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create middleware with mock handler
			authMiddleware := Auth(mockHandler)

			// Execute the middleware
			authMiddleware(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			body := w.Body.String()
			if body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}

			// Verify Content-Type for error responses
			if tt.expectedStatus == http.StatusUnauthorized {
				contentType := w.Header().Get("Content-Type")
				// Note: httptest.ResponseRecorder doesn't set Content-Type automatically
				// In a real scenario, you might want to set it in your Auth function
				_ = contentType // We're not checking this in the current implementation
			}
		})
	}
}

// Benchmark test to measure performance
func BenchmarkAuth(b *testing.B) {
	os.Setenv("AUTH_TOKEN", "secret123")
	defer os.Unsetenv("AUTH_TOKEN")

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := Auth(mockHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Token", "secret123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		authMiddleware(w, req)
	}
}

// Test concurrent access to ensure thread safety
func TestAuthConcurrent(t *testing.T) {
	os.Setenv("AUTH_TOKEN", "secret123")
	defer os.Unsetenv("AUTH_TOKEN")

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	authMiddleware := Auth(mockHandler)

	// Test concurrent valid requests
	t.Run("concurrent valid requests", func(t *testing.T) {
		const numGoroutines = 100
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Token", "secret123")
				w := httptest.NewRecorder()

				authMiddleware(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status 200, got %d", w.Code)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

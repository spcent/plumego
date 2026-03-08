package auth

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestSimpleAuthMiddlewareAuthenticate(t *testing.T) {
	tests := []struct {
		name            string
		configuredToken string
		authorization   string
		xToken          string
		url             string
		cookie          *http.Cookie
		expectedStatus  int
		expectedBody    string
	}{
		{
			name:            "valid x-token header",
			configuredToken: "secret123",
			xToken:          "secret123",
			expectedStatus:  http.StatusOK,
			expectedBody:    "success",
		},
		{
			name:            "valid bearer token",
			configuredToken: "secret123",
			authorization:   "Bearer secret123",
			expectedStatus:  http.StatusOK,
			expectedBody:    "success",
		},
		{
			name:            "missing configured token fails closed",
			configuredToken: "",
			expectedStatus:  http.StatusUnauthorized,
			expectedBody:    `{"error":{"code":"auth_unauthenticated","message":"invalid or missing authentication token","category":"auth_error"}}` + "\n",
		},
		{
			name:            "missing credentials",
			configuredToken: "secret123",
			expectedStatus:  http.StatusUnauthorized,
			expectedBody:    `{"error":{"code":"auth_unauthenticated","message":"invalid or missing authentication token","category":"auth_error"}}` + "\n",
		},
		{
			name:            "malformed authorization header",
			configuredToken: "secret123",
			authorization:   "Bearer",
			expectedStatus:  http.StatusUnauthorized,
			expectedBody:    `{"error":{"code":"auth_unauthenticated","message":"invalid or missing authentication token","category":"auth_error"}}` + "\n",
		},
		{
			name:            "query token ignored",
			configuredToken: "secret123",
			url:             "/test?token=secret123",
			expectedStatus:  http.StatusUnauthorized,
			expectedBody:    `{"error":{"code":"auth_unauthenticated","message":"invalid or missing authentication token","category":"auth_error"}}` + "\n",
		},
		{
			name:            "cookie token ignored",
			configuredToken: "secret123",
			cookie:          &http.Cookie{Name: "auth_token", Value: "secret123"},
			expectedStatus:  http.StatusUnauthorized,
			expectedBody:    `{"error":{"code":"auth_unauthenticated","message":"invalid or missing authentication token","category":"auth_error"}}` + "\n",
		},
		{
			name:            "invalid token",
			configuredToken: "secret123",
			xToken:          "wrong-token",
			expectedStatus:  http.StatusUnauthorized,
			expectedBody:    `{"error":{"code":"auth_unauthenticated","message":"invalid or missing authentication token","category":"auth_error"}}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hit := false
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hit = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
			})

			url := tt.url
			if url == "" {
				url = "/test"
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			if tt.xToken != "" {
				req.Header.Set("X-Token", tt.xToken)
			}
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			w := httptest.NewRecorder()
			authMiddleware := NewSimpleAuthMiddleware(tt.configuredToken).Authenticate(mockHandler)
			authMiddleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if body := w.Body.String(); body != tt.expectedBody {
				t.Fatalf("expected body %q, got %q", tt.expectedBody, body)
			}
			if tt.expectedStatus == http.StatusUnauthorized && w.Header().Get("WWW-Authenticate") == "" {
				t.Fatalf("expected WWW-Authenticate header")
			}
			if tt.expectedStatus == http.StatusOK && !hit {
				t.Fatalf("expected downstream handler call")
			}
			if tt.expectedStatus == http.StatusUnauthorized && hit {
				t.Fatalf("did not expect downstream handler call")
			}
		})
	}
}

func BenchmarkSimpleAuthMiddleware(b *testing.B) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := NewSimpleAuthMiddleware("secret123").Authenticate(mockHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Token", "secret123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		authMiddleware.ServeHTTP(w, req)
	}
}

func TestSimpleAuthMiddlewareConcurrent(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	authMiddleware := NewSimpleAuthMiddleware("secret123").Authenticate(mockHandler)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer secret123")
			w := httptest.NewRecorder()

			authMiddleware.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
		}()
	}

	wg.Wait()
}

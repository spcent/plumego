package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHTTPMethodMatching tests all HTTP methods and fallback behavior comprehensively.
func TestHTTPMethodMatching(t *testing.T) {
	tests := []struct {
		name              string
		registeredMethods []string
		requestMethod     string
		path              string
		expectMatch       bool
		expectStatus      int
	}{
		// Exact method matches
		{name: "GET matches GET", registeredMethods: []string{"GET"}, requestMethod: "GET", path: "/", expectMatch: true, expectStatus: 200},
		{name: "POST matches POST", registeredMethods: []string{"POST"}, requestMethod: "POST", path: "/", expectMatch: true, expectStatus: 200},
		{name: "PUT matches PUT", registeredMethods: []string{"PUT"}, requestMethod: "PUT", path: "/", expectMatch: true, expectStatus: 200},
		{name: "DELETE matches DELETE", registeredMethods: []string{"DELETE"}, requestMethod: "DELETE", path: "/", expectMatch: true, expectStatus: 200},
		{name: "PATCH matches PATCH", registeredMethods: []string{"PATCH"}, requestMethod: "PATCH", path: "/", expectMatch: true, expectStatus: 200},
		{name: "OPTIONS matches OPTIONS", registeredMethods: []string{"OPTIONS"}, requestMethod: "OPTIONS", path: "/", expectMatch: true, expectStatus: 200},
		{name: "TRACE matches TRACE", registeredMethods: []string{"TRACE"}, requestMethod: "TRACE", path: "/", expectMatch: true, expectStatus: 200},
		{name: "CONNECT matches CONNECT", registeredMethods: []string{"CONNECT"}, requestMethod: "CONNECT", path: "/", expectMatch: true, expectStatus: 200},

		// HEAD fallback to GET
		{name: "HEAD falls back to GET", registeredMethods: []string{"GET"}, requestMethod: "HEAD", path: "/", expectMatch: true, expectStatus: 200},
		{name: "HEAD no match without GET", registeredMethods: []string{"POST"}, requestMethod: "HEAD", path: "/", expectMatch: false, expectStatus: 404},

		// MethodAny matches all
		{name: "MethodAny matches GET", registeredMethods: []string{"ANY"}, requestMethod: "GET", path: "/", expectMatch: true, expectStatus: 200},
		{name: "MethodAny matches POST", registeredMethods: []string{"ANY"}, requestMethod: "POST", path: "/", expectMatch: true, expectStatus: 200},
		{name: "MethodAny matches HEAD", registeredMethods: []string{"ANY"}, requestMethod: "HEAD", path: "/", expectMatch: true, expectStatus: 200},

		// Method not found
		{name: "POST not found when only GET", registeredMethods: []string{"GET"}, requestMethod: "POST", path: "/", expectMatch: false, expectStatus: 404},
		{name: "DELETE not found when only PUT", registeredMethods: []string{"PUT"}, requestMethod: "DELETE", path: "/", expectMatch: false, expectStatus: 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()

			// Register routes for each method
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			for _, method := range tt.registeredMethods {
				if err := r.AddRoute(method, tt.path, handler); err != nil {
					t.Fatalf("AddRoute failed: %v", err)
				}
			}

			req := httptest.NewRequest(tt.requestMethod, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if tt.expectMatch && rec.Code == http.StatusNotFound {
				t.Errorf("expected match but got 404")
			}
			if !tt.expectMatch && rec.Code != http.StatusNotFound {
				t.Errorf("expected 404 but got %d", rec.Code)
			}
		})
	}
}

// TestPathNormalizationEdgeCases tests path normalization with comprehensive edge cases.
func TestPathNormalizationEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		registerPath   string
		requestPath    string
		shouldMatch    bool
		expectedStatus int
	}{
		// Root variations
		{name: "root normalizes empty to /", registerPath: "", requestPath: "/", shouldMatch: true, expectedStatus: 200},
		{name: "root with trailing slash", registerPath: "/", requestPath: "/", shouldMatch: true, expectedStatus: 200},

		// Segment variations
		{name: "segment matches exactly", registerPath: "/api/users", requestPath: "/api/users", shouldMatch: true, expectedStatus: 200},
		{name: "segment does not match with extra slash", registerPath: "/api/users", requestPath: "/api//users", shouldMatch: false, expectedStatus: 404},
		{name: "segment case sensitive", registerPath: "/users", requestPath: "/Users", shouldMatch: false, expectedStatus: 404},

		// Trailing slashes
		{name: "trailing slash stripped on register", registerPath: "/api/", requestPath: "/api", shouldMatch: true, expectedStatus: 200},
		{name: "no trailing slash on register matches root request", registerPath: "/api", requestPath: "/api", shouldMatch: true, expectedStatus: 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			if err := r.AddRoute(http.MethodGet, tt.registerPath, handler); err != nil {
				t.Fatalf("AddRoute failed: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if tt.shouldMatch && rec.Code != http.StatusOK {
				t.Errorf("expected match but got %d", rec.Code)
			}
			if !tt.shouldMatch && rec.Code != http.StatusNotFound {
				t.Errorf("expected 404 but got %d", rec.Code)
			}
		})
	}
}

// TestParameterExtractionComprehensive tests parameter extraction in various scenarios.
func TestParameterExtractionComprehensive(t *testing.T) {
	tests := []struct {
		name          string
		routePath     string
		requestPath   string
		paramName     string
		expectedValue string
		shouldMatch   bool
	}{
		// Single parameters
		{name: "single param at end", routePath: "/users/:id", requestPath: "/users/123", paramName: "id", expectedValue: "123", shouldMatch: true},
		{name: "single param in middle", routePath: "/:lang/docs", requestPath: "/en/docs", paramName: "lang", expectedValue: "en", shouldMatch: true},
		{name: "single param at start", routePath: "/:version/api", requestPath: "/v1/api", paramName: "version", expectedValue: "v1", shouldMatch: true},

		// Multiple parameters
		{name: "two params", routePath: "/users/:id/posts/:postId", requestPath: "/users/123/posts/456", paramName: "id", expectedValue: "123", shouldMatch: true},
		{name: "two params extract second", routePath: "/users/:id/posts/:postId", requestPath: "/users/123/posts/456", paramName: "postId", expectedValue: "456", shouldMatch: true},

		// Parameter values with special chars
		{name: "param with hyphen", routePath: "/files/:name", requestPath: "/files/my-file.txt", paramName: "name", expectedValue: "my-file.txt", shouldMatch: true},
		{name: "param with underscore", routePath: "/users/:slug", requestPath: "/users/john_doe", paramName: "slug", expectedValue: "john_doe", shouldMatch: true},
		{name: "param with numbers", routePath: "/posts/:id", requestPath: "/posts/2026", paramName: "id", expectedValue: "2026", shouldMatch: true},

		// Non-matching cases
		{name: "missing segment", routePath: "/users/:id/posts", requestPath: "/users/123", paramName: "id", expectedValue: "", shouldMatch: false},
		{name: "extra segment", routePath: "/users/:id", requestPath: "/users/123/posts", paramName: "id", expectedValue: "", shouldMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			var capturedValue string

			handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				capturedValue = Param(req, tt.paramName)
				w.WriteHeader(http.StatusOK)
			})

			if err := r.AddRoute(http.MethodGet, tt.routePath, handler); err != nil {
				t.Fatalf("AddRoute failed: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if tt.shouldMatch {
				if rec.Code != http.StatusOK {
					t.Errorf("expected match but got %d", rec.Code)
				}
				if capturedValue != tt.expectedValue {
					t.Errorf("expected param %q but got %q", tt.expectedValue, capturedValue)
				}
			} else {
				if rec.Code != http.StatusNotFound {
					t.Errorf("expected 404 but got %d", rec.Code)
				}
			}
		})
	}
}

// TestRouteMatchingPriority tests static > param > wildcard priority.
func TestRouteMatchingPriority(t *testing.T) {
	tests := []struct {
		name         string
		routePaths   map[string]string // path -> response
		requestPath  string
		expectedResp string
	}{
		{
			name: "static beats param",
			routePaths: map[string]string{
				"/users/:id":   "param",
				"/users/admin": "static",
			},
			requestPath:  "/users/admin",
			expectedResp: "static",
		},
		{
			name: "param beats wildcard",
			routePaths: map[string]string{
				"/files/*path": "wildcard",
				"/files/:id":   "param",
			},
			requestPath:  "/files/123",
			expectedResp: "param",
		},
		{
			name: "static beats wildcard",
			routePaths: map[string]string{
				"/api/*path":     "wildcard",
				"/api/users":     "static",
				"/api/users/:id": "param",
			},
			requestPath:  "/api/users",
			expectedResp: "static",
		},
		{
			name: "param specific beats wildcard",
			routePaths: map[string]string{
				"/api/v1/*path":     "wildcard",
				"/api/v1/users/*id": "param-wildcard",
			},
			requestPath:  "/api/v1/users/123/posts",
			expectedResp: "param-wildcard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()

			for path, resp := range tt.routePaths {
				respValue := resp // capture for closure
				handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(respValue))
				})
				if err := r.AddRoute(http.MethodGet, path, handler); err != nil {
					t.Fatalf("AddRoute failed: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected match but got %d", rec.Code)
			}
			if rec.Body.String() != tt.expectedResp {
				t.Errorf("expected %q but got %q", tt.expectedResp, rec.Body.String())
			}
		})
	}
}

// TestGroupPrefixBehavior tests group prefix handling comprehensively.
func TestGroupPrefixBehavior(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*Router)
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "single group",
			setup: func(r *Router) {
				api := r.Group("/api")
				mustAddRoute(api, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("users"))
				}))
			},
			requestPath:    "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "users",
		},
		{
			name: "nested groups",
			setup: func(r *Router) {
				api := r.Group("/api")
				v1 := api.Group("/v1")
				mustAddRoute(v1, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("v1-users"))
				}))
			},
			requestPath:    "/api/v1/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "v1-users",
		},
		{
			name: "group with trailing slash normalized",
			setup: func(r *Router) {
				api := r.Group("/api/")
				mustAddRoute(api, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("users"))
				}))
			},
			requestPath:    "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "users",
		},
		{
			name: "group root route",
			setup: func(r *Router) {
				api := r.Group("/api")
				mustAddRoute(api, http.MethodGet, "", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("api-root"))
				}))
			},
			requestPath:    "/api",
			expectedStatus: http.StatusOK,
			expectedBody:   "api-root",
		},
		{
			name: "multiple groups at same level",
			setup: func(r *Router) {
				api := r.Group("/api")
				admin := r.Group("/admin")
				mustAddRoute(api, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("api-users"))
				}))
				mustAddRoute(admin, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("admin-users"))
				}))
			},
			requestPath:    "/admin/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "admin-users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			tt.setup(r)

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d but got %d", tt.expectedStatus, rec.Code)
			}
			if rec.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q but got %q", tt.expectedBody, rec.Body.String())
			}
		})
	}
}

// TestWildcardParameterMatching tests wildcard route matching behavior.
func TestWildcardParameterMatching(t *testing.T) {
	tests := []struct {
		name           string
		routePath      string
		requestPath    string
		expectedStatus int
		expectedParam  string
	}{
		{
			name:           "wildcard captures single segment",
			routePath:      "/files/*path",
			requestPath:    "/files/document.pdf",
			expectedStatus: http.StatusOK,
			expectedParam:  "document.pdf",
		},
		{
			name:           "wildcard captures multiple segments",
			routePath:      "/files/*path",
			requestPath:    "/files/docs/2026/report.pdf",
			expectedStatus: http.StatusOK,
			expectedParam:  "docs/2026/report.pdf",
		},
		{
			name:           "wildcard with param in prefix",
			routePath:      "/users/:id/files/*path",
			requestPath:    "/users/123/files/docs/readme.txt",
			expectedStatus: http.StatusOK,
			expectedParam:  "docs/readme.txt",
		},
		{
			name:           "wildcard does not match empty",
			routePath:      "/files/*path",
			requestPath:    "/files",
			expectedStatus: http.StatusNotFound,
			expectedParam:  "",
		},
		{
			name:           "wildcard does not match with trailing slash only",
			routePath:      "/files/*path",
			requestPath:    "/files/",
			expectedStatus: http.StatusNotFound,
			expectedParam:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			var capturedParam string

			handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				capturedParam = Param(req, "path")
				w.WriteHeader(http.StatusOK)
			})

			if err := r.AddRoute(http.MethodGet, tt.routePath, handler); err != nil {
				t.Fatalf("AddRoute failed: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d but got %d", tt.expectedStatus, rec.Code)
			}
			if tt.expectedStatus == http.StatusOK && capturedParam != tt.expectedParam {
				t.Errorf("expected param %q but got %q", tt.expectedParam, capturedParam)
			}
		})
	}
}

// TestFreezeLifecycleBehavior tests router freeze behavior comprehensively.
func TestFreezeLifecycleBehavior(t *testing.T) {
	tests := []struct {
		name          string
		operation     func(*Router) error
		expectError   bool
		errorContains string
	}{
		{
			name: "AddRoute after Freeze returns error",
			operation: func(r *Router) error {
				r.Freeze()
				return r.AddRoute(http.MethodGet, "/test", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
			},
			expectError:   true,
			errorContains: "frozen",
		},
		{
			name: "Static after Freeze returns error",
			operation: func(r *Router) error {
				r.Freeze()
				return r.Static("/static", t.TempDir())
			},
			expectError:   true,
			errorContains: "frozen",
		},
		{
			name: "StaticFS after Freeze returns error",
			operation: func(r *Router) error {
				r.Freeze()
				return r.StaticFS("/static", nil)
			},
			expectError:   true,
			errorContains: "frozen",
		},
		{
			name: "Multiple Freeze calls are safe",
			operation: func(r *Router) error {
				r.Freeze()
				r.Freeze()
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			err := tt.operation(r)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.expectError && err != nil && !contains(err.Error(), tt.errorContains) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

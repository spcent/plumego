package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestBasicRoutes(t *testing.T) {
	// Reset global router
	r := NewRouter()

	r.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.PostFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("echo"))
	})

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/ping", "pong"},
		{"POST", "/echo", "echo"},
		{"GET", "/echo", "404 page not found\n"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		resp := w.Body.String()
		if resp != tt.expected {
			t.Errorf("[%s %s] expected %q, got %q", tt.method, tt.path, tt.expected, resp)
		}
	}
}

func TestParamRoutes(t *testing.T) {
	r := NewRouter()

	r.GetFunc("/hello/:name", func(w http.ResponseWriter, r *http.Request) {
		name, _ := contract.Param(r, "name")
		ctxParams := contract.ParamsFromContext(r.Context())
		if ctxParams["name"] != name {
			t.Fatalf("context params mismatch: got %s want %s", ctxParams["name"], name)
		}
		w.Write([]byte("Hello " + name))
	})

	r.GetFunc("/users/:id/books/:bookId", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		bookID, _ := contract.Param(r, "bookId")
		ctxParams := contract.ParamsFromContext(r.Context())
		if ctxParams["id"] != id || ctxParams["bookId"] != bookID {
			t.Fatalf("context params mismatch: %v", ctxParams)
		}
		w.Write([]byte("User " + id + " Book " + bookID))
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/hello/Alice", "Hello Alice"},
		{"/hello/Bob", "Hello Bob"},
		{"/users/123/books/456", "User 123 Book 456"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		resp := strings.TrimSpace(w.Body.String())
		if resp != tt.expected {
			t.Errorf("[%s] expected %q, got %q", tt.path, tt.expected, resp)
		}
	}
}

func TestParamsInjectedIntoContext(t *testing.T) {
	r := NewRouter()

	r.GetFunc("/hello/:name", func(w http.ResponseWriter, r *http.Request) {
		ctxParams := contract.ParamsFromContext(r.Context())
		if ctxParams == nil {
			t.Fatalf("expected params in context")
		}

		paramVal, ok := contract.Param(r, "name")
		if !ok {
			t.Fatalf("expected Param helper to find name")
		}

		if ctxParams["name"] != paramVal {
			t.Fatalf("context params mismatch: got %s want %s", ctxParams["name"], paramVal)
		}
		w.Write([]byte(ctxParams["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Alice", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "Alice" {
		t.Fatalf("expected context value to be written, got %q", got)
	}
}

func TestRequestContextHelpers(t *testing.T) {
	r := NewRouter()

	r.GetFunc("/hello/:name", func(w http.ResponseWriter, r *http.Request) {
		rc := contract.RequestContextFrom(r.Context())
		name, _ := contract.Param(r, "name")
		if rc.Params["name"] != name {
			t.Fatalf("request context params mismatch: got %s want %s", rc.Params["name"], name)
		}

		if val, ok := contract.Param(r, "name"); !ok || val != name {
			t.Fatalf("Param helper mismatch: got %s (exists=%t) want %s", val, ok, name)
		}

		if _, ok := contract.Param(r, "missing"); ok {
			t.Fatalf("expected missing parameter to return ok=false")
		}

		w.Write([]byte(rc.Params["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Carol", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "Carol" {
		t.Fatalf("expected Param helper to provide name, got %q", got)
	}
}

func TestContextHandlerRegistration(t *testing.T) {
	r := NewRouter()

	r.GetFunc("/ctx/:id", func(w http.ResponseWriter, r *http.Request) {
		rc := contract.RequestContextFrom(r.Context())
		if rc.Params == nil {
			t.Fatalf("expected RequestContext to be present")
		}

		paramVal, ok := contract.Param(r, "id")
		if !ok {
			t.Fatalf("expected Param helper to find id")
		}

		if rc.Params["id"] != paramVal {
			t.Fatalf("context param mismatch: got %s want %s", rc.Params["id"], paramVal)
		}

		w.Write([]byte(paramVal))
	})

	req := httptest.NewRequest(http.MethodGet, "/ctx/42", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "42" {
		t.Fatalf("expected context handler to write param, got %q", got)
	}
}

func TestAnyRoute(t *testing.T) {
	r := NewRouter()

	r.Any("/any", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("any"))
	}))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/any", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Body.String() != "any" {
			t.Errorf("[%s /any] expected %q, got %q", method, "any", w.Body.String())
		}
	}
}

func TestAnyRootFallback(t *testing.T) {
	r := NewRouter()

	r.GetFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("docs"))
	})

	r.Any("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("home"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "home" {
		t.Fatalf("expected any root to match, got %q", got)
	}
}

func TestPrintRoutes(t *testing.T) {
	// Reset global r
	r := NewRouter()

	r.Get("/print1", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	r.Post("/print2", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	r.Any("/print3", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	// Read captured output
	var outBuf bytes.Buffer
	r.Print(&outBuf)

	output := outBuf.String()
	if !strings.Contains(output, "GET    /print1") {
		t.Errorf("PrintRoutes output missing GET /print1: %s", output)
	}
	if !strings.Contains(output, "POST   /print2") {
		t.Errorf("PrintRoutes output missing POST /print2: %s", output)
	}
	if !strings.Contains(output, "/print3") {
		t.Errorf("PrintRoutes output missing /print3: %s", output)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	r := NewRouter()

	r.Any("/any", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("any"))
	}))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/any", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Body.String() != "any" {
			t.Errorf("[%s /any] expected %q, got %q", method, "any", w.Body.String())
		}
	}
}

func TestRouteGroup(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api")
	v1 := api.Group("/v1")
	v2 := api.Group("/v2")

	v1.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		ctxParams := contract.ParamsFromContext(r.Context())
		if ctxParams["id"] != id {
			t.Fatalf("expected id in context")
		}
		w.Write([]byte("User " + id))
	})
	v1.PostFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Create User"))
	})
	v2.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Users v2"))
	})

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/api/v1/users/123", "User 123"},
		{"POST", "/api/v1/users", "Create User"},
		{"GET", "/api/v1/users", "404 page not found\n"},
		{"GET", "/api/v2/users/123", "404 page not found\n"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		resp := w.Body.String()
		if resp != tt.expected {
			t.Errorf("[%s %s] expected %q, got %q", tt.method, tt.path, tt.expected, resp)
		}
	}
}

func TestRouteGroupMiddlewares(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api")
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Group", "api")
			next.ServeHTTP(w, r)
		})
	})

	api.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	req := httptest.NewRequest("GET", "/api/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Group") != "api" {
		t.Errorf("expected middleware to set X-Group header")
	}
	if w.Body.String() != "pong" {
		t.Errorf("expected response body 'pong', got %q", w.Body.String())
	}
}

func TestRouteGroupMiddlewareIsolation(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api")
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Group", "api")
			next.ServeHTTP(w, r)
		})
	})

	api.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.GetFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/api/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Group") != "api" {
		t.Errorf("expected group middleware on /api/ping")
	}

	req = httptest.NewRequest("GET", "/status", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Group") != "" {
		t.Errorf("expected group middleware not to run on /status")
	}
}

func TestNestedGroupMiddlewareOrder(t *testing.T) {
	r := NewRouter()

	var order []string

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "global")
			next.ServeHTTP(w, r)
		})
	})

	api := r.Group("/api")
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "api")
			next.ServeHTTP(w, r)
		})
	})

	v1 := api.Group("/v1")
	v1.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "v1")
			next.ServeHTTP(w, r)
		})
	})

	v1.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.Write([]byte("pong"))
	})

	req := httptest.NewRequest("GET", "/api/v1/ping", nil)
	w := httptest.NewRecorder()
	order = []string{}
	r.ServeHTTP(w, req)

	expected := []string{"global", "api", "v1", "handler"}
	if !slicesEqual(order, expected) {
		t.Fatalf("expected middleware order %v, got %v", expected, order)
	}
}

func TestRouterFreeze(t *testing.T) {
	r := NewRouter()

	r.Get("/ping", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	r.Freeze()

	err := r.AddRoute("GET", "/panic", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	if err == nil {
		t.Errorf("expected error when adding route after freeze")
	}
}

func TestRouteMetadata(t *testing.T) {
	r := NewRouter()

	err := r.AddRouteWithOptions("GET", "/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithRouteName("ping"), WithRouteTags("health", "status"))
	if err != nil {
		t.Fatalf("add route failed: %v", err)
	}

	routes := r.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Meta.Name != "ping" {
		t.Fatalf("expected name 'ping', got %q", route.Meta.Name)
	}
	if len(route.Meta.Tags) != 2 || route.Meta.Tags[0] != "health" {
		t.Fatalf("unexpected tags: %v", route.Meta.Tags)
	}
}

func TestRouterCtxHandler(t *testing.T) {
	r := NewRouter()
	r.GetCtx("/hello/:name", func(ctx *contract.Ctx) {
		_ = ctx.JSON(http.StatusOK, map[string]string{
			"name":  ctx.Params["name"],
			"trace": ctx.TraceID,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/gopher", nil)
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", recorder.Code)
	}

	var payload map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload["name"] != "gopher" {
		t.Fatalf("expected param to be available, got %+v", payload)
	}
}

// --- Enhanced Route Group Tests ---

func TestGroupPrefixNormalization(t *testing.T) {
	tests := []struct {
		name       string
		parent     string
		child      string
		wantPrefix string
	}{
		{"basic", "/api", "/v1", "/api/v1"},
		{"trailing slash parent", "/api/", "/v1", "/api/v1"},
		{"trailing slash child", "/api", "/v1/", "/api/v1"},
		{"both trailing slashes", "/api/", "/v1/", "/api/v1"},
		{"child missing leading slash", "/api", "v1", "/api/v1"},
		{"empty parent", "", "/v1", "/v1"},
		{"empty child", "/api", "", "/api"},
		{"both empty", "", "", ""},
		{"root child", "/api", "/", "/api"},
		{"deep nesting", "/a/b", "/c/d", "/a/b/c/d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeGroupPrefix(tt.parent, tt.child)
			if got != tt.wantPrefix {
				t.Errorf("normalizeGroupPrefix(%q, %q) = %q, want %q",
					tt.parent, tt.child, got, tt.wantPrefix)
			}
		})
	}
}

func TestGroupNoPrefixDoubleSlash(t *testing.T) {
	r := NewRouter()

	// Trailing slash on group prefix should not produce double slashes
	api := r.Group("/api/")
	v1 := api.Group("/v1/")

	v1.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Errorf("expected 200/ok, got %d/%q", w.Code, w.Body.String())
	}
}

func TestGroupMissingLeadingSlash(t *testing.T) {
	r := NewRouter()

	api := r.Group("api") // no leading slash
	api.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	req := httptest.NewRequest("GET", "/api/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.String() != "pong" {
		t.Errorf("expected 200/pong, got %d/%q", w.Code, w.Body.String())
	}
}

func TestGroupEmptyPrefix(t *testing.T) {
	r := NewRouter()

	// Empty prefix group should behave like root
	g := r.Group("")
	g.GetFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Errorf("expected 200/ok, got %d/%q", w.Code, w.Body.String())
	}
}

func TestDeepNestedGroups(t *testing.T) {
	r := NewRouter()

	// 4 levels of nesting
	api := r.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")
	settings := users.Group("/settings")

	settings.GetFunc("/theme", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("dark"))
	})

	req := httptest.NewRequest("GET", "/api/v1/users/settings/theme", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.String() != "dark" {
		t.Errorf("expected 200/dark, got %d/%q", w.Code, w.Body.String())
	}
}

func TestDeepNestedGroupMiddlewareOrder(t *testing.T) {
	r := NewRouter()
	var order []string

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "root")
			next.ServeHTTP(w, r)
		})
	})

	api := r.Group("/api")
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "api")
			next.ServeHTTP(w, r)
		})
	})

	v1 := api.Group("/v1")
	v1.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "v1")
			next.ServeHTTP(w, r)
		})
	})

	users := v1.Group("/users")
	users.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "users")
			next.ServeHTTP(w, r)
		})
	})

	users.GetFunc("/:id", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.Write([]byte("ok"))
	})

	order = []string{}
	req := httptest.NewRequest("GET", "/api/v1/users/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	expected := []string{"root", "api", "v1", "users", "handler"}
	if !slicesEqual(order, expected) {
		t.Fatalf("expected middleware order %v, got %v", expected, order)
	}
}

func TestSiblingGroupMiddlewareIsolation(t *testing.T) {
	r := NewRouter()

	admin := r.Group("/admin")
	admin.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Admin", "true")
			next.ServeHTTP(w, r)
		})
	})
	admin.GetFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("admin"))
	})

	public := r.Group("/public")
	public.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Public", "true")
			next.ServeHTTP(w, r)
		})
	})
	public.GetFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("public"))
	})

	// Admin group should have X-Admin but not X-Public
	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Admin") != "true" {
		t.Error("expected X-Admin header on admin route")
	}
	if w.Header().Get("X-Public") != "" {
		t.Error("admin route should not have X-Public header")
	}

	// Public group should have X-Public but not X-Admin
	req = httptest.NewRequest("GET", "/public/page", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Public") != "true" {
		t.Error("expected X-Public header on public route")
	}
	if w.Header().Get("X-Admin") != "" {
		t.Error("public route should not have X-Admin header")
	}
}

func TestGroupWithPathParams(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api/v1")
	users := api.Group("/users")
	users.GetFunc("/:id/posts/:postID", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		postID, _ := contract.Param(r, "postID")
		w.Write([]byte(id + ":" + postID))
	})

	req := httptest.NewRequest("GET", "/api/v1/users/42/posts/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.String() != "42:99" {
		t.Errorf("expected 200/42:99, got %d/%q", w.Code, w.Body.String())
	}
}

func TestGroupRootHandler(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api")
	// Register handler at the group root (empty path)
	api.GetFunc("", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api-root"))
	})
	api.GetFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api-info"))
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/api", "api-root"},
		{"/api/info", "api-info"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Body.String() != tt.expected {
			t.Errorf("[GET %s] expected %q, got %q", tt.path, tt.expected, w.Body.String())
		}
	}
}

func TestGroupMultipleMethods(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api")
	api.GetFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("list"))
	})
	api.PostFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("create"))
	})
	api.PutFunc("/items/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("update"))
	})
	api.DeleteFunc("/items/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("delete"))
	})

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/api/items", "list"},
		{"POST", "/api/items", "create"},
		{"PUT", "/api/items/1", "update"},
		{"DELETE", "/api/items/1", "delete"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Body.String() != tt.expected {
			t.Errorf("[%s %s] expected %q, got %q", tt.method, tt.path, tt.expected, w.Body.String())
		}
	}
}

func TestGroupFunc(t *testing.T) {
	r := NewRouter()

	r.GroupFunc("/api/v1", func(v1 *Router) {
		v1.GetFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

		v1.GroupFunc("/users", func(users *Router) {
			users.GetFunc("/:id", func(w http.ResponseWriter, r *http.Request) {
				id, _ := contract.Param(r, "id")
				w.Write([]byte("user-" + id))
			})
		})
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/status", "ok"},
		{"/api/v1/users/5", "user-5"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK || w.Body.String() != tt.expected {
			t.Errorf("[GET %s] expected 200/%q, got %d/%q",
				tt.path, tt.expected, w.Code, w.Body.String())
		}
	}
}

func TestGroupFuncWithMiddleware(t *testing.T) {
	r := NewRouter()
	var order []string

	r.GroupFunc("/api", func(api *Router) {
		api.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "api-mw")
				next.ServeHTTP(w, r)
			})
		})

		api.GroupFunc("/v1", func(v1 *Router) {
			v1.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, "v1-mw")
					next.ServeHTTP(w, r)
				})
			})

			v1.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "handler")
				w.Write([]byte("pong"))
			})
		})
	})

	order = []string{}
	req := httptest.NewRequest("GET", "/api/v1/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	expected := []string{"api-mw", "v1-mw", "handler"}
	if !slicesEqual(order, expected) {
		t.Fatalf("expected middleware order %v, got %v", expected, order)
	}
}

func TestGroupFuncReturnsGroup(t *testing.T) {
	r := NewRouter()

	v1 := r.GroupFunc("/api/v1", func(v1 *Router) {
		v1.GetFunc("/inside", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("inside"))
		})
	})

	// Can still add routes after GroupFunc returns
	v1.GetFunc("/outside", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("outside"))
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/inside", "inside"},
		{"/api/v1/outside", "outside"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Body.String() != tt.expected {
			t.Errorf("[GET %s] expected %q, got %q", tt.path, tt.expected, w.Body.String())
		}
	}
}

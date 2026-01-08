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

func TestRouterFreeze(t *testing.T) {
	r := NewRouter()

	r.Get("/ping", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	r.Freeze()

	err := r.AddRoute("GET", "/panic", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	if err == nil {
		t.Errorf("expected error when adding route after freeze")
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

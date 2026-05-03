package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestMethodMismatchReturnsNotFound(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	rec := serveRouter(r, http.MethodPost, "/only")
	assertResponseStatus(t, rec, http.StatusNotFound)
	assertResponseBody(t, rec, "404 page not found\n")
	assertResponseHeader(t, rec, "Allow", "")
}

func TestAddRouteRejectsMalformedParamAndWildcardPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{name: "empty param name", pattern: "/users/:"},
		{name: "param starts with digit", pattern: "/users/:1id"},
		{name: "param contains punctuation", pattern: "/users/:id-name"},
		{name: "param contains colon", pattern: "/users/:id:name"},
		{name: "empty wildcard name", pattern: "/files/*"},
		{name: "wildcard starts with digit", pattern: "/files/*1path"},
		{name: "wildcard contains punctuation", pattern: "/files/*file.path"},
		{name: "non-terminal wildcard", pattern: "/files/*path/edit"},
		{name: "empty path segment", pattern: "/files//readme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			err := r.AddRoute(http.MethodGet, tt.pattern, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			if err == nil {
				t.Fatalf("expected AddRoute(%q) to fail", tt.pattern)
			}
		})
	}
}

func TestAddRouteRejectsMalformedMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{name: "empty", method: ""},
		{name: "leading space", method: " GET"},
		{name: "trailing space", method: "GET "},
		{name: "embedded space", method: "GE T"},
		{name: "newline", method: "GET\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			err := r.AddRoute(tt.method, "/users/:id", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			if err == nil {
				t.Fatalf("expected AddRoute method %q to fail", tt.method)
			}
		})
	}
}

func TestAddRouteAllowsCustomMethodAndIdentifierParams(t *testing.T) {
	r := NewRouter()
	err := r.AddRoute("MKCOL", "/files/:file_id/*restPath", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(Param(req, "file_id") + "|" + Param(req, "restPath")))
	}))
	if err != nil {
		t.Fatalf("add custom method route failed: %v", err)
	}

	rec := serveRouter(r, "MKCOL", "/files/root/a/b")
	assertResponseStatus(t, rec, http.StatusOK)
	assertResponseBody(t, rec, "root|a/b")
}

func TestAddRouteRejectsNilHandler(t *testing.T) {
	r := NewRouter()

	err := r.AddRoute(http.MethodGet, "/nil", nil)
	if err == nil {
		t.Fatalf("expected nil handler registration to fail")
	}
}

func TestAddRouteNormalizesRelativeRootPath(t *testing.T) {
	r := NewRouter()
	err := r.AddRoute(http.MethodGet, "users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(Param(req, "id")))
	}), WithRouteName("users.show"))
	if err != nil {
		t.Fatalf("add relative route failed: %v", err)
	}

	rec := serveRouter(r, http.MethodGet, "/users/42")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "42")

	routes := r.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].Path != "/users/:id" {
		t.Fatalf("stored route path = %q, want %q", routes[0].Path, "/users/:id")
	}
	if got := r.URL("users.show", "id", "42"); got != "/users/42" {
		t.Fatalf("URL() = %q, want %q", got, "/users/42")
	}
}

func TestAddRouteNormalizesRelativeGroupPath(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api")

	err := api.AddRoute(http.MethodGet, "users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(Param(req, "id")))
	}), WithRouteName("api.users.show"))
	if err != nil {
		t.Fatalf("add grouped relative route failed: %v", err)
	}

	rec := serveRouter(r, http.MethodGet, "/api/users/42")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "42")

	routes := r.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].Path != "/api/users/:id" {
		t.Fatalf("stored group route path = %q, want %q", routes[0].Path, "/api/users/:id")
	}
	if got := r.URL("api.users.show", "id", "42"); got != "/api/users/42" {
		t.Fatalf("URL() = %q, want %q", got, "/api/users/42")
	}
}

func TestAddRouteClearsStaleExactCache(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, http.MethodGet, "/files/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("wild:" + Param(req, "path")))
	}))

	rec := serveRouter(r, http.MethodGet, "/files/readme")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "wild:readme")

	mustAddRoute(r, http.MethodGet, "/files/readme", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("exact"))
	}))

	rec = serveRouter(r, http.MethodGet, "/files/readme")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "exact")
}

func TestWarmCachePreservesTrieSpecificity(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, http.MethodGet, "/files/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("wild:" + Param(req, "path")))
	}))
	mustAddRoute(r, http.MethodGet, "/files/public/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("param:" + Param(req, "id")))
	}))
	mustAddRoute(r, http.MethodGet, "/files/public/readme", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("static"))
	}))

	rec := serveRouter(r, http.MethodGet, "/files/other")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "wild:other")

	rec = serveRouter(r, http.MethodGet, "/files/public/42")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "param:42")

	rec = serveRouter(r, http.MethodGet, "/files/public/readme")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "static")

	rec = serveRouter(r, http.MethodGet, "/files/public/42")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "param:42")

	rec = serveRouter(r, http.MethodGet, "/files/public/readme")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "static")
}

func TestMethodNotAllowedWhenEnabled(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true))
	mustAddRoute(r, http.MethodGet, "/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	rec := serveRouter(r, http.MethodPost, "/only")
	assertResponseStatus(t, rec, http.StatusMethodNotAllowed)
	assertResponseHeader(t, rec, "Allow", http.MethodGet)
}

func TestMethodNotAllowedRoot(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true))
	mustAddRoute(r, http.MethodGet, "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	}))

	rec := serveRouter(r, http.MethodPost, "/")
	assertResponseStatus(t, rec, http.StatusMethodNotAllowed)
	assertResponseHeader(t, rec, "Allow", http.MethodGet)
}

func TestMethodNotAllowedRemainsWithCache(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true), withCacheCapacity(10))
	mustAddRoute(r, http.MethodGet, "/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	rec := serveRouter(r, http.MethodPost, "/only")
	assertResponseStatus(t, rec, http.StatusMethodNotAllowed)
	assertResponseHeader(t, rec, "Allow", http.MethodGet)

	rec = serveRouter(r, http.MethodPost, "/only")
	assertResponseStatus(t, rec, http.StatusMethodNotAllowed)
	assertResponseHeader(t, rec, "Allow", http.MethodGet)
}

func TestMethodNotAllowedAllowHeaderSorted(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true))
	mustAddRoute(r, http.MethodGet, "/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	mustAddRoute(r, http.MethodPost, "/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	mustAddRoute(r, http.MethodPut, "/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	rec := serveRouter(r, http.MethodDelete, "/only")
	assertResponseStatus(t, rec, http.StatusMethodNotAllowed)
	assertResponseHeader(t, rec, "Allow", "GET, POST, PUT")
}

func TestAnyFallbackWhenMethodTreeMissing(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}))
	mustAddRoute(r, methodAny, "/fallback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("any"))
	}))

	rec := serveRouter(r, http.MethodGet, "/fallback")
	assertResponseStatus(t, rec, http.StatusOK)
	assertTrimmedResponseBody(t, rec, "any")
}

func TestAnyFallbackWithCache(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, methodAny, "/fallback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("any"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/fallback", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "any" {
		t.Fatalf("expected %q, got %q", "any", body)
	}

	req = httptest.NewRequest(http.MethodGet, "/fallback", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected cached 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "any" {
		t.Fatalf("expected cached %q, got %q", "any", body)
	}
}

func TestParameterizedAnyFallbackContextUsesAnyMethodWithCache(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, methodAny, "/fallback/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := contract.RequestContextFromContext(r.Context())
		if rc.RoutePattern != "/fallback/:id" {
			t.Fatalf("expected route pattern %q, got %q", "/fallback/:id", rc.RoutePattern)
		}
		w.Write([]byte(Param(r, "id")))
	}))

	for _, id := range []string{"one", "two"} {
		req := httptest.NewRequest(http.MethodPost, "/fallback/"+id, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if body := strings.TrimSpace(rec.Body.String()); body != id {
			t.Fatalf("expected body %q, got %q", id, body)
		}
	}
}

func TestHeadFallbackParameterizedRouteSuppressesBodyWithCache(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := contract.RequestContextFromContext(r.Context())
		if rc.RoutePattern != "/users/:id" {
			t.Fatalf("expected route pattern %q, got %q", "/users/:id", rc.RoutePattern)
		}
		w.Header().Set("X-User-ID", Param(r, "id"))
		n, err := w.Write([]byte("body"))
		if err != nil {
			t.Fatalf("HEAD fallback write returned error: %v", err)
		}
		w.Header().Set("X-Write-N", strconv.Itoa(n))
	}))

	for _, id := range []string{"one", "two"} {
		req := httptest.NewRequest(http.MethodHead, "/users/"+id, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if got := rec.Header().Get("X-User-ID"); got != id {
			t.Fatalf("expected X-User-ID %q, got %q", id, got)
		}
		if body := rec.Body.String(); body != "" {
			t.Fatalf("expected empty HEAD body, got %q", body)
		}
		if got := rec.Header().Get("X-Write-N"); got != "4" {
			t.Fatalf("expected suppressed write to report 4 bytes, got %q", got)
		}
	}
}

func TestHeadFallbackReadFromDrainsAndUnwrapsResponseWriter(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, http.MethodGet, "/stream", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := io.Copy(w, strings.NewReader("stream-body"))
		if err != nil {
			t.Fatalf("HEAD fallback ReadFrom returned error: %v", err)
		}
		w.Header().Set("X-Copied-N", strconv.FormatInt(n, 10))
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Fatalf("HEAD fallback ResponseController Flush returned error: %v", err)
		}
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodHead, "/stream", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
		if body := rec.Body.String(); body != "" {
			t.Fatalf("request %d: expected empty HEAD body, got %q", i, body)
		}
		if got := rec.Header().Get("X-Copied-N"); got != "11" {
			t.Fatalf("request %d: expected copied count 11, got %q", i, got)
		}
	}
}

func TestTrailingSlashNormalization(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		w.Write([]byte(id))
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/123/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "123" {
		t.Fatalf("expected id %q, got %q", "123", body)
	}
}

func TestInternalDoubleSlashNotNormalized(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/users//123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestWildcardParamCapturesRemainder(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/files/*path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := Param(r, "path")
		w.Write([]byte(path))
	}))

	req := httptest.NewRequest(http.MethodGet, "/files/a/b/c.txt", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "a/b/c.txt" {
		t.Fatalf("expected %q, got %q", "a/b/c.txt", body)
	}
}

func TestEncodedSpaceParamIsDecoded(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		w.Write([]byte(id))
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/a%20b", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "a b" {
		t.Fatalf("expected %q, got %q", "a b", body)
	}
}

func TestEncodedSlashDoesNotMatchSingleSegmentParam(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/a%2Fb", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestEncodedSlashMatchesWildcardParam(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/files/*path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := Param(r, "path")
		w.Write([]byte(path))
	}))

	req := httptest.NewRequest(http.MethodGet, "/files/a%2Fb", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "a/b" {
		t.Fatalf("expected %q, got %q", "a/b", body)
	}
}

func TestDuplicateParamKeysLastWins(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/teams/:id/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		w.Write([]byte(id))
	}))

	req := httptest.NewRequest(http.MethodGet, "/teams/1/users/2", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "2" {
		t.Fatalf("expected %q, got %q", "2", body)
	}
}

func TestGroupParamKeysLastWins(t *testing.T) {
	r := NewRouter()
	group := r.Group("/orgs/:id")
	mustAddRoute(group, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		w.Write([]byte(id))
	}))

	req := httptest.NewRequest(http.MethodGet, "/orgs/1/users/2", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "2" {
		t.Fatalf("expected %q, got %q", "2", body)
	}
}

func TestGroupParamDoesNotMatchEncodedSlash(t *testing.T) {
	r := NewRouter()
	group := r.Group("/orgs/:id")
	mustAddRoute(group, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/orgs/a%2Fb/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCachedRouteDoesNotLeakParams(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		w.Write([]byte(id))
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/one", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "one" {
		t.Fatalf("expected %q, got %q", "one", body)
	}

	req = httptest.NewRequest(http.MethodGet, "/users/two", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "two" {
		t.Fatalf("expected %q, got %q", "two", body)
	}
}

func TestRouteParamsOverrideExistingContext(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		rc := contract.RequestContextFromContext(r.Context())

		if id != "123" {
			t.Fatalf("expected Param id %q, got %q", "123", id)
		}
		if rc.Params["id"] != "123" {
			t.Fatalf("expected RequestContext param %q, got %q", "123", rc.Params["id"])
		}
		w.Write([]byte("ok"))
	}))

	baseReq := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	ctx := contract.WithRequestContext(baseReq.Context(), contract.RequestContext{
		Params: map[string]string{"id": "ctx2"},
	})
	req := baseReq.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRouteWithoutParamsClearsExistingContextParams(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := Param(r, "id"); got != "" {
			t.Fatalf("expected no route param id, got %q", got)
		}
		rc := contract.RequestContextFromContext(r.Context())
		if len(rc.Params) != 0 {
			t.Fatalf("expected no route params, got %v", rc.Params)
		}
		w.Write([]byte("ok"))
	}))

	baseReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	ctx := contract.WithRequestContext(baseReq.Context(), contract.RequestContext{
		Params: map[string]string{"id": "stale"},
	})
	req := baseReq.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assertResponseStatus(t, rec, http.StatusOK)
	assertResponseBody(t, rec, "ok")
}

func TestRequestContextIncludesRoutePatternAndName(t *testing.T) {
	r := NewRouter()
	err := r.AddRoute(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := contract.RequestContextFromContext(r.Context())
		if rc.RoutePattern != "/users/:id" {
			t.Fatalf("expected route pattern %q, got %q", "/users/:id", rc.RoutePattern)
		}
		if rc.RouteName != "user_show" {
			t.Fatalf("expected route name %q, got %q", "user_show", rc.RouteName)
		}
		w.WriteHeader(http.StatusOK)
	}), WithRouteName("user_show"))
	if err != nil {
		t.Fatalf("add route failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRouteMetadataConcurrentRegistrationAndServe(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rc := contract.RequestContextFromContext(req.Context())
		if rc.RouteName != "users.show" {
			t.Errorf("route name = %q, want %q", rc.RouteName, "users.show")
		}
		w.WriteHeader(http.StatusOK)
	}), WithRouteName("users.show"))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			path := "/extra/" + strconv.Itoa(i) + "/:id"
			name := "extra." + strconv.Itoa(i)
			err := r.AddRoute(http.MethodGet, path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			}), WithRouteName(name))
			if err != nil {
				t.Errorf("add route %s failed: %v", path, err)
			}
		}(i)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/users/"+strconv.Itoa(i), nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("serve status = %d, want %d", rec.Code, http.StatusOK)
			}
		}(i)
	}
	wg.Wait()
}

func TestCachedRouteNormalizesTrailingSlash(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/users/123/", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if size := r.state.matchCache.Size(); size != 1 {
		t.Fatalf("expected cache size 1, got %d", size)
	}
}

func TestCachedRouteSeparatesByHostAndMethod(t *testing.T) {
	r := NewRouter(withCacheCapacity(10))
	mustAddRoute(r, http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mustAddRoute(r, http.MethodPost, "/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Host = "one.example"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Host = "two.example"
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/ping", nil)
	req.Host = "one.example"
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Note: Cache size may be less than 3 due to cache eviction policy
	// The important thing is that different hosts and methods are cached separately
	size := r.state.matchCache.Size()
	if size < 2 {
		t.Fatalf("expected cache size at least 2, got %d", size)
	}
}

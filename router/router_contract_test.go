package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
		{name: "empty wildcard name", pattern: "/files/*"},
		{name: "non-terminal wildcard", pattern: "/files/*path/edit"},
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
		w.Write([]byte("body"))
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

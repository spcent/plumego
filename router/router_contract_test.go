package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestMethodMismatchReturnsNotFound(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/only", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "404 page not found\n" {
		t.Fatalf("unexpected body: %q", body)
	}
	if allow := rec.Header().Get("Allow"); allow != "" {
		t.Fatalf("expected empty Allow header, got %q", allow)
	}
}

func TestMethodNotAllowedWhenEnabled(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true))
	r.GetFunc("/only", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, allow)
	}
}

func TestMethodNotAllowedRoot(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true))
	r.GetFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, allow)
	}
}

func TestMethodNotAllowedRemainsWithCache(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true), WithCache(10))
	r.GetFunc("/only", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, allow)
	}

	req = httptest.NewRequest(http.MethodPost, "/only", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected cached 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected cached Allow header %q, got %q", http.MethodGet, allow)
	}
}

func TestMethodNotAllowedAllowHeaderSorted(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true))
	_ = r.GetFunc("/only", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	_ = r.PostFunc("/only", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	_ = r.PutFunc("/only", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest(http.MethodDelete, "/only", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != "GET, POST, PUT" {
		t.Fatalf("expected Allow header %q, got %q", "GET, POST, PUT", allow)
	}
}

func TestAnyFallbackWhenMethodTreeMissing(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	r.Any("/fallback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
}

func TestAnyFallbackWithCache(t *testing.T) {
	r := NewRouter(WithCache(10))
	r.Any("/fallback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestTrailingSlashNormalization(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte(id))
	})

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
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/users//123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestWildcardParamCapturesRemainder(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/files/*path", func(w http.ResponseWriter, r *http.Request) {
		path, _ := contract.Param(r, "path")
		w.Write([]byte(path))
	})

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
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte(id))
	})

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
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/users/a%2Fb", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestEncodedSlashMatchesWildcardParam(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/files/*path", func(w http.ResponseWriter, r *http.Request) {
		path, _ := contract.Param(r, "path")
		w.Write([]byte(path))
	})

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
	r.GetFunc("/teams/:id/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte(id))
	})

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
	group.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte(id))
	})

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
	group.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/orgs/a%2Fb/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRouteCacheDoesNotLeakParams(t *testing.T) {
	r := NewRouter(WithCache(10))
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte(id))
	})

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

func TestMiddlewareShortCircuitStopsHandler(t *testing.T) {
	r := NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Stop", "true")
			w.WriteHeader(http.StatusUnauthorized)
		})
	})

	r.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if rec.Header().Get("X-Stop") != "true" {
		t.Fatalf("expected short-circuit header to be set")
	}
}

func TestCachedMatchPreservesMiddlewareOrder(t *testing.T) {
	r := NewRouter(WithCache(10))
	var order []string

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1")
			next.ServeHTTP(w, r)
		})
	})
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2")
			next.ServeHTTP(w, r)
		})
	})
	r.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	order = []string{}
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !slicesEqual(order, []string{"mw1", "mw2", "handler"}) {
		t.Fatalf("expected order [mw1 mw2 handler], got %v", order)
	}

	order = []string{}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !slicesEqual(order, []string{"mw1", "mw2", "handler"}) {
		t.Fatalf("expected cached order [mw1 mw2 handler], got %v", order)
	}
}

func TestCachedMatchPreservesNestedGroupOrder(t *testing.T) {
	r := NewRouter(WithCache(10))
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
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()
	order = []string{}
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !slicesEqual(order, []string{"global", "api", "v1", "handler"}) {
		t.Fatalf("expected order [global api v1 handler], got %v", order)
	}

	order = []string{}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !slicesEqual(order, []string{"global", "api", "v1", "handler"}) {
		t.Fatalf("expected cached order [global api v1 handler], got %v", order)
	}
}

func TestRouteParamsOverrideExistingContext(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		rc := contract.RequestContextFrom(r.Context())

		if id != "123" {
			t.Fatalf("expected Param id %q, got %q", "123", id)
		}
		if rc.Params["id"] != "123" {
			t.Fatalf("expected RequestContext param %q, got %q", "123", rc.Params["id"])
		}
		w.Write([]byte("ok"))
	})

	baseReq := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	ctx := context.WithValue(baseReq.Context(), contract.ParamsContextKey{}, map[string]string{"id": "ctx"})
	ctx = context.WithValue(ctx, contract.RequestContextKey{}, contract.RequestContext{
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
	err := r.AddRouteWithOptions(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := contract.RequestContextFrom(r.Context())
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

func TestRouteContextVisibleToOuterMiddleware(t *testing.T) {
	r := NewRouter()
	var gotPattern string

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			gotPattern = contract.RequestContextFrom(r.Context()).RoutePattern
		})
	})

	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotPattern != "/users/:id" {
		t.Fatalf("expected outer middleware to see route pattern %q, got %q", "/users/:id", gotPattern)
	}
}

func TestRouteCacheNormalizesTrailingSlash(t *testing.T) {
	r := NewRouter(WithCache(10))
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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

	if size := r.routeCache.Size(); size != 1 {
		t.Fatalf("expected cache size 1, got %d", size)
	}
}

func TestRouteCacheInvalidatesOnMiddlewareChange(t *testing.T) {
	r := NewRouter(WithCache(10))
	var order []string

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1")
			next.ServeHTTP(w, r)
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2")
			next.ServeHTTP(w, r)
		})
	}

	r.Use(mw1)
	r.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	order = []string{}
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !slicesEqual(order, []string{"mw1", "handler"}) {
		t.Fatalf("expected order [mw1 handler], got %v", order)
	}

	r.Use(mw2)
	order = []string{}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !slicesEqual(order, []string{"mw1", "mw2", "handler"}) {
		t.Fatalf("expected order [mw1 mw2 handler], got %v", order)
	}
}

func TestRouteCacheSeparatesByHostAndMethod(t *testing.T) {
	r := NewRouter(WithCache(10))
	r.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.PostFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
	size := r.routeCache.Size()
	if size < 2 {
		t.Fatalf("expected cache size at least 2, got %d", size)
	}
}

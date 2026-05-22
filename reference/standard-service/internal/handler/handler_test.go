package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"standard-service/internal/domain/item"
)

// readinessCheckerFunc is a test helper that adapts a plain function to ReadinessChecker.
type readinessCheckerFunc func(context.Context) error

func (f readinessCheckerFunc) Ready(ctx context.Context) error { return f(ctx) }

func TestHealthHandlerResponses(t *testing.T) {
	h := HealthHandler{ServiceName: "svc"}

	tests := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		status  string
		check   string
	}{
		{name: "live", handler: h.Live, status: "ok", check: "liveness"},
		{name: "ready_no_checkers", handler: h.Ready, status: "ready", check: "readiness"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			resp := decodeReferenceData[healthResponse](t, rec)
			if resp.Status != tt.status || resp.Service != "svc" || resp.Check != tt.check || resp.Timestamp == "" {
				t.Fatalf("unexpected health response: %+v", resp)
			}
		})
	}
}

func TestHealthHandlerReadyWithCheckers(t *testing.T) {
	t.Run("passing checker returns 200", func(t *testing.T) {
		h := HealthHandler{
			ServiceName: "svc",
			Checkers: []ReadinessChecker{
				readinessCheckerFunc(func(_ context.Context) error { return nil }),
			},
		}
		rec := httptest.NewRecorder()
		h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		resp := decodeReferenceData[healthResponse](t, rec)
		if resp.Status != "ready" {
			t.Fatalf("status = %q, want %q", resp.Status, "ready")
		}
	})

	t.Run("failing checker returns 503", func(t *testing.T) {
		h := HealthHandler{
			ServiceName: "svc",
			Checkers: []ReadinessChecker{
				readinessCheckerFunc(func(_ context.Context) error {
					return errors.New("db not connected")
				}),
			},
		}
		rec := httptest.NewRecorder()
		h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
	})

	t.Run("second checker failing returns 503", func(t *testing.T) {
		h := HealthHandler{
			ServiceName: "svc",
			Checkers: []ReadinessChecker{
				readinessCheckerFunc(func(_ context.Context) error { return nil }),
				readinessCheckerFunc(func(_ context.Context) error {
					return errors.New("cache not connected")
				}),
			},
		}
		rec := httptest.NewRecorder()
		h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestAPIHandlerResponses(t *testing.T) {
	h := APIHandler{}

	tests := []struct {
		name   string
		path   string
		fn     func(http.ResponseWriter, *http.Request)
		assert func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "root",
			path: "/",
			fn:   h.Root,
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
				}
				got := decodeReferenceData[rootResponse](t, rec)
				if got.Service != "plumego-reference" || got.Docs == "" || got.Version != version {
					t.Fatalf("unexpected root response: %+v", got)
				}
			},
		},
		{
			name: "hello",
			path: "/api/hello",
			fn:   h.Hello,
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
				}
				got := decodeReferenceData[helloResponse](t, rec)
				if got.Service != "plumego-reference" || got.Mode != "canonical" || got.Version != version {
					t.Fatalf("unexpected hello response: %+v", got)
				}
				// Endpoints must be a non-empty slice; verify the discovery entry exists.
				if len(got.Endpoints) == 0 {
					t.Fatal("hello: Endpoints must not be empty")
				}
				// Verify each entry carries a non-empty Method, Path, and Name.
				for i, ep := range got.Endpoints {
					if ep.Method == "" || ep.Path == "" || ep.Name == "" {
						t.Fatalf("hello: endpoint[%d] missing field: %+v", i, ep)
					}
				}
				// Verify a specific known entry is present with the correct method.
				found := false
				for _, ep := range got.Endpoints {
					if ep.Name == "api_hello" && ep.Method == http.MethodGet && ep.Path == "/api/hello" {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("hello: endpoint api_hello/GET//api/hello not found in %+v", got.Endpoints)
				}
			},
		},
		{
			name: "greet",
			path: "/api/v1/greet?name=Alice",
			fn:   h.Greet,
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
				}
				got := decodeReferenceData[greetResponse](t, rec)
				if got.Message != "hello, Alice" {
					t.Fatalf("unexpected greet response: %+v", got)
				}
			},
		},
		{
			name: "status",
			path: "/api/status",
			fn:   h.Status,
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
				}
				got := decodeReferenceData[statusResponse](t, rec)
				if got.Status != "healthy" || got.Structure.Routes != "one_method_one_path_one_handler" || len(got.Modules) == 0 {
					t.Fatalf("unexpected status response: %+v", got)
				}
				if got.Version != version {
					t.Fatalf("status.Version = %q, want %q", got.Version, version)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			tt.assert(t, rec)
		})
	}
}

func TestItemHandlerCreate(t *testing.T) {
	h := ItemHandler{Repo: item.NewMemoryStore()}

	t.Run("valid body returns 201 with item", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"widget"}`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		item := decodeReferenceData[item.Item](t, rec)
		if item.ID == "" || item.Name != "widget" || item.CreatedAt == "" {
			t.Fatalf("unexpected item: %+v", item)
		}
	})

	t.Run("missing name returns 400 TypeRequired", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":""}`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid JSON returns 400 TypeBadRequest", func(t *testing.T) {
		body := bytes.NewBufferString(`not-json`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestItemHandlerList(t *testing.T) {
	t.Run("empty store returns empty items with zero total", func(t *testing.T) {
		h := ItemHandler{Repo: item.NewMemoryStore()}
		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		resp := decodeReferenceData[listResponse](t, rec)
		if len(resp.Items) != 0 || resp.Total != 0 || resp.Limit != 20 || resp.Offset != 0 {
			t.Fatalf("unexpected list response: %+v", resp)
		}
	})

	t.Run("populated store returns items in creation order", func(t *testing.T) {
		store := item.NewMemoryStore()
		store.Create("alpha")
		store.Create("beta")
		h := ItemHandler{Repo: store}

		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		resp := decodeReferenceData[listResponse](t, rec)
		if resp.Total != 2 || len(resp.Items) != 2 {
			t.Fatalf("total=%d items=%d, want 2/2", resp.Total, len(resp.Items))
		}
		// Verify stable creation order.
		if resp.Items[0].Name != "alpha" || resp.Items[1].Name != "beta" {
			t.Fatalf("unexpected order: [%s, %s], want [alpha, beta]",
				resp.Items[0].Name, resp.Items[1].Name)
		}
	})

	t.Run("limit param restricts returned items", func(t *testing.T) {
		store := item.NewMemoryStore()
		for _, name := range []string{"a", "b", "c", "d", "e"} {
			store.Create(name)
		}
		h := ItemHandler{Repo: store}

		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items?limit=3", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		resp := decodeReferenceData[listResponse](t, rec)
		if resp.Total != 5 || len(resp.Items) != 3 || resp.Limit != 3 {
			t.Fatalf("total=%d items=%d limit=%d, want 5/3/3", resp.Total, len(resp.Items), resp.Limit)
		}
		// First page must be the first 3 in creation order.
		if resp.Items[0].Name != "a" || resp.Items[2].Name != "c" {
			t.Fatalf("unexpected page items: %v", resp.Items)
		}
	})

	t.Run("offset param skips items in creation order", func(t *testing.T) {
		store := item.NewMemoryStore()
		for _, name := range []string{"a", "b", "c", "d", "e"} {
			store.Create(name)
		}
		h := ItemHandler{Repo: store}

		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items?offset=3", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		resp := decodeReferenceData[listResponse](t, rec)
		if resp.Total != 5 || len(resp.Items) != 2 || resp.Offset != 3 {
			t.Fatalf("total=%d items=%d offset=%d, want 5/2/3", resp.Total, len(resp.Items), resp.Offset)
		}
		// Items at offset 3 are "d" and "e".
		if resp.Items[0].Name != "d" || resp.Items[1].Name != "e" {
			t.Fatalf("unexpected offset items: %v", resp.Items)
		}
	})
}

func TestItemHandlerGetByID(t *testing.T) {
	store := item.NewMemoryStore()
	created := store.Create("gadget")
	h := ItemHandler{Repo: store}

	t.Run("existing id returns 200 with item", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+created.ID, nil)
		// Inject the path parameter the way the router does at runtime.
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.GetByID(rec, req.WithContext(ctx))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		item := decodeReferenceData[item.Item](t, rec)
		if item.ID != created.ID || item.Name != "gadget" {
			t.Fatalf("unexpected item: %+v", item)
		}
	})

	t.Run("missing id returns 404 TypeNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/no-such-item", nil)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": "no-such-item"},
		})
		rec := httptest.NewRecorder()
		h.GetByID(rec, req.WithContext(ctx))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestItemHandlerUpdate(t *testing.T) {
	t.Run("existing id returns 200 with updated item", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create("original")
		h := ItemHandler{Repo: store}

		body := bytes.NewBufferString(`{"name":"renamed"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+created.ID, body)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.Update(rec, req.WithContext(ctx))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		got := decodeReferenceData[item.Item](t, rec)
		if got.ID != created.ID || got.Name != "renamed" || got.CreatedAt != created.CreatedAt {
			t.Fatalf("unexpected updated item: %+v", got)
		}
	})

	t.Run("missing id returns 404 TypeNotFound", func(t *testing.T) {
		h := ItemHandler{Repo: item.NewMemoryStore()}
		body := bytes.NewBufferString(`{"name":"anything"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/no-such-item", body)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": "no-such-item"},
		})
		rec := httptest.NewRecorder()
		h.Update(rec, req.WithContext(ctx))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("missing name returns 400 TypeRequired", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create("original")
		h := ItemHandler{Repo: store}

		body := bytes.NewBufferString(`{"name":""}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+created.ID, body)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.Update(rec, req.WithContext(ctx))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid JSON returns 400 TypeBadRequest", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create("original")
		h := ItemHandler{Repo: store}

		body := bytes.NewBufferString(`not-json`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+created.ID, body)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.Update(rec, req.WithContext(ctx))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("put is idempotent", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create("original")
		h := ItemHandler{Repo: store}

		for i := 0; i < 3; i++ {
			body := bytes.NewBufferString(`{"name":"stable"}`)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+created.ID, body)
			ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
				Params: map[string]string{"id": created.ID},
			})
			rec := httptest.NewRecorder()
			h.Update(rec, req.WithContext(ctx))
			if rec.Code != http.StatusOK {
				t.Fatalf("attempt %d: status = %d, want %d", i+1, rec.Code, http.StatusOK)
			}
			got := decodeReferenceData[item.Item](t, rec)
			if got.Name != "stable" {
				t.Fatalf("attempt %d: name = %q, want stable", i+1, got.Name)
			}
		}
	})
}

func TestItemHandlerDelete(t *testing.T) {
	store := item.NewMemoryStore()
	created := store.Create("gadget")
	h := ItemHandler{Repo: store}

	t.Run("missing id returns 404 TypeNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/no-such-item", nil)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": "no-such-item"},
		})
		rec := httptest.NewRecorder()
		h.Delete(rec, req.WithContext(ctx))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("existing id returns 204 no content", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+created.ID, nil)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.Delete(rec, req.WithContext(ctx))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("expected empty body, got %q", rec.Body.String())
		}
	})

	t.Run("already deleted id returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+created.ID, nil)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.Delete(rec, req.WithContext(ctx))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestRequireWriteKey(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("empty key is a no-op — all requests pass", func(t *testing.T) {
		mw := RequireWriteKey("")(okHandler)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("correct key passes through to handler", func(t *testing.T) {
		mw := RequireWriteKey("secret")(okHandler)
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(WriteKeyHeader, "secret")
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("missing header returns 401", func(t *testing.T) {
		mw := RequireWriteKey("secret")(okHandler)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("wrong key returns 401", func(t *testing.T) {
		mw := RequireWriteKey("secret")(okHandler)
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(WriteKeyHeader, "wrong")
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

func decodeReferenceData[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != contract.ContentTypeJSON {
		t.Fatalf("content type = %q, want %q", got, contract.ContentTypeJSON)
	}

	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode success envelope: %v", err)
	}
	if len(env.Data) == 0 {
		t.Fatal("success envelope missing data")
	}

	var body T
	if err := json.Unmarshal(env.Data, &body); err != nil {
		t.Fatalf("decode success data: %v", err)
	}
	return body
}

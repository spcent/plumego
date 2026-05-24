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
	"github.com/spcent/plumego/health"
	plumelog "github.com/spcent/plumego/log"
	"standard-service/internal/domain/item"
)

// discardLogger returns a StructuredLogger that silently discards all output.
// Use it in tests that construct an APIHandler to satisfy the non-nil Logger requirement.
func discardLogger() plumelog.StructuredLogger {
	return plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatDiscard})
}

// componentChecker is a test helper that adapts a name and a function to
// health.ComponentChecker. This mirrors what production code does: a small
// struct wrapper that delegates Check() to an existing client method.
type componentChecker struct {
	name  string
	check func(context.Context) error
}

func (c componentChecker) Name() string                    { return c.name }
func (c componentChecker) Check(ctx context.Context) error { return c.check(ctx) }

// responseEnvelope holds the full success response structure for test decoding.
type responseEnvelope struct {
	Data      json.RawMessage `json:"data"`
	Meta      map[string]any  `json:"meta"`
	RequestID string          `json:"request_id"`
}

// decodeEnvelope decodes the full response envelope from rec and validates Content-Type.
func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) responseEnvelope {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != contract.ContentTypeJSON {
		t.Fatalf("content type = %q, want %q", got, contract.ContentTypeJSON)
	}
	var env responseEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	return env
}

// decodeReferenceData decodes the data field of a success envelope into T.
func decodeReferenceData[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	env := decodeEnvelope(t, rec)
	if len(env.Data) == 0 {
		t.Fatal("success envelope missing data")
	}
	var body T
	if err := json.Unmarshal(env.Data, &body); err != nil {
		t.Fatalf("decode data field: %v", err)
	}
	return body
}

// errorFields holds the code and details from an error response.
// Use decodeErrorPayload to read both in a single pass.
type errorFields struct {
	Code    string         `json:"code"`
	Details map[string]any `json:"details"`
}

// decodeErrorPayload decodes the error envelope's code and details map.
// It reads rec.Body.Bytes() directly so it is safe to call multiple times
// on the same recorder without the reader position advancing.
func decodeErrorPayload(t *testing.T, rec *httptest.ResponseRecorder) errorFields {
	t.Helper()
	var env struct {
		Error errorFields `json:"error"`
	}
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&env); err != nil {
		t.Fatalf("decode error payload: %v", err)
	}
	return env.Error
}

func TestHealthHandlerLive(t *testing.T) {
	h := HealthHandler{ServiceName: "svc"}
	rec := httptest.NewRecorder()
	h.Live(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	resp := decodeReferenceData[livenessResponse](t, rec)
	if resp.Status != "ok" || resp.Service != "svc" || resp.Timestamp == "" {
		t.Fatalf("unexpected liveness response: %+v", resp)
	}
}

func TestHealthHandlerReadyNoCheckers(t *testing.T) {
	h := HealthHandler{ServiceName: "svc"}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	resp := decodeReferenceData[health.ReadinessStatus](t, rec)
	if !resp.Ready {
		t.Fatalf("Ready = false, want true (no checkers registered)")
	}
	if resp.Timestamp.IsZero() {
		t.Fatal("Timestamp must not be zero")
	}
	if len(resp.Components) != 0 {
		t.Fatalf("Components = %v, want empty (no checkers)", resp.Components)
	}
}

func TestHealthHandlerReadyWithCheckers(t *testing.T) {
	t.Run("passing checker returns 200 with component map", func(t *testing.T) {
		h := HealthHandler{
			ServiceName: "svc",
			Checkers: []health.ComponentChecker{
				componentChecker{name: "database", check: func(_ context.Context) error { return nil }},
			},
		}
		rec := httptest.NewRecorder()
		h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		resp := decodeReferenceData[health.ReadinessStatus](t, rec)
		if !resp.Ready {
			t.Fatalf("Ready = false, want true")
		}
		if !resp.Components["database"] {
			t.Fatalf("Components[database] = false, want true; got %v", resp.Components)
		}
	})

	t.Run("failing checker returns 503 with component name as detail key", func(t *testing.T) {
		h := HealthHandler{
			ServiceName: "svc",
			Checkers: []health.ComponentChecker{
				componentChecker{name: "database", check: func(_ context.Context) error {
					return errors.New("connection refused")
				}},
			},
		}
		rec := httptest.NewRecorder()
		h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
		// The failing component name is the detail key; its error message is the value.
		// This lets operators see exactly which component failed without inspecting logs.
		ef := decodeErrorPayload(t, rec)
		if ef.Details["database"] != "connection refused" {
			t.Fatalf("error detail database = %v, want connection refused", ef.Details["database"])
		}
	})

	t.Run("all checkers are probed even if an earlier one fails", func(t *testing.T) {
		h := HealthHandler{
			ServiceName: "svc",
			Checkers: []health.ComponentChecker{
				componentChecker{name: "database", check: func(_ context.Context) error { return nil }},
				componentChecker{name: "cache", check: func(_ context.Context) error {
					return errors.New("cache offline")
				}},
			},
		}
		rec := httptest.NewRecorder()
		h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
		ef := decodeErrorPayload(t, rec)
		// "cache" failed — its detail must appear.
		if ef.Details["cache"] != "cache offline" {
			t.Fatalf("error detail cache = %v, want cache offline", ef.Details["cache"])
		}
		// "database" passed — it must not appear as a failure detail.
		if _, ok := ef.Details["database"]; ok {
			t.Fatalf("error details should not include passing component database: %v", ef.Details)
		}
	})

	t.Run("multiple failing checkers all appear in error details", func(t *testing.T) {
		h := HealthHandler{
			ServiceName: "svc",
			Checkers: []health.ComponentChecker{
				componentChecker{name: "database", check: func(_ context.Context) error {
					return errors.New("connection refused")
				}},
				componentChecker{name: "cache", check: func(_ context.Context) error {
					return errors.New("cache offline")
				}},
			},
		}
		rec := httptest.NewRecorder()
		h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
		ef := decodeErrorPayload(t, rec)
		if ef.Details["database"] != "connection refused" {
			t.Fatalf("error detail database = %v, want connection refused", ef.Details["database"])
		}
		if ef.Details["cache"] != "cache offline" {
			t.Fatalf("error detail cache = %v, want cache offline", ef.Details["cache"])
		}
	})

	t.Run("all checkers passing returns component map with all true", func(t *testing.T) {
		h := HealthHandler{
			ServiceName: "svc",
			Checkers: []health.ComponentChecker{
				componentChecker{name: "database", check: func(_ context.Context) error { return nil }},
				componentChecker{name: "cache", check: func(_ context.Context) error { return nil }},
			},
		}
		rec := httptest.NewRecorder()
		h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		resp := decodeReferenceData[health.ReadinessStatus](t, rec)
		if !resp.Ready {
			t.Fatalf("Ready = false, want true")
		}
		for _, name := range []string{"database", "cache"} {
			if !resp.Components[name] {
				t.Fatalf("Components[%s] = false, want true; full map: %v", name, resp.Components)
			}
		}
	})
}

func TestAPIHandlerResponses(t *testing.T) {
	h := APIHandler{Logger: discardLogger(), ServiceName: "test-service", Version: "test-version"}

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
				if got.Service != "test-service" || got.Docs == "" || got.Version != "test-version" {
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
				if got.Service != "test-service" || got.Mode != "canonical" || got.Version != "test-version" {
					t.Fatalf("unexpected hello response: %+v", got)
				}
				// Endpoints must be a non-empty slice with all required fields.
				if len(got.Endpoints) == 0 {
					t.Fatal("hello: Endpoints must not be empty")
				}
				for i, ep := range got.Endpoints {
					if ep.Name == "" || ep.Method == "" || ep.Path == "" {
						t.Fatalf("hello: endpoint[%d] missing field: %+v", i, ep)
					}
				}
				// Verify a well-known entry is present with the correct name, method, and path.
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
			name: "greet missing name",
			path: "/api/v1/greet",
			fn:   h.Greet,
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
				}
				if code := decodeErrorPayload(t, rec).Code; code != "greet.name.required" {
					t.Fatalf("error code = %q, want %q", code, "greet.name.required")
				}
			},
		},
		{
			name: "info",
			path: "/api/info",
			fn:   h.Info,
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
				}
				got := decodeReferenceData[infoResponse](t, rec)
				if got.Mode != "canonical" || len(got.Modules) == 0 {
					t.Fatalf("unexpected info response: %+v", got)
				}
				if got.Version != "test-version" {
					t.Fatalf("info.Version = %q, want %q", got.Version, "test-version")
				}
				if got.Service != "test-service" {
					t.Fatalf("info.Service = %q, want %q", got.Service, "test-service")
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
		body := bytes.NewBufferString(`{"name":"widget","description":"a widget"}`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		got := decodeReferenceData[item.Item](t, rec)
		if got.ID == "" || got.Name != "widget" || got.Description != "a widget" || got.CreatedAt.IsZero() {
			t.Fatalf("unexpected item: %+v", got)
		}
	})

	t.Run("both name and description missing returns 400 with both field details", func(t *testing.T) {
		body := bytes.NewBufferString(`{}`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		ef := decodeErrorPayload(t, rec)
		if ef.Code != codeItemFieldsRequired {
			t.Fatalf("error code = %q, want %q", ef.Code, codeItemFieldsRequired)
		}
		// Both field names must appear as keys in the details map.
		if _, ok := ef.Details["name"]; !ok {
			t.Error("error details missing 'name' key")
		}
		if _, ok := ef.Details["description"]; !ok {
			t.Error("error details missing 'description' key")
		}
	})

	t.Run("missing name only returns 400 with name detail", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"","description":"a widget"}`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		ef := decodeErrorPayload(t, rec)
		if ef.Code != codeItemFieldsRequired {
			t.Fatalf("error code = %q, want %q", ef.Code, codeItemFieldsRequired)
		}
		if _, ok := ef.Details["name"]; !ok {
			t.Error("error details missing 'name' key")
		}
		if _, ok := ef.Details["description"]; ok {
			t.Error("error details should not contain 'description' when description is present")
		}
	})

	t.Run("missing description only returns 400 with description detail", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"widget"}`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		ef := decodeErrorPayload(t, rec)
		if ef.Code != codeItemFieldsRequired {
			t.Fatalf("error code = %q, want %q", ef.Code, codeItemFieldsRequired)
		}
		if _, ok := ef.Details["description"]; !ok {
			t.Error("error details missing 'description' key")
		}
		if _, ok := ef.Details["name"]; ok {
			t.Error("error details should not contain 'name' when name is present")
		}
	})

	t.Run("empty body returns 400 TypeRequired with body_required code", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", http.NoBody))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if code := decodeErrorPayload(t, rec).Code; code != codeItemCreateBodyRequired {
			t.Fatalf("error code = %q, want %q", code, codeItemCreateBodyRequired)
		}
	})

	t.Run("invalid JSON returns 400 TypeBadRequest", func(t *testing.T) {
		body := bytes.NewBufferString(`not-json`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if code := decodeErrorPayload(t, rec).Code; code != codeItemCreateInvalidJSON {
			t.Fatalf("error code = %q, want %q", code, codeItemCreateInvalidJSON)
		}
	})
}

func TestItemHandlerList(t *testing.T) {
	t.Run("empty store returns empty data with zero total in meta", func(t *testing.T) {
		h := ItemHandler{Repo: item.NewMemoryStore()}
		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		env := decodeEnvelope(t, rec)
		var items []item.Item
		if err := json.Unmarshal(env.Data, &items); err != nil {
			t.Fatalf("decode items: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("items = %v, want empty", items)
		}
		assertMeta(t, env.Meta, 0, 20, 0)
	})

	t.Run("populated store returns items in creation order", func(t *testing.T) {
		store := item.NewMemoryStore()
		store.Create(context.Background(), "alpha", "alpha item")
		store.Create(context.Background(), "beta", "beta item")
		h := ItemHandler{Repo: store}

		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		env := decodeEnvelope(t, rec)
		var items []item.Item
		if err := json.Unmarshal(env.Data, &items); err != nil {
			t.Fatalf("decode items: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("len(items) = %d, want 2", len(items))
		}
		assertMeta(t, env.Meta, 2, 20, 0)
		// Verify stable creation order.
		if items[0].Name != "alpha" || items[1].Name != "beta" {
			t.Fatalf("unexpected order: [%s, %s], want [alpha, beta]",
				items[0].Name, items[1].Name)
		}
	})

	t.Run("limit param restricts returned items and is reflected in meta", func(t *testing.T) {
		store := item.NewMemoryStore()
		for _, name := range []string{"a", "b", "c", "d", "e"} {
			store.Create(context.Background(), name, name+" item")
		}
		h := ItemHandler{Repo: store}

		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items?limit=3", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		env := decodeEnvelope(t, rec)
		var items []item.Item
		if err := json.Unmarshal(env.Data, &items); err != nil {
			t.Fatalf("decode items: %v", err)
		}
		if len(items) != 3 {
			t.Fatalf("len(items) = %d, want 3", len(items))
		}
		assertMeta(t, env.Meta, 5, 3, 0)
		if items[0].Name != "a" || items[2].Name != "c" {
			t.Fatalf("unexpected page items: %v", items)
		}
	})

	t.Run("offset beyond total clamps and returns empty page", func(t *testing.T) {
		store := item.NewMemoryStore()
		store.Create(context.Background(), "a", "a item")
		store.Create(context.Background(), "b", "b item")
		store.Create(context.Background(), "c", "c item")
		h := ItemHandler{Repo: store}

		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items?offset=10", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		env := decodeEnvelope(t, rec)
		var items []item.Item
		if err := json.Unmarshal(env.Data, &items); err != nil {
			t.Fatalf("decode items: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("items = %v, want empty (offset beyond total)", items)
		}
		// total is 3; offset was clamped to 3.
		assertMeta(t, env.Meta, 3, 20, 3)
	})

	t.Run("offset param skips items in creation order", func(t *testing.T) {
		store := item.NewMemoryStore()
		for _, name := range []string{"a", "b", "c", "d", "e"} {
			store.Create(context.Background(), name, name+" item")
		}
		h := ItemHandler{Repo: store}

		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items?offset=3", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		env := decodeEnvelope(t, rec)
		var items []item.Item
		if err := json.Unmarshal(env.Data, &items); err != nil {
			t.Fatalf("decode items: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("len(items) = %d, want 2", len(items))
		}
		assertMeta(t, env.Meta, 5, 20, 3)
		if items[0].Name != "d" || items[1].Name != "e" {
			t.Fatalf("unexpected offset items: %v", items)
		}
	})

	t.Run("invalid limit returns 400 TypeBadRequest", func(t *testing.T) {
		h := ItemHandler{Repo: item.NewMemoryStore()}
		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items?limit=abc", nil))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if code := decodeErrorPayload(t, rec).Code; code != codeItemListInvalidParam {
			t.Fatalf("error code = %q, want %q", code, codeItemListInvalidParam)
		}
	})

	t.Run("invalid offset returns 400 TypeBadRequest", func(t *testing.T) {
		h := ItemHandler{Repo: item.NewMemoryStore()}
		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items?offset=xyz", nil))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if code := decodeErrorPayload(t, rec).Code; code != codeItemListInvalidParam {
			t.Fatalf("error code = %q, want %q", code, codeItemListInvalidParam)
		}
	})

	t.Run("negative limit returns 400 TypeBadRequest", func(t *testing.T) {
		h := ItemHandler{Repo: item.NewMemoryStore()}
		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items?limit=-1", nil))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestItemHandlerGetByID(t *testing.T) {
	store := item.NewMemoryStore()
	created := store.Create(context.Background(), "gadget", "a gadget")
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
		got := decodeReferenceData[item.Item](t, rec)
		if got.ID != created.ID || got.Name != "gadget" || got.Description != "a gadget" {
			t.Fatalf("unexpected item: %+v", got)
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
	t.Run("existing id returns 200 with fully updated item", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create(context.Background(), "original", "an original item")
		h := ItemHandler{Repo: store}

		body := bytes.NewBufferString(`{"name":"renamed","description":"updated description"}`)
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
		// id and created_at are immutable; name and description are replaced.
		if got.ID != created.ID || got.Name != "renamed" ||
			got.Description != "updated description" || !got.CreatedAt.Equal(created.CreatedAt) {
			t.Fatalf("unexpected updated item: %+v", got)
		}
	})

	t.Run("description is replaced by put", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create(context.Background(), "widget", "old description")
		h := ItemHandler{Repo: store}

		body := bytes.NewBufferString(`{"name":"widget","description":"new description"}`)
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
		if got.Description != "new description" {
			t.Fatalf("Description = %q, want new description", got.Description)
		}
	})

	t.Run("missing id returns 404 TypeNotFound", func(t *testing.T) {
		h := ItemHandler{Repo: item.NewMemoryStore()}
		body := bytes.NewBufferString(`{"name":"anything","description":"anything"}`)
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

	t.Run("both name and description missing returns 400 with both field details", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create(context.Background(), "original", "an original item")
		h := ItemHandler{Repo: store}

		body := bytes.NewBufferString(`{}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+created.ID, body)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.Update(rec, req.WithContext(ctx))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		ef := decodeErrorPayload(t, rec)
		if ef.Code != codeItemUpdateFieldsRequired {
			t.Fatalf("error code = %q, want %q", ef.Code, codeItemUpdateFieldsRequired)
		}
		if _, ok := ef.Details["name"]; !ok {
			t.Error("error details missing 'name' key")
		}
		if _, ok := ef.Details["description"]; !ok {
			t.Error("error details missing 'description' key")
		}
	})

	t.Run("missing name only returns 400 with name detail", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create(context.Background(), "original", "an original item")
		h := ItemHandler{Repo: store}

		body := bytes.NewBufferString(`{"name":"","description":"some desc"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+created.ID, body)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.Update(rec, req.WithContext(ctx))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		ef := decodeErrorPayload(t, rec)
		if ef.Code != codeItemUpdateFieldsRequired {
			t.Fatalf("error code = %q, want %q", ef.Code, codeItemUpdateFieldsRequired)
		}
		if _, ok := ef.Details["name"]; !ok {
			t.Error("error details missing 'name' key")
		}
		if _, ok := ef.Details["description"]; ok {
			t.Error("error details should not contain 'description' when description is present")
		}
	})

	t.Run("empty body returns 400 TypeRequired with body_required code", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create(context.Background(), "original", "an original item")
		h := ItemHandler{Repo: store}

		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+created.ID, http.NoBody)
		ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
			Params: map[string]string{"id": created.ID},
		})
		rec := httptest.NewRecorder()
		h.Update(rec, req.WithContext(ctx))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if code := decodeErrorPayload(t, rec).Code; code != codeItemUpdateBodyRequired {
			t.Fatalf("error code = %q, want %q", code, codeItemUpdateBodyRequired)
		}
	})

	t.Run("invalid JSON returns 400 TypeBadRequest", func(t *testing.T) {
		store := item.NewMemoryStore()
		created := store.Create(context.Background(), "original", "an original item")
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
		created := store.Create(context.Background(), "original", "an original item")
		h := ItemHandler{Repo: store}

		for i := 0; i < 3; i++ {
			body := bytes.NewBufferString(`{"name":"stable","description":"stable desc"}`)
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
			if got.Name != "stable" || got.Description != "stable desc" {
				t.Fatalf("attempt %d: name=%q desc=%q, want stable/stable desc", i+1, got.Name, got.Description)
			}
		}
	})
}

func TestItemHandlerDelete(t *testing.T) {
	store := item.NewMemoryStore()
	created := store.Create(context.Background(), "gadget", "a gadget")
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

// assertMeta verifies pagination metadata from the response meta field.
func assertMeta(t *testing.T, meta map[string]any, wantTotal, wantLimit, wantOffset int) {
	t.Helper()
	total := int(meta["total"].(float64))
	limit := int(meta["limit"].(float64))
	offset := int(meta["offset"].(float64))
	if total != wantTotal || limit != wantLimit || offset != wantOffset {
		t.Fatalf("meta total=%d limit=%d offset=%d, want %d/%d/%d",
			total, limit, offset, wantTotal, wantLimit, wantOffset)
	}
}

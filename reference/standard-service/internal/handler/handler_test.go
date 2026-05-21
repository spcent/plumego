package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestHealthHandlerResponses(t *testing.T) {
	h := HealthHandler{ServiceName: "svc"}

	tests := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		status  string
		check   string
	}{
		{name: "live", handler: h.Live, status: "ok", check: "liveness"},
		{name: "ready", handler: h.Ready, status: "ready", check: "readiness"},
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

func TestAPIHandlerResponses(t *testing.T) {
	h := APIHandler{}

	helloRec := httptest.NewRecorder()
	h.Hello(helloRec, httptest.NewRequest(http.MethodGet, "/", nil))
	if helloRec.Code != http.StatusOK {
		t.Fatalf("hello status = %d, want %d", helloRec.Code, http.StatusOK)
	}
	hello := decodeReferenceData[helloResponse](t, helloRec)
	if hello.Service != "plumego-reference" || hello.Mode != "canonical" || hello.Endpoints["api_hello"] == "" {
		t.Fatalf("unexpected hello response: %+v", hello)
	}

	greetRec := httptest.NewRecorder()
	h.Greet(greetRec, httptest.NewRequest(http.MethodGet, "/api/v1/greet?name=Alice", nil))
	if greetRec.Code != http.StatusOK {
		t.Fatalf("greet status = %d, want %d", greetRec.Code, http.StatusOK)
	}
	greet := decodeReferenceData[greetResponse](t, greetRec)
	if greet.Message != "hello, Alice" {
		t.Fatalf("unexpected greet response: %+v", greet)
	}

	statusRec := httptest.NewRecorder()
	h.Status(statusRec, httptest.NewRequest(http.MethodGet, "/api/status", nil))
	if statusRec.Code != http.StatusOK {
		t.Fatalf("status status = %d, want %d", statusRec.Code, http.StatusOK)
	}
	status := decodeReferenceData[statusResponse](t, statusRec)
	if status.Status != "healthy" || status.Structure.Routes != "one_method_one_path_one_handler" || len(status.Modules) == 0 {
		t.Fatalf("unexpected status response: %+v", status)
	}
}

func TestItemHandlerCreate(t *testing.T) {
	h := ItemHandler{Repo: NewMemoryItemStore()}

	t.Run("valid body returns 201 with item", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"widget"}`)
		rec := httptest.NewRecorder()
		h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		item := decodeReferenceData[Item](t, rec)
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

func TestItemHandlerGetByID(t *testing.T) {
	store := NewMemoryItemStore()
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
		item := decodeReferenceData[Item](t, rec)
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

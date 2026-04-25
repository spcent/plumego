package handler

import (
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

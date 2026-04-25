package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestWriteHealthResponseShape(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	writeHealthResponse(rec, req, "with-messaging")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != contract.ContentTypeJSON {
		t.Fatalf("content type = %q, want %q", got, contract.ContentTypeJSON)
	}

	var env struct {
		Data healthResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if env.Data.Status != "ok" || env.Data.Service != "with-messaging" || env.Data.Timestamp == "" {
		t.Fatalf("unexpected health response: %+v", env.Data)
	}
}

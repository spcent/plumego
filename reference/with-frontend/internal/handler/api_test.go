package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIStatusShape(t *testing.T) {
	h := APIHandler{Logger: nil}
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	h.Status(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp struct {
		Data struct {
			Status  string `json:"status"`
			Service string `json:"service"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.Status != "ok" {
		t.Errorf("status = %q, want ok", resp.Data.Status)
	}
	if resp.Data.Service != "with-frontend" {
		t.Errorf("service = %q, want with-frontend", resp.Data.Service)
	}
}

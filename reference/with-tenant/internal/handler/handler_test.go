package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModelsResponseShape(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	rec := httptest.NewRecorder()

	ModelsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp struct {
		Data struct {
			TenantID string   `json:"tenant_id"`
			Models   []string `json:"models"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Data.Models) == 0 {
		t.Error("models list should not be empty")
	}
	if resp.Data.Models[0] != "gpt-4o" {
		t.Errorf("models[0] = %q, want gpt-4o", resp.Data.Models[0])
	}
}

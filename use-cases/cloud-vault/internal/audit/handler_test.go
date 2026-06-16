package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	plumelog "github.com/spcent/plumego/log"
)

func TestHandler_List(t *testing.T) {
	db := openTestDB(t)
	logger := NewLogger(db.DB)
	logger.Log(context.Background(), "u1", "127.0.0.1", ActionCreate, ResourceDocument, "doc-1", nil)

	h := NewHandler(logger, plumelog.NewLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}

	var body struct {
		Data struct {
			Items []Event `json:"items"`
			Total int     `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Data.Total != 1 || len(body.Data.Items) != 1 {
		t.Fatalf("unexpected response body: %+v", body)
	}
}

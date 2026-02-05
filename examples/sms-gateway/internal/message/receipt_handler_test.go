package message

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/metrics/smsgateway"
)

func TestExampleReceiptHandlerRecordsDelay(t *testing.T) {
	repo := NewMemoryRepository()
	collector := metrics.NewMockCollector()
	reporter := smsgateway.NewReporter(collector)

	now := time.Now().UTC()
	msg := Message{
		ID:        "msg-1",
		TenantID:  "tenant-1",
		Provider:  "provider-a",
		Status:    StatusSent,
		SentAt:    now.Add(-10 * time.Second),
		UpdatedAt: now.Add(-2 * time.Second),
		CreatedAt: now.Add(-30 * time.Second),
	}
	if err := repo.Insert(context.Background(), msg); err != nil {
		t.Fatalf("insert message: %v", err)
	}

	payload := ReceiptRequest{
		MessageID:   msg.ID,
		DeliveredAt: now.Format(time.RFC3339),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	ctx := contract.NewCtx(rec, req, nil)

	handler := ExampleReceiptHandler(repo, reporter)
	handler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	stored, found, err := repo.Get(context.Background(), msg.ID)
	if err != nil || !found {
		t.Fatalf("expected message stored")
	}
	if stored.Status != StatusDelivered {
		t.Fatalf("expected delivered status, got %s", stored.Status)
	}

	records := collector.GetRecords()
	foundMetric := false
	for _, rec := range records {
		if rec.Type == metrics.MetricSMSGateway && rec.Name == smsgateway.MetricReceiptDelay {
			foundMetric = true
			if rec.Labels[smsgateway.LabelProvider] != "provider-a" {
				t.Fatalf("unexpected provider label")
			}
			if rec.Duration < 9*time.Second {
				t.Fatalf("expected receipt delay based on sent_at, got %v", rec.Duration)
			}
		}
	}
	if !foundMetric {
		t.Fatalf("expected receipt delay metric")
	}
}

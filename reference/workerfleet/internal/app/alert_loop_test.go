package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"workerfleet/internal/domain"
)

func TestEvaluateAndNotifyAlertsPersistsAndDispatches(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})
	if err := runtime.store.UpsertWorkerSnapshot(domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1"},
		Status:   domain.WorkerStatusOffline,
	}); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}
	cfg := DefaultConfig()
	cfg.Runtime.NotificationEnabled = true
	cfg.Runtime.NotifierDeliveryTimeout = time.Second
	cfg.Notifier.WebhookURL = server.URL

	emitted, err := runtime.EvaluateAndNotifyAlerts(context.Background(), cfg)
	if err != nil {
		t.Fatalf("evaluate alerts: %v", err)
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted = %d, want 1", len(emitted))
	}
	if calls != 1 {
		t.Fatalf("webhook calls = %d, want 1", calls)
	}
	records, err := runtime.store.ListAlertRecords()
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("stored alerts = %d, want 1", len(records))
	}
}

func TestEvaluateAndNotifyAlertsIgnoresDeliveryFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})
	if err := runtime.store.UpsertWorkerSnapshot(domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1"},
		Status:   domain.WorkerStatusOffline,
	}); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}
	cfg := DefaultConfig()
	cfg.Runtime.NotificationEnabled = true
	cfg.Runtime.NotifierDeliveryTimeout = time.Second
	cfg.Notifier.WebhookURL = server.URL

	if _, err := runtime.EvaluateAndNotifyAlerts(context.Background(), cfg); err != nil {
		t.Fatalf("delivery failure should not fail alert evaluation: %v", err)
	}
}

func TestStartAlertLoopDisabledIsSafe(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})

	stop, err := runtime.StartAlertLoop(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("start alert loop: %v", err)
	}
	stop()
}

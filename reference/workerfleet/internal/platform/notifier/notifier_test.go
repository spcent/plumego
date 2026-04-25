package notifier

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"workerfleet/internal/domain"
)

func assertErrorDoesNotLeak(t *testing.T, err error, secret string) {
	t.Helper()

	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaked secret value %q: %v", secret, err)
	}
}

func TestFeishuNotifierSendsTextPayload(t *testing.T) {
	var got feishuTextMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode feishu payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewFeishuNotifier(FeishuConfig{
		WebhookURL: server.URL,
		HTTPClient: server.Client(),
	})
	alert := domain.AlertRecord{
		AlertType: domain.AlertWorkerOffline,
		Status:    domain.AlertStatusFiring,
		WorkerID:  "worker-1",
		Message:   "worker is offline",
	}

	if err := notifier.Notify(context.Background(), alert); err != nil {
		t.Fatalf("notify: %v", err)
	}
	if got.MsgType != "text" {
		t.Fatalf("msg_type = %q, want text", got.MsgType)
	}
	if !strings.Contains(got.Content.Text, "worker_offline") {
		t.Fatalf("unexpected text payload %q", got.Content.Text)
	}
}

func TestWebhookNotifierSendsAlertEnvelope(t *testing.T) {
	var got webhookEnvelope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test-Token") != "secret-value" {
			t.Fatalf("missing custom header")
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode webhook payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier(WebhookConfig{
		URL: server.URL,
		Headers: map[string]string{
			"X-Test-Token": "secret-value",
		},
		HTTPClient: server.Client(),
	})
	alert := domain.AlertRecord{
		AlertType:   domain.AlertWorkerDegraded,
		Status:      domain.AlertStatusResolved,
		WorkerID:    "worker-1",
		Message:     "worker is degraded resolved",
		TriggeredAt: time.Date(2026, 4, 19, 16, 0, 0, 0, time.UTC),
		ResolvedAt:  time.Date(2026, 4, 19, 16, 5, 0, 0, time.UTC),
	}

	if err := notifier.Notify(context.Background(), alert); err != nil {
		t.Fatalf("notify: %v", err)
	}
	if got.Alert.Status != domain.AlertStatusResolved {
		t.Fatalf("status = %q, want %q", got.Alert.Status, domain.AlertStatusResolved)
	}
}

func TestDispatcherFanOutsToAllSinks(t *testing.T) {
	feishuCalls := 0
	webhookCalls := 0

	feishuServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		feishuCalls++
		w.WriteHeader(http.StatusOK)
	}))
	defer feishuServer.Close()

	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalls++
		w.WriteHeader(http.StatusOK)
	}))
	defer webhookServer.Close()

	dispatcher := NewDispatcher(
		NewFeishuNotifier(FeishuConfig{WebhookURL: feishuServer.URL, HTTPClient: feishuServer.Client()}),
		NewWebhookNotifier(WebhookConfig{URL: webhookServer.URL, HTTPClient: webhookServer.Client()}),
	)

	if err := dispatcher.Notify(context.Background(), domain.AlertRecord{
		AlertType: domain.AlertWorkerOffline,
		Status:    domain.AlertStatusFiring,
		WorkerID:  "worker-1",
		Message:   "worker is offline",
	}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if feishuCalls != 1 || webhookCalls != 1 {
		t.Fatalf("fan-out counts feishu=%d webhook=%d", feishuCalls, webhookCalls)
	}
}

func TestNotifierErrorDoesNotLeakHeaderSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	notifier := NewWebhookNotifier(WebhookConfig{
		URL: server.URL,
		Headers: map[string]string{
			"X-Test-Token": "super-secret-token",
		},
		HTTPClient: server.Client(),
	})

	err := notifier.Notify(context.Background(), domain.AlertRecord{
		AlertType: domain.AlertWorkerOffline,
		Status:    domain.AlertStatusFiring,
		WorkerID:  "worker-1",
		Message:   "worker is offline",
	})
	if err == nil {
		t.Fatalf("expected notify to fail")
	}
	assertErrorDoesNotLeak(t, err, "super-secret-token")
}

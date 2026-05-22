package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"workerfleet/internal/domain"
	"workerfleet/internal/platform/notifier"
	platformstore "workerfleet/internal/platform/store"
)

func TestEvaluateAndNotifyAlertsPersistsEnqueuesAndDelivers(t *testing.T) {
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
	if err := runtime.shell.alerts.store.UpsertWorkerSnapshot(context.Background(), domain.WorkerSnapshot{
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
	if calls != 0 {
		t.Fatalf("webhook calls before delivery = %d, want 0", calls)
	}
	records, err := runtime.shell.alerts.store.ListAlertRecords(context.Background())
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("stored alerts = %d, want 1", len(records))
	}
	if err := runtime.shell.alerts.DeliverNotificationOutbox(context.Background(), cfg, 10); err != nil {
		t.Fatalf("deliver notification outbox: %v", err)
	}
	if calls != 1 {
		t.Fatalf("webhook calls = %d, want 1", calls)
	}
}

func TestEvaluateAndNotifyAlertsEnqueuesOutbox(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})
	runtime.shell.alerts.engineFactory = func() domainAlertEngine {
		return domainAlertEngineFunc(func(context.Context) ([]domain.AlertRecord, error) {
			return []domain.AlertRecord{{
				AlertID:   "alert-1",
				WorkerID:  "worker-1",
				AlertType: domain.AlertWorkerOffline,
			}}, nil
		})
	}

	cfg := DefaultConfig()
	cfg.Runtime.NotificationEnabled = true
	cfg.Notifier.FeishuWebhookURL = "https://feishu.example/hook"
	cfg.Notifier.WebhookURL = "https://webhook.example/hook"

	if _, err := runtime.EvaluateAndNotifyAlerts(context.Background(), cfg); err != nil {
		t.Fatalf("evaluate alerts: %v", err)
	}
	jobs, err := runtime.shell.alerts.store.ClaimNotificationJobs(context.Background(), time.Now().UTC(), 10)
	if err != nil {
		t.Fatalf("claim notification jobs: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("jobs = %d, want 2", len(jobs))
	}
	seen := map[platformstore.NotificationSinkType]bool{}
	for _, job := range jobs {
		seen[job.SinkType] = true
	}
	if !seen[platformstore.NotificationSinkFeishu] || !seen[platformstore.NotificationSinkWebhook] {
		t.Fatalf("unexpected sinks %#v", seen)
	}
}

func TestNotificationOutboxRetriesDeliveryFailure(t *testing.T) {
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
	observer := &recordingRuntimeErrorObserver{}
	runtime.shell.alerts.errors = observer
	if err := runtime.shell.alerts.store.UpsertWorkerSnapshot(context.Background(), domain.WorkerSnapshot{
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
	if err := runtime.shell.alerts.DeliverNotificationOutbox(context.Background(), cfg, 10); err != nil {
		t.Fatalf("deliver notification outbox: %v", err)
	}
	operations, _ := observer.snapshot()
	if len(operations) != 1 || operations[0] != "alert_notify" {
		t.Fatalf("reported operations = %#v, want alert_notify", operations)
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

func TestAlertRunnerUsesInjectedDependencies(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})
	observer := &recordingRuntimeErrorObserver{}
	runtime.shell.alerts.errors = observer
	runtime.shell.alerts.engineFactory = func() domainAlertEngine {
		return domainAlertEngineFunc(func(context.Context) ([]domain.AlertRecord, error) {
			return []domain.AlertRecord{{
				AlertID:   "alert-1",
				WorkerID:  "worker-1",
				AlertType: domain.AlertWorkerOffline,
			}}, nil
		})
	}
	dispatched := 0
	runtime.shell.alerts.dispatcherFn = func(Config) alertDispatcher {
		return notifier.NewDispatcherWithBindings(notifier.SinkBinding{
			Type: platformstore.NotificationSinkWebhook,
			Sink: alertDispatcherFunc(func(context.Context, domain.AlertRecord) error {
				dispatched++
				return errors.New("notify failed")
			}),
		})
	}

	cfg := DefaultConfig()
	cfg.Runtime.NotificationEnabled = true
	cfg.Runtime.NotifierDeliveryTimeout = time.Second
	cfg.Notifier.WebhookURL = "https://webhook.example/hook"

	if _, err := runtime.EvaluateAndNotifyAlerts(context.Background(), cfg); err != nil {
		t.Fatalf("evaluate and notify alerts: %v", err)
	}
	if err := runtime.shell.alerts.DeliverNotificationOutbox(context.Background(), cfg, 10); err != nil {
		t.Fatalf("deliver notification outbox: %v", err)
	}
	if dispatched != 1 {
		t.Fatalf("dispatched = %d, want 1", dispatched)
	}
	operations, errs := observer.snapshot()
	if len(operations) != 1 || operations[0] != "alert_notify" {
		t.Fatalf("reported operations = %#v, want alert_notify", operations)
	}
	if errs[0] == nil || errs[0].Error() != "notify failed" {
		t.Fatalf("error = %v, want notify failed", errs[0])
	}
}

type alertDispatcherFunc func(context.Context, domain.AlertRecord) error

func (fn alertDispatcherFunc) Notify(ctx context.Context, alert domain.AlertRecord) error {
	return fn(ctx, alert)
}

type domainAlertEngineFunc func(context.Context) ([]domain.AlertRecord, error)

func (fn domainAlertEngineFunc) Evaluate(ctx context.Context) ([]domain.AlertRecord, error) {
	return fn(ctx)
}

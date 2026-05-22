package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"workerfleet/internal/domain"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
	"workerfleet/internal/platform/notifier"
	platformstore "workerfleet/internal/platform/store"
)

const (
	defaultNotificationBatchSize = 25
	maxNotificationAttempts      = 5
)

func NewAlertRunner(store runtimeStore, policy domain.StatusPolicy, alertPolicy domain.AlertPolicy, metrics *workerfleetmetrics.Observer, errors RuntimeErrorObserver) *AlertRunner {
	runner := &AlertRunner{
		store:       store,
		policy:      policy,
		alertPolicy: alertPolicy,
		metrics:     metrics,
		errors:      errors,
		lease:       nopLoopLease{},
	}
	runner.dispatcherFn = func(cfg Config) alertDispatcher {
		return newAlertDispatcher(cfg)
	}
	runner.engineFactory = func() domainAlertEngine {
		return domain.NewAlertEngine(
			runner.store,
			runner.store,
			runner.policy,
			nil,
			domain.WithAlertPolicy(runner.alertPolicy),
			domain.WithAlertMetrics(runner.metrics),
		)
	}
	return runner
}

func (r *Runtime) StartAlertLoop(ctx context.Context, cfg Config) (func(), error) {
	if r == nil || r.shell.alerts == nil {
		return func() {}, nil
	}
	return r.shell.alerts.Start(ctx, cfg)
}

func (a *AlertRunner) Start(ctx context.Context, cfg Config) (func(), error) {
	if a == nil || !cfg.Runtime.AlertEvaluationEnabled {
		return func() {}, nil
	}
	loopCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	settings := cfg.Runtime.alertEvaluationLoopSettings()
	settings.Lease = a.lease
	startManagedLoop(loopCtx, &wg, settings, a.reportRuntimeError, func(ctx context.Context) error {
		_, err := a.EvaluateAndNotifyAlerts(ctx, cfg)
		return err
	})
	if cfg.Runtime.NotificationEnabled {
		deliverySettings := cfg.Runtime.notificationDeliveryLoopSettings()
		deliverySettings.Lease = a.lease
		startManagedLoop(loopCtx, &wg, deliverySettings, a.reportRuntimeError, func(ctx context.Context) error {
			return a.DeliverNotificationOutbox(ctx, cfg, defaultNotificationBatchSize)
		})
	}
	return func() {
		cancel()
		wg.Wait()
	}, nil
}

func (r *Runtime) EvaluateAndNotifyAlerts(ctx context.Context, cfg Config) ([]domain.AlertRecord, error) {
	if r == nil || r.shell.alerts == nil {
		return nil, errWorkerfleetStoreNotConfigured
	}
	return r.shell.alerts.EvaluateAndNotifyAlerts(ctx, cfg)
}

func (a *AlertRunner) EvaluateAndNotifyAlerts(ctx context.Context, cfg Config) ([]domain.AlertRecord, error) {
	if a == nil || a.store == nil {
		return nil, errWorkerfleetStoreNotConfigured
	}
	engine := a.engineFactory()
	emitted, err := engine.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Runtime.NotificationEnabled || len(emitted) == 0 {
		return emitted, nil
	}
	if err := a.store.EnqueueNotificationJobs(ctx, notificationJobsForAlerts(cfg, emitted, time.Now().UTC())); err != nil {
		return nil, err
	}
	return emitted, nil
}

func (a *AlertRunner) DeliverNotificationOutbox(ctx context.Context, cfg Config, limit int) error {
	if a == nil || a.store == nil {
		return errWorkerfleetStoreNotConfigured
	}
	jobs, err := a.store.ClaimNotificationJobs(ctx, time.Now().UTC(), limit)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return nil
	}

	dispatcher := a.dispatcherFn(cfg)
	bindings := dispatcher.Bindings()
	for _, job := range jobs {
		binding, ok := findNotificationSink(bindings, job.SinkType)
		if !ok {
			if err := a.markNotificationFailed(ctx, job, notifier.ErrNoSinks); err != nil {
				return err
			}
			continue
		}
		deliveryCtx, cancel := context.WithTimeout(ctx, cfg.Runtime.NotifierDeliveryTimeout)
		err := binding.Sink.Notify(deliveryCtx, job.Alert)
		cancel()
		if err != nil {
			a.reportRuntimeError("alert_notify", err)
			if err := a.markNotificationFailed(ctx, job, err); err != nil {
				return err
			}
			continue
		}
		if err := a.store.MarkNotificationDelivered(ctx, job.JobID, time.Now().UTC()); err != nil {
			return err
		}
	}
	return nil
}

func (a *AlertRunner) markNotificationFailed(ctx context.Context, job platformstore.NotificationJob, err error) error {
	errorClass, permanent := notifier.ClassifyError(err)
	if job.Attempts >= maxNotificationAttempts {
		permanent = true
	}
	return a.store.MarkNotificationFailed(ctx, job.JobID, platformstore.NotificationFailure{
		ErrorClass:    errorClass,
		ErrorMessage:  err.Error(),
		Permanent:     permanent,
		NextAttemptAt: time.Now().UTC().Add(notificationRetryDelay(job.Attempts)),
	})
}

func (a *AlertRunner) reportRuntimeError(operation string, err error) {
	if a == nil || a.errors == nil || err == nil {
		return
	}
	a.errors.ObserveRuntimeError(operation, err)
}

func newAlertDispatcher(cfg Config) *notifier.Dispatcher {
	client := &http.Client{Timeout: cfg.Runtime.NotifierDeliveryTimeout}
	var sinks []notifier.SinkBinding
	if strings.TrimSpace(cfg.Notifier.FeishuWebhookURL) != "" {
		sinks = append(sinks, notifier.SinkBinding{
			Type: platformstore.NotificationSinkFeishu,
			Sink: notifier.NewFeishuNotifier(notifier.FeishuConfig{
				WebhookURL: cfg.Notifier.FeishuWebhookURL,
				HTTPClient: client,
			}),
		})
	}
	if strings.TrimSpace(cfg.Notifier.WebhookURL) != "" {
		sinks = append(sinks, notifier.SinkBinding{
			Type: platformstore.NotificationSinkWebhook,
			Sink: notifier.NewWebhookNotifier(notifier.WebhookConfig{
				URL:        cfg.Notifier.WebhookURL,
				Headers:    cfg.Notifier.WebhookHeaders,
				HTTPClient: client,
			}),
		})
	}
	return notifier.NewDispatcherWithBindings(sinks...)
}

func notificationJobsForAlerts(cfg Config, alerts []domain.AlertRecord, now time.Time) []platformstore.NotificationJob {
	sinkTypes := configuredNotificationSinkTypes(cfg)
	jobs := make([]platformstore.NotificationJob, 0, len(alerts)*len(sinkTypes))
	for _, alert := range alerts {
		for _, sinkType := range sinkTypes {
			jobs = append(jobs, platformstore.NotificationJob{
				JobID:         notificationJobID(alert.AlertID, sinkType),
				AlertID:       alert.AlertID,
				SinkType:      sinkType,
				Alert:         alert,
				Status:        platformstore.NotificationJobPending,
				NextAttemptAt: now,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		}
	}
	return jobs
}

func configuredNotificationSinkTypes(cfg Config) []platformstore.NotificationSinkType {
	var sinkTypes []platformstore.NotificationSinkType
	if strings.TrimSpace(cfg.Notifier.FeishuWebhookURL) != "" {
		sinkTypes = append(sinkTypes, platformstore.NotificationSinkFeishu)
	}
	if strings.TrimSpace(cfg.Notifier.WebhookURL) != "" {
		sinkTypes = append(sinkTypes, platformstore.NotificationSinkWebhook)
	}
	return sinkTypes
}

func notificationJobID(alertID string, sinkType platformstore.NotificationSinkType) string {
	return fmt.Sprintf("%s:%s", alertID, sinkType)
}

func findNotificationSink(bindings []notifier.SinkBinding, sinkType platformstore.NotificationSinkType) (notifier.SinkBinding, bool) {
	for _, binding := range bindings {
		if binding.Type == sinkType {
			return binding, true
		}
	}
	return notifier.SinkBinding{}, false
}

func notificationRetryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	delay := 30 * time.Second
	for i := 1; i < attempts; i++ {
		delay *= 2
		if delay >= 5*time.Minute {
			return 5 * time.Minute
		}
	}
	return delay
}

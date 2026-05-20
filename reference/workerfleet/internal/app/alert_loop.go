package app

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"workerfleet/internal/domain"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
	"workerfleet/internal/platform/notifier"
)

func NewAlertRunner(store runtimeStore, policy domain.StatusPolicy, metrics *workerfleetmetrics.Observer, errors RuntimeErrorObserver) *AlertRunner {
	runner := &AlertRunner{
		store:   store,
		policy:  policy,
		metrics: metrics,
		errors:  errors,
	}
	runner.dispatcherFn = func(cfg Config) alertDispatcher {
		return newAlertDispatcher(cfg)
	}
	runner.engineFactory = func() domainAlertEngine {
		return domain.NewAlertEngine(runner.store, runner.store, runner.policy, nil, domain.WithAlertMetrics(runner.metrics))
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
	startLoop(loopCtx, &wg, cfg.Runtime.AlertEvaluationInterval, func(ctx context.Context) {
		_, err := a.EvaluateAndNotifyAlerts(ctx, cfg)
		if err != nil {
			a.reportRuntimeError("alert_evaluate", err)
		}
	})
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
	dispatcher := a.dispatcherFn(cfg)
	for _, alert := range emitted {
		deliveryCtx, cancel := context.WithTimeout(ctx, cfg.Runtime.NotifierDeliveryTimeout)
		if err := dispatcher.Notify(deliveryCtx, alert); err != nil {
			a.reportRuntimeError("alert_notify", err)
		}
		cancel()
	}
	return emitted, nil
}

func (a *AlertRunner) reportRuntimeError(operation string, err error) {
	if a == nil || a.errors == nil || err == nil {
		return
	}
	a.errors.ObserveRuntimeError(operation, err)
}

func newAlertDispatcher(cfg Config) *notifier.Dispatcher {
	client := &http.Client{Timeout: cfg.Runtime.NotifierDeliveryTimeout}
	var sinks []notifier.Sink
	if strings.TrimSpace(cfg.Notifier.FeishuWebhookURL) != "" {
		sinks = append(sinks, notifier.NewFeishuNotifier(notifier.FeishuConfig{
			WebhookURL: cfg.Notifier.FeishuWebhookURL,
			HTTPClient: client,
		}))
	}
	if strings.TrimSpace(cfg.Notifier.WebhookURL) != "" {
		sinks = append(sinks, notifier.NewWebhookNotifier(notifier.WebhookConfig{
			URL:        cfg.Notifier.WebhookURL,
			Headers:    cfg.Notifier.WebhookHeaders,
			HTTPClient: client,
		}))
	}
	return notifier.NewDispatcher(sinks...)
}

package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"workerfleet/internal/domain"
	"workerfleet/internal/platform/notifier"
)

func (r *Runtime) StartAlertLoop(ctx context.Context, cfg Config) (func(), error) {
	if r == nil || !cfg.Runtime.AlertEvaluationEnabled {
		return func() {}, nil
	}
	loopCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	startLoop(loopCtx, &wg, cfg.Runtime.AlertEvaluationInterval, func(ctx context.Context) {
		_, _ = r.EvaluateAndNotifyAlerts(ctx, cfg)
	})
	return func() {
		cancel()
		wg.Wait()
	}, nil
}

func (r *Runtime) EvaluateAndNotifyAlerts(ctx context.Context, cfg Config) ([]domain.AlertRecord, error) {
	if r == nil || r.store == nil {
		return nil, fmt.Errorf("workerfleet store is not configured")
	}
	engine := domain.NewAlertEngine(r.store, r.store, r.policy, nil, domain.WithAlertMetrics(r.metrics))
	emitted, err := engine.Evaluate()
	if err != nil {
		return nil, err
	}
	if !cfg.Runtime.NotificationEnabled || len(emitted) == 0 {
		return emitted, nil
	}
	dispatcher := newAlertDispatcher(cfg)
	for _, alert := range emitted {
		deliveryCtx, cancel := context.WithTimeout(ctx, cfg.Runtime.NotifierDeliveryTimeout)
		_ = dispatcher.Notify(deliveryCtx, alert)
		cancel()
	}
	return emitted, nil
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

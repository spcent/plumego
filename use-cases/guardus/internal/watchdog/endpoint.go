package watchdog

import (
	"context"
	"time"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/domain/endpoint"
)

// monitorEndpoint runs the probe loop for a single endpoint until ctx is
// canceled.
func (w *Watchdog) monitorEndpoint(ctx context.Context, ep *endpoint.Endpoint, extraLabels []string) {
	w.executeEndpoint(ctx, ep, extraLabels)
	ticker := time.NewTicker(ep.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			w.log.Warn("monitor endpoint cancelled", plumelog.Fields{
				"group":    ep.Group,
				"endpoint": ep.Name,
				"key":      ep.Key(),
			})
			return
		case <-ticker.C:
			w.executeEndpoint(ctx, ep, extraLabels)
		}
	}
}

func (w *Watchdog) executeEndpoint(ctx context.Context, ep *endpoint.Endpoint, extraLabels []string) {
	if err := w.sem.Acquire(ctx, 1); err != nil {
		w.log.Debug("execute endpoint cancelled", plumelog.Fields{"err": err.Error()})
		return
	}
	defer w.sem.Release(1)
	if w.cfg.Connectivity != nil && w.cfg.Connectivity.Checker != nil && !w.cfg.Connectivity.Checker.IsConnected() {
		w.log.Info("no connectivity; skipping endpoint execution")
		return
	}
	w.log.Debug("monitoring endpoint", plumelog.Fields{
		"group":    ep.Group,
		"endpoint": ep.Name,
		"key":      ep.Key(),
	})
	result := ep.EvaluateHealth()
	if w.cfg.Metrics && w.metrics != nil {
		w.metrics.PublishResult(ep, result, extraLabels)
	}
	w.persistResult(ep, result)
	fields := plumelog.Fields{
		"group":    ep.Group,
		"endpoint": ep.Name,
		"key":      ep.Key(),
		"success":  result.Success,
		"errors":   len(result.Errors),
		"duration": result.Duration.Round(time.Millisecond).String(),
	}
	if !result.Success {
		// On failure include the response body so operators can diagnose
		// without bumping the global log level.
		fields["body"] = result.Body
	}
	w.log.Info("endpoint monitored", fields)
	inEndpointMaintenanceWindow := false
	for _, mw := range ep.MaintenanceWindows {
		if mw.IsUnderMaintenance() {
			w.log.Debug("endpoint under maintenance window")
			inEndpointMaintenanceWindow = true
		}
	}
	if !w.cfg.Maintenance.IsUnderMaintenance() && !inEndpointMaintenanceWindow {
		w.handleAlerting(ep, result)
	} else {
		w.log.Debug("alerting skipped: maintenance window active")
	}
}

func (w *Watchdog) persistResult(ep *endpoint.Endpoint, result *endpoint.Result) {
	if err := w.store.InsertEndpointResult(ep, result); err != nil {
		w.log.Error("failed to insert endpoint result", plumelog.Fields{
			"key": ep.Key(),
			"err": err.Error(),
		})
	}
}

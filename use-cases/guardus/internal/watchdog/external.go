package watchdog

import (
	"context"
	"time"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/domain/endpoint"
)

func (w *Watchdog) monitorExternalEndpointHeartbeat(ctx context.Context, ee *endpoint.ExternalEndpoint, extraLabels []string) {
	ticker := time.NewTicker(ee.Heartbeat.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			w.log.Warn("external heartbeat monitor cancelled", plumelog.Fields{
				"group":    ee.Group,
				"endpoint": ee.Name,
				"key":      ee.Key(),
			})
			return
		case <-ticker.C:
			w.executeExternalEndpointHeartbeat(ctx, ee, extraLabels)
		}
	}
}

func (w *Watchdog) executeExternalEndpointHeartbeat(ctx context.Context, ee *endpoint.ExternalEndpoint, extraLabels []string) {
	if err := w.sem.Acquire(ctx, 1); err != nil {
		w.log.Debug("external heartbeat cancelled", plumelog.Fields{"err": err.Error()})
		return
	}
	defer w.sem.Release(1)
	if w.cfg.Connectivity != nil && w.cfg.Connectivity.Checker != nil && !w.cfg.Connectivity.Checker.IsConnected() {
		w.log.Info("no connectivity; skipping external heartbeat")
		return
	}
	w.log.Debug("checking external heartbeat", plumelog.Fields{
		"group":    ee.Group,
		"endpoint": ee.Name,
		"key":      ee.Key(),
	})
	convertedEndpoint := ee.ToEndpoint()
	hasReceivedResultWithinHeartbeatInterval, err := w.store.HasEndpointStatusNewerThan(ee.Key(), time.Now().Add(-ee.Heartbeat.Interval))
	if err != nil {
		w.log.Error("failed to check heartbeat freshness", plumelog.Fields{
			"key": ee.Key(),
			"err": err.Error(),
		})
		return
	}
	if hasReceivedResultWithinHeartbeatInterval {
		// A push within the interval already updated state; nothing more to do.
		w.log.Info("external heartbeat fresh", plumelog.Fields{
			"group":    ee.Group,
			"endpoint": ee.Name,
			"key":      ee.Key(),
		})
		return
	}
	result := &endpoint.Result{
		Timestamp: time.Now(),
		Success:   false,
		Errors:    []string{"heartbeat: no update received within " + ee.Heartbeat.Interval.String()},
	}
	if w.cfg.Metrics && w.metrics != nil {
		w.metrics.PublishResult(convertedEndpoint, result, extraLabels)
	}
	w.persistResult(convertedEndpoint, result)
	w.log.Info("external heartbeat MISSED", plumelog.Fields{
		"group":    ee.Group,
		"endpoint": ee.Name,
		"key":      ee.Key(),
	})
	inEndpointMaintenanceWindow := false
	for _, mw := range ee.MaintenanceWindows {
		if mw.IsUnderMaintenance() {
			inEndpointMaintenanceWindow = true
			break
		}
	}
	if !w.cfg.Maintenance.IsUnderMaintenance() && !inEndpointMaintenanceWindow {
		w.handleAlerting(convertedEndpoint, result)
		// Sync counters back so external pushes see the updated streaks.
		ee.NumberOfSuccessesInARow = convertedEndpoint.NumberOfSuccessesInARow
		ee.NumberOfFailuresInARow = convertedEndpoint.NumberOfFailuresInARow
	}
}

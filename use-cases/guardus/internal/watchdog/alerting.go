package watchdog

import (
	"errors"
	"os"
	"time"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/domain/endpoint"
)

// HandleAlertingExternal is the entry point for external-endpoint pushes.
//
// It drives the same trigger/resolve state machine as periodic probing and
// syncs the success/failure counters back to the originating ExternalEndpoint
// so that subsequent pushes see the up-to-date streak.
func (w *Watchdog) HandleAlertingExternal(ee *endpoint.ExternalEndpoint, converted *endpoint.Endpoint, result *endpoint.Result) {
	w.handleAlerting(converted, result)
	ee.NumberOfSuccessesInARow = converted.NumberOfSuccessesInARow
	ee.NumberOfFailuresInARow = converted.NumberOfFailuresInARow
}

// handleAlerting drives the trigger/resolve state machine for an endpoint
// against the watchdog's alerting configuration.
func (w *Watchdog) handleAlerting(ep *endpoint.Endpoint, result *endpoint.Result) {
	if w.cfg.Alerting == nil {
		return
	}
	if result.Success {
		w.handleAlertsToResolve(ep, result)
	} else {
		w.handleAlertsToTrigger(ep, result)
	}
}

func (w *Watchdog) handleAlertsToTrigger(ep *endpoint.Endpoint, result *endpoint.Result) {
	ep.NumberOfSuccessesInARow = 0
	ep.NumberOfFailuresInARow++
	// Snapshot LastReminderSent so multiple alerts on the same endpoint use
	// the same reference for reminder-due checks.
	lastReminderSent := ep.LastReminderSent
	for _, ea := range ep.Alerts {
		if !ea.IsEnabled() || ea.FailureThreshold > ep.NumberOfFailuresInARow {
			continue
		}
		sendInitialAlert := !ea.Triggered
		sendReminder := ea.Triggered && ea.MinimumReminderInterval > 0 && time.Since(lastReminderSent) >= ea.MinimumReminderInterval
		if !sendInitialAlert && !sendReminder {
			w.log.Debug("trigger alert skipped (not due)", plumelog.Fields{
				"endpoint":    ep.Name,
				"description": ea.GetDescription(),
			})
			continue
		}
		ap := w.cfg.Alerting.GetAlertingProviderByAlertType(ea.Type)
		if ap == nil {
			w.log.Warn("no provider configured; skipping TRIGGERED alert", plumelog.Fields{
				"type":     string(ea.Type),
				"endpoint": ep.Key(),
			})
			continue
		}
		alertKind := "reminder"
		if sendInitialAlert {
			alertKind = "initial"
		}
		w.log.Info("sending alert", plumelog.Fields{
			"kind":        alertKind,
			"type":        string(ea.Type),
			"endpoint":    ep.Key(),
			"description": ea.GetDescription(),
		})
		var err error
		if os.Getenv("MOCK_ALERT_PROVIDER") == "true" {
			if os.Getenv("MOCK_ALERT_PROVIDER_ERROR") == "true" {
				err = errors.New("error")
			}
		} else {
			err = ap.Send(ep, ea, result, false)
		}
		if err != nil {
			w.log.Error("failed to send alert", plumelog.Fields{
				"endpoint": ep.Key(),
				"err":      err.Error(),
			})
			continue
		}
		if sendInitialAlert {
			ea.Triggered = true
		}
		ep.LastReminderSent = time.Now()
		if err := w.store.UpsertTriggeredEndpointAlert(ep, ea); err != nil {
			w.log.Error("failed to persist triggered alert", plumelog.Fields{
				"endpoint": ep.Key(),
				"err":      err.Error(),
			})
		}
	}
}

func (w *Watchdog) handleAlertsToResolve(ep *endpoint.Endpoint, result *endpoint.Result) {
	ep.NumberOfSuccessesInARow++
	for _, ea := range ep.Alerts {
		isStillBelowSuccessThreshold := ea.SuccessThreshold > ep.NumberOfSuccessesInARow
		if isStillBelowSuccessThreshold && ea.IsEnabled() && ea.Triggered {
			if err := w.store.UpsertTriggeredEndpointAlert(ep, ea); err != nil {
				w.log.Error("failed to refresh triggered alert", plumelog.Fields{
					"endpoint": ep.Key(),
					"err":      err.Error(),
				})
			}
		}
		if !ea.IsEnabled() || !ea.Triggered || isStillBelowSuccessThreshold {
			continue
		}
		// Always clear Triggered, even on send failure; further explanation
		// in alert.Alert.Triggered.
		ea.Triggered = false
		if err := w.store.DeleteTriggeredEndpointAlert(ep, ea); err != nil {
			w.log.Error("failed to delete triggered alert", plumelog.Fields{
				"endpoint": ep.Key(),
				"err":      err.Error(),
			})
		}
		if !ea.IsSendingOnResolved() {
			w.log.Debug("send-on-resolved disabled", plumelog.Fields{
				"type":     string(ea.Type),
				"endpoint": ep.Key(),
			})
			continue
		}
		ap := w.cfg.Alerting.GetAlertingProviderByAlertType(ea.Type)
		if ap == nil {
			w.log.Warn("no provider configured; skipping RESOLVED alert", plumelog.Fields{
				"type":     string(ea.Type),
				"endpoint": ep.Key(),
			})
			continue
		}
		w.log.Info("sending RESOLVED alert", plumelog.Fields{
			"type":        string(ea.Type),
			"endpoint":    ep.Key(),
			"description": ea.GetDescription(),
		})
		if err := ap.Send(ep, ea, result, true); err != nil {
			w.log.Error("failed to send resolved alert", plumelog.Fields{
				"endpoint": ep.Key(),
				"err":      err.Error(),
			})
		}
	}
	ep.NumberOfFailuresInARow = 0
}

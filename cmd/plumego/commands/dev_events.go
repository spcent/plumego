package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spcent/plumego/cmd/plumego/internal/devserver"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
	"github.com/spcent/plumego/x/pubsub"
)

func startDevEventForwarder(ctx context.Context, out *output.Formatter, ps *pubsub.InProcBroker) (func(), error) {
	patterns := []string{"app.*", "build.*"}
	opts := pubsub.DefaultSubOptions()
	subs := make([]pubsub.Subscription, 0, len(patterns))

	for _, pattern := range patterns {
		sub, err := ps.SubscribePatternWithContext(ctx, pattern, opts)
		if err != nil {
			for _, existing := range subs {
				existing.Cancel()
			}
			return func() {}, err
		}
		subs = append(subs, sub)
		go forwardDevEvents(ctx, out, sub)
	}

	return func() {
		for _, sub := range subs {
			sub.Cancel()
		}
	}, nil
}

func forwardDevEvents(ctx context.Context, out *output.Formatter, sub pubsub.Subscription) {
	for {
		select {
		case <-ctx.Done():
			sub.Cancel()
			return
		case msg, ok := <-sub.C():
			if !ok {
				return
			}
			emitPubSubEvent(out, msg)
		}
	}
}

func emitPubSubEvent(out *output.Formatter, msg pubsub.Message) {
	event := output.Event{
		Event: msg.Topic,
	}
	if !msg.Time.IsZero() {
		event.Time = msg.Time.Format(time.RFC3339)
	}

	switch msg.Topic {
	case devserver.EventBuildStart:
		event.Message = "Build started"
		if data, ok := msg.Data.(devserver.BuildEvent); ok {
			event.Data = buildEventData(data)
		}
	case devserver.EventBuildSuccess:
		event.Message = "Build succeeded"
		if data, ok := msg.Data.(devserver.BuildEvent); ok {
			event.Data = buildEventData(data)
		}
	case devserver.EventBuildFail:
		event.Message = "Build failed"
		event.Level = "error"
		if data, ok := msg.Data.(devserver.BuildEvent); ok {
			event.Data = buildEventData(data)
		}
	case devserver.EventAppStart, devserver.EventAppStop, devserver.EventAppRestart:
		if data, ok := msg.Data.(devserver.AppLifecycleEvent); ok {
			event.Data = map[string]any{
				"state": data.State,
			}
			if data.PID != 0 {
				event.Data["pid"] = data.PID
			}
			if data.Error != "" {
				event.Data["error"] = data.Error
				event.Level = "error"
			}
			event.Message = "Application " + data.State
		}
	case devserver.EventAppLog, devserver.EventAppError:
		if data, ok := msg.Data.(devserver.LogEvent); ok {
			event.Level = data.Level
			event.Message = data.Message
			event.Data = map[string]any{
				"source": data.Source,
			}
		}
	default:
		event.Data = toEventMap(msg.Data)
	}

	_ = out.Event(event)
}

func buildEventData(data devserver.BuildEvent) map[string]any {
	payload := map[string]any{
		"success": data.Success,
	}
	if data.Duration > 0 {
		payload["duration_ms"] = data.Duration.Milliseconds()
	}
	if data.Output != "" {
		payload["output"] = data.Output
	}
	if data.Error != "" {
		payload["error"] = data.Error
	}
	return payload
}

func toEventMap(data any) map[string]any {
	switch v := data.(type) {
	case map[string]any:
		return v
	case map[string]string:
		out := make(map[string]any, len(v))
		for k, val := range v {
			out[k] = val
		}
		return out
	default:
		raw, err := json.Marshal(data)
		if err != nil {
			return map[string]any{"value": fmt.Sprintf("%v", data)}
		}
		var out map[string]any
		if err := json.Unmarshal(raw, &out); err != nil {
			return map[string]any{"value": fmt.Sprintf("%v", data)}
		}
		return out
	}
}

func emitDevStart(out *output.Formatter, absDir, addr, dashboardAddr string) error {
	if out.Format() == "text" {
		message := fmt.Sprintf(
			"Starting Plumego Dev Server\n   Project: %s\n   App URL: http://localhost%s\n   Dashboard URL: http://localhost%s\n",
			absDir,
			addr,
			dashboardAddr,
		)
		return out.Event(output.Event{
			Event:   "starting",
			Message: message,
		})
	}

	return out.Event(output.Event{
		Event:   "starting",
		Message: "Plumego Dev Server starting",
		Data: map[string]any{
			"project":        absDir,
			"app_addr":       addr,
			"dashboard_addr": dashboardAddr,
		},
	})
}

func emitDashboardStarted(out *output.Formatter, dashboardAddr string) error {
	if out.Format() == "text" {
		return out.Event(output.Event{
			Event:   "dashboard_started",
			Message: fmt.Sprintf("Dashboard started at http://localhost%s\n", dashboardAddr),
		})
	}

	return out.Event(output.Event{
		Event:   "dashboard_started",
		Message: "Dashboard started",
		Data: map[string]any{
			"url": fmt.Sprintf("http://localhost%s", dashboardAddr),
		},
	})
}

func emitAppReady(out *output.Formatter, addr string) error {
	if out.Format() == "text" {
		return nil
	}

	return out.Event(output.Event{
		Event:   "ready",
		Message: "Application ready",
		Data: map[string]any{
			"url": fmt.Sprintf("http://localhost%s", addr),
		},
	})
}

func emitWatching(out *output.Formatter) error {
	if out.Format() == "text" {
		return out.Event(output.Event{
			Event:   "watching",
			Message: "Watching for changes...\n   Press Ctrl+C to stop",
		})
	}

	return out.Event(output.Event{
		Event:   "watching",
		Message: "Watching for changes",
	})
}

func emitReloadDisabled(out *output.Formatter) error {
	return out.Event(output.Event{
		Event:   "reload_disabled",
		Message: "Auto reload disabled.\nPress Ctrl+C to stop",
		Data: map[string]any{
			"auto_reload": false,
		},
	})
}

func emitFileChanged(out *output.Formatter, path string) error {
	if out.Format() == "text" {
		return out.Event(output.Event{
			Event:   "file_changed",
			Message: fmt.Sprintf("\nFile changed: %s", path),
			Data: map[string]any{
				"path": path,
			},
		})
	}

	return out.Event(output.Event{
		Event:   "file_changed",
		Message: "File changed",
		Data: map[string]any{
			"path": path,
		},
	})
}

func emitReloadFailed(out *output.Formatter, reloadErr error) error {
	message := "Reload failed"
	if out.Format() == "text" {
		message = fmt.Sprintf("Reload failed: %v", reloadErr)
	}

	return out.Event(output.Event{
		Event:   "reload_failed",
		Level:   "error",
		Message: message,
		Data: map[string]any{
			"error": reloadErr.Error(),
		},
	})
}

func emitReloadComplete(out *output.Formatter) error {
	if out.Format() == "text" {
		return out.Event(output.Event{
			Event:   "reload_complete",
			Message: "Reload complete\n",
		})
	}

	return out.Event(output.Event{
		Event:   "reload_complete",
		Message: "Reload complete",
	})
}

func emitWatcherError(out *output.Formatter, watchErr error) error {
	message := "Watcher error"
	if out.Format() == "text" {
		message = fmt.Sprintf("Watcher error: %v", watchErr)
	}

	return out.Event(output.Event{
		Event:   "watcher_error",
		Level:   "warn",
		Message: message,
		Data: map[string]any{
			"error": watchErr.Error(),
		},
	})
}

func emitShutdown(out *output.Formatter) error {
	if out.Format() == "text" {
		return out.Event(output.Event{
			Event:   "stopped",
			Message: "\nShutting down...",
			Data: map[string]any{
				"code": 0,
			},
		})
	}

	return out.Event(output.Event{
		Event:   "stopped",
		Message: "Shutting down",
		Data: map[string]any{
			"code": 0,
		},
	})
}

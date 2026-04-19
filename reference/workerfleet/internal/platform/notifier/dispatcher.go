package notifier

import (
	"context"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

type Sink interface {
	Notify(ctx context.Context, alert domain.AlertRecord) error
}

type Dispatcher struct {
	sinks []Sink
}

func NewDispatcher(sinks ...Sink) *Dispatcher {
	filtered := make([]Sink, 0, len(sinks))
	for _, sink := range sinks {
		if sink == nil {
			continue
		}
		filtered = append(filtered, sink)
	}
	return &Dispatcher{sinks: filtered}
}

func (d *Dispatcher) Notify(ctx context.Context, alert domain.AlertRecord) error {
	for _, sink := range d.sinks {
		if err := sink.Notify(ctx, alert); err != nil {
			return err
		}
	}
	return nil
}

func renderAlertText(alert domain.AlertRecord) string {
	status := string(alert.Status)
	scope := string(alert.WorkerID)
	if scope == "" {
		scope = string(alert.TaskID)
	}
	return "[" + status + "] " + string(alert.AlertType) + " " + scope + " " + alert.Message
}

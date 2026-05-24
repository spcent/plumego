package notifier

import (
	"context"
	"errors"
	"fmt"

	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

var ErrNoSinks = errors.New("workerfleet notifier sink is required")

const (
	ErrorClassConfiguration = "configuration"
	ErrorClassHTTP4xx       = "http_4xx"
	ErrorClassHTTP429       = "http_429"
	ErrorClassHTTP5xx       = "http_5xx"
	ErrorClassNetwork       = "network"
	ErrorClassUnknown       = "unknown"
)

type Sink interface {
	Notify(ctx context.Context, alert domain.AlertRecord) error
}

type SinkBinding struct {
	Type platformstore.NotificationSinkType
	Sink Sink
}

type Dispatcher struct {
	sinks []SinkBinding
}

func NewDispatcher(sinks ...Sink) *Dispatcher {
	bindings := make([]SinkBinding, 0, len(sinks))
	for _, sink := range sinks {
		if sink == nil {
			continue
		}
		bindings = append(bindings, SinkBinding{Sink: sink})
	}
	return NewDispatcherWithBindings(bindings...)
}

func NewDispatcherWithBindings(bindings ...SinkBinding) *Dispatcher {
	filtered := make([]SinkBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.Sink == nil {
			continue
		}
		filtered = append(filtered, binding)
	}
	return &Dispatcher{sinks: filtered}
}

func (d *Dispatcher) Bindings() []SinkBinding {
	if d == nil || len(d.sinks) == 0 {
		return nil
	}
	return append([]SinkBinding(nil), d.sinks...)
}

func (d *Dispatcher) Notify(ctx context.Context, alert domain.AlertRecord) error {
	if d == nil || len(d.sinks) == 0 {
		return ErrNoSinks
	}
	for _, binding := range d.sinks {
		if err := binding.Sink.Notify(ctx, alert); err != nil {
			return err
		}
	}
	return nil
}

type DeliveryError struct {
	Class      string
	Permanent  bool
	Sink       string
	StatusCode int
	Err        error
}

func (e DeliveryError) Error() string {
	if e.Err == nil {
		if e.Sink != "" && e.StatusCode > 0 {
			return fmt.Sprintf("%s notify failed: status %d", e.Sink, e.StatusCode)
		}
		return e.Class
	}
	return e.Err.Error()
}

func (e DeliveryError) Unwrap() error {
	return e.Err
}

func (e DeliveryError) RuntimeErrorClass() string {
	if e.Class == "" {
		return ErrorClassUnknown
	}
	return e.Class
}

func HTTPStatusError(sink string, statusCode int) error {
	class := ErrorClassHTTP5xx
	permanent := false
	switch {
	case statusCode == 429:
		class = ErrorClassHTTP429
	case statusCode >= 400 && statusCode < 500:
		class = ErrorClassHTTP4xx
		permanent = true
	}
	return DeliveryError{
		Class:      class,
		Permanent:  permanent,
		Sink:       sink,
		StatusCode: statusCode,
	}
}

func ClassifyError(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	var deliveryErr DeliveryError
	if errors.As(err, &deliveryErr) {
		return deliveryErr.Class, deliveryErr.Permanent
	}
	if errors.Is(err, ErrNoSinks) {
		return ErrorClassConfiguration, true
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrorClassNetwork, false
	}
	return ErrorClassUnknown, false
}

func renderAlertText(alert domain.AlertRecord) string {
	status := string(alert.Status)
	scope := string(alert.WorkerID)
	if scope == "" {
		scope = string(alert.TaskID)
	}
	return "[" + status + "] " + string(alert.AlertType) + " " + scope + " " + alert.Message
}

package metrics

import (
	"context"
	"time"
)

type multiHTTPObserver struct {
	observers []HTTPObserver
}

// NewMultiHTTPObserver fans out HTTP metrics to all non-nil observers.
func NewMultiHTTPObserver(observers ...HTTPObserver) HTTPObserver {
	filtered := make([]HTTPObserver, 0, len(observers))
	for _, observer := range observers {
		if observer != nil {
			filtered = append(filtered, observer)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return multiHTTPObserver{observers: filtered}
}

func (m multiHTTPObserver) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	for _, observer := range m.observers {
		observer.ObserveHTTP(ctx, method, path, status, bytes, duration)
	}
}

var _ HTTPObserver = multiHTTPObserver{}

// Package watchdog probes endpoints, persists results, and drives alerting.
//
// Unlike upstream gatus, the watchdog state (semaphore, cancel func) is
// scoped to a Watchdog struct rather than package globals, so reload can
// dispose the old watchdog and start a new one without action-at-a-distance.
package watchdog

import (
	"context"
	"sync"
	"time"

	plumelog "github.com/spcent/plumego/log"
	"golang.org/x/sync/semaphore"

	"guardus/internal/config"
	"guardus/internal/domain/endpoint"
	"guardus/internal/metrics"
	"guardus/internal/storage"
)

// UnlimitedConcurrencyWeight is the semaphore weight used when concurrency
// is configured as 0 (unlimited).
const UnlimitedConcurrencyWeight = 10000

// Watchdog drives endpoint probing for a single config snapshot.
//
// One Watchdog instance corresponds to one *config.Config. On reload, the
// caller must Shutdown the old Watchdog before starting a new one — the
// store is the only shared state and is owned by App.
type Watchdog struct {
	cfg     *config.Config
	store   storage.Store
	sem     *semaphore.Weighted
	metrics *metrics.Endpoints
	log     plumelog.StructuredLogger

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

// New constructs a Watchdog. The store, metrics, and logger are injected (no
// global store.Get(), no package-level logger). metrics may be nil to skip
// Prometheus exposition; logger must not be nil.
func New(cfg *config.Config, store storage.Store, m *metrics.Endpoints, logger plumelog.StructuredLogger) *Watchdog {
	return &Watchdog{cfg: cfg, store: store, metrics: m, log: logger}
}

// Monitor spawns goroutines for every enabled endpoint and external endpoint
// heartbeat. It returns immediately; the goroutines stop when ctx is canceled
// or Shutdown is called.
func (w *Watchdog) Monitor(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.done = make(chan struct{})

	if w.cfg.Concurrency == 0 {
		w.sem = semaphore.NewWeighted(UnlimitedConcurrencyWeight)
	} else {
		w.sem = semaphore.NewWeighted(int64(w.cfg.Concurrency))
	}
	extraLabels := w.cfg.GetUniqueExtraMetricLabels()

	var wg sync.WaitGroup
	for _, ep := range w.cfg.Endpoints {
		if !ep.IsEnabled() {
			continue
		}
		// Stagger startup so probes don't all fire at once.
		time.Sleep(222 * time.Millisecond)
		wg.Add(1)
		go func(ep *endpoint.Endpoint) {
			defer wg.Done()
			w.monitorEndpoint(runCtx, ep, extraLabels)
		}(ep)
	}
	for _, ee := range w.cfg.ExternalEndpoints {
		if !ee.IsEnabled() || ee.Heartbeat.Interval <= 0 {
			continue
		}
		wg.Add(1)
		go func(ee *endpoint.ExternalEndpoint) {
			defer wg.Done()
			w.monitorExternalEndpointHeartbeat(runCtx, ee, extraLabels)
		}(ee)
	}
	go func() {
		wg.Wait()
		close(w.done)
	}()
}

// Shutdown cancels in-flight probes and closes endpoint connections.
//
// It is safe to call Shutdown more than once; subsequent calls are no-ops.
func (w *Watchdog) Shutdown() {
	w.mu.Lock()
	cancel := w.cancel
	w.cancel = nil
	done := w.done
	w.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	for _, ep := range w.cfg.Endpoints {
		ep.Close()
	}
	if done != nil {
		// Bound the wait — a probe stuck on a syscall shouldn't block reload.
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			w.log.Warn("watchdog shutdown timed out waiting for probes to finish")
		}
	}
}

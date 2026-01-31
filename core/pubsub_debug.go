package core

import (
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/pubsub"
)

// ConfigurePubSub registers a snapshot endpoint when enabled.
func (a *App) ConfigurePubSub() {
	if err := a.ensureMutable("configure_pubsub", "configure pubsub"); err != nil {
		a.logError("ConfigurePubSub failed", err, nil)
		return
	}

	cfg := a.configSnapshot().PubSub
	if !cfg.Enabled {
		return
	}

	pub := cfg.Pub
	if pub == nil {
		a.mu.RLock()
		pub = a.pub
		a.mu.RUnlock()
	}
	path := strings.TrimSpace(cfg.Path)
	if path == "" {
		path = "/_debug/pubsub"
	}

	a.ensureRouter().GetCtx(path, func(ctx *contract.Ctx) {
		if pub == nil {
			writeContractError(ctx, http.StatusInternalServerError, "missing_pubsub", "pubsub is not configured")
			return
		}

		type snapshoter interface{ Snapshot() pubsub.MetricsSnapshot }

		if ps, ok := pub.(snapshoter); ok {
			writeContractResponse(ctx, http.StatusOK, ps.Snapshot())
			return
		}

		writeContractError(ctx, http.StatusNotImplemented, "not_supported", "pubsub snapshot not supported by this implementation")
	})
}

package pubsubdebug

import (
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/x/pubsub"
)

// Configure registers a snapshot endpoint when enabled.
func Configure(hooks Hooks) {
	if hooks.EnsureMutable == nil || hooks.ConfigSnapshot == nil || hooks.EnsureRouter == nil {
		return
	}

	if err := hooks.EnsureMutable("configure_pubsub", "configure pubsub"); err != nil {
		if hooks.LogError != nil {
			hooks.LogError("ConfigurePubSub failed", err)
		}
		return
	}

	cfg := hooks.ConfigSnapshot()
	if !cfg.Enabled {
		return
	}

	pub := cfg.Pub
	if pub == nil && hooks.DefaultPubSub != nil {
		pub = hooks.DefaultPubSub()
	}
	path := strings.TrimSpace(cfg.Path)
	if path == "" {
		path = "/_debug/pubsub"
	}

	r := hooks.EnsureRouter()
	r.Get(path, contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
		if pub == nil {
			contract.WriteContractError(ctx, http.StatusInternalServerError, "missing_pubsub", "pubsub is not configured")
			return
		}

		type snapshoter interface{ Snapshot() pubsub.MetricsSnapshot }

		if ps, ok := pub.(snapshoter); ok {
			contract.WriteContractResponse(ctx, http.StatusOK, ps.Snapshot())
			return
		}

		contract.WriteContractError(ctx, http.StatusNotImplemented, "not_supported", "pubsub snapshot not supported by this implementation")
	}, r.Logger()))
}

// Hooks provide the minimal integration points for Configure.
type Hooks struct {
	EnsureMutable  func(op, desc string) error
	LogError       func(msg string, err error)
	ConfigSnapshot func() PubSubConfig
	DefaultPubSub  func() pubsub.Broker
	EnsureRouter   func() *router.Router
}

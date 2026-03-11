package devserver

import (
	"context"

	"github.com/spcent/plumego/pubsub"
)

// DashboardAPI defines the methods used by the dev command.
type DashboardAPI interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	BuildAndRun(ctx context.Context) error
	Rebuild(ctx context.Context) error
	PublishEvent(eventType string, data any)
	GetPubSub() *pubsub.InProcBroker
	GetBuilder() BuilderAPI
	GetRunner() RunnerAPI
}

package devserver

import (
	"context"

	"github.com/spcent/plumego/pubsub"
)

// BuilderAPI defines the methods used by the dev command.
type BuilderAPI interface {
	SetCustomBuild(cmd string, args []string)
}

// RunnerAPI defines the methods used by the dev command.
type RunnerAPI interface {
	SetCustomCommand(cmd string, args []string)
	SetOutputPassthrough(enabled bool)
}

// DashboardAPI defines the methods used by the dev command.
type DashboardAPI interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	BuildAndRun(ctx context.Context) error
	Rebuild(ctx context.Context) error
	PublishEvent(eventType string, data any)
	GetPubSub() *pubsub.InProcPubSub
	GetBuilder() BuilderAPI
	GetRunner() RunnerAPI
}

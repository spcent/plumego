package webhookout

import (
	"context"
	"time"
)

type Store interface {
	CreateTarget(ctx context.Context, t Target) (Target, error)
	UpdateTarget(ctx context.Context, id string, patch TargetPatch) (Target, error)
	GetTarget(ctx context.Context, id string) (Target, bool)
	ListTargets(ctx context.Context, filter TargetFilter) ([]Target, error)

	CreateDelivery(ctx context.Context, d Delivery) (Delivery, error)
	UpdateDelivery(ctx context.Context, id string, patch DeliveryPatch) (Delivery, error)
	GetDelivery(ctx context.Context, id string) (Delivery, bool)
	ListDeliveries(ctx context.Context, filter DeliveryFilter) ([]Delivery, error)
}

type DeliveryPatch struct {
	Attempt *int
	Status  *DeliveryStatus
	NextAt  *time.Time

	LastHTTPStatus  *int
	LastError       *string
	LastDurationMs  *int
	LastRespSnippet *string

	// optional: update payload (rare)
	PayloadJSON *[]byte
}

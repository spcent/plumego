package mongo

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LoopLeaseCoordinator struct {
	backend          loopLeaseBackend
	ownerID          string
	ttl              time.Duration
	operationTimeout time.Duration
	now              func() time.Time
}

type loopLeaseBackend interface {
	TryAcquireLease(ctx context.Context, input loopLeaseAcquireInput) (bool, error)
}

type loopLeaseAcquireInput struct {
	LoopName  string
	OwnerID   string
	Now       time.Time
	ExpiresAt time.Time
}

type loopLeaseBackendFunc func(context.Context, loopLeaseAcquireInput) (bool, error)

func (fn loopLeaseBackendFunc) TryAcquireLease(ctx context.Context, input loopLeaseAcquireInput) (bool, error) {
	return fn(ctx, input)
}

type mongoLoopLeaseBackend struct {
	collection *mongo.Collection
}

type LoopLeaseOption func(*LoopLeaseCoordinator)

func WithLoopLeaseOperationTimeout(timeout time.Duration) LoopLeaseOption {
	return func(c *LoopLeaseCoordinator) {
		if timeout > 0 {
			c.operationTimeout = timeout
		}
	}
}

func WithLoopLeaseClock(now func() time.Time) LoopLeaseOption {
	return func(c *LoopLeaseCoordinator) {
		if now != nil {
			c.now = now
		}
	}
}

func NewLoopLeaseCoordinator(collection *mongo.Collection, ownerID string, ttl time.Duration, opts ...LoopLeaseOption) (*LoopLeaseCoordinator, error) {
	if collection == nil {
		return nil, errors.New("mongo loop lease collection is required")
	}
	return newLoopLeaseCoordinator(mongoLoopLeaseBackend{collection: collection}, ownerID, ttl, opts...)
}

func newLoopLeaseCoordinator(backend loopLeaseBackend, ownerID string, ttl time.Duration, opts ...LoopLeaseOption) (*LoopLeaseCoordinator, error) {
	if backend == nil {
		return nil, errors.New("loop lease backend is required")
	}
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, errors.New("loop lease owner is required")
	}
	if ttl <= 0 {
		return nil, errors.New("loop lease ttl must be greater than zero")
	}
	coordinator := &LoopLeaseCoordinator{
		backend:          backend,
		ownerID:          ownerID,
		ttl:              ttl,
		operationTimeout: defaultOperationTimeout,
		now:              func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(coordinator)
	}
	return coordinator, nil
}

func (s *Store) LoopLeaseCoordinator(ownerID string, ttl time.Duration) (*LoopLeaseCoordinator, error) {
	if s == nil {
		return nil, errors.New("mongo store is required")
	}
	return NewLoopLeaseCoordinator(
		s.collections.LoopLeases,
		ownerID,
		ttl,
		WithLoopLeaseOperationTimeout(s.operationTimeout),
		WithLoopLeaseClock(s.now),
	)
}

func (c *LoopLeaseCoordinator) TryAcquire(ctx context.Context, loopName string) (func(), bool, error) {
	if c == nil {
		return nil, false, errors.New("loop lease coordinator is required")
	}
	loopName = strings.TrimSpace(loopName)
	if loopName == "" {
		return nil, false, errors.New("loop lease name is required")
	}

	ctx, cancel := c.operationContext(ctx)
	defer cancel()

	now := c.now().UTC()
	acquired, err := c.backend.TryAcquireLease(ctx, loopLeaseAcquireInput{
		LoopName:  loopName,
		OwnerID:   c.ownerID,
		Now:       now,
		ExpiresAt: now.Add(c.ttl),
	})
	if err != nil || !acquired {
		return nil, acquired, err
	}
	return func() {}, true, nil
}

func (c *LoopLeaseCoordinator) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if c.operationTimeout <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, c.operationTimeout)
}

func (b mongoLoopLeaseBackend) TryAcquireLease(ctx context.Context, input loopLeaseAcquireInput) (bool, error) {
	filter := bson.D{
		{Key: "_id", Value: input.LoopName},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "owner_id", Value: input.OwnerID}},
			bson.D{{Key: "expires_at", Value: bson.D{{Key: "$lte", Value: input.Now}}}},
		}},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "owner_id", Value: input.OwnerID},
			{Key: "expires_at", Value: input.ExpiresAt},
			{Key: "updated_at", Value: input.Now},
		}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: input.LoopName},
			{Key: "created_at", Value: input.Now},
		}},
	}
	result, err := b.collection.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if mongo.IsDuplicateKeyError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result.MatchedCount > 0 || result.UpsertedCount > 0, nil
}

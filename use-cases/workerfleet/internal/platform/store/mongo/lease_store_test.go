package mongo

import (
	"context"
	"testing"
	"time"
)

func TestLeaseCoordinatorAcquireRenewSteal(t *testing.T) {
	base := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	now := base
	backend := &memoryLoopLeaseBackend{docs: map[string]memoryLoopLeaseDoc{}}

	ownerA, err := newLoopLeaseCoordinator(backend, "owner-a", time.Minute, WithLoopLeaseClock(func() time.Time {
		return now
	}))
	if err != nil {
		t.Fatalf("new owner-a coordinator: %v", err)
	}
	ownerB, err := newLoopLeaseCoordinator(backend, "owner-b", time.Minute, WithLoopLeaseClock(func() time.Time {
		return now
	}))
	if err != nil {
		t.Fatalf("new owner-b coordinator: %v", err)
	}

	_, acquired, err := ownerA.TryAcquire(context.Background(), "status_sweep")
	if err != nil {
		t.Fatalf("owner-a acquire: %v", err)
	}
	if !acquired {
		t.Fatalf("owner-a should acquire an empty lease")
	}

	now = base.Add(30 * time.Second)
	_, acquired, err = ownerB.TryAcquire(context.Background(), "status_sweep")
	if err != nil {
		t.Fatalf("owner-b acquire active lease: %v", err)
	}
	if acquired {
		t.Fatalf("owner-b should not acquire an unexpired owner-a lease")
	}

	_, acquired, err = ownerA.TryAcquire(context.Background(), "status_sweep")
	if err != nil {
		t.Fatalf("owner-a renew: %v", err)
	}
	if !acquired {
		t.Fatalf("owner-a should renew its own lease")
	}

	now = base.Add(80 * time.Second)
	_, acquired, err = ownerB.TryAcquire(context.Background(), "status_sweep")
	if err != nil {
		t.Fatalf("owner-b acquire renewed lease: %v", err)
	}
	if acquired {
		t.Fatalf("owner-b should not steal before renewed lease expiry")
	}

	now = base.Add(91 * time.Second)
	_, acquired, err = ownerB.TryAcquire(context.Background(), "status_sweep")
	if err != nil {
		t.Fatalf("owner-b steal expired lease: %v", err)
	}
	if !acquired {
		t.Fatalf("owner-b should steal after lease expiry")
	}
}

type memoryLoopLeaseDoc struct {
	OwnerID   string
	ExpiresAt time.Time
}

type memoryLoopLeaseBackend struct {
	docs map[string]memoryLoopLeaseDoc
}

func (b *memoryLoopLeaseBackend) TryAcquireLease(ctx context.Context, input loopLeaseAcquireInput) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	doc, exists := b.docs[input.LoopName]
	if exists && doc.OwnerID != input.OwnerID && input.Now.Before(doc.ExpiresAt) {
		return false, nil
	}
	b.docs[input.LoopName] = memoryLoopLeaseDoc{
		OwnerID:   input.OwnerID,
		ExpiresAt: input.ExpiresAt,
	}
	return true, nil
}

package session

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeKV is an in-memory KV for tests; TTLs are honored.
type fakeKV struct {
	mu    sync.Mutex
	data  map[string][]byte
	exp   map[string]time.Time
	clock func() time.Time
}

func newFakeKV() *fakeKV {
	return &fakeKV{
		data:  make(map[string][]byte),
		exp:   make(map[string]time.Time),
		clock: time.Now,
	}
}

func (f *fakeKV) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = append([]byte(nil), value...)
	if ttl > 0 {
		f.exp[key] = f.clock().Add(ttl)
	} else {
		delete(f.exp, key)
	}
	return nil
}

func (f *fakeKV) Get(_ context.Context, key string) ([]byte, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[key]
	if !ok {
		return nil, false, nil
	}
	if deadline, has := f.exp[key]; has && f.clock().After(deadline) {
		delete(f.data, key)
		delete(f.exp, key)
		return nil, false, nil
	}
	return append([]byte(nil), v...), true, nil
}

func (f *fakeKV) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, key)
	delete(f.exp, key)
	return nil
}

func TestIssueAndRotate(t *testing.T) {
	store := NewStore(newFakeKV(), time.Hour)
	ctx := context.Background()

	tok, err := store.Issue(ctx, "user-1", "tenant-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	rec, next, err := store.Rotate(ctx, tok)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if rec.UserID != "user-1" || rec.TenantID != "tenant-1" {
		t.Fatalf("record = %+v", rec)
	}
	if next == "" || next == tok {
		t.Fatal("rotation must return a fresh token")
	}

	// The replacement stays in the same family and rotates again.
	rec2, _, err := store.Rotate(ctx, next)
	if err != nil {
		t.Fatalf("second rotate: %v", err)
	}
	if rec2.FamilyID != rec.FamilyID {
		t.Fatal("rotation must preserve the token family")
	}
}

func TestRotateUnknownToken(t *testing.T) {
	store := NewStore(newFakeKV(), time.Hour)
	if _, _, err := store.Rotate(context.Background(), "never-issued"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestReuseRevokesFamily(t *testing.T) {
	store := NewStore(newFakeKV(), time.Hour)
	ctx := context.Background()

	tok, err := store.Issue(ctx, "u", "t")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	_, next, err := store.Rotate(ctx, tok)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}

	// Reusing the consumed token is theft evidence → ErrReused.
	if _, _, err := store.Rotate(ctx, tok); !errors.Is(err, ErrReused) {
		t.Fatalf("reuse: expected ErrReused, got %v", err)
	}
	// The whole family is revoked: the still-unrotated successor dies too.
	if _, _, err := store.Rotate(ctx, next); !errors.Is(err, ErrReused) {
		t.Fatalf("successor after revoke: expected ErrReused, got %v", err)
	}
}

func TestExpiredTokenInvalid(t *testing.T) {
	kv := newFakeKV()
	now := time.Now()
	kv.clock = func() time.Time { return now }
	store := NewStore(kv, time.Minute)
	ctx := context.Background()

	tok, err := store.Issue(ctx, "u", "t")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	now = now.Add(2 * time.Minute) // past TTL
	if _, _, err := store.Rotate(ctx, tok); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken after expiry, got %v", err)
	}
}

func TestTokensAreOpaqueAndHashed(t *testing.T) {
	kv := newFakeKV()
	store := NewStore(kv, time.Hour)
	tok, err := store.Issue(context.Background(), "u", "t")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	// The raw token must not appear as a storage key (only its hash may).
	kv.mu.Lock()
	defer kv.mu.Unlock()
	for key := range kv.data {
		if key == keyToken+tok {
			t.Fatal("raw refresh token stored as key; must store hash only")
		}
	}
}

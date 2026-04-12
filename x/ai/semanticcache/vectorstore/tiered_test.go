package vectorstore

import (
	"context"
	"errors"
	"testing"
)

type failingBackend struct {
	name        string
	addErr      error
	addBatchErr error
}

func (b *failingBackend) Add(context.Context, *Entry) error { return b.addErr }

func (b *failingBackend) AddBatch(context.Context, []*Entry) error { return b.addBatchErr }

func (b *failingBackend) Search(context.Context, []float64, int, float64) ([]*SearchResult, error) {
	return nil, nil
}

func (b *failingBackend) Get(context.Context, string) (*Entry, error) {
	return nil, errors.New("not found")
}

func (b *failingBackend) Delete(context.Context, string) error { return nil }

func (b *failingBackend) DeleteBatch(context.Context, []string) error { return nil }

func (b *failingBackend) Update(context.Context, *Entry) error { return nil }

func (b *failingBackend) Count(context.Context) (int64, error) { return 0, nil }

func (b *failingBackend) Clear(context.Context) error { return nil }

func (b *failingBackend) Close() error { return nil }

func (b *failingBackend) Name() string { return b.name }

func (b *failingBackend) Stats(context.Context) (*BackendStats, error) {
	return &BackendStats{Name: b.name}, nil
}

func TestTieredBackendReportsBackgroundAddFailures(t *testing.T) {
	ctx := context.Background()
	l1 := NewMemoryBackend(MemoryConfig{Dimensions: 2, MaxSize: 10})
	l2Err := errors.New("l2 add failed")
	l3Err := errors.New("l3 add failed")
	var got []TieredBackgroundError

	backend, err := NewTieredBackend(TieredConfig{
		L1: l1,
		L2: &failingBackend{name: "l2", addErr: l2Err},
		L3: &failingBackend{name: "l3", addErr: l3Err},
		OnError: func(entry TieredBackgroundError) {
			got = append(got, entry)
		},
		Dimensions: 2,
	})
	if err != nil {
		t.Fatalf("NewTieredBackend: %v", err)
	}

	entry := &Entry{ID: "e1", Vector: []float64{1, 0}}
	if err := backend.Add(ctx, entry); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("background error count = %d, want 2", len(got))
	}
	if got[0].Tier != "l2" || got[0].Operation != "add" || !errors.Is(got[0].Err, l2Err) {
		t.Fatalf("unexpected first background error: %+v", got[0])
	}
	if got[1].Tier != "l3" || got[1].Operation != "add" || !errors.Is(got[1].Err, l3Err) {
		t.Fatalf("unexpected second background error: %+v", got[1])
	}
}

func TestTieredBackendReportsBackgroundBatchFailures(t *testing.T) {
	ctx := context.Background()
	l1 := NewMemoryBackend(MemoryConfig{Dimensions: 2, MaxSize: 10})
	l2Err := errors.New("l2 add batch failed")
	var got []TieredBackgroundError

	backend, err := NewTieredBackend(TieredConfig{
		L1: l1,
		L2: &failingBackend{name: "l2", addBatchErr: l2Err},
		OnError: func(entry TieredBackgroundError) {
			got = append(got, entry)
		},
		Dimensions: 2,
	})
	if err != nil {
		t.Fatalf("NewTieredBackend: %v", err)
	}

	entries := []*Entry{
		{ID: "e1", Vector: []float64{1, 0}},
		{ID: "e2", Vector: []float64{0, 1}},
	}
	if err := backend.AddBatch(ctx, entries); err != nil {
		t.Fatalf("AddBatch: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("background error count = %d, want 1", len(got))
	}
	if got[0].Tier != "l2" || got[0].Operation != "add_batch" || !errors.Is(got[0].Err, l2Err) {
		t.Fatalf("unexpected background error: %+v", got[0])
	}
}

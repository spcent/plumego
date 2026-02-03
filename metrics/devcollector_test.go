package metrics

import (
	"context"
	"testing"
	"time"
)

func TestDevCollectorSnapshot(t *testing.T) {
	cfg := DevCollectorConfig{
		Window:     time.Minute,
		MaxSamples: 2,
		MaxSeries:  10,
		MaxValues:  100,
		DBMaxSlow:  5,
		DBSlowMS:   1,
	}

	collector := NewDevCollector(cfg)
	ctx := context.Background()

	collector.ObserveHTTP(ctx, "GET", "/hello", 200, 128, 10*time.Millisecond)
	collector.ObserveHTTP(ctx, "GET", "/hello", 404, 64, 20*time.Millisecond)
	collector.ObserveHTTP(ctx, "POST", "/submit", 500, 32, 30*time.Millisecond)

	snapshot := collector.Snapshot()

	if snapshot.Total.Count != 3 {
		t.Fatalf("expected total count 3, got %d", snapshot.Total.Count)
	}
	if snapshot.Total.ErrorCount != 2 {
		t.Fatalf("expected error count 2, got %d", snapshot.Total.ErrorCount)
	}
	if snapshot.Total.Duration.Mean < 19.9 || snapshot.Total.Duration.Mean > 20.1 {
		t.Fatalf("expected mean duration ~20ms, got %.2f", snapshot.Total.Duration.Mean)
	}
	if len(snapshot.Recent) != 2 {
		t.Fatalf("expected recent samples limited to 2, got %d", len(snapshot.Recent))
	}

	collector.ObserveDB(ctx, "query", "sqlite", "select 1", 1, 5*time.Millisecond, nil)
	dbSnapshot := collector.DBSnapshot()
	if dbSnapshot.Total.Count != 1 {
		t.Fatalf("expected db total count 1, got %d", dbSnapshot.Total.Count)
	}
	if len(dbSnapshot.Slow) == 0 {
		t.Fatalf("expected slow query sample")
	}
}

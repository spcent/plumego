package metrics

import (
	"context"
	"testing"
	"time"
)

func TestRedactSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single-quoted string",
			input:    "SELECT * FROM users WHERE email = 'john@example.com'",
			expected: "SELECT * FROM users WHERE email = '?'",
		},
		{
			name:     "double-quoted string",
			input:    `SELECT * FROM users WHERE name = "John"`,
			expected: `SELECT * FROM users WHERE name = "?"`,
		},
		{
			name:     "backtick identifier",
			input:    "SELECT `password_hash` FROM `users`",
			expected: "SELECT `?` FROM `?`",
		},
		{
			name:     "single-line comment",
			input:    "SELECT * FROM users -- admin password is s3cr3t",
			expected: "SELECT * FROM users -- ?",
		},
		{
			name:     "block comment",
			input:    "SELECT /* secret_column */ * FROM users",
			expected: "SELECT /* ? */ * FROM users",
		},
		{
			name:     "numeric literal",
			input:    "SELECT * FROM users WHERE id = 12345",
			expected: "SELECT * FROM users WHERE id = ?",
		},
		{
			name:     "mixed quotes and numbers",
			input:    `SELECT * FROM users WHERE name = "John" AND age = 30 AND email = 'j@x.com'`,
			expected: `SELECT * FROM users WHERE name = "?" AND age = ? AND email = '?'`,
		},
		{
			name:     "escaped single quote",
			input:    "SELECT * FROM users WHERE name = 'O\\'Brien'",
			expected: "SELECT * FROM users WHERE name = '?'",
		},
		{
			name:     "empty query",
			input:    "",
			expected: "",
		},
		{
			name:     "comment at end with newline",
			input:    "SELECT 1 -- comment\nSELECT 2",
			expected: "SELECT ? -- ? SELECT ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactSQL(tt.input)
			if got != tt.expected {
				t.Errorf("redactSQL(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.expected)
			}
		})
	}
}

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

func TestDevCollectorDBTableSeries(t *testing.T) {
	cfg := DevCollectorConfig{
		Window:      time.Minute,
		MaxSamples:  10,
		MaxSeries:   10,
		MaxValues:   100,
		DBMaxSlow:   10,
		DBSlowMS:    1000, // high threshold so we don't trigger slow query
		DBMaxSeries: 10,
	}

	collector := NewDevCollector(cfg)
	ctx := context.Background()

	// Record queries against different tables
	collector.ObserveDB(ctx, "query", "postgres", "SELECT * FROM users WHERE id = 1", 1, 10*time.Millisecond, nil)
	collector.ObserveDB(ctx, "query", "postgres", "SELECT * FROM users WHERE active = true", 5, 20*time.Millisecond, nil)
	collector.ObserveDB(ctx, "exec", "postgres", "INSERT INTO orders (user_id, total) VALUES (1, 100)", 1, 15*time.Millisecond, nil)
	collector.ObserveDB(ctx, "exec", "postgres", "UPDATE products SET stock = 0 WHERE id = 5", 1, 25*time.Millisecond, nil)

	snap := collector.DBSnapshot()

	// Verify total
	if snap.Total.Count != 4 {
		t.Fatalf("expected total count 4, got %d", snap.Total.Count)
	}

	// Verify table series exist
	if len(snap.Tables) != 3 {
		t.Fatalf("expected 3 table series (users, orders, products), got %d", len(snap.Tables))
	}

	// Tables should be sorted by count descending
	// "users" has 2 queries, "orders" and "products" have 1 each
	if snap.Tables[0].Table != "users" {
		t.Errorf("expected first table to be 'users', got %q", snap.Tables[0].Table)
	}
	if snap.Tables[0].Count != 2 {
		t.Errorf("expected users count 2, got %d", snap.Tables[0].Count)
	}
}

func TestDevCollectorDBSlowQueryTable(t *testing.T) {
	cfg := DevCollectorConfig{
		Window:      time.Minute,
		MaxSamples:  10,
		MaxSeries:   10,
		MaxValues:   100,
		DBMaxSlow:   10,
		DBSlowMS:    1, // 1ms threshold
		DBMaxSeries: 10,
	}

	collector := NewDevCollector(cfg)
	ctx := context.Background()

	collector.ObserveDB(ctx, "query", "postgres", "SELECT * FROM users WHERE id = 1", 1, 5*time.Millisecond, nil)

	snap := collector.DBSnapshot()
	if len(snap.Slow) == 0 {
		t.Fatalf("expected slow query sample")
	}

	if snap.Slow[0].Table != "users" {
		t.Errorf("expected slow query table 'users', got %q", snap.Slow[0].Table)
	}
}

func TestDevCollectorDBTableSeriesClear(t *testing.T) {
	cfg := DefaultDevCollectorConfig()
	collector := NewDevCollector(cfg)
	ctx := context.Background()

	collector.ObserveDB(ctx, "query", "postgres", "SELECT * FROM users", 1, 10*time.Millisecond, nil)

	snap := collector.DBSnapshot()
	if len(snap.Tables) == 0 {
		t.Fatalf("expected table series before clear")
	}

	collector.Clear()

	snap = collector.DBSnapshot()
	if len(snap.Tables) != 0 {
		t.Fatalf("expected no table series after clear, got %d", len(snap.Tables))
	}
}

func TestDevCollectorDBTableNoQuery(t *testing.T) {
	cfg := DefaultDevCollectorConfig()
	collector := NewDevCollector(cfg)
	ctx := context.Background()

	// Operations without SQL (ping, connect, close) should not create table series
	collector.ObserveDB(ctx, "ping", "postgres", "", 0, 2*time.Millisecond, nil)
	collector.ObserveDB(ctx, "connect", "postgres", "", 0, 5*time.Millisecond, nil)

	snap := collector.DBSnapshot()
	if len(snap.Tables) != 0 {
		t.Fatalf("expected no table series for operations without queries, got %d", len(snap.Tables))
	}
}

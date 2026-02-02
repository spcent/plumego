package db

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/metrics"
)

func TestSlowQueryDetector_Check(t *testing.T) {
	detector := NewSlowQueryDetector(
		WithSlowQueryThreshold(100 * time.Millisecond),
	)

	// Fast query - should not be detected
	isSlow := detector.Check("query", "postgres", "SELECT * FROM users", 50*time.Millisecond, nil)
	if isSlow {
		t.Error("expected fast query not to be detected as slow")
	}

	// Slow query - should be detected
	isSlow = detector.Check("query", "postgres", "SELECT * FROM huge_table", 200*time.Millisecond, nil)
	if !isSlow {
		t.Error("expected slow query to be detected")
	}

	// Verify it was recorded
	slowQueries := detector.GetSlowQueries()
	if len(slowQueries) != 1 {
		t.Fatalf("expected 1 slow query, got %d", len(slowQueries))
	}

	if slowQueries[0].Query != "SELECT * FROM huge_table" {
		t.Errorf("unexpected query: %s", slowQueries[0].Query)
	}
}

func TestSlowQueryDetector_MaxRecords(t *testing.T) {
	detector := NewSlowQueryDetector(
		WithSlowQueryThreshold(50*time.Millisecond),
		WithSlowQueryMaxRecords(5),
	)

	// Add 10 slow queries
	for i := 0; i < 10; i++ {
		detector.Check("query", "postgres", "SELECT * FROM users", 100*time.Millisecond, nil)
	}

	// Should only keep the last 5
	slowQueries := detector.GetSlowQueries()
	if len(slowQueries) != 5 {
		t.Errorf("expected 5 slow queries, got %d", len(slowQueries))
	}

	// But total detected should be 10
	if detector.GetTotalDetected() != 10 {
		t.Errorf("expected total detected to be 10, got %d", detector.GetTotalDetected())
	}
}

func TestSlowQueryDetector_Callback(t *testing.T) {
	callbackCalled := false
	var recordedQuery SlowQuery

	detector := NewSlowQueryDetector(
		WithSlowQueryThreshold(50*time.Millisecond),
		WithSlowQueryCallback(func(sq SlowQuery) {
			callbackCalled = true
			recordedQuery = sq
		}),
	)

	detector.Check("query", "postgres", "SELECT * FROM users", 100*time.Millisecond, nil)

	if !callbackCalled {
		t.Error("expected callback to be called")
	}

	if recordedQuery.Query != "SELECT * FROM users" {
		t.Errorf("unexpected query in callback: %s", recordedQuery.Query)
	}
}

func TestSlowQueryDetector_GetRecentSlowQueries(t *testing.T) {
	detector := NewSlowQueryDetector(
		WithSlowQueryThreshold(50 * time.Millisecond),
	)

	// Add 10 slow queries
	for i := 0; i < 10; i++ {
		query := "SELECT * FROM users WHERE id = " + string(rune('0'+i))
		detector.Check("query", "postgres", query, 100*time.Millisecond, nil)
	}

	// Get recent 3
	recent := detector.GetRecentSlowQueries(3)
	if len(recent) != 3 {
		t.Fatalf("expected 3 recent queries, got %d", len(recent))
	}

	// Should be the last 3 (7, 8, 9)
	for i, sq := range recent {
		expectedID := '7' + rune(i)
		if !strings.Contains(sq.Query, string(expectedID)) {
			t.Errorf("unexpected query at index %d: %s", i, sq.Query)
		}
	}
}

func TestSlowQueryDetector_Clear(t *testing.T) {
	detector := NewSlowQueryDetector(
		WithSlowQueryThreshold(50 * time.Millisecond),
	)

	// Add some slow queries
	detector.Check("query", "postgres", "SELECT * FROM users", 100*time.Millisecond, nil)
	detector.Check("query", "postgres", "SELECT * FROM posts", 150*time.Millisecond, nil)

	if detector.GetTotalDetected() != 2 {
		t.Errorf("expected 2 total detected, got %d", detector.GetTotalDetected())
	}

	// Clear
	detector.Clear()

	if detector.GetTotalDetected() != 0 {
		t.Errorf("expected 0 total detected after clear, got %d", detector.GetTotalDetected())
	}

	slowQueries := detector.GetSlowQueries()
	if len(slowQueries) != 0 {
		t.Errorf("expected 0 slow queries after clear, got %d", len(slowQueries))
	}
}

func TestSlowQueryDetector_Summary(t *testing.T) {
	detector := NewSlowQueryDetector(
		WithSlowQueryThreshold(100 * time.Millisecond),
	)

	// No slow queries
	summary := detector.Summary()
	if !strings.Contains(summary, "No slow queries detected") {
		t.Error("expected summary to indicate no slow queries")
	}

	// Add a slow query
	detector.Check("query", "postgres", "SELECT * FROM users", 200*time.Millisecond, nil)

	summary = detector.Summary()
	if !strings.Contains(summary, "Total detected: 1") {
		t.Error("expected summary to show 1 detected query")
	}
	if !strings.Contains(summary, "SELECT * FROM users") {
		t.Error("expected summary to contain the query")
	}
}

func TestSlowQueryDetector_WithError(t *testing.T) {
	detector := NewSlowQueryDetector(
		WithSlowQueryThreshold(50 * time.Millisecond),
	)

	testErr := errors.New("query failed")
	detector.Check("query", "postgres", "SELECT * FROM invalid", 100*time.Millisecond, testErr)

	slowQueries := detector.GetSlowQueries()
	if len(slowQueries) != 1 {
		t.Fatalf("expected 1 slow query, got %d", len(slowQueries))
	}

	if slowQueries[0].Error == nil {
		t.Error("expected error to be recorded")
	}
	if slowQueries[0].Error.Error() != testErr.Error() {
		t.Errorf("unexpected error: %v", slowQueries[0].Error)
	}
}

func TestSlowQueryDetector_DefaultThreshold(t *testing.T) {
	detector := NewSlowQueryDetector()

	// Default threshold is 1 second
	isSlow := detector.Check("query", "postgres", "SELECT * FROM users", 500*time.Millisecond, nil)
	if isSlow {
		t.Error("expected query under 1s not to be slow with default threshold")
	}

	isSlow = detector.Check("query", "postgres", "SELECT * FROM huge_table", 1500*time.Millisecond, nil)
	if !isSlow {
		t.Error("expected query over 1s to be slow with default threshold")
	}
}

func TestSlowQueryDetector_Concurrency(t *testing.T) {
	detector := NewSlowQueryDetector(
		WithSlowQueryThreshold(50*time.Millisecond),
		WithSlowQueryMaxRecords(1000),
	)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				detector.Check("query", "postgres", "SELECT * FROM users", 100*time.Millisecond, nil)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if detector.GetTotalDetected() != 100 {
		t.Errorf("expected 100 total detected, got %d", detector.GetTotalDetected())
	}
}

func TestMetricsCollectorWithSlowQueryDetection(t *testing.T) {
	callbackCalled := false
	base := metrics.NewMockCollector()

	collector := NewMetricsCollectorWithSlowQueryDetection(
		base,
		WithSlowQueryThreshold(100*time.Millisecond),
		WithSlowQueryCallback(func(sq SlowQuery) {
			callbackCalled = true
		}),
	)

	// Record a slow query
	collector.ObserveDB(nil, "query", "postgres", "SELECT * FROM users", 10, 200*time.Millisecond, nil)

	// Verify it was forwarded to base collector
	if base.DBCallCount() != 1 {
		t.Errorf("expected base collector to receive 1 call, got %d", base.DBCallCount())
	}

	// Verify slow query was detected
	if !callbackCalled {
		t.Error("expected slow query callback to be called")
	}

	detector := collector.GetSlowQueryDetector()
	if detector.GetTotalDetected() != 1 {
		t.Errorf("expected 1 slow query detected, got %d", detector.GetTotalDetected())
	}
}

func TestTruncateQuery(t *testing.T) {
	tests := []struct {
		query  string
		maxLen int
		want   string
	}{
		{"SELECT * FROM users", 50, "SELECT * FROM users"},
		{"SELECT * FROM users WHERE name = 'very long name that exceeds limit'", 30, "SELECT * FROM users WHERE name..."},
		{"", 10, ""},
		{"SHORT", 100, "SHORT"},
	}

	for _, tt := range tests {
		got := truncateQuery(tt.query, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateQuery(%q, %d) = %q, want %q", tt.query, tt.maxLen, got, tt.want)
		}
	}
}

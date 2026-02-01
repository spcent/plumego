package db

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDBStatsAggregator_RecordQuery(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	// Record some queries
	agg.RecordQuery("query", "SELECT * FROM users", 50*time.Millisecond, nil)
	agg.RecordQuery("exec", "INSERT INTO users (name) VALUES ('test')", 30*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM posts", 150*time.Millisecond, nil) // slow query

	// Check global stats
	global := agg.GetGlobalStats()
	if global.TotalQueries != 3 {
		t.Errorf("expected 3 total queries, got %d", global.TotalQueries)
	}
	if global.SuccessQueries != 3 {
		t.Errorf("expected 3 success queries, got %d", global.SuccessQueries)
	}
	if global.SlowQueries != 1 {
		t.Errorf("expected 1 slow query, got %d", global.SlowQueries)
	}
}

func TestDBStatsAggregator_OperationStats(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	agg.RecordQuery("query", "SELECT * FROM users", 50*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM posts", 60*time.Millisecond, nil)
	agg.RecordQuery("exec", "INSERT INTO users (name) VALUES ('test')", 30*time.Millisecond, nil)

	// Check query operation stats
	queryStats, ok := agg.GetOperationStats("query")
	if !ok {
		t.Fatal("expected query operation stats to exist")
	}
	if queryStats.TotalQueries != 2 {
		t.Errorf("expected 2 query operations, got %d", queryStats.TotalQueries)
	}

	// Check exec operation stats
	execStats, ok := agg.GetOperationStats("exec")
	if !ok {
		t.Fatal("expected exec operation stats to exist")
	}
	if execStats.TotalQueries != 1 {
		t.Errorf("expected 1 exec operation, got %d", execStats.TotalQueries)
	}

	// Check all operation stats
	allStats := agg.GetAllOperationStats()
	if len(allStats) != 2 {
		t.Errorf("expected 2 operation types, got %d", len(allStats))
	}
}

func TestDBStatsAggregator_TableStats(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	agg.RecordQuery("query", "SELECT * FROM users", 50*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM users WHERE id = 1", 60*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM posts", 70*time.Millisecond, nil)

	// Check users table stats
	usersStats, ok := agg.GetTableStats("users")
	if !ok {
		t.Fatal("expected users table stats to exist")
	}
	if usersStats.TotalQueries != 2 {
		t.Errorf("expected 2 queries on users table, got %d", usersStats.TotalQueries)
	}

	// Check posts table stats
	postsStats, ok := agg.GetTableStats("posts")
	if !ok {
		t.Fatal("expected posts table stats to exist")
	}
	if postsStats.TotalQueries != 1 {
		t.Errorf("expected 1 query on posts table, got %d", postsStats.TotalQueries)
	}

	// Check all table stats
	allStats := agg.GetAllTableStats()
	if len(allStats) != 2 {
		t.Errorf("expected 2 tables, got %d", len(allStats))
	}
}

func TestDBStatsAggregator_ErrorTracking(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	agg.RecordQuery("query", "SELECT * FROM users", 50*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM invalid", 60*time.Millisecond, errors.New("table not found"))

	global := agg.GetGlobalStats()
	if global.SuccessQueries != 1 {
		t.Errorf("expected 1 success query, got %d", global.SuccessQueries)
	}
	if global.FailedQueries != 1 {
		t.Errorf("expected 1 failed query, got %d", global.FailedQueries)
	}
}

func TestDBStatsAggregator_DurationTracking(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	agg.RecordQuery("query", "SELECT * FROM users", 30*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM posts", 70*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM comments", 50*time.Millisecond, nil)

	global := agg.GetGlobalStats()

	if global.MinDuration != 30*time.Millisecond {
		t.Errorf("expected min duration 30ms, got %s", global.MinDuration)
	}
	if global.MaxDuration != 70*time.Millisecond {
		t.Errorf("expected max duration 70ms, got %s", global.MaxDuration)
	}

	expectedAvg := 50 * time.Millisecond // (30 + 70 + 50) / 3
	if global.AvgDuration != expectedAvg {
		t.Errorf("expected avg duration %s, got %s", expectedAvg, global.AvgDuration)
	}
}

func TestDBStatsAggregator_QueryTypeTracking(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	agg.RecordQuery("query", "SELECT * FROM users", 50*time.Millisecond, nil)
	agg.RecordQuery("exec", "INSERT INTO users (name) VALUES ('test')", 30*time.Millisecond, nil)
	agg.RecordQuery("exec", "UPDATE users SET name = 'test2' WHERE id = 1", 40*time.Millisecond, nil)
	agg.RecordQuery("exec", "DELETE FROM users WHERE id = 1", 20*time.Millisecond, nil)

	global := agg.GetGlobalStats()

	if global.SelectQueries != 1 {
		t.Errorf("expected 1 SELECT query, got %d", global.SelectQueries)
	}
	if global.InsertQueries != 1 {
		t.Errorf("expected 1 INSERT query, got %d", global.InsertQueries)
	}
	if global.UpdateQueries != 1 {
		t.Errorf("expected 1 UPDATE query, got %d", global.UpdateQueries)
	}
	if global.DeleteQueries != 1 {
		t.Errorf("expected 1 DELETE query, got %d", global.DeleteQueries)
	}
}

func TestDBStatsAggregator_TopSlowTables(t *testing.T) {
	agg := NewDBStatsAggregator(50 * time.Millisecond)

	// Add some slow queries to different tables
	agg.RecordQuery("query", "SELECT * FROM users", 100*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM users", 150*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM posts", 80*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM comments", 200*time.Millisecond, nil)

	topTables := agg.GetTopSlowTables(2)

	if len(topTables) != 2 {
		t.Fatalf("expected 2 top tables, got %d", len(topTables))
	}

	// First should be users (2 slow queries)
	if topTables[0].Table != "users" {
		t.Errorf("expected first table to be 'users', got %q", topTables[0].Table)
	}
	if topTables[0].SlowQueries != 2 {
		t.Errorf("expected 2 slow queries for users, got %d", topTables[0].SlowQueries)
	}

	// Second should be comments or posts (1 slow query each)
	if topTables[1].SlowQueries != 1 {
		t.Errorf("expected 1 slow query for second table, got %d", topTables[1].SlowQueries)
	}
}

func TestDBStatsAggregator_Reset(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	agg.RecordQuery("query", "SELECT * FROM users", 50*time.Millisecond, nil)
	agg.RecordQuery("exec", "INSERT INTO users (name) VALUES ('test')", 30*time.Millisecond, nil)

	// Verify stats exist
	global := agg.GetGlobalStats()
	if global.TotalQueries != 2 {
		t.Errorf("expected 2 total queries, got %d", global.TotalQueries)
	}

	// Reset
	agg.Reset()

	// Verify stats are cleared
	global = agg.GetGlobalStats()
	if global.TotalQueries != 0 {
		t.Errorf("expected 0 total queries after reset, got %d", global.TotalQueries)
	}

	allOps := agg.GetAllOperationStats()
	if len(allOps) != 0 {
		t.Errorf("expected 0 operation stats after reset, got %d", len(allOps))
	}

	allTables := agg.GetAllTableStats()
	if len(allTables) != 0 {
		t.Errorf("expected 0 table stats after reset, got %d", len(allTables))
	}
}

func TestDBStatsAggregator_Summary(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	agg.RecordQuery("query", "SELECT * FROM users", 50*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM posts", 150*time.Millisecond, nil)
	agg.RecordQuery("exec", "INSERT INTO users (name) VALUES ('test')", 30*time.Millisecond, nil)

	summary := agg.Summary()

	// Verify summary contains expected sections
	if !strings.Contains(summary, "Database Statistics Summary") {
		t.Error("expected summary to contain title")
	}
	if !strings.Contains(summary, "Global Stats") {
		t.Error("expected summary to contain global stats")
	}
	if !strings.Contains(summary, "Stats by Operation") {
		t.Error("expected summary to contain operation stats")
	}
	if !strings.Contains(summary, "Top Slow Tables") {
		t.Error("expected summary to contain table stats")
	}
}

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		query     string
		wantTable string
	}{
		{"SELECT * FROM users", "users"},
		{"SELECT id, name FROM users WHERE id = 1", "users"},
		{"SELECT * FROM `users`", "users"},
		{"SELECT * FROM \"users\"", "users"},
		{"INSERT INTO users (name) VALUES ('test')", "users"},
		{"INSERT INTO posts (title, content) VALUES ('test', 'content')", "posts"},
		{"UPDATE users SET name = 'test' WHERE id = 1", "users"},
		{"DELETE FROM users WHERE id = 1", "users"},
		{"", ""},
		{"INVALID QUERY", ""},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := extractTableName(tt.query)
			if got != tt.wantTable {
				t.Errorf("extractTableName(%q) = %q, want %q", tt.query, got, tt.wantTable)
			}
		})
	}
}

func TestDBStatsAggregator_DefaultSlowQueryThreshold(t *testing.T) {
	// Create with zero threshold - should default to 1 second
	agg := NewDBStatsAggregator(0)

	agg.RecordQuery("query", "SELECT * FROM users", 500*time.Millisecond, nil)
	agg.RecordQuery("query", "SELECT * FROM posts", 1500*time.Millisecond, nil)

	global := agg.GetGlobalStats()
	if global.SlowQueries != 1 {
		t.Errorf("expected 1 slow query with default threshold, got %d", global.SlowQueries)
	}
	if global.SlowQueryThresh != 1*time.Second {
		t.Errorf("expected default threshold of 1s, got %s", global.SlowQueryThresh)
	}
}

func TestDBStatsAggregator_Concurrency(t *testing.T) {
	agg := NewDBStatsAggregator(100 * time.Millisecond)

	// Record queries concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				agg.RecordQuery("query", "SELECT * FROM users", 50*time.Millisecond, nil)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	global := agg.GetGlobalStats()
	if global.TotalQueries != 1000 {
		t.Errorf("expected 1000 total queries, got %d", global.TotalQueries)
	}
}

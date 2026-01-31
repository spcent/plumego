package sharding

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMetricsCollector_RecordQuery(t *testing.T) {
	collector := NewMetricsCollector(4)
	ctx := context.Background()

	t.Run("record successful query", func(t *testing.T) {
		collector.RecordQuery(ctx, "SELECT", 0, 1*time.Millisecond, nil)

		snapshot := collector.Snapshot()
		if snapshot.TotalQueries != 1 {
			t.Errorf("expected 1 total query, got %d", snapshot.TotalQueries)
		}

		if snapshot.SelectQueries != 1 {
			t.Errorf("expected 1 SELECT query, got %d", snapshot.SelectQueries)
		}

		if snapshot.ShardQueryCounts[0] != 1 {
			t.Errorf("expected 1 query on shard 0, got %d", snapshot.ShardQueryCounts[0])
		}

		if snapshot.FailedQueries != 0 {
			t.Errorf("expected 0 failed queries, got %d", snapshot.FailedQueries)
		}
	})

	t.Run("record failed query", func(t *testing.T) {
		collector.Reset()
		err := ErrShardNotFound
		collector.RecordQuery(ctx, "INSERT", 1, 2*time.Millisecond, err)

		snapshot := collector.Snapshot()
		if snapshot.FailedQueries != 1 {
			t.Errorf("expected 1 failed query, got %d", snapshot.FailedQueries)
		}

		if snapshot.InsertQueries != 1 {
			t.Errorf("expected 1 INSERT query, got %d", snapshot.InsertQueries)
		}

		if snapshot.ShardErrorCounts[1] != 1 {
			t.Errorf("expected 1 error on shard 1, got %d", snapshot.ShardErrorCounts[1])
		}
	})

	t.Run("record different query types", func(t *testing.T) {
		collector.Reset()

		collector.RecordQuery(ctx, "SELECT", 0, 1*time.Millisecond, nil)
		collector.RecordQuery(ctx, "INSERT", 1, 2*time.Millisecond, nil)
		collector.RecordQuery(ctx, "UPDATE", 2, 3*time.Millisecond, nil)
		collector.RecordQuery(ctx, "DELETE", 3, 4*time.Millisecond, nil)

		snapshot := collector.Snapshot()
		if snapshot.SelectQueries != 1 {
			t.Errorf("expected 1 SELECT, got %d", snapshot.SelectQueries)
		}
		if snapshot.InsertQueries != 1 {
			t.Errorf("expected 1 INSERT, got %d", snapshot.InsertQueries)
		}
		if snapshot.UpdateQueries != 1 {
			t.Errorf("expected 1 UPDATE, got %d", snapshot.UpdateQueries)
		}
		if snapshot.DeleteQueries != 1 {
			t.Errorf("expected 1 DELETE, got %d", snapshot.DeleteQueries)
		}
	})
}

func TestMetricsCollector_Latency(t *testing.T) {
	collector := NewMetricsCollector(2)
	ctx := context.Background()

	collector.RecordQuery(ctx, "SELECT", 0, 100*time.Microsecond, nil)
	collector.RecordQuery(ctx, "SELECT", 0, 200*time.Microsecond, nil)
	collector.RecordQuery(ctx, "SELECT", 0, 300*time.Microsecond, nil)

	snapshot := collector.Snapshot()

	if snapshot.MinLatency != 100 {
		t.Errorf("expected min latency 100us, got %d", snapshot.MinLatency)
	}

	if snapshot.MaxLatency != 300 {
		t.Errorf("expected max latency 300us, got %d", snapshot.MaxLatency)
	}

	expectedAvg := uint64(200) // (100 + 200 + 300) / 3
	if snapshot.AvgLatency != expectedAvg {
		t.Errorf("expected avg latency %dus, got %d", expectedAvg, snapshot.AvgLatency)
	}
}

func TestMetricsCollector_ShardDistribution(t *testing.T) {
	collector := NewMetricsCollector(4)
	ctx := context.Background()

	// Send queries to different shards
	collector.RecordQuery(ctx, "SELECT", 0, 1*time.Millisecond, nil)
	collector.RecordQuery(ctx, "SELECT", 0, 1*time.Millisecond, nil)
	collector.RecordQuery(ctx, "SELECT", 1, 1*time.Millisecond, nil)
	collector.RecordQuery(ctx, "SELECT", 2, 1*time.Millisecond, nil)
	collector.RecordQuery(ctx, "SELECT", 2, 1*time.Millisecond, nil)
	collector.RecordQuery(ctx, "SELECT", 2, 1*time.Millisecond, nil)

	snapshot := collector.Snapshot()

	if snapshot.ShardQueryCounts[0] != 2 {
		t.Errorf("expected 2 queries on shard 0, got %d", snapshot.ShardQueryCounts[0])
	}
	if snapshot.ShardQueryCounts[1] != 1 {
		t.Errorf("expected 1 query on shard 1, got %d", snapshot.ShardQueryCounts[1])
	}
	if snapshot.ShardQueryCounts[2] != 3 {
		t.Errorf("expected 3 queries on shard 2, got %d", snapshot.ShardQueryCounts[2])
	}
	if snapshot.ShardQueryCounts[3] != 0 {
		t.Errorf("expected 0 queries on shard 3, got %d", snapshot.ShardQueryCounts[3])
	}
}

func TestMetricsCollector_CrossShardQueries(t *testing.T) {
	collector := NewMetricsCollector(4)

	collector.RecordSingleShardQuery()
	collector.RecordSingleShardQuery()
	collector.RecordCrossShardQuery()

	snapshot := collector.Snapshot()

	if snapshot.SingleShardQueries != 2 {
		t.Errorf("expected 2 single-shard queries, got %d", snapshot.SingleShardQueries)
	}

	if snapshot.CrossShardQueries != 1 {
		t.Errorf("expected 1 cross-shard query, got %d", snapshot.CrossShardQueries)
	}
}

func TestMetricsCollector_Rewrite(t *testing.T) {
	collector := NewMetricsCollector(2)

	// Record cache hit
	collector.RecordRewrite(true, nil)

	// Record cache miss
	collector.RecordRewrite(false, nil)
	collector.RecordRewrite(false, nil)

	// Record rewrite error
	collector.RecordRewrite(false, ErrNoShardingRule)

	snapshot := collector.Snapshot()

	if snapshot.RewriteCount != 4 {
		t.Errorf("expected 4 rewrites, got %d", snapshot.RewriteCount)
	}

	if snapshot.CacheHits != 1 {
		t.Errorf("expected 1 cache hit, got %d", snapshot.CacheHits)
	}

	if snapshot.CacheMisses != 3 {
		t.Errorf("expected 3 cache misses, got %d", snapshot.CacheMisses)
	}

	if snapshot.RewriteErrors != 1 {
		t.Errorf("expected 1 rewrite error, got %d", snapshot.RewriteErrors)
	}

	expectedHitRate := float64(25.0) // 1 / 4 * 100
	if snapshot.CacheHitRate != expectedHitRate {
		t.Errorf("expected cache hit rate %.2f%%, got %.2f%%", expectedHitRate, snapshot.CacheHitRate)
	}
}

func TestMetricsCollector_Reset(t *testing.T) {
	collector := NewMetricsCollector(2)
	ctx := context.Background()

	// Add some metrics
	collector.RecordQuery(ctx, "SELECT", 0, 1*time.Millisecond, nil)
	collector.RecordSingleShardQuery()
	collector.RecordCrossShardQuery()
	collector.RecordRewrite(true, nil)

	// Verify metrics exist
	snapshot := collector.Snapshot()
	if snapshot.TotalQueries == 0 {
		t.Fatal("expected metrics to be recorded")
	}

	// Reset
	collector.Reset()

	// Verify all metrics are zero
	snapshot = collector.Snapshot()
	if snapshot.TotalQueries != 0 {
		t.Errorf("expected 0 total queries after reset, got %d", snapshot.TotalQueries)
	}
	if snapshot.SingleShardQueries != 0 {
		t.Errorf("expected 0 single-shard queries after reset, got %d", snapshot.SingleShardQueries)
	}
	if snapshot.CrossShardQueries != 0 {
		t.Errorf("expected 0 cross-shard queries after reset, got %d", snapshot.CrossShardQueries)
	}
	if snapshot.RewriteCount != 0 {
		t.Errorf("expected 0 rewrites after reset, got %d", snapshot.RewriteCount)
	}
}

func TestMetricsCollector_EnableDisable(t *testing.T) {
	collector := NewMetricsCollector(2)
	ctx := context.Background()

	// Disable metrics
	collector.Disable()

	if collector.IsEnabled() {
		t.Error("expected metrics to be disabled")
	}

	// Record query (should be ignored)
	collector.RecordQuery(ctx, "SELECT", 0, 1*time.Millisecond, nil)

	snapshot := collector.Snapshot()
	if snapshot.TotalQueries != 0 {
		t.Errorf("expected 0 queries when disabled, got %d", snapshot.TotalQueries)
	}

	// Re-enable
	collector.Enable()

	if !collector.IsEnabled() {
		t.Error("expected metrics to be enabled")
	}

	// Record query (should be counted)
	collector.RecordQuery(ctx, "SELECT", 0, 1*time.Millisecond, nil)

	snapshot = collector.Snapshot()
	if snapshot.TotalQueries != 1 {
		t.Errorf("expected 1 query when enabled, got %d", snapshot.TotalQueries)
	}
}

func TestMetricsCollector_PrometheusFormat(t *testing.T) {
	collector := NewMetricsCollector(2)
	ctx := context.Background()

	collector.RecordQuery(ctx, "SELECT", 0, 1*time.Millisecond, nil)
	collector.RecordSingleShardQuery()

	metrics := collector.PrometheusMetrics()

	// Verify format contains expected metrics
	expectedMetrics := []string{
		"db_sharding_queries_total",
		"db_sharding_single_shard_queries_total",
		"db_sharding_cross_shard_queries_total",
		"db_sharding_failed_queries_total",
	}

	for _, expected := range expectedMetrics {
		if !strings.Contains(metrics, expected) {
			t.Errorf("expected metrics to contain %s", expected)
		}
	}
}

func BenchmarkMetricsCollector_RecordQuery(b *testing.B) {
	collector := NewMetricsCollector(4)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordQuery(ctx, "SELECT", i%4, 1*time.Millisecond, nil)
	}
}

func BenchmarkMetricsCollector_Snapshot(b *testing.B) {
	collector := NewMetricsCollector(4)
	ctx := context.Background()

	// Add some data
	for i := 0; i < 1000; i++ {
		collector.RecordQuery(ctx, "SELECT", i%4, 1*time.Millisecond, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Snapshot()
	}
}

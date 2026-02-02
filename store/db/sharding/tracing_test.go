package sharding

import (
	"context"
	"database/sql"
	"testing"

	"github.com/spcent/plumego/store/db/rw"
)

func TestTracer_StartSpan(t *testing.T) {
	config := DefaultTracingConfig()
	config.Enabled = true
	tracer := NewTracer(config)

	ctx := context.Background()

	t.Run("create span", func(t *testing.T) {
		newCtx, span := tracer.StartSpan(ctx, "test.operation")

		if span == nil {
			t.Fatal("expected span to be created")
		}

		if span.name != "test.operation" {
			t.Errorf("expected span name 'test.operation', got %s", span.name)
		}

		if newCtx == nil {
			t.Error("expected context to be returned")
		}
	})

	t.Run("disabled tracer", func(t *testing.T) {
		disabledConfig := DefaultTracingConfig()
		disabledConfig.Enabled = false
		disabledTracer := NewTracer(disabledConfig)

		_, span := disabledTracer.StartSpan(ctx, "test.operation")

		if span == nil {
			t.Fatal("expected span to be created even when disabled")
		}
	})
}

func TestSpan_SetAttribute(t *testing.T) {
	tracer := NewTracer(DefaultTracingConfig())
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "test")

	span.SetAttribute("key1", "value1")
	span.SetAttribute("key2", 123)

	if span.attributes["key1"] != "value1" {
		t.Errorf("expected attribute 'key1' to be 'value1', got %v", span.attributes["key1"])
	}

	if span.attributes["key2"] != 123 {
		t.Errorf("expected attribute 'key2' to be 123, got %v", span.attributes["key2"])
	}
}

func TestSpan_SetAttributes(t *testing.T) {
	tracer := NewTracer(DefaultTracingConfig())
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "test")

	attrs := map[string]any{
		"db.system":    "mysql",
		"db.statement": "SELECT * FROM users",
		"db.operation": "SELECT",
	}

	span.SetAttributes(attrs)

	for k, v := range attrs {
		if span.attributes[k] != v {
			t.Errorf("expected attribute %s to be %v, got %v", k, v, span.attributes[k])
		}
	}
}

func TestSpan_AddEvent(t *testing.T) {
	tracer := NewTracer(DefaultTracingConfig())
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "test")

	span.AddEvent("query.start", map[string]any{
		"query": "SELECT * FROM users",
	})

	if len(span.events) != 1 {
		t.Errorf("expected 1 event, got %d", len(span.events))
	}

	if span.events[0].Name != "query.start" {
		t.Errorf("expected event name 'query.start', got %s", span.events[0].Name)
	}
}

func TestSpan_SetStatus(t *testing.T) {
	tracer := NewTracer(DefaultTracingConfig())
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "test")

	span.SetStatus(SpanStatusOK, "success")

	if span.status.Code != SpanStatusOK {
		t.Errorf("expected status code OK, got %v", span.status.Code)
	}

	if span.status.Message != "success" {
		t.Errorf("expected status message 'success', got %s", span.status.Message)
	}
}

func TestSpan_RecordError(t *testing.T) {
	tracer := NewTracer(DefaultTracingConfig())
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "test")

	err := ErrShardNotFound
	span.RecordError(err)

	if span.status.Code != SpanStatusError {
		t.Errorf("expected status code Error, got %v", span.status.Code)
	}

	if len(span.events) != 1 {
		t.Errorf("expected 1 error event, got %d", len(span.events))
	}

	if span.events[0].Name != "error" {
		t.Errorf("expected event name 'error', got %s", span.events[0].Name)
	}
}

func TestTracingHelper_TraceQuery(t *testing.T) {
	config := DefaultTracingConfig()
	config.Enabled = true
	helper := NewTracingHelper(config)

	ctx := context.Background()
	query := "SELECT * FROM users WHERE id = ?"
	args := []any{123}

	newCtx, span := helper.TraceQuery(ctx, query, args)

	if newCtx == nil {
		t.Error("expected context to be returned")
	}

	if span == nil {
		t.Fatal("expected span to be created")
	}

	if span.attributes["db.statement"] != query {
		t.Errorf("expected db.statement to be %s, got %v", query, span.attributes["db.statement"])
	}

	if span.attributes["db.args.count"] != len(args) {
		t.Errorf("expected db.args.count to be %d, got %v", len(args), span.attributes["db.args.count"])
	}
}

func TestTracingHelper_TraceShardResolve(t *testing.T) {
	helper := NewTracingHelper(DefaultTracingConfig())
	ctx := context.Background()

	newCtx, span := helper.TraceShardResolve(ctx, "users")

	if newCtx == nil {
		t.Error("expected context to be returned")
	}

	if span == nil {
		t.Fatal("expected span to be created")
	}

	if span.attributes["shard.table"] != "users" {
		t.Errorf("expected shard.table to be 'users', got %v", span.attributes["shard.table"])
	}
}

func TestTracingHelper_TraceSQLRewrite(t *testing.T) {
	helper := NewTracingHelper(DefaultTracingConfig())
	ctx := context.Background()

	originalSQL := "SELECT * FROM users"
	newCtx, span := helper.TraceSQLRewrite(ctx, originalSQL)

	if newCtx == nil {
		t.Error("expected context to be returned")
	}

	if span == nil {
		t.Fatal("expected span to be created")
	}

	if span.attributes["sql.original"] != originalSQL {
		t.Errorf("expected sql.original to be %s, got %v", originalSQL, span.attributes["sql.original"])
	}
}

func TestTracingHelper_TraceShardQuery(t *testing.T) {
	helper := NewTracingHelper(DefaultTracingConfig())
	ctx := context.Background()

	query := "SELECT * FROM users_0"
	newCtx, span := helper.TraceShardQuery(ctx, 0, query)

	if newCtx == nil {
		t.Error("expected context to be returned")
	}

	if span == nil {
		t.Fatal("expected span to be created")
	}

	if span.attributes["shard.index"] != 0 {
		t.Errorf("expected shard.index to be 0, got %v", span.attributes["shard.index"])
	}

	if span.attributes["db.statement"] != query {
		t.Errorf("expected db.statement to be %s, got %v", query, span.attributes["db.statement"])
	}
}

func TestTracingHelper_TraceCrossShardQuery(t *testing.T) {
	helper := NewTracingHelper(DefaultTracingConfig())
	ctx := context.Background()

	newCtx, span := helper.TraceCrossShardQuery(ctx, 4, "all")

	if newCtx == nil {
		t.Error("expected context to be returned")
	}

	if span == nil {
		t.Fatal("expected span to be created")
	}

	if span.attributes["shard.count"] != 4 {
		t.Errorf("expected shard.count to be 4, got %v", span.attributes["shard.count"])
	}

	if span.attributes["shard.policy"] != "all" {
		t.Errorf("expected shard.policy to be 'all', got %v", span.attributes["shard.policy"])
	}
}

func TestInstrumentedRouter(t *testing.T) {
	// Create a test router
	shards := make([]*rw.Cluster, 2)
	for i := 0; i < 2; i++ {
		connector := &stubConnector{conn: &stubConn{}}
		primary := sql.OpenDB(connector)
		cluster, _ := rw.New(rw.Config{
			Primary: primary,
			HealthCheck: rw.HealthCheckConfig{
				Enabled: false,
			},
		})
		shards[i] = cluster
	}
	defer func() {
		for _, shard := range shards {
			shard.Close()
		}
	}()

	registry := NewShardingRuleRegistry()
	rule, _ := NewShardingRule("users", "user_id", NewModStrategy(), 2)
	registry.Register(rule)

	router, _ := NewRouter(shards, registry)

	tracingConfig := DefaultTracingConfig()
	tracingConfig.Enabled = true

	instrumented := NewInstrumentedRouter(router, true, tracingConfig)

	if instrumented.MetricsCollector() == nil {
		t.Error("expected metrics collector to be created")
	}

	if instrumented.Router() != router {
		t.Error("expected router to be the same")
	}
}

func BenchmarkTracer_StartSpan(b *testing.B) {
	config := DefaultTracingConfig()
	config.Enabled = true
	tracer := NewTracer(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := tracer.StartSpan(ctx, "test.operation")
		span.End()
	}
}

func BenchmarkSpan_SetAttribute(b *testing.B) {
	tracer := NewTracer(DefaultTracingConfig())
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span.SetAttribute("key", "value")
	}
}

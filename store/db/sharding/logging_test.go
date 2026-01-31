package sharding

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	glog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/store/db/rw"
)

func TestLoggingRouter_Creation(t *testing.T) {
	// Create a test router
	connector := &stubConnector{conn: &stubConn{}}
	primary := sql.OpenDB(connector)
	cluster, _ := rw.New(rw.Config{
		Primary: primary,
		HealthCheck: rw.HealthCheckConfig{
			Enabled: false,
		},
	})
	defer cluster.Close()

	registry := NewShardingRuleRegistry()
	router, _ := NewRouter([]*rw.Cluster{cluster}, registry)

	t.Run("with nil logger creates default JSONLogger", func(t *testing.T) {
		loggingRouter := NewLoggingRouter(router, nil)

		if loggingRouter.Logger() == nil {
			t.Error("expected logger to be set")
		}

		if loggingRouter.Router() != router {
			t.Error("expected router to be the same")
		}
	})

	t.Run("with custom logger", func(t *testing.T) {
		var buf bytes.Buffer
		logger := glog.NewJSONLogger(glog.JSONLoggerConfig{
			Output: &buf,
			Level:  glog.INFO,
		})

		loggingRouter := NewLoggingRouter(router, logger)

		if loggingRouter.Logger() != logger {
			t.Error("expected custom logger to be used")
		}
	})
}

func TestLoggingRouter_LogQuery(t *testing.T) {
	var buf bytes.Buffer
	logger := glog.NewJSONLogger(glog.JSONLoggerConfig{
		Output: &buf,
		Level:  glog.INFO,
	})

	// Create minimal router
	connector := &stubConnector{conn: &stubConn{}}
	primary := sql.OpenDB(connector)
	cluster, _ := rw.New(rw.Config{Primary: primary})
	defer cluster.Close()

	registry := NewShardingRuleRegistry()
	router, _ := NewRouter([]*rw.Cluster{cluster}, registry)

	loggingRouter := NewLoggingRouter(router, logger)
	ctx := context.Background()

	t.Run("successful query", func(t *testing.T) {
		buf.Reset()
		loggingRouter.LogQuery(ctx, "SELECT * FROM users", 0, 10*time.Millisecond, nil)

		output := buf.String()
		if !strings.Contains(output, "query executed") {
			t.Error("expected 'query executed' message")
		}
		if !strings.Contains(output, "SELECT * FROM users") {
			t.Error("expected query in output")
		}

		// Verify JSON structure
		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		if entry["query"] != "SELECT * FROM users" {
			t.Errorf("expected query field, got %v", entry["query"])
		}

		if entry["shard_index"] != float64(0) {
			t.Errorf("expected shard_index 0, got %v", entry["shard_index"])
		}
	})

	t.Run("failed query", func(t *testing.T) {
		buf.Reset()
		loggingRouter.LogQuery(ctx, "SELECT * FROM users", 0, 10*time.Millisecond, ErrShardNotFound)

		output := buf.String()
		if !strings.Contains(output, "query failed") {
			t.Error("expected 'query failed' message")
		}
		if !strings.Contains(output, "ERROR") {
			t.Error("expected ERROR level for failed query")
		}

		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		if _, ok := entry["error"]; !ok {
			t.Error("expected error field in output")
		}
	})
}

func TestLoggingRouter_LogShardResolution(t *testing.T) {
	var buf bytes.Buffer
	logger := glog.NewJSONLogger(glog.JSONLoggerConfig{
		Output: &buf,
		Level:  glog.INFO,
	})

	connector := &stubConnector{conn: &stubConn{}}
	primary := sql.OpenDB(connector)
	cluster, _ := rw.New(rw.Config{Primary: primary})
	defer cluster.Close()

	registry := NewShardingRuleRegistry()
	router, _ := NewRouter([]*rw.Cluster{cluster}, registry)

	loggingRouter := NewLoggingRouter(router, logger)
	ctx := context.Background()

	loggingRouter.LogShardResolution(ctx, "users", 123, 0)

	output := buf.String()
	if !strings.Contains(output, "shard resolved") {
		t.Error("expected 'shard resolved' message")
	}
	if !strings.Contains(output, "users") {
		t.Error("expected table name in output")
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["table"] != "users" {
		t.Errorf("expected table 'users', got %v", entry["table"])
	}

	if entry["shard_index"] != float64(0) {
		t.Errorf("expected shard_index 0, got %v", entry["shard_index"])
	}
}

func TestLoggingRouter_LogCrossShardQuery(t *testing.T) {
	var buf bytes.Buffer
	logger := glog.NewJSONLogger(glog.JSONLoggerConfig{
		Output: &buf,
		Level:  glog.WARNING,
	})

	connector := &stubConnector{conn: &stubConn{}}
	primary := sql.OpenDB(connector)
	cluster, _ := rw.New(rw.Config{Primary: primary})
	defer cluster.Close()

	registry := NewShardingRuleRegistry()
	router, _ := NewRouter([]*rw.Cluster{cluster}, registry)

	loggingRouter := NewLoggingRouter(router, logger)
	ctx := context.Background()

	loggingRouter.LogCrossShardQuery(ctx, "SELECT * FROM users", "all")

	output := buf.String()
	if !strings.Contains(output, "cross-shard query") {
		t.Error("expected 'cross-shard query' message")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("expected WARN level for cross-shard query")
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["level"] != "WARN" {
		t.Errorf("expected level WARN, got %v", entry["level"])
	}

	if entry["policy"] != "all" {
		t.Errorf("expected policy 'all', got %v", entry["policy"])
	}
}

func TestLoggingRouter_LogRewrite(t *testing.T) {
	var buf bytes.Buffer
	logger := glog.NewJSONLogger(glog.JSONLoggerConfig{
		Output: &buf,
		Level:  glog.INFO,
	})

	connector := &stubConnector{conn: &stubConn{}}
	primary := sql.OpenDB(connector)
	cluster, _ := rw.New(rw.Config{Primary: primary})
	defer cluster.Close()

	registry := NewShardingRuleRegistry()
	router, _ := NewRouter([]*rw.Cluster{cluster}, registry)

	loggingRouter := NewLoggingRouter(router, logger)
	ctx := context.Background()

	loggingRouter.LogRewrite(ctx, "SELECT * FROM users", "SELECT * FROM users_0", true)

	output := buf.String()
	if !strings.Contains(output, "SQL rewritten") {
		t.Error("expected 'SQL rewritten' message")
	}
	if !strings.Contains(output, "users_0") {
		t.Error("expected rewritten table name in output")
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["original"] != "SELECT * FROM users" {
		t.Errorf("expected original query, got %v", entry["original"])
	}

	if entry["rewritten"] != "SELECT * FROM users_0" {
		t.Errorf("expected rewritten query, got %v", entry["rewritten"])
	}

	if entry["cached"] != true {
		t.Errorf("expected cached true, got %v", entry["cached"])
	}
}

func TestLoggingRouter_WithTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := glog.NewJSONLogger(glog.JSONLoggerConfig{
		Output: &buf,
		Level:  glog.INFO,
	})

	connector := &stubConnector{conn: &stubConn{}}
	primary := sql.OpenDB(connector)
	cluster, _ := rw.New(rw.Config{Primary: primary})
	defer cluster.Close()

	registry := NewShardingRuleRegistry()
	router, _ := NewRouter([]*rw.Cluster{cluster}, registry)

	loggingRouter := NewLoggingRouter(router, logger)

	// Create context with trace ID
	ctx := context.Background()
	traceID := "test-trace-123"
	ctx = glog.WithTraceID(ctx, traceID)

	loggingRouter.LogQuery(ctx, "SELECT * FROM users", 0, 10*time.Millisecond, nil)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["trace_id"] != traceID {
		t.Errorf("expected trace_id %q, got %v", traceID, entry["trace_id"])
	}
}

func TestLoggingRouter_DefaultFields(t *testing.T) {
	var buf bytes.Buffer

	connector := &stubConnector{conn: &stubConn{}}
	primary := sql.OpenDB(connector)
	cluster, _ := rw.New(rw.Config{Primary: primary})
	defer cluster.Close()

	registry := NewShardingRuleRegistry()
	router, _ := NewRouter([]*rw.Cluster{cluster}, registry)

	// Create a custom logger with buffer and default fields
	customLogger := glog.NewJSONLogger(glog.JSONLoggerConfig{
		Output: &buf,
		Level:  glog.INFO,
		Fields: glog.Fields{"component": "sharding"},
	})
	loggingRouter := NewLoggingRouter(router, customLogger)

	ctx := context.Background()
	loggingRouter.LogQuery(ctx, "SELECT * FROM users", 0, 10*time.Millisecond, nil)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["component"] != "sharding" {
		t.Errorf("expected component field 'sharding', got %v", entry["component"])
	}
}

func BenchmarkLoggingRouter_LogQuery(b *testing.B) {
	var buf bytes.Buffer
	logger := glog.NewJSONLogger(glog.JSONLoggerConfig{
		Output: &buf,
		Level:  glog.INFO,
	})

	connector := &stubConnector{conn: &stubConn{}}
	primary := sql.OpenDB(connector)
	cluster, _ := rw.New(rw.Config{Primary: primary})
	defer cluster.Close()

	registry := NewShardingRuleRegistry()
	router, _ := NewRouter([]*rw.Cluster{cluster}, registry)

	loggingRouter := NewLoggingRouter(router, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loggingRouter.LogQuery(ctx, "SELECT * FROM users", 0, 10*time.Millisecond, nil)
	}
}

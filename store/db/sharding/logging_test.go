package sharding

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/store/db/rw"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	config := DefaultLoggerConfig()
	config.Output = &buf
	config.Level = LogLevelInfo

	logger := NewLogger(config)
	ctx := context.Background()

	t.Run("debug below threshold", func(t *testing.T) {
		buf.Reset()
		logger.Debug(ctx, "debug message")

		if buf.Len() > 0 {
			t.Error("expected debug message to be filtered out")
		}
	})

	t.Run("info at threshold", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "info message")

		output := buf.String()
		if !strings.Contains(output, "info message") {
			t.Error("expected info message to be logged")
		}
		if !strings.Contains(output, `"level":"INFO"`) {
			t.Error("expected INFO level in output")
		}
	})

	t.Run("warn above threshold", func(t *testing.T) {
		buf.Reset()
		logger.Warn(ctx, "warn message")

		output := buf.String()
		if !strings.Contains(output, "warn message") {
			t.Error("expected warn message to be logged")
		}
		if !strings.Contains(output, `"level":"WARN"`) {
			t.Error("expected WARN level in output")
		}
	})

	t.Run("error above threshold", func(t *testing.T) {
		buf.Reset()
		logger.Error(ctx, "error message")

		output := buf.String()
		if !strings.Contains(output, "error message") {
			t.Error("expected error message to be logged")
		}
		if !strings.Contains(output, `"level":"ERROR"`) {
			t.Error("expected ERROR level in output")
		}
	})
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	config := DefaultLoggerConfig()
	config.Output = &buf
	config.Level = LogLevelInfo

	logger := NewLogger(config)
	ctx := context.Background()

	loggerWithFields := logger.WithFields(map[string]interface{}{
		"component": "sharding",
		"version":   "1.0",
	})

	loggerWithFields.Info(ctx, "test message")

	output := buf.String()
	if !strings.Contains(output, `"component":"sharding"`) {
		t.Error("expected component field in output")
	}
	if !strings.Contains(output, `"version":"1.0"`) {
		t.Error("expected version field in output")
	}
}

func TestLogger_WithField(t *testing.T) {
	var buf bytes.Buffer
	config := DefaultLoggerConfig()
	config.Output = &buf

	logger := NewLogger(config)
	ctx := context.Background()

	loggerWithField := logger.WithField("request_id", "12345")
	loggerWithField.Info(ctx, "test message")

	output := buf.String()
	if !strings.Contains(output, `"request_id":"12345"`) {
		t.Error("expected request_id field in output")
	}
}

func TestLogger_Debugf(t *testing.T) {
	var buf bytes.Buffer
	config := DefaultLoggerConfig()
	config.Output = &buf
	config.Level = LogLevelDebug

	logger := NewLogger(config)
	ctx := context.Background()

	logger.Debugf(ctx, "user %s logged in at %d", "alice", 123456)

	output := buf.String()
	if !strings.Contains(output, "user alice logged in at 123456") {
		t.Error("expected formatted message in output")
	}
}

func TestLogger_InfoWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := DefaultLoggerConfig()
	config.Output = &buf

	logger := NewLogger(config)
	ctx := context.Background()

	logger.InfoWithFields(ctx, "query executed", map[string]interface{}{
		"query":    "SELECT * FROM users",
		"duration": "10ms",
	})

	output := buf.String()
	if !strings.Contains(output, "query executed") {
		t.Error("expected message in output")
	}
	if !strings.Contains(output, `"query":"SELECT * FROM users"`) {
		t.Error("expected query field in output")
	}
	if !strings.Contains(output, `"duration":"10ms"`) {
		t.Error("expected duration field in output")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	config := DefaultLoggerConfig()
	config.Output = &buf
	config.Level = LogLevelInfo

	logger := NewLogger(config)
	ctx := context.Background()

	// Debug should be filtered
	logger.Debug(ctx, "debug message")
	if buf.Len() > 0 {
		t.Error("expected debug to be filtered")
	}

	// Change level to DEBUG
	logger.SetLevel(LogLevelDebug)

	// Debug should now be logged
	buf.Reset()
	logger.Debug(ctx, "debug message")
	if buf.Len() == 0 {
		t.Error("expected debug to be logged after level change")
	}
}

func TestLogger_GetLevel(t *testing.T) {
	config := DefaultLoggerConfig()
	config.Level = LogLevelWarn

	logger := NewLogger(config)

	if logger.GetLevel() != LogLevelWarn {
		t.Errorf("expected level WARN, got %v", logger.GetLevel())
	}

	logger.SetLevel(LogLevelError)

	if logger.GetLevel() != LogLevelError {
		t.Errorf("expected level ERROR, got %v", logger.GetLevel())
	}
}

func TestLoggingRouter(t *testing.T) {
	var buf bytes.Buffer
	logConfig := DefaultLoggerConfig()
	logConfig.Output = &buf
	logConfig.Level = LogLevelDebug

	logger := NewLogger(logConfig)

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

	loggingRouter := NewLoggingRouter(router, logger)

	if loggingRouter.Logger() == nil {
		t.Error("expected logger to be set")
	}

	if loggingRouter.Router() != router {
		t.Error("expected router to be the same")
	}
}

func TestLoggingRouter_LogQuery(t *testing.T) {
	var buf bytes.Buffer
	logConfig := DefaultLoggerConfig()
	logConfig.Output = &buf
	logConfig.Level = LogLevelDebug

	logger := NewLogger(logConfig)

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
	})
}

func TestLoggingRouter_LogShardResolution(t *testing.T) {
	var buf bytes.Buffer
	logConfig := DefaultLoggerConfig()
	logConfig.Output = &buf
	logConfig.Level = LogLevelDebug

	logger := NewLogger(logConfig)

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
}

func TestLoggingRouter_LogCrossShardQuery(t *testing.T) {
	var buf bytes.Buffer
	logConfig := DefaultLoggerConfig()
	logConfig.Output = &buf
	logConfig.Level = LogLevelWarn

	logger := NewLogger(logConfig)

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
}

func TestLoggingRouter_LogRewrite(t *testing.T) {
	var buf bytes.Buffer
	logConfig := DefaultLoggerConfig()
	logConfig.Output = &buf
	logConfig.Level = LogLevelDebug

	logger := NewLogger(logConfig)

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
}

func BenchmarkLogger_Info(b *testing.B) {
	var buf bytes.Buffer
	config := DefaultLoggerConfig()
	config.Output = &buf

	logger := NewLogger(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "test message")
	}
}

func BenchmarkLogger_InfoWithFields(b *testing.B) {
	var buf bytes.Buffer
	config := DefaultLoggerConfig()
	config.Output = &buf

	logger := NewLogger(config)
	ctx := context.Background()

	fields := map[string]interface{}{
		"query":    "SELECT * FROM users",
		"duration": "10ms",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoWithFields(ctx, "query executed", fields)
	}
}

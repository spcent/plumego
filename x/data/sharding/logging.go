package sharding

import (
	"context"
	"fmt"
	"time"

	"github.com/spcent/plumego/log"
)

// LoggingRouter wraps a router with structured logging for database sharding operations.
// It uses the plumego/log/log.StructuredLogger interface for flexible logging backends.
type LoggingRouter struct {
	router *Router
	logger log.StructuredLogger
}

// NewLoggingRouter creates a new logging router with optional logger.
// If logger is nil, it creates a default JSONLogger suitable for production use.
func NewLoggingRouter(router *Router, logger log.StructuredLogger) *LoggingRouter {
	if logger == nil {
		// Default to JSON logger for sharding operations
		logger = log.NewJSONLogger(log.JSONLoggerConfig{
			Level: log.INFO,
			Fields: log.Fields{
				"component": "sharding",
			},
		})
	}

	return &LoggingRouter{
		router: router,
		logger: logger,
	}
}

// Logger returns the underlying logger instance.
func (lr *LoggingRouter) Logger() log.StructuredLogger {
	return lr.logger
}

// Router returns the underlying router instance.
func (lr *LoggingRouter) Router() *Router {
	return lr.router
}

// LogQuery logs query execution with context, automatically capturing trace IDs.
func (lr *LoggingRouter) LogQuery(ctx context.Context, query string, shardIndex int, latency time.Duration, err error) {
	fields := log.Fields{
		"query":       query,
		"shard_index": shardIndex,
		"latency_ms":  latency.Milliseconds(),
	}

	if err != nil {
		fields["error"] = err.Error()
		lr.logger.ErrorCtx(ctx, "query failed", fields)
	} else {
		lr.logger.InfoCtx(ctx, "query executed", fields)
	}
}

// LogShardResolution logs shard resolution with context.
func (lr *LoggingRouter) LogShardResolution(ctx context.Context, tableName string, shardKey any, shardIndex int) {
	lr.logger.InfoCtx(ctx, "shard resolved", log.Fields{
		"table":       tableName,
		"shard_key":   fmt.Sprintf("%v", shardKey),
		"shard_index": shardIndex,
	})
}

// LogCrossShardQuery logs cross-shard queries with context.
// These are typically less performant and should be monitored.
func (lr *LoggingRouter) LogCrossShardQuery(ctx context.Context, query string, policy string) {
	lr.logger.WarnCtx(ctx, "cross-shard query", log.Fields{
		"query":  query,
		"policy": policy,
	})
}

// LogRewrite logs SQL rewriting with context.
func (lr *LoggingRouter) LogRewrite(ctx context.Context, original string, rewritten string, cached bool) {
	lr.logger.InfoCtx(ctx, "SQL rewritten", log.Fields{
		"original":  original,
		"rewritten": rewritten,
		"cached":    cached,
	})
}

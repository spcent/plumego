package sharding

import (
	"context"
	"fmt"
	"time"

	glog "github.com/spcent/plumego/log"
)

// LoggingRouter wraps a router with structured logging for database sharding operations.
// It uses the plumego/log/glog.StructuredLogger interface for flexible logging backends.
type LoggingRouter struct {
	router *Router
	logger glog.StructuredLogger
}

// NewLoggingRouter creates a new logging router with optional logger.
// If logger is nil, it creates a default JSONLogger suitable for production use.
func NewLoggingRouter(router *Router, logger glog.StructuredLogger) *LoggingRouter {
	if logger == nil {
		// Default to JSON logger for sharding operations
		logger = glog.NewJSONLogger(glog.JSONLoggerConfig{
			Level: glog.INFO,
			Fields: glog.Fields{
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
func (lr *LoggingRouter) Logger() glog.StructuredLogger {
	return lr.logger
}

// Router returns the underlying router instance.
func (lr *LoggingRouter) Router() *Router {
	return lr.router
}

// LogQuery logs query execution with context, automatically capturing trace IDs.
func (lr *LoggingRouter) LogQuery(ctx context.Context, query string, shardIndex int, latency time.Duration, err error) {
	fields := glog.Fields{
		"query":       query,
		"shard_index": shardIndex,
		"latency_ms":  latency.Milliseconds(),
	}

	if err != nil {
		fields["error"] = err.Error()
		lr.logger.ErrorCtx(ctx, "query failed", fields)
	} else {
		lr.logger.DebugCtx(ctx, "query executed", fields)
	}
}

// LogShardResolution logs shard resolution with context.
func (lr *LoggingRouter) LogShardResolution(ctx context.Context, tableName string, shardKey any, shardIndex int) {
	lr.logger.DebugCtx(ctx, "shard resolved", glog.Fields{
		"table":       tableName,
		"shard_key":   fmt.Sprintf("%v", shardKey),
		"shard_index": shardIndex,
	})
}

// LogCrossShardQuery logs cross-shard queries with context.
// These are typically less performant and should be monitored.
func (lr *LoggingRouter) LogCrossShardQuery(ctx context.Context, query string, policy string) {
	lr.logger.WarnCtx(ctx, "cross-shard query", glog.Fields{
		"query":  query,
		"policy": policy,
	})
}

// LogRewrite logs SQL rewriting with context.
func (lr *LoggingRouter) LogRewrite(ctx context.Context, original string, rewritten string, cached bool) {
	lr.logger.DebugCtx(ctx, "SQL rewritten", glog.Fields{
		"original":  original,
		"rewritten": rewritten,
		"cached":    cached,
	})
}

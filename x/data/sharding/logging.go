package sharding

import (
	"context"
	"fmt"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
)

// LoggingRouter wraps a router with structured logging for database sharding operations.
// It uses the plumego/log/log.StructuredLogger interface for flexible logging backends.
type LoggingRouter struct {
	router *Router
	logger log.StructuredLogger
}

// NewLoggingRouter creates a new logging router with optional logger.
// If logger is nil, it creates a JSON logger through the canonical NewLogger path.
func NewLoggingRouter(router *Router, logger log.StructuredLogger) *LoggingRouter {
	if logger == nil {
		logger = log.NewLogger(log.LoggerConfig{
			Format: log.LoggerFormatJSON,
			Level:  log.INFO,
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

func fieldsWithRequestID(ctx context.Context, fields log.Fields) log.Fields {
	merged := log.Fields{}
	for k, v := range fields {
		merged[k] = v
	}
	if requestID := contract.RequestIDFromContext(ctx); requestID != "" {
		merged["request_id"] = requestID
	}
	return merged
}

// LogQuery logs query execution with context, attaching request correlation explicitly.
func (lr *LoggingRouter) LogQuery(ctx context.Context, query string, shardIndex int, latency time.Duration, err error) {
	fields := log.Fields{
		"query":       query,
		"shard_index": shardIndex,
		"latency_ms":  latency.Milliseconds(),
	}

	if err != nil {
		fields["error"] = err.Error()
		lr.logger.ErrorCtx(ctx, "query failed", fieldsWithRequestID(ctx, fields))
	} else {
		lr.logger.InfoCtx(ctx, "query executed", fieldsWithRequestID(ctx, fields))
	}
}

// LogShardResolution logs shard resolution with context.
func (lr *LoggingRouter) LogShardResolution(ctx context.Context, tableName string, shardKey any, shardIndex int) {
	lr.logger.InfoCtx(ctx, "shard resolved", fieldsWithRequestID(ctx, log.Fields{
		"table":       tableName,
		"shard_key":   fmt.Sprintf("%v", shardKey),
		"shard_index": shardIndex,
	}))
}

// LogCrossShardQuery logs cross-shard queries with context.
// These are typically less performant and should be monitored.
func (lr *LoggingRouter) LogCrossShardQuery(ctx context.Context, query string, policy string) {
	lr.logger.WarnCtx(ctx, "cross-shard query", fieldsWithRequestID(ctx, log.Fields{
		"query":  query,
		"policy": policy,
	}))
}

// LogRewrite logs SQL rewriting with context.
func (lr *LoggingRouter) LogRewrite(ctx context.Context, original string, rewritten string, cached bool) {
	lr.logger.InfoCtx(ctx, "SQL rewritten", fieldsWithRequestID(ctx, log.Fields{
		"original":  original,
		"rewritten": rewritten,
		"cached":    cached,
	}))
}

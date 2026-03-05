package logging

import (
	"context"

	glog "github.com/spcent/plumego/log"
)

// StructuredLoggerAdapter wraps an ai/logging.Logger to implement glog.StructuredLogger.
// Use this when a component accepts glog.StructuredLogger but you have an ai/logging.Logger.
type StructuredLoggerAdapter struct {
	inner Logger
}

// NewStructuredLoggerAdapter wraps an ai/logging.Logger as a glog.StructuredLogger.
func NewStructuredLoggerAdapter(l Logger) glog.StructuredLogger {
	return &StructuredLoggerAdapter{inner: l}
}

func (a *StructuredLoggerAdapter) WithFields(fields glog.Fields) glog.StructuredLogger {
	return &StructuredLoggerAdapter{inner: a.inner.With(fieldsFromMap(fields)...)}
}

func (a *StructuredLoggerAdapter) Debug(msg string, fields glog.Fields) {
	a.inner.Debug(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) Info(msg string, fields glog.Fields) {
	a.inner.Info(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) Warn(msg string, fields glog.Fields) {
	a.inner.Warn(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) Error(msg string, fields glog.Fields) {
	a.inner.Error(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) Fatal(msg string, fields glog.Fields) {
	a.inner.Fatal(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) DebugCtx(ctx context.Context, msg string, fields glog.Fields) {
	a.inner.WithContext(ctx).Debug(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) InfoCtx(ctx context.Context, msg string, fields glog.Fields) {
	a.inner.WithContext(ctx).Info(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) WarnCtx(ctx context.Context, msg string, fields glog.Fields) {
	a.inner.WithContext(ctx).Warn(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) ErrorCtx(ctx context.Context, msg string, fields glog.Fields) {
	a.inner.WithContext(ctx).Error(msg, fieldsFromMap(fields)...)
}

func (a *StructuredLoggerAdapter) FatalCtx(ctx context.Context, msg string, fields glog.Fields) {
	a.inner.WithContext(ctx).Fatal(msg, fieldsFromMap(fields)...)
}

// fieldsFromMap converts glog.Fields (map) to []Field (slice).
func fieldsFromMap(m glog.Fields) []Field {
	if len(m) == 0 {
		return nil
	}
	out := make([]Field, 0, len(m))
	for k, v := range m {
		out = append(out, Field{Key: k, Value: v})
	}
	return out
}

// AILoggerAdapter wraps a glog.StructuredLogger to implement ai/logging.Logger.
// Use this when a component accepts ai/logging.Logger but you have a glog.StructuredLogger.
type AILoggerAdapter struct {
	inner glog.StructuredLogger
	level Level
}

// NewAILoggerAdapter wraps a glog.StructuredLogger as an ai/logging.Logger.
func NewAILoggerAdapter(l glog.StructuredLogger) Logger {
	return &AILoggerAdapter{inner: l, level: DebugLevel}
}

func (a *AILoggerAdapter) Debug(msg string, fields ...Field) {
	if a.level > DebugLevel {
		return
	}
	a.inner.Debug(msg, fieldsToMap(fields))
}

func (a *AILoggerAdapter) Info(msg string, fields ...Field) {
	if a.level > InfoLevel {
		return
	}
	a.inner.Info(msg, fieldsToMap(fields))
}

func (a *AILoggerAdapter) Warn(msg string, fields ...Field) {
	if a.level > WarnLevel {
		return
	}
	a.inner.Warn(msg, fieldsToMap(fields))
}

func (a *AILoggerAdapter) Error(msg string, fields ...Field) {
	if a.level > ErrorLevel {
		return
	}
	a.inner.Error(msg, fieldsToMap(fields))
}

func (a *AILoggerAdapter) Fatal(msg string, fields ...Field) {
	a.inner.Fatal(msg, fieldsToMap(fields))
}

func (a *AILoggerAdapter) With(fields ...Field) Logger {
	return &AILoggerAdapter{
		inner: a.inner.WithFields(fieldsToMap(fields)),
		level: a.level,
	}
}

func (a *AILoggerAdapter) WithContext(ctx context.Context) Logger {
	traceID := glog.TraceIDFromContext(ctx)
	if traceID == "" {
		return a
	}
	return &AILoggerAdapter{
		inner: a.inner.WithFields(glog.Fields{"trace_id": traceID}),
		level: a.level,
	}
}

func (a *AILoggerAdapter) SetLevel(level Level) {
	a.level = level
}

// fieldsToMap converts []Field (slice) to glog.Fields (map).
func fieldsToMap(fields []Field) glog.Fields {
	if len(fields) == 0 {
		return nil
	}
	m := make(glog.Fields, len(fields))
	for _, f := range fields {
		m[f.Key] = f.Value
	}
	return m
}

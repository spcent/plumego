package glog

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Fields represents structured log fields to attach to a log entry.
type Fields map[string]any

// StructuredLogger defines the minimal logging interface used by the application.
//
// Implementations can wrap zap/logrus/zerolog/slog or any other logger.
// The logger should treat a nil Fields map the same as an empty one.
type StructuredLogger interface {
	WithFields(fields Fields) StructuredLogger

	Debug(msg string, fields Fields)
	Info(msg string, fields Fields)
	Warn(msg string, fields Fields)
	Error(msg string, fields Fields)
}

// Lifecycle allows a logger to participate in application start/stop hooks
// (e.g. to initialize flags or flush buffers). Methods are optional.
type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// gLogger adapts the existing glog implementation to the StructuredLogger interface.
type gLogger struct {
	fields Fields
}

// NewGLogger creates a StructuredLogger backed by the package-level glog implementation.
// It implements Lifecycle to run glog.Init/Flush/Close automatically.
func NewGLogger() StructuredLogger {
	return &gLogger{fields: Fields{}}
}

// Start initializes the underlying glog system.
func (l *gLogger) Start(ctx context.Context) error {
	Init()
	return nil
}

// Stop flushes and closes glog resources.
func (l *gLogger) Stop(ctx context.Context) error {
	Flush()
	Close()
	return nil
}

func (l *gLogger) WithFields(fields Fields) StructuredLogger {
	merged := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &gLogger{fields: merged}
}

func (l *gLogger) Debug(msg string, fields Fields) {
	if !V(1) {
		return
	}
	l.logWithFields("DEBUG", msg, fields)
}

func (l *gLogger) Info(msg string, fields Fields) {
	l.logWithFields("INFO", msg, fields)
}

func (l *gLogger) Warn(msg string, fields Fields) {
	l.logWithFields("WARN", msg, fields)
}

func (l *gLogger) Error(msg string, fields Fields) {
	l.logWithFields("ERROR", msg, fields)
}

func (l *gLogger) logWithFields(level string, msg string, fields Fields) {
	combined := l.mergeFields(fields)
	formatted := l.formatFields(combined)
	if formatted != "" {
		msg = msg + " " + formatted
	}

	switch level {
	case "DEBUG":
		Info(msg)
	case "INFO":
		Info(msg)
	case "WARN":
		Warning(msg)
	case "ERROR":
		Error(msg)
	}
}

func (l *gLogger) mergeFields(fields Fields) Fields {
	merged := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return merged
}

func (l *gLogger) formatFields(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, fields[k]))
	}

	return strings.Join(parts, " ")
}

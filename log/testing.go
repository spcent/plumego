package glog

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"
)

// LogEntry holds a single captured log entry.
type LogEntry struct {
	Level   Level
	Message string
	Fields  Fields
	Time    time.Time
}

// TestLogger is an in-memory StructuredLogger designed for unit tests.
// It captures every log entry so assertions can inspect what was logged.
// TestLogger is safe for concurrent use.
type TestLogger struct {
	mu         sync.RWMutex
	entries    []LogEntry
	level      Level
	baseFields Fields
}

// NewTestLogger creates a TestLogger that captures all levels.
func NewTestLogger() *TestLogger {
	return &TestLogger{
		level:      DEBUG,
		baseFields: Fields{},
	}
}

// --- StructuredLogger interface ---

func (l *TestLogger) WithFields(fields Fields) StructuredLogger {
	l.mu.RLock()
	merged := mergeFields(l.baseFields, fields)
	level := l.level
	l.mu.RUnlock()

	return &TestLogger{
		level:      level,
		baseFields: merged,
	}
}

func (l *TestLogger) Debug(msg string, fields Fields) { l.capture(DEBUG, msg, fields) }
func (l *TestLogger) Info(msg string, fields Fields)  { l.capture(INFO, msg, fields) }
func (l *TestLogger) Warn(msg string, fields Fields)  { l.capture(WARNING, msg, fields) }
func (l *TestLogger) Error(msg string, fields Fields) { l.capture(ERROR, msg, fields) }
func (l *TestLogger) Fatal(msg string, fields Fields) {
	l.capture(FATAL, msg, fields)
	os.Exit(1)
}

func (l *TestLogger) DebugCtx(ctx context.Context, msg string, fields Fields) {
	l.capture(DEBUG, msg, l.withTrace(ctx, fields))
}
func (l *TestLogger) InfoCtx(ctx context.Context, msg string, fields Fields) {
	l.capture(INFO, msg, l.withTrace(ctx, fields))
}
func (l *TestLogger) WarnCtx(ctx context.Context, msg string, fields Fields) {
	l.capture(WARNING, msg, l.withTrace(ctx, fields))
}
func (l *TestLogger) ErrorCtx(ctx context.Context, msg string, fields Fields) {
	l.capture(ERROR, msg, l.withTrace(ctx, fields))
}
func (l *TestLogger) FatalCtx(ctx context.Context, msg string, fields Fields) {
	l.capture(FATAL, msg, l.withTrace(ctx, fields))
	os.Exit(1)
}

// --- Test assertion helpers ---

// Entries returns a snapshot of all captured log entries.
func (l *TestLogger) Entries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	cp := make([]LogEntry, len(l.entries))
	copy(cp, l.entries)
	return cp
}

// Clear removes all captured entries.
func (l *TestLogger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = l.entries[:0]
}

// Count returns the total number of captured entries.
func (l *TestLogger) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}

// CountByLevel returns the number of entries at the given level.
func (l *TestLogger) CountByLevel(level Level) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	n := 0
	for _, e := range l.entries {
		if e.Level == level {
			n++
		}
	}
	return n
}

// HasEntry returns true if at least one entry matches the level and contains
// substr in its message.
func (l *TestLogger) HasEntry(level Level, substr string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, e := range l.entries {
		if e.Level == level && strings.Contains(e.Message, substr) {
			return true
		}
	}
	return false
}

// LastEntry returns the most recently captured entry and true.
// Returns a zero LogEntry and false when nothing has been logged yet.
func (l *TestLogger) LastEntry() (LogEntry, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.entries) == 0 {
		return LogEntry{}, false
	}
	return l.entries[len(l.entries)-1], true
}

// --- internals ---

func (l *TestLogger) capture(level Level, msg string, fields Fields) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level < l.level {
		return
	}
	l.entries = append(l.entries, LogEntry{
		Level:   level,
		Message: msg,
		Fields:  mergeFields(l.baseFields, fields),
		Time:    time.Now(),
	})
}

func (l *TestLogger) withTrace(ctx context.Context, fields Fields) Fields {
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		out := mergeFields(fields, Fields{"trace_id": traceID})
		return out
	}
	return fields
}

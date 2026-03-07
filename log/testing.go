package log

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

// sharedEntries holds the captured log entries shared between a root TestLogger
// and all loggers derived from it via WithFields.
type sharedEntries struct {
	mu      sync.RWMutex
	entries []LogEntry
}

// TestLogger is an in-memory StructuredLogger designed for unit tests.
// It captures every log entry so assertions can inspect what was logged.
// All loggers derived via WithFields share the same entry store as the root,
// so assertions on the root logger see output from all derived loggers.
// TestLogger is safe for concurrent use.
//
// By default Fatal/FatalCtx call os.Exit(1) after capturing the entry.
// Set FatalHook to override this behaviour in tests that need to assert on
// fatal messages without terminating the process:
//
//	tl := log.NewTestLogger()
//	tl.FatalHook = func(msg string) { panic("fatal: " + msg) }
type TestLogger struct {
	shared     *sharedEntries
	level      Level
	baseFields Fields
	// FatalHook is called instead of os.Exit(1) when a Fatal/FatalCtx message
	// is logged. When nil, os.Exit(1) is called (production-safe default).
	FatalHook func(msg string)
}

// NewTestLogger creates a TestLogger that captures all levels.
func NewTestLogger() *TestLogger {
	return &TestLogger{
		shared:     &sharedEntries{},
		level:      DEBUG,
		baseFields: Fields{},
	}
}

// --- StructuredLogger interface ---

// With is a convenience shortcut for WithFields(Fields{key: value}).
func (l *TestLogger) With(key string, value any) StructuredLogger {
	return l.WithFields(Fields{key: value})
}

// WithFields returns a derived logger that shares the same entry store and FatalHook.
// Entries logged via the child are visible through the root's Entries() etc.
func (l *TestLogger) WithFields(fields Fields) StructuredLogger {
	return &TestLogger{
		shared:     l.shared, // shared with parent – child entries are visible on root
		level:      l.level,
		baseFields: mergeFields(l.baseFields, fields),
		FatalHook:  l.FatalHook,
	}
}

func (l *TestLogger) Debug(msg string, fields ...Fields) { l.capture(DEBUG, msg, firstFields(fields)) }
func (l *TestLogger) Info(msg string, fields ...Fields)  { l.capture(INFO, msg, firstFields(fields)) }
func (l *TestLogger) Warn(msg string, fields ...Fields)  { l.capture(WARNING, msg, firstFields(fields)) }
func (l *TestLogger) Error(msg string, fields ...Fields) { l.capture(ERROR, msg, firstFields(fields)) }
func (l *TestLogger) Fatal(msg string, fields ...Fields) {
	l.capture(FATAL, msg, firstFields(fields))
	l.handleFatal(msg)
}

func (l *TestLogger) DebugCtx(ctx context.Context, msg string, fields ...Fields) {
	l.capture(DEBUG, msg, l.withTrace(ctx, firstFields(fields)))
}
func (l *TestLogger) InfoCtx(ctx context.Context, msg string, fields ...Fields) {
	l.capture(INFO, msg, l.withTrace(ctx, firstFields(fields)))
}
func (l *TestLogger) WarnCtx(ctx context.Context, msg string, fields ...Fields) {
	l.capture(WARNING, msg, l.withTrace(ctx, firstFields(fields)))
}
func (l *TestLogger) ErrorCtx(ctx context.Context, msg string, fields ...Fields) {
	l.capture(ERROR, msg, l.withTrace(ctx, firstFields(fields)))
}
func (l *TestLogger) FatalCtx(ctx context.Context, msg string, fields ...Fields) {
	l.capture(FATAL, msg, l.withTrace(ctx, firstFields(fields)))
	l.handleFatal(msg)
}

// --- Test assertion helpers ---

// Entries returns a snapshot of all captured log entries, including those
// written by loggers derived via WithFields.
func (l *TestLogger) Entries() []LogEntry {
	l.shared.mu.RLock()
	defer l.shared.mu.RUnlock()
	cp := make([]LogEntry, len(l.shared.entries))
	copy(cp, l.shared.entries)
	return cp
}

// Clear removes all captured entries from the shared store.
func (l *TestLogger) Clear() {
	l.shared.mu.Lock()
	defer l.shared.mu.Unlock()
	l.shared.entries = l.shared.entries[:0]
}

// Count returns the total number of captured entries.
func (l *TestLogger) Count() int {
	l.shared.mu.RLock()
	defer l.shared.mu.RUnlock()
	return len(l.shared.entries)
}

// CountByLevel returns the number of entries at the given level.
func (l *TestLogger) CountByLevel(level Level) int {
	l.shared.mu.RLock()
	defer l.shared.mu.RUnlock()
	n := 0
	for _, e := range l.shared.entries {
		if e.Level == level {
			n++
		}
	}
	return n
}

// HasEntry returns true if at least one entry matches the level and contains
// substr in its message.
func (l *TestLogger) HasEntry(level Level, substr string) bool {
	l.shared.mu.RLock()
	defer l.shared.mu.RUnlock()
	for _, e := range l.shared.entries {
		if e.Level == level && strings.Contains(e.Message, substr) {
			return true
		}
	}
	return false
}

// LastEntry returns the most recently captured entry and true.
// Returns a zero LogEntry and false when nothing has been logged yet.
func (l *TestLogger) LastEntry() (LogEntry, bool) {
	l.shared.mu.RLock()
	defer l.shared.mu.RUnlock()
	if len(l.shared.entries) == 0 {
		return LogEntry{}, false
	}
	return l.shared.entries[len(l.shared.entries)-1], true
}

// --- internals ---

// handleFatal invokes FatalHook when set; otherwise terminates the process.
func (l *TestLogger) handleFatal(msg string) {
	if l.FatalHook != nil {
		l.FatalHook(msg)
		return
	}
	os.Exit(1)
}

func (l *TestLogger) capture(level Level, msg string, fields Fields) {
	if level < l.level {
		return
	}
	entry := LogEntry{
		Level:   level,
		Message: msg,
		Fields:  mergeFields(l.baseFields, fields),
		Time:    time.Now(),
	}
	l.shared.mu.Lock()
	defer l.shared.mu.Unlock()
	l.shared.entries = append(l.shared.entries, entry)
}

func (l *TestLogger) withTrace(ctx context.Context, fields Fields) Fields {
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		out := mergeFields(fields, Fields{"trace_id": traceID})
		return out
	}
	return fields
}

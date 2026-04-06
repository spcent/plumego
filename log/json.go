package log

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
)

// JSONLogger implements StructuredLogger with JSON output format.
// It's thread-safe and suitable for structured logging in production environments.
type JSONLogger struct {
	mu               sync.Mutex
	writeErrOnce     sync.Once
	output           io.Writer
	errorOutput      io.Writer // optional; Error/Fatal go here when non-nil
	level            Level
	verbosity        int
	fields           Fields
	respectVerbosity bool
}

// JSONLoggerConfig configures a JSONLogger instance.
type JSONLoggerConfig struct {
	// Output is the destination for Debug/Info/Warn entries (defaults to os.Stdout).
	Output io.Writer
	// ErrorOutput is an optional separate destination for Error and Fatal entries.
	// When nil, Error/Fatal entries are written to Output.
	// Set to os.Stderr to mirror the typical console-logger convention.
	ErrorOutput io.Writer
	// Level is the minimum log level (defaults to INFO).
	Level Level
	// Fields contains default fields added to every log entry.
	Fields Fields
	// RespectVerbosity applies V(1) filtering to Debug/DebugCtx when enabled.
	// Default false keeps backward-compatible behaviour.
	RespectVerbosity bool
	// Verbosity controls local debug gating when RespectVerbosity is true.
	Verbosity int
}

// NewJSONLogger creates a new JSONLogger with the given configuration.
func NewJSONLogger(config JSONLoggerConfig) *JSONLogger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	return &JSONLogger{
		output:           config.Output,
		errorOutput:      config.ErrorOutput,
		level:            config.Level,
		verbosity:        config.Verbosity,
		fields:           cloneFields(config.Fields),
		respectVerbosity: config.RespectVerbosity,
	}
}

// With is a convenience shortcut for WithFields(Fields{key: value}).
func (l *JSONLogger) With(key string, value any) StructuredLogger {
	return l.WithFields(Fields{key: value})
}

// WithFields returns a new logger with additional fields merged in.
// Each derived logger gets its own independent write-error tracker.
// The fields, outputs, level and verbosity are snapshotted at call time.
func (l *JSONLogger) WithFields(fields Fields) StructuredLogger {
	// These fields are immutable after construction, so a single lock snapshot
	// is sufficient; no risk of partial reads.
	l.mu.Lock()
	merged := mergeFields(l.fields, fields)
	child := &JSONLogger{
		output:           l.output,
		errorOutput:      l.errorOutput,
		level:            l.level,
		verbosity:        l.verbosity,
		fields:           merged,
		respectVerbosity: l.respectVerbosity,
		// writeErrOnce is zero-value (fresh) for every derived logger
	}
	l.mu.Unlock()
	return child
}

// Debug logs a debug message with optional fields.
// When RespectVerbosity is enabled in JSONLoggerConfig, Debug is gated on the
// local Verbosity field (default 0). This differs from gLogger which uses the
// global glog -v flag; the two implementations intentionally use separate
// verbosity sources to remain decoupled.
func (l *JSONLogger) Debug(msg string, fields ...Fields) {
	if l.respectVerbosity && !l.vAt(1) {
		return
	}
	l.log(DEBUG, msg, firstFields(fields), nil)
}

// Info logs an info message with optional fields.
func (l *JSONLogger) Info(msg string, fields ...Fields) {
	l.log(INFO, msg, firstFields(fields), nil)
}

// Warn logs a warning message with optional fields.
func (l *JSONLogger) Warn(msg string, fields ...Fields) {
	l.log(WARNING, msg, firstFields(fields), nil)
}

// Error logs an error message with optional fields.
func (l *JSONLogger) Error(msg string, fields ...Fields) {
	l.log(ERROR, msg, firstFields(fields), nil)
}

// Fatal logs a fatal message then calls os.Exit(1).
func (l *JSONLogger) Fatal(msg string, fields ...Fields) {
	l.log(FATAL, msg, firstFields(fields), nil)
	os.Exit(1)
}

// DebugCtx logs a debug message with context and optional fields.
func (l *JSONLogger) DebugCtx(ctx context.Context, msg string, fields ...Fields) {
	if l.respectVerbosity && !l.vAt(1) {
		return
	}
	l.log(DEBUG, msg, firstFields(fields), ctx)
}

// InfoCtx logs an info message with context and optional fields.
func (l *JSONLogger) InfoCtx(ctx context.Context, msg string, fields ...Fields) {
	l.log(INFO, msg, firstFields(fields), ctx)
}

// WarnCtx logs a warning message with context and optional fields.
func (l *JSONLogger) WarnCtx(ctx context.Context, msg string, fields ...Fields) {
	l.log(WARNING, msg, firstFields(fields), ctx)
}

// ErrorCtx logs an error message with context and optional fields.
func (l *JSONLogger) ErrorCtx(ctx context.Context, msg string, fields ...Fields) {
	l.log(ERROR, msg, firstFields(fields), ctx)
}

// FatalCtx logs a fatal message with context then calls os.Exit(1).
func (l *JSONLogger) FatalCtx(ctx context.Context, msg string, fields ...Fields) {
	l.log(FATAL, msg, firstFields(fields), ctx)
	os.Exit(1)
}

func (l *JSONLogger) log(level Level, msg string, fields Fields, ctx context.Context) {
	// Snapshot the minimum level under the lock so a concurrent SetLevel()
	// cannot race with this read (go race detector would flag the plain read).
	l.mu.Lock()
	minLevel := l.level
	l.mu.Unlock()
	if level < minLevel {
		return
	}

	// Build the entry and marshal it outside the lock to reduce contention.
	entry := l.buildEntry(level, msg, fields, ctx)
	data, err := json.Marshal(entry)
	if err != nil {
		data = []byte(`{"level":"ERROR","msg":"failed to marshal log entry"}`)
	}

	// Only the actual write to the io.Writer is serialised.
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writeRaw(level, data)
}

// buildEntry constructs the log entry map. Called without the mutex held.
func (l *JSONLogger) buildEntry(level Level, msg string, fields Fields, ctx context.Context) map[string]any {
	// Snapshot base fields under lock so WithFields-derived loggers are safe.
	l.mu.Lock()
	baseFields := l.fields
	l.mu.Unlock()

	combined := mergeFields(baseFields, fields)
	entry := make(map[string]any, len(combined)+4)
	for k, v := range combined {
		entry[k] = v
	}

	// Reserved keys are always controlled by logger internals.
	entry["time"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = levelName(level)
	entry["msg"] = msg
	if requestID := contract.RequestIDFromContext(ctx); requestID != "" {
		entry["request_id"] = requestID
	}

	return entry
}

// writeRaw writes pre-marshalled JSON data to the appropriate output.
// Must be called with l.mu held.
func (l *JSONLogger) writeRaw(level Level, data []byte) {
	w := l.writerFor(level)
	if err := writeFull(w, data); err != nil {
		l.reportWriteError(err)
		return
	}
	if err := writeFull(w, []byte("\n")); err != nil {
		l.reportWriteError(err)
	}
}

// writerFor returns the appropriate writer for the given level.
// Error and Fatal go to errorOutput when configured; everything else uses output.
func (l *JSONLogger) writerFor(level Level) io.Writer {
	if l.errorOutput != nil && (level == ERROR || level == FATAL) {
		return l.errorOutput
	}
	return l.output
}

// Start implements the Lifecycle interface (no-op for JSONLogger).
func (l *JSONLogger) Start(ctx context.Context) error {
	return nil
}

// Stop implements the Lifecycle interface (flushes both outputs if they support it).
func (l *JSONLogger) Stop(ctx context.Context) error {
	l.mu.Lock()
	out := l.output
	errOut := l.errorOutput
	l.mu.Unlock()

	var firstErr error
	if syncer, ok := out.(interface{ Sync() error }); ok {
		if err := syncer.Sync(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if errOut != nil && errOut != out {
		if syncer, ok := errOut.(interface{ Sync() error }); ok {
			if err := syncer.Sync(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// SetLevel changes the minimum log level at runtime.
// It is safe to call concurrently with logging.
func (l *JSONLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Level returns the current minimum log level.
func (l *JSONLogger) Level() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *JSONLogger) vAt(level int) bool {
	return level <= l.verbosity
}

func (l *JSONLogger) reportWriteError(err error) {
	l.writeErrOnce.Do(func() {
		_, _ = os.Stderr.WriteString("json logger: failed to write log output: " + err.Error() + "\n")
	})
}

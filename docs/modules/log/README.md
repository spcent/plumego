# Log Module

> **Package Path**: `github.com/spcent/plumego/log` | **Stability**: High | **Priority**: P1

## Overview

The `log/` package provides structured logging capabilities for Plumego applications. Unlike traditional text-based logging, structured logging outputs logs in a machine-readable format (JSON) with typed fields, making logs easier to search, filter, and analyze.

**Key Features**:
- **Structured Output**: JSON-formatted logs with typed fields
- **Log Levels**: Debug, Info, Warn, Error with filtering
- **Contextual Fields**: Attach metadata to log entries
- **Performance**: Minimal allocation, suitable for high-throughput
- **Pluggable Backends**: Console, file, syslog, external services
- **Request Correlation**: Automatic request ID tracking

## Quick Start

### Basic Logging

```go
import "github.com/spcent/plumego/log"

// Simple logging
log.Info("Server started", "port", 8080)
log.Warn("High memory usage", "usage_mb", 512)
log.Error("Database connection failed", "error", err)

// Debug logging (only in debug mode)
log.Debug("Processing request", "user_id", "user-123", "path", "/api/users")
```

**Output** (JSON):
```json
{"level":"info","time":"2025-01-15T10:30:00Z","message":"Server started","port":8080}
{"level":"warn","time":"2025-01-15T10:30:05Z","message":"High memory usage","usage_mb":512}
{"level":"error","time":"2025-01-15T10:30:10Z","message":"Database connection failed","error":"connection refused"}
```

### With Context

```go
// Create logger with default fields
logger := log.With(
    "service", "api",
    "version", "1.0.0",
)

// All logs include service and version
logger.Info("Request received", "path", "/api/users")
// Output: {"level":"info","service":"api","version":"1.0.0","message":"Request received","path":"/api/users"}

logger.Error("Request failed", "error", err)
// Output: {"level":"error","service":"api","version":"1.0.0","message":"Request failed","error":"..."}
```

## Log Levels

### Level Hierarchy

```
DEBUG < INFO < WARN < ERROR
```

- **DEBUG**: Detailed diagnostic information (development only)
- **INFO**: General informational messages
- **WARN**: Warning messages (potential issues)
- **ERROR**: Error messages (actual failures)

### Setting Log Level

```go
// Set minimum level (filters out lower levels)
log.SetLevel(log.LevelInfo) // Only Info, Warn, Error

log.Debug("This won't appear")  // Filtered out
log.Info("This appears")        // Shown
log.Error("This appears")       // Shown
```

### Environment-Based Level

```go
import "os"

// Development: DEBUG
// Production: INFO
level := log.LevelInfo
if os.Getenv("APP_ENV") == "development" {
    level = log.LevelDebug
}
log.SetLevel(level)
```

## Structured Fields

### Adding Fields

```go
// Key-value pairs
log.Info("User login",
    "user_id", "user-123",
    "email", "user@example.com",
    "ip", "192.168.1.1",
)

// Output:
// {"level":"info","message":"User login","user_id":"user-123","email":"user@example.com","ip":"192.168.1.1"}
```

### Field Types

All Go types are supported:

```go
log.Info("Request completed",
    "method", "GET",              // string
    "status", 200,                // int
    "duration_ms", 125.5,         // float64
    "success", true,              // bool
    "timestamp", time.Now(),      // time.Time
    "headers", map[string]string{ // map
        "User-Agent": "curl/7.68.0",
    },
    "tags", []string{"api", "v1"}, // slice
)
```

### Nested Fields

```go
log.Info("Database query",
    "query", "SELECT * FROM users",
    "params", map[string]interface{}{
        "limit":  10,
        "offset": 0,
    },
    "result", map[string]interface{}{
        "rows":     10,
        "duration": "5ms",
    },
)
```

## Request Logging

### HTTP Request Context

```go
func logMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := getRequestID(r)

        // Log request
        log.Info("Request started",
            "request_id", requestID,
            "method", r.Method,
            "path", r.URL.Path,
            "remote_addr", r.RemoteAddr,
            "user_agent", r.Header.Get("User-Agent"),
        )

        // Process request
        next.ServeHTTP(w, r)

        // Log completion
        log.Info("Request completed",
            "request_id", requestID,
            "method", r.Method,
            "path", r.URL.Path,
            "duration_ms", time.Since(start).Milliseconds(),
        )
    })
}
```

### With Logger in Context

```go
// Attach logger to request context
func withLogger(r *http.Request) *http.Request {
    logger := log.With(
        "request_id", getRequestID(r),
        "method", r.Method,
        "path", r.URL.Path,
    )

    ctx := context.WithValue(r.Context(), "logger", logger)
    return r.WithContext(ctx)
}

// Extract logger from context
func loggerFromContext(ctx context.Context) *log.Logger {
    if logger, ok := ctx.Value("logger").(*log.Logger); ok {
        return logger
    }
    return log.Default()
}

// Usage in handler
func handleRequest(w http.ResponseWriter, r *http.Request) {
    logger := loggerFromContext(r.Context())

    logger.Info("Processing request")
    // All logs include request_id, method, path
}
```

## Error Logging

### Error with Context

```go
if err := db.Query(sql); err != nil {
    log.Error("Database query failed",
        "error", err,
        "query", sql,
        "user_id", userID,
    )
    return err
}
```

### Error with Stack Trace

```go
import "runtime/debug"

if err != nil {
    log.Error("Panic recovered",
        "error", err,
        "stack", string(debug.Stack()),
    )
}
```

### Panic Logging

```go
func panicRecovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Error("Panic recovered",
                    "error", err,
                    "path", r.URL.Path,
                    "stack", string(debug.Stack()),
                )

                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()

        next.ServeHTTP(w, r)
    })
}
```

## Performance Considerations

### Conditional Logging

```go
// ❌ Bad: Always formats string (even if debug disabled)
log.Debug(fmt.Sprintf("Processing %d items", len(items)))

// ✅ Good: Only formats if debug enabled
if log.IsDebugEnabled() {
    log.Debug("Processing items", "count", len(items))
}
```

### Field Allocation

```go
// ✅ Efficient: Reuse logger with common fields
logger := log.With("service", "api", "version", "1.0")

for _, item := range items {
    logger.Info("Processing item", "id", item.ID)
}
```

## Security Best Practices

### Never Log Secrets

```go
// ❌ NEVER: Logging passwords, tokens, keys
log.Info("User login", "password", password)           // NEVER!
log.Info("API call", "api_key", apiKey)               // NEVER!
log.Info("JWT token", "token", token)                 // NEVER!

// ✅ Good: Log non-sensitive data only
log.Info("User login", "user_id", userID, "email", email)
log.Info("API call", "endpoint", endpoint, "status", status)
log.Info("JWT issued", "user_id", userID, "expires_in", "15m")
```

### Sanitize PII

```go
func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    return parts[0][:1] + "***@" + parts[1]
}

log.Info("User registered",
    "email", maskEmail(user.Email), // u***@example.com
    "user_id", user.ID,
)
```

## Output Formats

### JSON (Default)

```go
log.SetFormat(log.FormatJSON)

log.Info("Server started", "port", 8080)
// {"level":"info","time":"2025-01-15T10:30:00Z","message":"Server started","port":8080}
```

### Console (Human-Readable)

```go
log.SetFormat(log.FormatConsole)

log.Info("Server started", "port", 8080)
// 2025-01-15 10:30:00 [INFO] Server started port=8080
```

### Custom Format

```go
log.SetFormatter(func(level, message string, fields map[string]interface{}) string {
    // Custom formatting logic
    return fmt.Sprintf("[%s] %s %v", level, message, fields)
})
```

## Output Destinations

### Standard Output (Default)

```go
log.SetOutput(os.Stdout) // Default
```

### File Output

```go
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

log.SetOutput(file)
```

### Multiple Outputs

```go
// Write to both stdout and file
multiWriter := io.MultiWriter(os.Stdout, file)
log.SetOutput(multiWriter)
```

### Rotating Logs

```go
import "gopkg.in/natefinch/lumberjack.v2"

log.SetOutput(&lumberjack.Logger{
    Filename:   "app.log",
    MaxSize:    100,  // MB
    MaxBackups: 3,
    MaxAge:     28,   // days
    Compress:   true,
})
```

## Integration Examples

### With Plumego App

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/log"
)

func main() {
    // Configure logging
    if os.Getenv("APP_ENV") == "production" {
        log.SetLevel(log.LevelInfo)
        log.SetFormat(log.FormatJSON)
    } else {
        log.SetLevel(log.LevelDebug)
        log.SetFormat(log.FormatConsole)
    }

    // Create app
    app := core.New(
        core.WithAddr(":8080"),
        core.WithDebug(os.Getenv("APP_ENV") != "production"),
    )

    log.Info("Starting application",
        "version", "1.0.0",
        "env", os.Getenv("APP_ENV"),
    )

    if err := app.Boot(); err != nil {
        log.Error("Application failed to start", "error", err)
        os.Exit(1)
    }
}
```

## Testing

### Capture Log Output

```go
func TestLogging(t *testing.T) {
    // Capture output
    var buf bytes.Buffer
    log.SetOutput(&buf)
    defer log.SetOutput(os.Stdout)

    // Log something
    log.Info("Test message", "key", "value")

    // Verify output
    output := buf.String()
    if !strings.Contains(output, "Test message") {
        t.Error("Expected log message not found")
    }
}
```

### Mock Logger

```go
type MockLogger struct {
    entries []LogEntry
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
    m.entries = append(m.entries, LogEntry{
        Level:   "info",
        Message: msg,
        Fields:  fields,
    })
}

// Use in tests
func TestHandler(t *testing.T) {
    mockLog := &MockLogger{}
    handler := NewHandler(mockLog)

    // Test handler...

    // Verify logs
    if len(mockLog.entries) != 1 {
        t.Errorf("Expected 1 log entry, got %d", len(mockLog.entries))
    }
}
```

## Module Documentation

- **[Structured Logging](structured-logging.md)** — Detailed guide on structured logging patterns

## Related Documentation

- [Middleware: Observability](../middleware/observability.md) — Request logging middleware
- [Metrics](../metrics/) — Metrics collection and monitoring

## Reference Implementation

See examples:
- `examples/reference/` — Production logging setup
- `examples/api-gateway/` — Request logging patterns

---

**Next Steps**:
1. Read [Structured Logging](structured-logging.md) for advanced patterns
2. Configure log levels for your environment
3. Set up log aggregation (ELK, Splunk, CloudWatch)

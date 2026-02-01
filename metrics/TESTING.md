# Testing with MetricsCollector

This document explains how to write tests with the `MetricsCollector` interface in a maintainable way.

## Problem

The `MetricsCollector` interface has many methods. When writing tests, creating mock implementations can be tedious and fragile. Every time a new method is added to the interface, all mock implementations must be updated.

## Solutions

We provide two approaches for creating test mocks, depending on your needs:

### 1. Simple Mocks with NoopCollector (Recommended for Simple Tests)

For tests that don't need to verify metrics calls, embed `NoopCollector`:

```go
import "github.com/spcent/plumego/metrics"

// Simple mock that does nothing
type myMock struct {
    *metrics.NoopCollector
}

func TestMyFeature(t *testing.T) {
    mock := &myMock{
        NoopCollector: metrics.NewNoopCollector(),
    }

    // Use the mock
    // All MetricsCollector methods are available and do nothing
}
```

**Benefits:**
- Minimal boilerplate
- Future-proof: when new methods are added to the interface, no changes needed
- No methods to implement manually

**Override specific methods if needed:**

```go
type customMock struct {
    *metrics.NoopCollector
    httpCallCount int
}

func (m *customMock) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
    m.httpCallCount++
}
```

### 2. MockCollector for Verification (Recommended for Complex Tests)

For tests that need to verify metrics calls, use `MockCollector`:

```go
import "github.com/spcent/plumego/metrics"

func TestDatabaseMetrics(t *testing.T) {
    mock := metrics.NewMockCollector()

    // Use your component that records metrics
    db := NewInstrumentedDB(realDB, mock, "postgres")
    db.Query("SELECT 1")

    // Verify metrics were recorded
    if mock.DBCallCount() != 1 {
        t.Errorf("expected 1 DB call, got %d", mock.DBCallCount())
    }

    // Check specific call details
    lastCall := mock.GetLastDBCall()
    if lastCall.Operation != "query" {
        t.Errorf("expected query operation, got %s", lastCall.Operation)
    }
}
```

**Features:**
- Automatic call recording for all metric types
- Convenient methods: `HTTPCallCount()`, `DBCallCount()`, `GetLastDBCall()`, etc.
- Thread-safe concurrent access
- Optional custom hooks for advanced scenarios

### 3. Custom Hooks for Advanced Testing

For complex verification logic, use custom hooks:

```go
func TestWithCustomVerification(t *testing.T) {
    mock := metrics.NewMockCollector()

    // Set up custom hook
    var capturedQuery string
    mock.OnObserveDB = func(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
        capturedQuery = query

        // Custom assertions
        if duration > 100*time.Millisecond {
            t.Error("query too slow")
        }
    }

    // Test your code
    // The hook will be called automatically
}
```

## Migration Guide

### Before (Old Pattern)

```go
type mockMetricsCollector struct{}

func (m *mockMetricsCollector) Record(ctx context.Context, record metrics.MetricRecord) {}
func (m *mockMetricsCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {}
func (m *mockMetricsCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {}
func (m *mockMetricsCollector) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {}
func (m *mockMetricsCollector) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {}
func (m *mockMetricsCollector) ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error) {}
func (m *mockMetricsCollector) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {}
func (m *mockMetricsCollector) GetStats() metrics.CollectorStats { return metrics.CollectorStats{} }
func (m *mockMetricsCollector) Clear() {}
```

### After (New Pattern)

**Option 1: Simple embedding**
```go
type mockMetricsCollector struct {
    *metrics.NoopCollector
}

func newMockMetricsCollector() *mockMetricsCollector {
    return &mockMetricsCollector{
        NoopCollector: metrics.NewNoopCollector(),
    }
}
```

**Option 2: Use built-in MockCollector**
```go
// Just use the built-in one
mock := metrics.NewMockCollector()
```

## Best Practices

1. **Use NoopCollector embedding for simple no-op mocks**
   - When you just need a valid MetricsCollector instance
   - When you don't care about verifying metrics calls

2. **Use MockCollector for verification**
   - When you need to assert metrics were recorded
   - When you need to check call counts or parameters

3. **Override specific methods only when needed**
   - Don't override methods you don't test
   - Keep overrides minimal and focused

4. **Use custom hooks for complex logic**
   - When you need custom validation per call
   - When you need to track state across calls

## Complete Example

```go
package mypackage

import (
    "testing"
    "time"
    "github.com/spcent/plumego/metrics"
)

func TestMyComponent_Simple(t *testing.T) {
    // Simple: just need a valid collector
    mock := &mockCollector{
        NoopCollector: metrics.NewNoopCollector(),
    }

    component := NewMyComponent(mock)
    component.DoSomething()

    // No verification needed
}

func TestMyComponent_WithVerification(t *testing.T) {
    // Advanced: verify metrics
    mock := metrics.NewMockCollector()

    component := NewMyComponent(mock)
    component.RecordData()

    // Verify calls
    if mock.HTTPCallCount() != 1 {
        t.Error("expected HTTP call")
    }

    calls := mock.GetHTTPCalls()
    if calls[0].Method != "POST" {
        t.Errorf("expected POST, got %s", calls[0].Method)
    }
}

func TestMyComponent_WithHooks(t *testing.T) {
    // Expert: custom validation
    mock := metrics.NewMockCollector()

    sawSlowQuery := false
    mock.OnObserveDB = func(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
        if duration > 100*time.Millisecond {
            sawSlowQuery = true
        }
    }

    component := NewMyComponent(mock)
    component.ExecuteQuery()

    if !sawSlowQuery {
        t.Error("expected to see slow query")
    }
}

// Simple mock type
type mockCollector struct {
    *metrics.NoopCollector
}
```

## Thread Safety

Both `NoopCollector` and `MockCollector` are thread-safe and can be used in concurrent tests.

```go
func TestConcurrent(t *testing.T) {
    mock := metrics.NewMockCollector()

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mock.ObserveHTTP(ctx, "GET", "/", 200, 100, 10*time.Millisecond)
        }()
    }
    wg.Wait()

    if mock.HTTPCallCount() != 100 {
        t.Errorf("expected 100 calls, got %d", mock.HTTPCallCount())
    }
}
```

## Summary

| Scenario | Recommended Approach |
|----------|---------------------|
| Simple no-op mock | Embed `NoopCollector` |
| Verify call counts | Use `MockCollector` |
| Verify call parameters | Use `MockCollector.GetXXXCalls()` |
| Custom validation logic | Use `MockCollector` with hooks |
| Override single method | Embed `NoopCollector`, override one method |

The new approach eliminates boilerplate and makes tests more maintainable. When new methods are added to `MetricsCollector`, your existing test mocks will continue to work without modifications.

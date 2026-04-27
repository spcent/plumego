# Testing `metrics`

This guide covers tests for the stable `metrics` root only.

## Stable Surface

The stable aggregate collector contract is intentionally small:

- `Record(ctx, MetricRecord)`
- `ObserveHTTP(ctx, method, path, status, bytes, duration)`
- `GetStats() CollectorStats`
- `Clear()`

`BaseMetricsCollector` stores aggregate stats only. It increments
`ErrorRecords` for explicit `MetricRecord.Error` values and for stable HTTP
records with status codes `>= 400`. Use `metrics.NewHTTPRecord(...)` when a
test or owner-side collector needs the canonical stable HTTP record shape,
including its timestamp. Generic records passed to `BaseMetricsCollector` are
not normalized or retained.

Feature-specific helper methods such as DB, MQ, KV, IPC, or PubSub observation
do not belong to stable `metrics`. Use the owning extension packages for those
surfaces.

## Recommended Patterns

### 1. Use `NoopCollector` for no-op dependencies

When a test only needs a valid collector instance, embed `NoopCollector`.

```go
type mockCollector struct {
	*metrics.NoopCollector
}

func TestComponent(t *testing.T) {
	collector := &mockCollector{
		NoopCollector: metrics.NewNoopCollector(),
	}

	component := NewComponent(collector)
	component.Run()
}
```

### 2. Use a small spy for stable-call verification

When a test needs to verify stable collector calls, embed `NoopCollector` and
override only the method you care about.

```go
type spyCollector struct {
	*metrics.NoopCollector

	mu         sync.Mutex
	httpCalls  int
	lastMethod string
}

func (s *spyCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.httpCalls++
	s.lastMethod = method
}

func TestHandlerMetrics(t *testing.T) {
	collector := &spyCollector{NoopCollector: metrics.NewNoopCollector()}

	handler := NewHandler(collector)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if collector.httpCalls != 1 {
		t.Fatalf("httpCalls = %d, want 1", collector.httpCalls)
	}
}
```

### 3. Assert stable behavior through `GetStats()`

`BaseMetricsCollector` tests should assert on `CollectorStats`, not on hidden
internal buffers. When a fan-out collector has deterministic child collectors,
assert exact aggregate totals rather than lower bounds. Multi collectors sum
child-maintained active series; if a child omits `ActiveSeries` but includes a
name breakdown, the breakdown size is used for that child.

```go
func TestBaseCollectorStats(t *testing.T) {
	collector := metrics.NewBaseMetricsCollector()

	collector.Record(context.Background(), metrics.MetricRecord{Name: "jobs_total"})
	collector.ObserveHTTP(context.Background(), http.MethodGet, "/health", 200, 0, 5*time.Millisecond)

	stats := collector.GetStats()
	if stats.TotalRecords != 2 {
		t.Fatalf("TotalRecords = %d, want 2", stats.TotalRecords)
	}
	if stats.ErrorRecords != 0 {
		t.Fatalf("ErrorRecords = %d, want 0", stats.ErrorRecords)
	}
	if stats.NameBreakdown["jobs_total"] != 1 {
		t.Fatalf("jobs_total breakdown = %d, want 1", stats.NameBreakdown["jobs_total"])
	}
}
```

### 4. Treat fan-out constructors as optional wiring helpers

`NewMultiCollector(...)` and `NewMultiHTTPObserver(...)` filter nil inputs. They
return nil when no non-nil targets are supplied, return the single target
unchanged when only one target remains, skip empty internal slots, and fan out
in order otherwise.

## Extension-Owned Helpers

If the code under test uses extension-owned metrics helpers:

- use `x/observability/featuremetrics` for generic record builders outside the
  stable root
- use `x/observability/testmetrics` for richer feature-specific mocks and call
  inspection
- read the owning extension's README or testing guide instead of widening stable
  `metrics/TESTING.md`

Stable `metrics` documentation should not teach extension-only observer methods
as if they were part of `AggregateCollector`.

## Validation

```bash
go test -timeout 20s ./metrics/...
go test -race -timeout 60s ./metrics/...
go vet ./metrics/...
```

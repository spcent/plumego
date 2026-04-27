# Card 5104

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: metrics
Owned Files:
- metrics/multi.go
- metrics/http_observer.go
- metrics/multi_test.go
- metrics/collector.go
- docs/modules/metrics/README.md
Depends On: 5103

Goal:
Make collector and HTTP observer fan-out behavior consistently nil-safe and
well covered.

Scope:
- Keep constructors filtering nil inputs and returning nil for no fan-out target.
- Guard fan-out methods against nil receiver or empty internal slices.
- Add tests for `NewMultiHTTPObserver(...)` and typed nil multi collector usage.
- Keep fan-out ordering unchanged.

Non-goals:
- Do not add asynchronous fan-out, error aggregation, or recovery behavior.
- Do not widen stable metrics into an exporter registry.
- Do not change existing constructor return types.

Files:
- metrics/multi.go
- metrics/http_observer.go
- metrics/multi_test.go
- metrics/collector.go
- docs/modules/metrics/README.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Keep the fan-out nil-safety language in `docs/modules/metrics/README.md`
  accurate.

Done Definition:
- Fan-out constructors and methods handle nil/empty inputs predictably.
- Tests cover collector and HTTP observer fan-out behavior.
- Targeted metrics tests and vet pass.

Outcome:
- Made `MultiCollector` methods no-op on nil receivers and skip nil internal
  collectors defensively.
- Reused the normalized empty stats helper for empty/nil fan-out stats.
- Added coverage for `NewMultiHTTPObserver(...)` empty, single, and multi-target
  behavior plus compile-time interface checks.

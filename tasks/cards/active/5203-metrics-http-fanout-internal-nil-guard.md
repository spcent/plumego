# Card 5203

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: metrics
Owned Files:
- metrics/http_observer.go
- metrics/multi_test.go
- docs/modules/metrics/README.md
Depends On: 5202

Goal:
Align HTTP observer fan-out internals with the defensive behavior already used
by collector fan-out.

Scope:
- Skip nil observers inside `multiHTTPObserver.ObserveHTTP(...)`.
- Add regression coverage for manually constructed internal nil slots.
- Keep constructor filtering, nil return behavior, and fan-out ordering
  unchanged.

Non-goals:
- Do not add async fan-out or panic recovery.
- Do not change `HTTPObserver` or `AggregateCollector` signatures.
- Do not widen stable metrics into an exporter registry.

Files:
- metrics/http_observer.go
- metrics/multi_test.go
- docs/modules/metrics/README.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Keep fan-out nil-safety wording accurate.

Done Definition:
- HTTP fan-out ignores nil internal observer slots.
- Tests cover constructor and internal defensive behavior.
- Targeted metrics tests and vet pass.

Outcome:

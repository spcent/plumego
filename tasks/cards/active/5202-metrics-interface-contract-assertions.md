# Card 5202

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: metrics
Owned Files:
- metrics/collector.go
- metrics/noop.go
- metrics/multi.go
- metrics/collector_test.go
- metrics/noop_test.go
Depends On: 5201

Goal:
Make the stable metrics interface contracts explicit and consistently checked.

Scope:
- Add compile-time assertions for base, no-op, and multi collectors against the
  relevant stable narrow interfaces.
- Remove redundant test-only interface assertions where production assertions are
  clearer.
- Keep the stable public interfaces unchanged.

Non-goals:
- Do not introduce a new facade or registry.
- Do not add extension-owned observer interfaces to stable `metrics`.
- Do not change collector behavior.

Files:
- metrics/collector.go
- metrics/noop.go
- metrics/multi.go
- metrics/collector_test.go
- metrics/noop_test.go

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- None expected; this is a contract assertion cleanup.

Done Definition:
- Production code contains explicit assertions for supported stable interfaces.
- Tests no longer carry redundant assertion-only cases.
- Targeted metrics tests and vet pass.

Outcome:

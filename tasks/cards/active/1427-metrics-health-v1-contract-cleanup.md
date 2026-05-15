# Card 1427

Milestone: v1-breaking-normalization
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: metrics, health
Owned Files:
- metrics/*
- health/*
Depends On:
- 1421

Goal:
- Keep metrics and health stable roots minimal by removing redundant aliases or
  helper surfaces before v1 freeze.

Scope:
- Enumerate no-op implementations, label helpers, health status aliases, and
  readiness/liveness helpers.
- Remove redundant aliases that are not part of the final v1 contract.
- Keep exporters, collectors, tracing, and sampling out of stable metrics.
- Keep HTTP route ownership out of health.
- Add focused tests if label validation or status semantics change.

Non-goals:
- Do not move observability exporter behavior into `metrics`.
- Do not add route registration to `health`.
- Do not introduce dependencies.

Files:
- metrics/*.go
- metrics/*_test.go
- health/*.go
- health/*_test.go

Tests:
- go test -timeout 20s ./metrics ./health
- go vet ./metrics ./health
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update metrics or health docs if public helper surfaces are removed.

Done Definition:
- Redundant helpers are removed or confirmed absent in the card outcome.
- Metrics and health tests pass.
- Boundary checks pass.

Outcome:


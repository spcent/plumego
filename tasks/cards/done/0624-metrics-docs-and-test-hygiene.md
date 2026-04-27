# Card 0624

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: metrics
Owned Files:
- metrics/doc.go
- metrics/noop_test.go
- metrics/TESTING.md
- docs/modules/metrics/README.md
- metrics/module.yaml
Depends On: 5304

Goal:
Polish remaining docs and low-level test hygiene after the nil-safety and stats
normalization work.

Scope:
- Replace ad hoc no-op concurrency signaling with `sync.WaitGroup`.
- Clarify package/module docs for nil-safe collectors, stats construction, and
  active-series semantics.
- Keep module manifest list sizes within schema limits.

Non-goals:
- Do not change runtime behavior.
- Do not document extension-only APIs as stable APIs.
- Do not alter module ownership or dependency rules.

Files:
- metrics/doc.go
- metrics/noop_test.go
- metrics/TESTING.md
- docs/modules/metrics/README.md
- metrics/module.yaml

Tests:
- go test -timeout 20s ./metrics/...
- go vet ./metrics/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Sync package docs, testing guide, module README, and manifest hints with the
  final implemented behavior.

Done Definition:
- No-op concurrency test uses structured synchronization.
- Docs describe implemented behavior only.
- Metrics tests, vet, and module manifest checks pass.

Outcome:
- Replaced ad hoc no-op concurrency signaling with `sync.WaitGroup`.
- Synced package docs, module README, testing guide, and module manifest hints
  with nil-safe collectors, active-series normalization, caller-owned stats, and
  HTTP record timestamp behavior.
- Kept module manifest list sizes within schema limits.

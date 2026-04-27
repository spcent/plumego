# Card 0614

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: metrics
Owned Files:
- metrics/doc.go
- metrics/TESTING.md
- metrics/example_test.go
- docs/modules/metrics/README.md
- metrics/module.yaml
Depends On: 5104

Goal:
Sync metrics package documentation, testing guidance, and examples with the
implemented stable metrics surface.

Scope:
- Document the final stable metrics semantics from cards 5101-5104.
- Keep examples focused on `Recorder`, `HTTPObserver`, base/noop/multi
  collectors, and owner-side extension behavior.
- Make module guidance clear about HTTP status classification, record helper
  usage, and fan-out nil behavior.

Non-goals:
- Do not document extension-only helper APIs as stable package APIs.
- Do not change code behavior unless a documentation example must compile.
- Do not alter module ownership or allowed imports.

Files:
- metrics/doc.go
- metrics/TESTING.md
- metrics/example_test.go
- docs/modules/metrics/README.md
- metrics/module.yaml

Tests:
- go test -timeout 20s ./metrics/...
- go vet ./metrics/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Sync `metrics/doc.go`, `metrics/TESTING.md`, `docs/modules/metrics/README.md`,
  and `metrics/module.yaml`.

Done Definition:
- Package docs and examples describe implemented behavior only.
- Metrics examples compile and pass.
- Targeted tests, vet, and module manifest checks pass.

Outcome:
- Synced package docs, module README, testing guide, examples, and module
  manifest with the implemented HTTP record, error classification, stats
  snapshot, and fan-out semantics.
- Added a compiling `NewHTTPRecord` example.
- Kept extension-only collectors and exporters documented as extension-owned.

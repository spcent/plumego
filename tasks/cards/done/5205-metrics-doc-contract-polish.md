# Card 5205

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: metrics
Owned Files:
- metrics/doc.go
- metrics/noop.go
- metrics/TESTING.md
- docs/modules/metrics/README.md
- metrics/module.yaml
Depends On: 5204

Goal:
Polish metrics documentation so package docs, no-op comments, testing guidance,
and module hints consistently describe the final stable behavior.

Scope:
- Fix stale no-op stats wording now that empty breakdown maps are initialized.
- Clarify base aggregate-only behavior, HTTP record timestamp ownership, and
  fan-out nil behavior.
- Keep module manifest lists within schema limits.

Non-goals:
- Do not document extension-only APIs as stable APIs.
- Do not change runtime behavior.
- Do not alter module ownership or dependency rules.

Files:
- metrics/doc.go
- metrics/noop.go
- metrics/TESTING.md
- docs/modules/metrics/README.md
- metrics/module.yaml

Tests:
- go test -timeout 20s ./metrics/...
- go vet ./metrics/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Sync package comments, no-op comments, testing guide, module README, and
  module manifest hints.

Done Definition:
- Documentation describes implemented stable behavior only.
- Metrics tests, vet, and module manifest checks pass.

Outcome:
- Updated package docs, no-op comments, testing guide, module README, and module
  manifest hints to describe aggregate-only base behavior, timestamped HTTP
  records, initialized no-op stats maps, and nil-safe fan-out.
- Kept manifest list sizes within module-manifest schema limits.

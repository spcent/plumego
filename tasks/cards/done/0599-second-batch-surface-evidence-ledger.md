# Card 0599

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: specs
Owned Files:
- specs/extension-beta-evidence.yaml
- internal/checks/extension-beta-evidence/main.go
- docs/extension-evidence/x-data.md
- docs/extension-evidence/x-discovery.md
- docs/extension-evidence/x-messaging.md
Depends On: 2301

Goal:
Start the second-batch candidate ledger for small surfaces only:
`x/data/file`, `x/data/idempotency`, `x/discovery` core/static, and
`x/messaging` app-facing service.

Scope:
- Add checker support for package/surface candidates that are not root module
  promotions.
- Add ledger entries for the selected small surfaces.
- Keep blockers and current decisions experimental.

Non-goals:
- Do not evaluate root `x/data`, `x/discovery`, or `x/messaging` for beta.
- Do not add subordinate primitive promotion for mq/pubsub/scheduler/webhook.
- Do not change module statuses.

Files:
- `specs/extension-beta-evidence.yaml`
- `internal/checks/extension-beta-evidence/main.go`
- `docs/extension-evidence/x-data.md`
- `docs/extension-evidence/x-discovery.md`
- `docs/extension-evidence/x-messaging.md`

Tests:
- `go test ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/0599-second-batch-surface-evidence-ledger.md`

Docs Sync:
- Required because promotion candidate evidence changes.

Done Definition:
- Second-batch small surfaces are represented in the evidence ledger with
  explicit blockers and without root-family promotion.

Outcome:
- Added `surface_candidates` support to `extension-beta-evidence` for small
  package or feature surfaces that do not imply root-module promotion.
- Registered second-batch surfaces for `x/data:file`, `x/data:idempotency`,
  `x/discovery:core-static`, and `x/messaging:app-facing-service`.
- Updated second-batch evidence docs to point at the ledger entries and preserve
  root-family experimental status.

Validations:
- `go test ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/0599-second-batch-surface-evidence-ledger.md`

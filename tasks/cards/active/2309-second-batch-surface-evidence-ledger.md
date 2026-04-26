# Card 2309

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
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
- `scripts/check-spec tasks/cards/done/2309-second-batch-surface-evidence-ledger.md`

Docs Sync:
- Required because promotion candidate evidence changes.

Done Definition:
- Second-batch small surfaces are represented in the evidence ledger with
  explicit blockers and without root-family promotion.

Outcome:

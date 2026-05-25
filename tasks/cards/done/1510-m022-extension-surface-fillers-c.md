# Card 1510

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/messaging
Owned Files:
- `x/messaging/mq/module.yaml`
- `x/messaging/pubsub/module.yaml`
- `x/messaging/scheduler/module.yaml`
- `x/observability/module.yaml`
- `x/observability/devtools/module.yaml`
- `x/observability/ops/module.yaml`
Depends On: 1503

Goal:
- Add missing `public_entrypoints` to the listed `x/messaging` and
  `x/observability` manifests.
- Remove the ghost `core/components/**` boundary rule from those manifests.

Scope:
- Touch only the manifests in `Owned Files`.

Non-goals:
- Do not change runtime behavior.
- Do not resolve MQTT/AMQP placeholder behavior here; that belongs to card
  `1516`.

Files:
- `x/messaging/mq/module.yaml`
- `x/messaging/pubsub/module.yaml`
- `x/messaging/scheduler/module.yaml`
- `x/observability/module.yaml`
- `x/observability/devtools/module.yaml`
- `x/observability/ops/module.yaml`

Acceptance Tests:
<!-- none; manifest-only card -->

Tests:
- `go test -timeout 20s ./x/messaging/... ./x/observability/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- None expected.

Validation:
- `go test -timeout 20s ./x/messaging/... ./x/observability/...`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Added explicit `public_entrypoints` to `x/messaging/mq`,
  `x/messaging/pubsub`, `x/messaging/scheduler`, `x/observability`,
  `x/observability/devtools`, and `x/observability/ops` so the machine-readable
  surface now exposes their canonical constructors, configs, and core exported
  types.
- Removed the stale `core/components/**` forbidden import from all six
  manifests because the path does not exist and was only historical template
  residue.
- Validation:
  - `go test -timeout 20s ./x/messaging/... ./x/observability/...`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/public-entrypoints-sync`
  - `gofmt -l .`

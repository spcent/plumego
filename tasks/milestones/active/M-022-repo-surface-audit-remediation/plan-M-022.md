# Plan for M-022: Repo Surface Audit Remediation

Milestone: `M-022`
Objective: Turn the verified 2026-05 repo-surface audit into a bounded queue that first fixes machine-readable drift, then converges duplicate extension stacks, then closes placeholder and compatibility debt.
Constraints: one primary module per card; max 5 files per card; max 3 validation commands per card; no new dependencies; no unplanned stable API renames.
Affected Modules: `core`, `log`, `contract`, `middleware`, `store`, `x/ai`, `x/data`, `x/frontend`, `x/gateway`, `x/messaging`, `x/observability`, `x/openapi`, `x/resilience`, `x/rpc`, `x/validate`, `x/websocket`, `reference/workerfleet`, `specs`, `tasks`

## Phase Map

- Phase 1: stable-root manifest and docs drift cleanup
- Phase 2: extension manifest inventory normalization
- Phase 3: duplicate resilience, metrics, error, and placeholder surface convergence

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1503 | Sync `core/module.yaml` with `Group`, `Run`, `RouteGroup`, and remove redundant ghost `forbidden_imports`. | `core` | `core/module.yaml` | none | `module-manifests`, `public-entrypoints-sync` |
| 1504 | Expand `log/module.yaml` to match exported construction types and fix the `glog.Fields` doc comment. | `log` | `log/module.yaml`, `log/logger.go` | none | `go test ./log/...`, `module-manifests` |
| 1505 | Replace `contract/module.yaml` wildcard sentinel inventory with concrete exported names and clearer wording. | `contract` | `contract/module.yaml` | none | `go test ./contract/...`, `public-entrypoints-sync` |
| 1506 | Clarify `middleware/module.yaml` selection and constructor guidance around rate limiting and documented exceptions. | `middleware` | `middleware/module.yaml`, `docs/modules/middleware/README.md` | none | `module-manifests`, `agent-workflow` |
| 1507 | Normalize `x/data` family entrypoints: remove concept labels, add missing subpackage surfaces, and decide whether `pgx`/`sqlx` need docs-only clarification or a later rename plan. | `x/data` | `x/data/module.yaml`, `x/data/migrate/module.yaml`, `x/data/pgx/module.yaml`, `x/data/sqlx/module.yaml`, `docs/modules/x/data/README.md` | 1503 | `go test ./x/data/...`, `module-manifests` |
| 1508 | Add missing `public_entrypoints` to `x/frontend`, `x/openapi`, and `x/validate` manifests. | `x/frontend` | `x/frontend/module.yaml`, `x/openapi/module.yaml`, `x/validate/module.yaml`, `docs/modules/x/frontend/README.md`, `docs/modules/x/openapi/README.md` | 1503 | `public-entrypoints-sync`, `module-manifests` |
| 1509 | Add missing subordinate manifest entrypoints and remove ghost `core/components/**` rules in `x/gateway` and `x/rpc`. | `x/gateway` | `x/gateway/discovery/module.yaml`, `x/gateway/ipc/module.yaml`, `x/rpc/client/module.yaml`, `x/rpc/gateway/module.yaml`, `x/rpc/server/module.yaml` | 1503 | `go test ./x/gateway/... ./x/rpc/...`, `module-manifests` |
| 1510 | Add missing manifest entrypoints and clean ghost path rules in `x/messaging` and `x/observability`. | `x/messaging` | `x/messaging/mq/module.yaml`, `x/messaging/pubsub/module.yaml`, `x/messaging/scheduler/module.yaml`, `x/observability/module.yaml`, `x/observability/devtools/module.yaml` | 1503 | `go test ./x/messaging/... ./x/observability/...`, `module-manifests` |
| 1511 | Design and land the first migration slice from `x/ai/circuitbreaker` and `x/ai/ratelimit` toward shared `x/resilience/*` primitives. | `x/ai` | `x/ai/module.yaml`, `x/ai/resilience/provider.go`, `x/ai/circuitbreaker/*`, `x/ai/ratelimit/*`, `x/resilience/module.yaml` | 1507 | `go test ./x/ai/... ./x/resilience/...`, `dependency-rules` |
| 1512 | Align `x/ai/metrics` with stable `metrics` contracts or document a deliberate adapter boundary. | `x/ai` | `x/ai/metrics/*`, `metrics/module.yaml`, `docs/modules/metrics/README.md`, `docs/modules/x/ai/README.md` | 1511 | `go test ./x/ai/... ./metrics/...`, `dependency-rules` |
| 1513 | Replace duplicate `x/rpc/gateway.ErrHandlerNil` with one canonical error path and update the manifest boundary if required. | `x/rpc/gateway` | `x/rpc/gateway/transcoder.go`, `x/rpc/gateway/module.yaml`, `contract/module.yaml` | 1505 | `go test ./x/rpc/...`, `dependency-rules` |
| 1514 | Correct the verified panic-wrapper scope in `x/ai`: constructors stay as-is unless fallible, but wrapper APIs such as `Register`, `Tags`, and `New*` panic shims must either gain migration windows or removal plans. | `x/ai` | `x/ai/provider/manager.go`, `x/ai/streaming/streaming.go`, `x/ai/resilience/provider.go`, `x/ai/semanticcache/provider.go`, `x/ai/metrics/metrics.go` | 1511 | `go test ./x/ai/...`, `deprecation-inventory -strict` |
| 1515 | Make `store/kv` defaults discoverable without forcing a breaking `Options` rename. | `store` | `store/kv/options.go`, `store/kv/kv.go`, `store/module.yaml`, `docs/modules/store/README.md` | none | `go test ./store/...`, `module-manifests` |
| 1516 | Resolve the unsupported MQTT/AMQP config surface by either documenting the placeholder sharply or removing the fields with a migration note. | `x/messaging/mq` | `x/messaging/mq/config.go`, `x/messaging/mq/module.yaml`, `docs/modules/x/messaging/README.md`, `specs/deprecation-inventory.yaml` | 1510 | `go test ./x/messaging/mq/...`, `deprecation-inventory -strict` |
| 1517 | Give `x/websocket` compatibility aliases an explicit removal window or remove them after caller migration. | `x/websocket` | `x/websocket/auth.go`, `x/websocket/module.yaml`, `specs/deprecation-inventory.yaml`, `docs/modules/x/websocket/README.md` | none | `go test ./x/websocket/...`, `deprecation-inventory -strict` |
| 1518 | Reconcile `reference/workerfleet` placeholder status with current implementation reality and docs positioning. | `reference/workerfleet` | `reference/workerfleet/internal/app/service.go`, `reference/workerfleet/README.md`, `specs/deprecation-inventory.yaml` | none | `go test ./reference/workerfleet/...`, `deprecation-inventory -strict` |

## Dependency Edges

- `1503 -> 1507`
- `1503 -> 1508`
- `1503 -> 1509`
- `1503 -> 1510`
- `1505 -> 1513`
- `1507 -> 1511`
- `1511 -> 1512`
- `1511 -> 1514`
- `1510 -> 1516`

## Parallel Groups

- Group A: `1503`, `1504`, `1505`, `1506`
- Group B: `1507`, `1508`, `1509`, `1510`
- Group C: `1515`, `1517`, `1518`

## Risk Register

- Risk: manifest cleanup lands with inconsistent `public_entrypoints` style again.
  Mitigation: pick one explicit convention per card and update the owning family docs when the convention changes.
- Risk: `x/ai` resilience consolidation changes behavior instead of just ownership.
  Mitigation: require migration tests first; prefer adapters and caller migration before deleting old primitives.
- Risk: deprecation cleanup crosses too many modules for one pass.
  Mitigation: keep alias and placeholder work in separate cards and require explicit removal-window text when deletion is deferred.

## Verification Strategy

- Card-level checks: the quick gates listed per card, plus focused module tests when code changes.
- Milestone-level checks: `dependency-rules`, `agent-workflow`, `module-manifests`, `reference-layout`, `public-entrypoints-sync`, `deprecation-inventory -strict`, then full `go test`, `go vet`, and `gofmt -l .`.

## Finding Disposition

- Verified but corrected: `x/ai/provider.NewManager` and `x/ai/streaming.NewStreamManager` are not panic constructors; the panic wrappers are `Register`, plus `x/ai/metrics.Tags`, `NewResilientProvider`, and `NewSemanticCachingProvider`.
- Verified but narrowed: `middleware` constructor naming drift is already partly covered by current `agent_hints`; queue only wording cleanup, not a mechanical rename campaign.
- Verified but low-confidence product issue: `x/data/pgx` and `x/data/sqlx` names can mislead, but package comments already disclaim concrete dependencies; keep this as docs-or-later rename planning, not immediate API churn.
- Not queued as a standalone defect: `gLogger` is private and not public API drift on its own.

## Checkpoints

| Phase | Checkpoint Gate | Status |
|-------|-----------------|--------|
| Phase 1 | `go run ./internal/checks/module-manifests` | pending |
| Phase 2 | `go run ./internal/checks/public-entrypoints-sync` | pending |
| Phase 3 | `go run ./internal/checks/deprecation-inventory -strict` | pending |

## Exit Condition

- all planned cards completed or explicitly superseded
- all phase checkpoints recorded as passed
- verify report shows pass
- milestone acceptance criteria ready for PR packaging

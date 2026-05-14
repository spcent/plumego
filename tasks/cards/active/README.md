# Active Task Cards

`tasks/cards/active/` is the live execution queue for near-term Plumego work.

Rules:

- keep only actionable cards here
- move completed cards to `tasks/cards/done/`
- prefer one primary module
- keep each card small enough for one focused implementation pass
- make the next card obvious without rereading the whole roadmap

State model:

- `active` means the card is executable now
- `blocked` means the card still matters but cannot be executed yet
- `done` means the work is complete and validated
- `superseded` means the card is no longer the right execution unit

Current implementation:

- only `active/` and `done/` exist as physical directories today
- if a card is blocked or superseded, do not leave it looking active by omission
- mark the state in the card body until a dedicated blocked or superseded queue is introduced
- remove stale items from the front of the queue quickly so the live queue stays trustworthy

Maintenance rules:

- the front of `active/` should represent the real next work, not an outdated snapshot
- completed cards should be archived promptly with the actual validations that ran
- new cards should be added only after they are small, bounded, and aligned with the current roadmap
- if a card expands beyond one focused pass, replace it with smaller cards rather than stretching its scope

Read order:

1. `docs/CODEX_WORKFLOW.md`
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
4. `specs/repo.yaml`
5. `specs/task-routing.yaml`
6. the active card itself

The active queue is an execution surface, not an archive.

## Active Queue

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| [1367](1367-x-tenant-beta-evidence-closure.md) | P2 | x/tenant | Complete beta evidence closure for `x/tenant` when real release refs and owner sign-off are available |
| [1370](1370-x-ai-stable-tier-beta-evidence-closure.md) | P2 | x/ai | Complete beta evidence closure for `x/ai` stable-tier subpackages |
| [1371](1371-x-data-surface-beta-evidence-closure.md) | P2 | x/data | Complete beta evidence closure for selected `x/data` surfaces |
| [1372](1372-x-discovery-surface-beta-evidence-closure.md) | P2 | x/discovery | Complete beta evidence closure for the `x/discovery` core/static surface |
| [1373](1373-x-messaging-service-beta-evidence-closure.md) | P2 | x/messaging | Complete beta evidence closure for the `x/messaging` app-facing service surface |
| [1375](1375-v1-rc-tag-and-observation-window.md) | P0 | release | Tag `v1.0.0-rc.1`, observe the release candidate, and define the exact path from rc.1 to final `v1.0.0` |
| [1387](1387-workerfleet-runtime-loop-error-observability.md) | P0 | reference/workerfleet/internal/app | Stop silently dropping runtime-loop, alert-evaluation, and notification errors |
| [1388](1388-workerfleet-store-query-context-propagation.md) | P0 | reference/workerfleet/internal/platform/store | Propagate request cancellation and deadlines through workerfleet read/query store paths |
| [1389](1389-workerfleet-ingest-store-context-propagation.md) | P0 | reference/workerfleet/internal/domain | Propagate context through worker registration, heartbeat ingest, and write-side persistence |
| [1390](1390-workerfleet-service-handler-dto-boundary.md) | P1 | reference/workerfleet/internal/app | Remove the reverse dependency where app service methods expose handler DTOs |
| [1391](1391-workerfleet-app-entrypoint-assembly-split.md) | P1 | reference/workerfleet/internal/app | Move workerfleet toward the standard-service thin-entrypoint shape without changing behavior |
| [1392](1392-workerfleet-design-doc-runtime-sync.md) | P2 | reference/workerfleet | Synchronize workerfleet docs with implemented runtime loops, alert loop, metrics, and shutdown behavior |
| [1393](1393-workerfleet-worker-ingress-auth-hardening.md) | P0 | reference/workerfleet/internal/handler | Fail closed on worker registration and heartbeat ingress when production auth is configured |
| [1413](1413-x-webhook-inbound-outbound-split.md) | P2 | x/webhook | Split outbound dispatch and inbound provider error mapping |
| [1414](1414-x-data-sharding-router-resolver-split.md) | P2 | x/data/sharding | Split sharding router planning and resolver rule helpers |
| [1415](1415-x-cache-redis-compat-constructor-cleanup.md) | P2 | x/cache/redis | Clarify Redis compatibility constructors and mutable adapter options |
| [1416](1416-x-ai-stable-tier-constructor-cleanup.md) | P1 | x/ai | Prefer error-returning dynamic registration in stable-tier AI subpackages |
| [1417](1417-x-tenant-config-manager-split.md) | P1 | x/tenant/config | Split tenant config SQL and legacy quota migration helpers |
| [1418](1418-x-frontend-response-negotiation-cleanup.md) | P2 | x/frontend | Isolate response interception and content negotiation behavior |
| [1419](1419-x-discovery-backend-boundary-cleanup.md) | P2 | x/discovery | Clarify core/static discovery boundary and unsupported backend behavior |
| [1420](1420-x-fileapi-transport-boundary-audit.md) | P2 | x/fileapi | Audit file API transport ownership over data/file and store/file contracts |

## Execution Completeness Checklist

For any card that removes or changes an exported symbol, the executor MUST:

1. Run `rg -n --glob '*.go' 'OldSymbolName' .` **before** editing to get
   the full caller list.
2. Address every file in the list in the same PR (migrate, update, or explicitly
   discard with `_ =`).
3. After editing, re-run the same search — for deletions the result must be empty.
4. Update tests that assert on the old behaviour in the same commit.
5. Mark Done only after `go build ./...` and `go test` both pass.

See `AGENTS.md §7.1` for the full protocol.

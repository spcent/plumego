# Card 0706

Milestone: M-002
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: core
Owned Files: tasks/cards/done/0706-stable-root-api-inventory.md, core/module.yaml, router/module.yaml, contract/module.yaml, middleware/module.yaml
Depends On: M-001

Goal:
Produce the stable-root API and behavior inventory needed to freeze M-002 change boundaries before implementation cards start.

Scope:
Inventory exported symbols and high-priority behavior contracts for `core`, `router`, `contract`, `middleware`, and `security`, then record the invariants and forbidden-change list in this card Outcome.

Non-goals:
Do not change runtime code.
Do not rename, remove, or add exported symbols.
Do not audit `x/*` packages.

Files:
core/module.yaml
router/module.yaml
contract/module.yaml
middleware/module.yaml
security/module.yaml
tasks/cards/done/0706-stable-root-api-inventory.md

Tests:
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests

Docs Sync:
Record the matrix in this card Outcome. Do not update user-facing docs unless the audit finds misleading stable-root claims.

Done Definition:
Outcome lists exported-surface invariants for `core`, `router`, `contract`, `middleware`, and `security`.
Outcome identifies forbidden changes for M-002.
Next implementation cards have clear boundaries.

Outcome:
Completed.

Stable-root invariants:

| Module | Public surface to preserve | High-priority behavior contracts |
| --- | --- | --- |
| `core` | `New`, `App`, `AppDependencies`, typed config structs, route helpers, `Use`, `Prepare`, `Server`, `Shutdown`, `URL`, `Routes`, `Logger` | Explicit app construction, explicit middleware and route registration, stdlib `http.Handler` compatibility, no hidden registration, lifecycle errors returned instead of swallowed |
| `router` | `NewRouter`, `Router`, `AddRoute`, `Group`, `Param`, `WithRouteName`, `WithMethodNotAllowed`, `Static`, `StaticFS` | Deterministic method/path matching, param extraction, group prefix behavior, route metadata, reverse routing, no response-writing or auth policy |
| `contract` | `WriteResponse`, `WriteError`, `WriteJSON`, `NewErrorBuilder`, `APIError`, `ErrorResponse`, bind helpers, request/trace context helpers | One canonical success path, one canonical error path, explicit context accessor pairs, no mutable request bag or protocol/session ownership |
| `middleware` | `Middleware`, `Chain`, `NewChain`, `Apply`, narrow subpackage constructors | Standard `func(http.Handler) http.Handler`, exactly-once next behavior, transport-only cross-cutting concerns, no business DTO or persistence lookup |
| `security` | `authn`, `jwt`, `headers`, `input`, `abuse`, `password` subpackage entrypoints | Fail closed on verification errors, no secret logging, timing-safe secret checks, explicit auth primitives, no tenant/session lifecycle ownership |

Forbidden changes for M-002:

- Do not add new stable public APIs unless a later card explicitly supersedes this inventory.
- Do not move `x/*`, tenant, gateway, observability exporter, or persistence topology behavior into stable roots.
- Do not create alternate response envelopes, alternate error-helper families, hidden globals, `init()` registration, or context service locators.
- Do not change exported symbol names, return values, or semantic behavior without following the exported symbol completeness protocol.
- Do not introduce new dependencies.

Implementation-card boundaries:

- Card 0707 may add `core` regression tests and compatibility fixes only.
- Card 0708 may add `router` regression tests and matching-behavior fixes only.
- Card 0709 may add `contract` regression tests and canonical response/error/context fixes only.
- Follow-up cards should cover `middleware` and `security` separately before expanding to `store`, `health`, `log`, or `metrics`.

Validation:

- `go run ./internal/checks/dependency-rules` passed.
- `go run ./internal/checks/module-manifests` passed.


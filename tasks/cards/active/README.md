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

- `active/`, `blocked/`, and `done/` exist as physical directories today
- if a card is blocked, move it to `tasks/cards/blocked/`
- if a card is superseded, replace it with a clearer active card and retire the stale unit intentionally
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

Cards are grouped by milestone. Execute M-008 first; M-009/M-010/M-011/M-012/M-013
become executable after M-008 completes. Later milestones follow the dependency
chain in tasks/milestones/ROADMAP.md.

### M-008 — v1.1.0 Release Execution (P0, execute first)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

### M-011 — Benchmark Suite (P0, parallel with M-008)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1512 | P0 | benchmark | Write benchmark/go.mod, benchmark/README.md, and docs/benchmarks/results-v1.1.0.md |

### M-009 — Beta Promotions Round 1 (P1, depends on M-008)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1370 | P1 | x/ai | Promote x/ai stable-tier subpackages (provider, session, tool, streaming) to beta |
| 1367 | P1 | x/tenant | Promote x/tenant to beta |
| 1372 | P1 | x/gateway/discovery | Promote x/gateway/discovery core-static to beta |
| 1371 | P1 | x/data | Promote x/data/file and x/data/idempotency to beta |
| 1513 | P1 | extension evidence | Update docs/EXTENSION_MATURITY.md and DEPRECATION.md for Round 1 promotions |

### M-010 — Beta Promotions Round 2 (P1, depends on M-009)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1373 | P1 | x/messaging | Promote x/messaging app-facing service to beta |
| 1514 | P1 | x/frontend | Evaluate x/frontend for beta; record promotion or explicit blocker |
| 1515 | P1 | extension evidence | Update docs/EXTENSION_MATURITY.md and DEPRECATION.md for Round 2 |

### M-012 — Input Validation Bridge (P1, depends on M-008)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1520 | P1 | x/validate | Create x/validate with Validator interface, Bind[T], and BindJSON[T] |
| 1521 | P1 | x/validate | Create x/validate/playground adapter wrapping go-playground/validator v10 |
| 1522 | P1 | reference/with-rest | Add x/validate usage example to with-rest create-item handler |

### M-013 — Migration Guides (P2, depends on M-008)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1530 | P2 | docs | Write docs/migration/from-gin.md with concept mapping table and code snippets |
| 1531 | P2 | docs | Write docs/migration/from-echo.md with Context abstraction translation |
| 1532 | P2 | docs | Write docs/migration/from-chi.md focused on minimal-delta conventions |
| 1533 | P2 | docs | Write docs/migration/middleware-compat.md and update docs/ADOPTION_PATH.md |

### M-014 — OpenAPI 3.1 Generation (P1, depends on M-009)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1540 | P1 | x/openapi | Create x/openapi Generator with Op hint type and Document struct |
| 1541 | P1 | x/openapi | Add JSON and YAML serialisation to x/openapi (no external YAML library) |
| 1542 | P1 | cmd/plumego | Add `plumego generate spec` subcommand wiring x/openapi.Generator |

### M-015 — Database Adapters (P2, depends on M-009)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1550 | P2 | x/data | Create x/data/pgx adapting pgx v5 to store/db interfaces; offline tests |
| 1551 | P2 | x/data | Create x/data/sqlx adapting sqlx to store/db interfaces; offline tests |
| 1552 | P2 | x/data | Create x/data/migrate wrapping goose; add plumego migrate up/down/status |

### M-016 — Event-Driven Reference Architecture (P2, depends on M-010)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1560 | P2 | reference/with-events | Scaffold reference/with-events with core.App and messaging service wiring |
| 1561 | P2 | reference/with-events | Implement order publisher (outbox) and idempotent consumer |
| 1562 | P2 | reference/with-events | Implement delayed retry job using x/messaging/scheduler |
| 1563 | P2 | reference/with-events | Implement webhook sender with backoff retry; write README.md |

### M-017 — gRPC Support (P3, depends on M-014)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1570 | P3 | x/rpc | Create x/rpc/server wrapping grpc.Server with context-aligned lifecycle |
| 1571 | P3 | x/rpc | Create x/rpc/client connection pool with logging/retry/tracing interceptors |
| 1572 | P3 | x/rpc | Create x/rpc/gateway HTTPTranscoder; add reference/with-rpc example |

### M-018 — Multi-Tenant Admin Reference (P3, depends on M-010)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1580 | P3 | reference/with-tenant-admin | Scaffold with-tenant-admin with core.App, admin auth middleware, route groups |
| 1581 | P3 | reference/with-tenant-admin | Implement tenant CRUD handlers (create, get, suspend, delete) |
| 1582 | P3 | reference/with-tenant-admin | Implement quota admin handlers (get, set, reset) using x/tenant/quota |
| 1583 | P3 | reference/with-tenant-admin | Implement usage recording handlers; write README.md |

### M-019 — Module Ecosystem Foundation (P3, depends on M-017)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1590 | P3 | specs | Create specs/community-extension.schema.yaml and validation check tool |
| 1591 | P3 | cmd/plumego | Implement `plumego add` command with schema validation before go get |
| 1592 | P3 | docs | Write docs/EXTENSION_AUTHORING.md with x/rpc as worked example |

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

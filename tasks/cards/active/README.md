# Active Task Cards

`tasks/cards/active/` is the live execution queue for near-term Plumego work.

Rules:

- keep only actionable cards here
- move completed cards to `tasks/cards/done/`
- prefer one primary module
- keep each card small enough for one focused implementation pass
- keep each card within the context budget limits: five files, three validation commands, and one reversible commit
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

1. `AGENTS.md`
2. the matching `specs/task-routing.yaml` entry
3. the active card itself
4. the card's `Recipe:` file, if present
5. the target module manifest, when module behavior changes

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

### M-009 — Beta Promotions Round 1 (P1, depends on M-008)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1500-m009-x-tenant-ga-evaluation.md | P1 | x/tenant | GA promotion evaluation: release evidence, snapshot comparison, owner sign-off |
| 1501-m009-x-ai-stable-tier-beta-evaluation.md | P1 | x/ai | Beta evaluation for provider, session, streaming, tool subpackages individually |

### M-010 — Beta Promotions Round 2 (P1, depends on M-009)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

### M-012 — Input Validation Bridge (P1, depends on M-008)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

### M-013 — Migration Guides (P2, depends on M-008)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

### M-014 — OpenAPI 3.1 Generation (P1, depends on M-009)

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| 1502-m014-x-openapi-cli-integration.md | P1 | x/openapi | Add `plumego openapi generate` subcommand, JSON serialization, expand primer |

### M-015 — Database Adapters (P2, depends on M-009)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

### M-016 — Event-Driven Reference Architecture (P2, depends on M-010)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

### M-017 — gRPC Support (P3, depends on M-014)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

### M-018 — Multi-Tenant Admin Reference (P3, depends on M-010)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

### M-019 — Module Ecosystem Foundation (P3, depends on M-017)

| Card | Priority | Primary module | Focus |
|---|---|---|---|

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

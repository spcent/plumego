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
| 2259 | P1 | x/rest | Extension beta readiness decision |
| 2260 | P1 | x/websocket | Extension beta readiness decision |
| 2261 | P1 | x/tenant | Extension beta readiness decision |
| 2262 | P1 | x/observability | Extension beta readiness decision |
| 2263 | P1 | x/gateway | Extension beta readiness decision |
| 2264 | P2 | core | Scenario entrypoint map |
| 2265 | P2 | x/rest | Runnable REST resource example |
| 2266 | P2 | x/tenant | Multi-tenant API example |
| 2267 | P2 | x/gateway | Edge gateway example |
| 2268 | P1 | cmd/plumego | Canonical scaffold sync |
| 2269 | P2 | cmd/plumego | REST API scaffold profile |
| 2270 | P2 | middleware | Production security profile |
| 2271 | P2 | x/observability | Observability and ops profile |
| 2272 | P2 | x/ai | Stable-tier AI adoption path |

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

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

## Blocked Follow-Up Queue

These cards are intentionally kept in `active/` with `State: blocked` because
the missing evidence depends on real release refs and owner sign-off.

| Card | Priority | Primary module | Focus |
|---|---|---|---|
| [0723-x-rest-beta-evidence-closure.md](./0723-x-rest-beta-evidence-closure.md) | P2 | x/rest | Complete REST beta evidence when release refs and sign-off exist |
| [0724-x-websocket-beta-evidence-closure.md](./0724-x-websocket-beta-evidence-closure.md) | P2 | x/websocket | Complete WebSocket beta evidence when release refs and sign-off exist |
| [0725-x-tenant-beta-evidence-closure.md](./0725-x-tenant-beta-evidence-closure.md) | P2 | x/tenant | Complete tenant beta evidence when release refs and sign-off exist |
| [0726-x-observability-beta-evidence-closure.md](./0726-x-observability-beta-evidence-closure.md) | P2 | x/observability | Complete observability beta evidence when release refs and sign-off exist |
| [0727-x-gateway-beta-evidence-closure.md](./0727-x-gateway-beta-evidence-closure.md) | P2 | x/gateway | Complete gateway beta evidence when release refs and sign-off exist |
| [0728-x-ai-stable-tier-beta-evidence-closure.md](./0728-x-ai-stable-tier-beta-evidence-closure.md) | P2 | x/ai | Complete AI stable-tier beta evidence when release refs and sign-off exist |
| [0729-x-data-surface-beta-evidence-closure.md](./0729-x-data-surface-beta-evidence-closure.md) | P2 | x/data | Complete selected data surface beta evidence |
| [0730-x-discovery-surface-beta-evidence-closure.md](./0730-x-discovery-surface-beta-evidence-closure.md) | P2 | x/discovery | Complete discovery core/static beta evidence |
| [0731-x-messaging-service-beta-evidence-closure.md](./0731-x-messaging-service-beta-evidence-closure.md) | P2 | x/messaging | Complete messaging service beta evidence |

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

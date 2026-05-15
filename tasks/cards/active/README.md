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
| [1423](1423-middleware-responsewriter-helper-unification.md) | P1 | middleware | Consolidate duplicate response writer recording helpers while preserving transport-only middleware behavior |
| [1424](1424-router-metadata-path-rule-cleanup.md) | P1 | router | Normalize route metadata/path error handling and cover negative paths without adding HTTP concerns |
| [1425](1425-security-fail-closed-breaking-cleanup.md) | P0 | security | Remove lenient security fallbacks and make invalid auth/input/policy paths fail closed |
| [1426](1426-log-v1-output-contract-cleanup.md) | P1 | log | Normalize logger configuration and caller/output contract before v1 freeze |
| [1427](1427-metrics-health-v1-contract-cleanup.md) | P2 | metrics, health | Keep metrics and health contracts minimal, deleting redundant aliases or helper surfaces |
| [1428](1428-core-v1-minimal-kernel-freeze.md) | P1 | core | Freeze core as a minimal kernel after router, middleware, and contract cleanup land |
| [1367](1367-x-tenant-beta-evidence-closure.md) | P2 | x/tenant | Complete beta evidence closure for `x/tenant` when real release refs and owner sign-off are available |
| [1370](1370-x-ai-stable-tier-beta-evidence-closure.md) | P2 | x/ai | Complete beta evidence closure for `x/ai` stable-tier subpackages |
| [1371](1371-x-data-surface-beta-evidence-closure.md) | P2 | x/data | Complete beta evidence closure for selected `x/data` surfaces |
| [1372](1372-x-discovery-surface-beta-evidence-closure.md) | P2 | x/discovery | Complete beta evidence closure for the `x/discovery` core/static surface |
| [1373](1373-x-messaging-service-beta-evidence-closure.md) | P2 | x/messaging | Complete beta evidence closure for the `x/messaging` app-facing service surface |
| [1375](1375-v1-rc-tag-and-observation-window.md) | P0 | release | Tag `v1.0.0-rc.1`, observe the release candidate, and define the exact path from rc.1 to final `v1.0.0` |

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

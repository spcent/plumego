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
| [1365](1365-x-rest-beta-evidence-closure.md) | P2 | x/rest | Complete beta evidence closure for `x/rest` when real release refs and owner sign-off are available |
| [1366](1366-x-websocket-beta-evidence-closure.md) | P2 | x/websocket | Complete beta evidence closure for `x/websocket` when real release refs and owner sign-off are available |
| [1367](1367-x-tenant-beta-evidence-closure.md) | P2 | x/tenant | Complete beta evidence closure for `x/tenant` when real release refs and owner sign-off are available |
| [1368](1368-x-observability-beta-evidence-closure.md) | P2 | x/observability | Complete beta evidence closure for `x/observability` when real release refs and owner sign-off are available |
| [1369](1369-x-gateway-beta-evidence-closure.md) | P2 | x/gateway | Complete beta evidence closure for `x/gateway` when real release refs and owner sign-off are available |
| [1370](1370-x-ai-stable-tier-beta-evidence-closure.md) | P2 | x/ai | Complete beta evidence closure for `x/ai` stable-tier subpackages |
| [1371](1371-x-data-surface-beta-evidence-closure.md) | P2 | x/data | Complete beta evidence closure for selected `x/data` surfaces |
| [1372](1372-x-discovery-surface-beta-evidence-closure.md) | P2 | x/discovery | Complete beta evidence closure for the `x/discovery` core/static surface |
| [1373](1373-x-messaging-service-beta-evidence-closure.md) | P2 | x/messaging | Complete beta evidence closure for the `x/messaging` app-facing service surface |
| [1374](1374-x-websocket-stable-evidence-readiness.md) | P2 | x/websocket | Close WebSocket maturity evidence only after the runtime stable-readiness cards and release governance evidence are complete |
| [1375](1375-v1-rc-tag-and-observation-window.md) | P0 | release | Tag `v1.0.0-rc.1`, observe the release candidate, and define the exact path from rc.1 to final `v1.0.0` |
| [1376](1376-contract-error-details-deep-clone.md) | P1 | contract | Clarify and harden `APIError.Details` cloning for stable error payloads. |
| [1377](1377-contract-request-id-length-boundary.md) | P1 | contract | Prevent oversized request ids from being echoed into contract JSON responses. |
| [1378](1378-contract-bindjson-cache-semantics.md) | P2 | contract | Freeze the misleading `EnableBodyCache` compatibility semantics. |
| [1379](1379-contract-validatestruct-compat-users.md) | P2 | contract | Finalize the stable decision for current external production `ValidateStruct` users. |
| [1380](1380-contract-release-gate-evidence.md) | P2 | contract | Record final validation evidence for the contract stable hardening pass. |
| [1381](1381-x-websocket-stable-governance-closure.md) | P2 | x/websocket | Close the remaining governance evidence needed before any stable promotion decision |
| [1382](1382-x-websocket-auth-interface-convergence.md) | P0 | x/websocket | Converge auth interfaces before stable by removing JWT/password-specific method names from the server-facing contracts |
| [1383](1383-x-websocket-hub-config-naming-contract.md) | P1 | x/websocket | Rename misleading hub and server config knobs before API freeze |
| [1384](1384-x-websocket-conn-close-api-contract.md) | P1 | x/websocket | Clarify low-level connection and handler close contracts before API freeze |
| [1385](1385-x-websocket-stable-doc-evidence-sync.md) | P2 | x/websocket | Synchronize WebSocket docs and evidence after cards 0773-0784 while keeping governance blockers explicit |
| [1386](1386-x-websocket-release-governance-blocker.md) | P2 | x/websocket | Close WebSocket release-governance evidence only when real release refs and owner sign-off exist |

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

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
| [0715](0715-x-websocket-auth-boundary-convergence.md) | P1 | x/websocket | Split room authorization from token authentication and close query credential gaps |
| [0716](0716-x-websocket-room-identity-validation.md) | P1 | x/websocket | Validate room names and remove URL room-password transport |
| [0717](0717-x-websocket-capacity-and-lifecycle-contract.md) | P1 | x/websocket | Clarify room registration capacity and shutdown semantics |
| [0718](0718-x-websocket-broadcast-stop-write-path.md) | P1 | x/websocket | Harden broadcast versus stop and add socket write deadline coverage |
| [0719](0719-x-websocket-close-and-stream-semantics.md) | P2 | x/websocket | Align close handshake wording and bounded-reader API semantics |
| [0720](0720-x-websocket-metrics-logging-events.md) | P2 | x/websocket | Remove unused state and expose observable security events consistently |
| [0721](0721-x-websocket-security-helper-secret-hygiene.md) | P2 | x/websocket | Avoid secret string copies and normalize secure auth construction |
| [0722](0722-x-websocket-validation-scope-and-log-safety.md) | P2 | x/websocket | Tighten log sanitization and remove transport-level heuristic content scanning |
| [0723](0723-x-websocket-protocol-compliance-coverage.md) | P1 | x/websocket | Add WebSocket protocol negative and boundary coverage |
| [0724](0724-x-websocket-doc-manifest-api-inventory.md) | P1 | x/websocket | Sync manifest, primer, public API inventory, and examples |
| [0725](0725-x-websocket-release-governance-blockers.md) | P3 | x/websocket | Record remaining release evidence blockers without promoting status |
| [0785](0785-x-data-rw-replica-ownership.md) | P2 | x/data/rw | Copy replica slices at cluster API boundaries |
| [0786](0786-x-data-file-metadata-input-validation.md) | P2 | x/data/file | Validate direct metadata manager tenant and path inputs |
| [0787](0787-x-data-idempotency-duplicate-classifier.md) | P2 | x/data/idempotency | Narrow duplicate-key fallback classification |
| [0788](0788-x-data-sharding-postgres-sslmode-config.md) | P2 | x/data/sharding/config | Make PostgreSQL sslmode explicit in config |
| [0789](0789-x-data-stable-readiness-sixth-gate.md) | P3 | x/data | Sync docs/manifests and record sixth readiness gate |

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

# Card 0593

Milestone:
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: docs
Owned Files:
- specs/extension-beta-evidence.yaml
- docs/extension-evidence/x-rest.md
- docs/extension-evidence/x-websocket.md
- docs/extension-evidence/snapshots/first-batch/x-rest-head.snapshot
- docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot
Depends On: 2302

Goal:
Add checked-in current-head API snapshot artifacts for the first two beta
candidates without claiming release-history completion.

Scope:
- Generate current-head exported API snapshots for `x/rest` and `x/websocket`.
- Record artifact paths in the evidence ledger.
- Keep `release_history_missing`, `api_snapshot_missing`, and
  `owner_signoff_missing` blockers until two real release refs exist.

Non-goals:
- Do not promote modules.
- Do not use branch heads as release refs.
- Do not add snapshots for other first-batch candidates in this card.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-rest.md`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/snapshots/first-batch/x-rest-head.snapshot`
- `docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`

Tests:
- `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-rest-head.snapshot docs/extension-evidence/snapshots/first-batch/x-rest-head.snapshot`
- `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/0593-first-batch-rest-websocket-snapshots.md`

Docs Sync:
- Required because beta evidence docs and artifacts change.

Done Definition:
- `x/rest` and `x/websocket` have checked-in current-head snapshot artifacts
  referenced by the ledger, while promotion blockers remain accurate.

Outcome:
- Generated checked-in current-head API snapshots for `x/rest` and
  `x/websocket`.
- Recorded both snapshot artifact paths in `specs/extension-beta-evidence.yaml`.
- Kept release and snapshot blockers because no real release refs are available
  and each candidate still has fewer than two release-backed snapshots.

Validations:
- `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-rest-head.snapshot docs/extension-evidence/snapshots/first-batch/x-rest-head.snapshot`
- `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go run ./internal/checks/extension-beta-evidence`

# Card 0735: x/frontend Release Evidence Blockers

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files:
- `docs/EXTENSION_MATURITY.md`
- `docs/modules/x-frontend/README.md`
- `x/frontend/module.yaml`
- `tasks/cards/active/README.md`
Depends On: 0734

Goal:
Record the remaining non-code blockers for stable promotion without promoting
`x/frontend` prematurely.

Scope:
- Confirm `module.yaml` remains `experimental`.
- Record exact promotion blockers: exported API snapshot, release evidence, and
  owner sign-off.
- Keep active queue accurate after completing the frontend hardening sequence.
- Run manifest and workflow checks.

Non-goals:
- Do not invent owner sign-off.
- Do not promote `x/frontend` to stable.
- Do not change code behavior.

Files:
- `docs/EXTENSION_MATURITY.md`
- `docs/modules/x-frontend/README.md`
- `x/frontend/module.yaml`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-maturity`

Docs Sync:
This card is the final docs and evidence sync.

Done Definition:
- Remaining stable blockers are explicit and concrete.
- No status promotion occurs.
- The listed validation commands pass.

Outcome:
- Kept `x/frontend` status as `experimental`.
- Recorded the exact stable promotion blockers: exported API snapshot for the
  current registrar/config surface, two-release no-churn evidence, and frontend
  owner sign-off.
- Removed the completed frontend hardening card from the active queue.
- Validation passed:
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/agent-workflow`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-maturity`

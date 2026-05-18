# Card 1504

Milestone: M-008
Recipe: specs/change-recipes/update-docs.yaml
Priority: P0
State: active
Primary Module: extension evidence
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `tasks/cards/active/1367-x-tenant-beta-evidence-closure.md`
- `tasks/cards/active/1370-x-ai-stable-tier-beta-evidence-closure.md`
- `tasks/cards/active/1371-x-data-surface-beta-evidence-closure.md`
- `tasks/cards/active/1372-x-discovery-surface-beta-evidence-closure.md`
- `tasks/cards/active/1373-x-messaging-service-beta-evidence-closure.md`

Goal:
- Update specs/extension-beta-evidence.yaml with second_release_ref = v1.1.0 for all five
  candidate surfaces: x/tenant, x/ai stable-tier, x/data file and idempotency, x/gateway/discovery,
  and x/messaging app-facing service.
- Move the five blocked cards (1367, 1370, 1371, 1372, 1373) from tasks/cards/blocked/ to
  tasks/cards/active/.

Scope:
- Confirm `git tag --list v1.1.0` outputs v1.1.0 before editing any file.
- Set second_release_ref to the string "v1.1.0" (not a commit SHA) in the evidence ledger
  for each of the five surfaces.
- Move each blocked card file from tasks/cards/blocked/ to tasks/cards/active/ and update
  the State field from blocked to active.
- Update tasks/cards/blocked/README.md to remove the five unblocked cards from its queue table.
- Update tasks/cards/active/README.md to add the five unblocked cards to the Active Queue table.

Non-goals:
- Do not promote any extension to beta in this card; that happens in M-009 and M-010.
- Do not change any module.yaml status fields.
- Do not add second_release_ref for surfaces not in the five candidate list.

Files:
- `specs/extension-beta-evidence.yaml`
- `tasks/cards/blocked/1367-x-tenant-beta-evidence-closure.md`
- `tasks/cards/blocked/1370-x-ai-stable-tier-beta-evidence-closure.md`
- `tasks/cards/blocked/1371-x-data-surface-beta-evidence-closure.md`
- `tasks/cards/blocked/1372-x-discovery-surface-beta-evidence-closure.md`
- `tasks/cards/blocked/1373-x-messaging-service-beta-evidence-closure.md`
- `tasks/cards/active/README.md`
- `tasks/cards/blocked/README.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- tasks/cards/active/README.md and tasks/cards/blocked/README.md updated to reflect card moves.

Done Definition:
- specs/extension-beta-evidence.yaml shows second_release_ref = v1.1.0 for all five surfaces.
- Cards 1367, 1370, 1371, 1372, 1373 exist in tasks/cards/active/ with State = active.
- `go run ./internal/checks/extension-beta-evidence` exits 0.
- tasks/cards/blocked/ no longer contains the five moved cards.

Outcome:
-

# Plan for M-010: Beta Promotions Round 2

Milestone: `M-010`
Objective: Promote x/messaging app-facing service surface to beta and evaluate
x/frontend for beta status, closing the final blocked evidence cards from the
post-v1 queue. If x/frontend has outstanding API instability, record an explicit
blocker rather than leaving the result ambiguous.
Constraints: same two-release evidence rule applies (v1.0.0 + v1.1.0), no
runtime behavior changes, x/frontend result must be explicit (promote or block),
no stable-root public API additions.
Affected Modules: extension evidence, x/messaging, x/frontend.

## Phase Map

- Phase 1: Orient — confirm M-009 is merged and prior round promotions are visible
  in docs/EXTENSION_MATURITY.md.
- Phase 2: Parallel Evaluations — promote x/messaging and evaluate x/frontend concurrently.
- Phase 3: Dashboard and Validate — update EXTENSION_MATURITY.md, DEPRECATION.md, run
  acceptance criteria, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1373 | Promote x/messaging app-facing service to beta | x/messaging | `specs/extension-beta-evidence.yaml`, `docs/extension-evidence/x-messaging.md`, `x/messaging/module.yaml` | M-009 | `extension-beta-evidence`, `extension-maturity` |
| 1374 | Evaluate x/frontend for beta; promote or record explicit blocker | x/frontend | `specs/extension-beta-evidence.yaml`, `docs/extension-evidence/x-frontend.md`, `x/frontend/module.yaml` | M-009 | `extension-beta-evidence`, `extension-maturity` |

## Dependency Edges

- Cards 1373 and 1374 both depend on M-009 being merged.
- Cards 1373 and 1374 are independent of each other.

## Parallel Groups

- Group A (parallel): cards 1373 and 1374 — different modules, no file overlap.
- Group B (sequential after A): dashboard update (EXTENSION_MATURITY.md, DEPRECATION.md).

## Risk Register

- Risk: x/frontend API snapshot shows instability between v1.0.0 and v1.1.0.
  Mitigation: record an explicit blocker in specs/extension-beta-evidence.yaml and
  docs/extension-evidence/x-frontend.md; set status to experimental; do not promote.
- Risk: messaging owner sign-off is not on record from M-007/M-009 intake.
  Mitigation: card 1373 requires an explicit owner_signoff field from the messaging
  team before marking done.

## Verification Strategy

- Card-level checks: run `go run ./internal/checks/extension-beta-evidence` and
  `go run ./internal/checks/extension-maturity` after each card.
- Milestone-level checks: run full Acceptance Criteria suite before committing.
- Explicit result check: x/frontend must appear in evidence ledger as either
  status = beta or with a documented blocker entry; ambiguity is a gate failure.

## Exit Condition

- card 1373 completed; x/messaging module.yaml status = beta
- card 1374 completed; x/frontend is either status = beta or has an explicit documented blocker
- docs/EXTENSION_MATURITY.md updated for both evaluated surfaces
- verify report shows pass
- milestone acceptance criteria ready for PR packaging

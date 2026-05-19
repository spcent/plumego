# Plan for M-002: Stable Roots Freeze and Reliability

Milestone: `M-002`
Objective: Preserve the legacy stable-root freeze execution record in the new
self-contained milestone directory shape.
Constraints: This milestone was completed before plan artifacts were required;
do not reinterpret or widen the historical scope.
Affected Modules: stable roots, stable API evidence, reliability tests.

## Phase Map

- Phase 1: Stable-root API and reliability inventory.
- Phase 2: Freeze cleanup and focused validation.
- Phase 3: Final verification and archive.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| legacy | Preserve completed milestone evidence | stable roots | `M-002.md`, `verify-M-002.md` | M-001 | recorded in verify |

## Dependency Edges

- (none)

## Parallel Groups

- None recorded for the legacy plan.

## Risk Register

- Risk: reconstructing card-level detail could imply execution that did not
  happen under the newer plan protocol.
  Mitigation: keep this plan as an archival index and treat `verify-M-002.md`
  as the source of validation evidence.

## Verification Strategy

- Card-level checks: legacy evidence only.
- Milestone-level checks: use `verify-M-002.md` and the `## Outcome` section in
  `M-002.md`.

## Exit Condition

- milestone spec, plan, and verify artifacts live in one directory
- historical outcome remains unchanged
- verify report remains the evidence source

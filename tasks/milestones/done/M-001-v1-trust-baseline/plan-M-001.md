# Plan for M-001: v1 Trust Baseline

Milestone: `M-001`
Objective: Preserve the legacy v1 trust baseline execution record in the new
self-contained milestone directory shape.
Constraints: This milestone was completed before plan artifacts were required;
do not reinterpret or widen the historical scope.
Affected Modules: `cmd/plumego`, docs, release evidence.

## Phase Map

- Phase 1: Legacy discovery and trust matrix.
- Phase 2: CLI template and documentation alignment.
- Phase 3: Final verification and archive.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| legacy | Preserve completed milestone evidence | `cmd/plumego` | `M-001.md`, `verify-M-001.md` | none | recorded in verify |

## Dependency Edges

- (none)

## Parallel Groups

- None recorded for the legacy plan.

## Risk Register

- Risk: reconstructing card-level detail could imply execution that did not
  happen under the newer plan protocol.
  Mitigation: keep this plan as an archival index and treat `verify-M-001.md`
  as the source of validation evidence.

## Verification Strategy

- Card-level checks: legacy evidence only.
- Milestone-level checks: use `verify-M-001.md` and the `## Outcome` section in
  `M-001.md`.

## Exit Condition

- milestone spec, plan, and verify artifacts live in one directory
- historical outcome remains unchanged
- verify report remains the evidence source

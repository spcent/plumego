# Plan for M-003: Extension Evidence Pipeline

Milestone: `M-003`
Objective: Preserve the legacy extension evidence pipeline execution record in
the new self-contained milestone directory shape.
Constraints: This milestone was completed before plan artifacts were required;
do not reinterpret or widen the historical scope.
Affected Modules: extension evidence, extension maturity, tasks.

## Phase Map

- Phase 1: Extension evidence model and blocker inventory.
- Phase 2: Evidence checks, snapshots, and documentation.
- Phase 3: Final verification and archive.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| legacy | Preserve completed milestone evidence | extension evidence | `M-003.md`, `verify-M-003.md` | M-002 | recorded in verify |

## Dependency Edges

- (none)

## Parallel Groups

- None recorded for the legacy plan.

## Risk Register

- Risk: reconstructing card-level detail could imply execution that did not
  happen under the newer plan protocol.
  Mitigation: keep this plan as an archival index and treat `verify-M-003.md`
  as the source of validation evidence.

## Verification Strategy

- Card-level checks: legacy evidence only.
- Milestone-level checks: use `verify-M-003.md` and the `## Outcome` section in
  `M-003.md`.

## Exit Condition

- milestone spec, plan, and verify artifacts live in one directory
- historical outcome remains unchanged
- verify report remains the evidence source

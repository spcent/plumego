# Plan for M-XXX: <Title>

Milestone: `M-XXX`
Objective:
Constraints:
Affected Modules:

## Phase Map

- Phase 1:
- Phase 2:
- Phase 3:

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 0001 | | | | | |

## Dependency Edges

- `0001 -> 0002`

## Parallel Groups

- Group A:
- Group B:

## Risk Register

- Risk:
  Mitigation:

## Verification Strategy

- Card-level checks:
- Milestone-level checks:

## Checkpoints

<!-- One checkpoint per phase. Agent writes checkpoint file after each phase passes.
     Checkpoint file: tasks/milestones/M-XXX.checkpoint.json
     Format per phase entry: {"phase": "phase_1", "passed": true, "timestamp": "2026-01-01T00:00:00Z", "gate": "<command run>"}
     Use `make milestone-status M=active/M-XXX` to read checkpoint state. -->

| Phase | Checkpoint Gate | Status |
|-------|-----------------|--------|
| Phase 1 | <!-- e.g. go test ./core/... --> | pending |
| Phase 2 | <!-- e.g. go run ./internal/checks/dependency-rules --> | pending |
| Phase 3 | <!-- e.g. make gates --> | pending |

## Exit Condition

- all planned cards completed or explicitly superseded
- all phase checkpoints recorded as passed
- verify report shows pass
- milestone acceptance criteria ready for PR packaging

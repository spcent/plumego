# Agent Workflow — Plumego

This document defines the default working loop for agents in `plumego`.

Use it with:

- `AGENTS.md` for hard rules and validation order
- `docs/CANONICAL_STYLE_GUIDE.md` for code shape
- `docs/AGENT_CODE_QUALITY_RULES.md` for preflight, review, and gate selection
- `docs/AGENT_CONTEXT_BUDGET.md` for bounded reads and resume discipline
- `specs/agent-quality-rules.yaml` for the machine-readable quality contract
- `specs/*` and `<module>/module.yaml` for routing, ownership, and validation

Repo-native change recipes live under `specs/change-recipes/`.

## 1. Working Modes

### Analysis

Use when scope, ownership, or architecture is unclear.

Return:

- owning module or `x/*` family
- in-scope paths
- out-of-scope paths
- likely touched files
- main risks
- validation plan

Do not edit code in this mode.

### Implementation

Use when the task contract is clear.

Expected behavior:

- confirm the owning module first
- make the smallest coherent change
- keep the diff inside one primary module when possible
- add or update focused tests
- run validation before handoff

### Review

Use when the user asks for review, audit, or risk assessment.

Expected output:

- findings first
- severity-ordered issues
- file references
- missing tests and regression risks

Do not patch unless the user explicitly asks for fixes.

## 2. Default Task Contract

When the task contract is incomplete, assume:

- one primary module per change
- no stable public API changes
- no new dependencies
- focused tests for behavior changes
- docs sync only for implemented behavior changes

Every task should still be framed in terms of:

- Goal
- In Scope
- Out of Scope
- Target Module
- API and Dependency Policy
- Behavior, Security, and Docs Impact
- Tests
- Validation
- Done Definition

## 3. Default Loop

1. Read `AGENTS.md`.
2. Select the matching `specs/task-routing.yaml` task entry.
3. Read that entry's `start_with` files.
4. Identify the owning module or family.
5. Read the target `<module>/module.yaml` before editing module behavior.
6. State the context package, scope, and impact assumptions.
7. Implement the smallest coherent change.
8. Add or update focused tests.
9. Run module validation first.
10. Run boundary and repo-wide checks only when the gate profile requires them.
11. Report validation as a compact command and status summary with residual risk.

Use targeted `rg` searches when a symbol, package, or rule needs confirmation.
Avoid broad file dumps after ownership and validation are already clear.

## 4. Context Budget

Use the smallest context package that can safely complete the task.

- `startup`: initial orientation
- `implementation`: normal code or docs changes
- `review`: findings-first review
- `control-plane`: workflow, architecture, quality, or spec changes

See `docs/AGENT_CONTEXT_BUDGET.md` for package contents, split thresholds, and
resume discipline.

Split work before implementation when the expected scope spans more than one
primary module, more than five files, more than three validation commands, or
unclear API, dependency, security, or boundary impact.

## 5. Stop Conditions

Stop and surface the issue before coding when:

- the owning module is unclear
- the task would force a stable root to import `x/*`
- the task needs a stable public API change that was not requested
- the task needs a new dependency that was not approved
- the task is broad but lacks acceptance criteria
- a repo spec, module manifest, and local pattern conflict in a behavior-changing way

Use `specs/stop-condition-handlers.yaml` for the deterministic resolution path.

## 6. Recipes

Before writing a bespoke workflow, check whether one of these recipes already
matches the task:

- `specs/change-recipes/analysis-only.yaml`
- `specs/change-recipes/fix-bug.yaml`
- `specs/change-recipes/http-endpoint-bugfix.yaml`
- `specs/change-recipes/review-only.yaml`
- `specs/change-recipes/add-http-endpoint.yaml`
- `specs/change-recipes/add-middleware.yaml`
- `specs/change-recipes/add-acceptance-tests.yaml`
- `specs/change-recipes/new-stable-module.yaml`
- `specs/change-recipes/new-extension-module.yaml`
- `specs/change-recipes/stable-root-boundary-review.yaml`
- `specs/change-recipes/symbol-change.yaml`
- `specs/change-recipes/tenant-policy-change.yaml`
- `specs/change-recipes/add-websocket-room.yaml`
- `specs/change-recipes/add-ai-tool.yaml`
- `specs/change-recipes/add-grpc-method.yaml`

## 7. Prompt Shapes

### Analysis prompt

```text
Do not edit code.

Read the relevant control-plane files and determine the best landing zone for:
[task]

Return:
- context package
- owning module or x/* family
- in-scope paths
- out-of-scope paths
- likely touched files
- risks
- validation plan
- smallest reversible task split
```

### Implementation prompt

```text
Implement this change in plumego:
[task]

Constraints:
- In scope: [paths]
- Out of scope: [paths]
- Target module: [module]
- No new dependencies
- Do not change stable public APIs unless explicitly requested
- Follow AGENTS.md and the canonical style guide

Requirements:
- complete the quality preflight before editing
- use the smallest matching context package
- state which files you will touch before broad edits
- make the smallest coherent change
- add or update focused tests
- run the required validation commands
- report residual risks at the end
```

### Review prompt

```text
Review this change only. Do not modify code.

Priorities:
1. boundary violations
2. stable-root to x/* drift
3. net/http compatibility risks
4. hidden globals or context service-location
5. fail-open behavior
6. response or error path drift
7. missing tests
8. docs/config/example drift

Output findings first, ordered by severity, with file references.
Use the review output contract in docs/AGENT_CODE_QUALITY_RULES.md.
```

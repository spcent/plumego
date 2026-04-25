# Codex Workflow — Plumego

This document defines the recommended way to use Codex in `plumego`.

It is a companion to:

- `AGENTS.md` for hard rules and validation order
- `docs/CANONICAL_STYLE_GUIDE.md` for code shape
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md` for repository layout
- `specs/*` and `<module>/module.yaml` for machine-readable ownership and boundaries

Use this document when you want a repeatable prompting and execution pattern,
not just a one-off instruction.

Repo-native recipe assets live under `specs/change-recipes/` and should be used
when a task matches one of the standard shapes.

## 1. Default Operating Model

Codex should work in one of three modes:

### Analysis Mode

Use when scope, ownership, or architecture is unclear.

Expected output:

- owning module or `x/*` family
- in-scope paths
- out-of-scope paths
- likely touched files
- main risks
- validation plan

Do not edit code in this mode.

### Implementation Mode

Use when the task contract is already clear.

Expected behavior:

- confirm the owning module first
- make the smallest coherent change
- keep the diff inside one primary module when possible
- add or update focused tests
- run validation before handoff

### Review Mode

Use when the user asks for review, audit, or risk assessment.

Expected output:

- findings first
- severity-ordered issues
- file-level references
- missing tests and regression risks

Do not patch code unless the user explicitly asks for fixes.

## 2. Task Contract

For stable and predictable runs, every task should define:

- Goal
- In Scope
- Out of Scope
- Target Module
- API and Dependency Policy
- Tests
- Validation
- Done Definition

When humans omit these details, Codex should assume:

- one primary module per change
- no stable public API changes
- no new dependencies
- focused tests are required for behavior changes
- docs sync is required only for implemented behavior changes

## 3. Stop Conditions

Codex should stop and surface the issue before coding when:

- the owning module is unclear
- the task would force a stable root to import `x/*`
- the task needs a stable public API change that was not requested
- the task needs a new dependency that was not approved
- the task is broad but lacks acceptance criteria or validation commands
- a repo spec, module manifest, and local pattern conflict in a behavior-changing way

## 4. Daily Workflow

Use this loop for normal feature work, bug fixes, and refactors:

1. Read the control plane in canonical order.
2. Identify the owning module or family.
3. State in-scope and out-of-scope paths.
4. State public API and dependency assumptions.
5. Implement the smallest coherent change.
6. Add or update focused tests.
7. Run module validation first.
8. Run boundary and repo-wide checks second.
9. Report residual risks and doc-sync impacts.

Default read order:

1. `AGENTS.md`
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
4. `specs/repo.yaml`
5. `specs/task-routing.yaml`
6. `specs/extension-taxonomy.yaml`
7. `specs/package-hotspots.yaml`
8. `specs/dependency-rules.yaml`
9. target `<module>/module.yaml`
10. `reference/standard-service`

## 5. Prompt Templates

Before writing a bespoke prompt, check whether one of these repo-native recipes
already matches the task:

- `specs/change-recipes/analysis-only.yaml`
- `specs/change-recipes/fix-bug.yaml`
- `specs/change-recipes/http-endpoint-bugfix.yaml`
- `specs/change-recipes/review-only.yaml`
- `specs/change-recipes/add-http-endpoint.yaml`
- `specs/change-recipes/add-middleware.yaml`
- `specs/change-recipes/new-stable-module.yaml`
- `specs/change-recipes/new-extension-module.yaml`
- `specs/change-recipes/stable-root-boundary-review.yaml`
- `specs/change-recipes/symbol-change.yaml`
- `specs/change-recipes/tenant-policy-change.yaml`

### Analysis Prompt

```text
Do not edit code.

Read the relevant control-plane files and determine the best landing zone for:
[task]

Return:
- owning module or x/* family
- in-scope paths
- out-of-scope paths
- likely touched files
- risks
- validation plan
- smallest reversible task split
```

### Implementation Prompt

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
- state which files you will touch before broad edits
- make the smallest coherent change
- add or update focused tests
- run the required validation commands
- report residual risks at the end
```

### Review Prompt

```text
Review this change only. Do not modify code.

Priorities:
1. boundary violations
2. stable-root to x/* drift
3. net/http compatibility risks
4. hidden globals or context service-location
5. fail-open behavior
6. missing tests

Output findings first, ordered by severity, with file references.
```

### Exported Symbol Change Prompt

```text
Change this exported symbol:
[symbol and requested change]

You must:
- enumerate every caller first with rg
- update every caller in the same change
- re-run the same search and verify no stale references remain
- update tests in the same change
- run go build ./... and go test ./...
```

### Milestone Prompt

```text
Execute this milestone:
[tasks/milestones/active/M-NNN.md]

Follow AGENTS.md milestone rules exactly:
- read every Context file first
- follow Tasks in order
- stay inside Affected Modules
- stop and record blockers in the spec if discovered
- run the full validation sequence before push
```

## 6. Symbol Change Protocol

When removing, renaming, or changing the behavior of an exported symbol:

1. enumerate callers first with `rg -n --glob '*.go' 'SymbolName' .`
2. decide how every site will be handled
3. edit all callers in the same change
4. re-run the same search
5. update tests in the same change
6. finish only after build, tests, and residual-reference checks pass

Do not leave dead wrappers, deprecated compatibility layers, or silent discard
sites behind.

## 7. Validation

Default validation order:

1. run the target module tests from `<module>/module.yaml`
2. run boundary and manifest checks
3. run repo-wide gates when the change is code-bearing, cross-module, or release relevant

Required full gate:

```bash
make gates
```

`make gates` mirrors `.github/workflows/quality-gates.yml`: boundary checks,
`go vet ./...`, a non-mutating `gofmt -l .` check, `go test -race -timeout 60s
./...`, and `go test -timeout 20s ./...`. Run `gofmt -w <paths>` before the
gate when formatting is required.

## 8. Milestone Workflow

Use the milestone path for multi-step, single-PR scopes with explicit human
authorship and autonomous Codex execution.

Companion workflow assets:

- `docs/MILESTONE_PIPELINE.md`
- `docs/github-workflows/milestone-pr-template.md`
- `tasks/milestones/ROADMAP.md`

High-level loop:

1. scaffold the spec with `make new-milestone`
2. fill Goal, Architecture Decisions, Context, Tasks, Acceptance Criteria, and Out of Scope
3. validate with `make check-spec`
4. launch with `make milestone M=active/M-NNN`
5. let Codex execute autonomously on the milestone branch
6. review the PR as the only manual checkpoint
7. archive the spec and update the roadmap

Use milestones when:

- the work spans multiple stable intermediate steps
- there are non-obvious architecture decisions to lock down
- the task needs explicit sequencing or parallel phases
- the reviewer wants one PR with a spec-backed contract

Do not use milestones for every small fix. Daily implementation prompts are the
default.

## 9. Worked Examples

These examples show how real Plumego work maps onto the control plane.

### Example A: Exported Symbol Cleanup

Task shape:

- rename or remove an exported `contract` or `core` symbol
- migrate every caller in one change
- update tests and verify zero stale references

Use:

- recipe: `specs/change-recipes/symbol-change.yaml`
- routing entry: `symbol_change`

Representative cards:

- `tasks/cards/done/0907-rename-errtype-constants-to-type.md`
- `tasks/cards/done/0918-contract-ctx-getter-rename.md`

### Example B: HTTP Handler or Route Bug

Task shape:

- broken request decode
- inconsistent error response
- route registration or transport regression

Use:

- recipe: `specs/change-recipes/http-endpoint-bugfix.yaml`
- routing entry: `http_endpoint_bugfix`

Representative cards:

- `tasks/cards/done/0743-writeerror-buffer-before-headers.md`
- `tasks/cards/done/0902-redirect-safe-by-default.md`

### Example C: Tenant Policy or Quota Change

Task shape:

- tenant resolution behavior
- quota feedback headers
- deny-path, isolation, or tenant-session changes

Use:

- recipe: `specs/change-recipes/tenant-policy-change.yaml`
- routing entry: `tenant_policy_change`

Representative cards:

- `tasks/cards/done/0796-x-tenant-quota-retry-after-coverage.md`
- `tasks/cards/done/0797-x-tenant-policy-and-isolation-coverage.md`
- `tasks/cards/done/0914-tenant-core-ratelimit-provider-unknown-tenant.md`

### Example D: Stable Root Boundary Audit

Task shape:

- review-only request
- stable-root ownership drift
- hidden coupling or x/* leakage concerns

Use:

- recipe: `specs/change-recipes/stable-root-boundary-review.yaml`
- routing entry: `stable_root_boundary_review`

Representative cards:

- `tasks/cards/done/0303-block-tenant-leakage-into-stable-roots.md`
- `tasks/cards/done/0848-security-jwt-session-lifecycle-pruning.md`
- `tasks/cards/done/0850-security-resilience-boundary-pruning.md`

## 10. Review Checklist for Humans

When reviewing Codex output, check:

1. the owning module was correct
2. the diff stayed inside scope
3. stable roots did not learn extension-family internals
4. control flow stayed explicit and `net/http` compatible
5. tests cover failure paths and regressions
6. docs changed only where implemented behavior changed
7. `go.mod` did not change unless explicitly approved

## 11. Anti-Patterns

Do not ask Codex to:

- "optimize this" without a target module or success bar
- "refactor broadly" across multiple primary modules in one pass
- "just make it cleaner" without boundaries
- preserve deprecated wrappers indefinitely
- infer whether public APIs or dependencies may change

Do not let Codex:

- widen stable roots into feature catalogs
- hide dependency flow in context or globals
- add one-off response helpers or route-registration idioms
- treat subordinate `x/*` packages as competing family entrypoints

## 12. Recommended Usage

For best stability, use this sequence:

1. analysis prompt
2. implementation prompt
3. review prompt

For larger work:

1. analysis prompt
2. split into cards or a milestone
3. execute one card or milestone at a time
4. review before merge

# Agent Context Budget

This document defines the token-bounded working model for Codex and other
coding agents in Plumego. The goal is to keep each run focused, reversible, and
cheap in context without weakening boundary, security, or validation rules.

Use this document with `AGENTS.md`, `docs/CODEX_WORKFLOW.md`,
`docs/AGENT_CODE_QUALITY_RULES.md`, and `specs/task-routing.yaml`.

## 1. Objective

Default agent runs should load the smallest context package that can answer the
task safely. Do not scan the entire control plane for routine work.

Target budget:

- Startup package: `AGENTS.md` plus the matching `specs/task-routing.yaml`
  entry.
- Implementation package: startup package plus the owning `module.yaml`, one
  task recipe when applicable, and only the directly relevant module primer or
  style section.
- Full control-plane package: only for architecture, boundary, release, or
  workflow-rule changes.

The full control plane remains authoritative, but it is not the default read
set.

## 2. Loading Order

Use this order for normal work:

1. Read `AGENTS.md`.
2. Read `specs/task-routing.yaml` only far enough to select the task entry.
3. Read the selected task entry's `start_with` files.
4. Read the target `<module>/module.yaml` before editing module behavior.
5. Read additional docs or specs only when the preflight identifies a concrete
   reason.

Stop reading when the task contract, owner, boundaries, and validation plan are
clear. Prefer `rg`-targeted searches over broad file dumps.

## 3. Context Packages

### `startup`

Use for first-pass orientation:

- `AGENTS.md`
- matching `specs/task-routing.yaml` task entry

### `implementation`

Use for normal code or doc changes:

- `startup`
- target `<module>/module.yaml`, when module behavior changes
- one matching `specs/change-recipes/*.yaml`, when a recipe exists
- one module primer or style guide section, only if needed for the edit shape
- touched files and focused tests

### `review`

Use when the user asks for review:

- `startup`
- `docs/AGENT_CODE_QUALITY_RULES.md` review output contract
- relevant module manifests or specs for the diff area
- changed files or PR diff

### `control-plane`

Use for changes to workflow, architecture rules, specs, quality gates, or
boundary definitions:

- `startup`
- `docs/CODEX_WORKFLOW.md`
- `docs/AGENT_CODE_QUALITY_RULES.md`
- directly changed specs
- only the architecture docs that own the changed rule

## 4. Task Card Limits

Split broad work before implementation when the expected edit set exceeds one
of these limits:

- more than one primary module
- more than five files
- more than three validation commands
- unclear public API, dependency, security, or boundary impact
- work that cannot be reverted in one commit

Each card should state:

- Goal
- Scope
- Non-goals
- Files
- Tests
- Docs Sync
- Done Definition

## 5. Preflight Ledger

Every implementation preflight should include the context package:

```text
Context package:
Owning module:
Target module.yaml read:
In-scope paths:
Out-of-scope paths:
Public API impact: none / yes
Dependency impact: none / yes
Behavior impact: none / yes
Security impact: none / yes
Docs impact: none / yes
Validation plan:
```

Use `Context package: control-plane` for edits to this document or other
workflow/spec authority.

## 6. Output Compression

Agent handoffs should summarize validation instead of pasting full logs:

```text
Validation:
- command: go test ./router/...
  status: pass
- command: go run ./internal/checks/agent-workflow
  status: fail
  key failure: specs/task-routing.yaml task x references missing start_with path
  next step: add the missing file or correct the route entry
```

For failures, include the command, status, first meaningful error, and next
step. Keep full logs in the terminal output, not the conversational context.

## 7. Resume Discipline

After context compaction or a long-running task, resume from the ledger rather
than rereading the full repository:

1. Re-read the current task card or preflight ledger.
2. Check `git status --short`.
3. Read only changed files and the directly relevant `start_with` entries.
4. Continue validation from the last incomplete step.

If the ledger is missing or stale, return to `startup` and rebuild the smallest
safe package.

# Agent Context Budget

This document defines the default bounded-read model for agents working in
Plumego. The goal is to keep each run focused, reversible, and cheap in
context without weakening boundary, security, or validation rules.

Use this document with `AGENTS.md`, `docs/CODEX_WORKFLOW.md`,
`docs/AGENT_CODE_QUALITY_RULES.md`, and `specs/task-routing.yaml`.

## 1. Default

Start with the smallest safe context package.

Normal read order:

1. Read `AGENTS.md`.
2. Read only enough of `specs/task-routing.yaml` to select the task entry.
3. Read that entry's `start_with` files.
4. Read the target `<module>/module.yaml` before editing module behavior.
5. Read extra docs or specs only when preflight identifies a concrete need.

Stop reading when ownership, boundaries, touched files, and validation are
clear. Prefer targeted `rg` searches over broad file dumps.

## 2. Context Packages

### `startup`

Use for first-pass orientation:

- `AGENTS.md`
- matching `specs/task-routing.yaml` task entry

### `implementation`

Use for normal code or docs changes:

- `startup`
- target `<module>/module.yaml`, when module behavior changes
- one matching `specs/change-recipes/*.yaml`, when a recipe exists
- one directly relevant module primer or style section, only if needed
- touched files and focused tests

### `review`

Use when the user asks for review:

- `startup`
- `docs/AGENT_CODE_QUALITY_RULES.md` review output contract
- relevant module manifests or specs for the diff area
- changed files or PR diff

### `control-plane`

Use for workflow, architecture, quality, boundary, or spec changes:

- `startup`
- `docs/CODEX_WORKFLOW.md`
- `docs/AGENT_CODE_QUALITY_RULES.md`
- directly changed specs or docs
- only the architecture docs that own the changed rule

## 3. Split Thresholds

Split work before implementation when any of these is true:

- More than one primary module
- More than five files
- More than three validation commands
- Unclear public API, dependency, security, or boundary impact
- Work cannot be cleanly reverted in one commit

Each task card should state:

- Goal
- Scope
- Non-goals
- Files
- Tests
- Docs Sync
- Done Definition

## 4. Preflight Ledger

Every implementation preflight should include:

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

Use `Context package: control-plane` for workflow, architecture, and spec edits.

## 5. Validation Summaries

Summarize validation instead of pasting full logs:

```text
Validation:
- command: go test ./router/...
  status: pass
- command: go run ./internal/checks/agent-workflow
  status: fail
  key failure: specs/task-routing.yaml entry points at a missing start_with path
  next step: fix the path or update the routing entry
```

For failures, include:

- command
- status
- first meaningful failure
- next step

Keep full logs in the terminal, not in the conversational context.

## 6. Resume Discipline

After compaction or a long-running task:

1. Re-read the current task card or preflight ledger.
2. Check `git status --short`.
3. Read only changed files and the directly relevant `start_with` entries.
4. Continue from the last incomplete validation step.

If the ledger is missing or stale, return to `startup` and rebuild the smallest
safe package.

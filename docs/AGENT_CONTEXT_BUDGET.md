# Agent Context Budget

This document defines the default bounded-read model for agents working in
Plumego. The goal is to keep each run focused, reversible, and cheap in
context without weakening boundary, security, or validation rules.

Use this document with `AGENTS.md`, `docs/CODEX_WORKFLOW.md`,
`docs/AGENT_CODE_QUALITY_RULES.md`, and `specs/task-routing.yaml`.

## 1. Default

Start with the smallest safe context package. The default read order is in `AGENTS.md §1`. Stop reading when ownership, boundaries, touched files, and validation are clear. Prefer targeted `rg` searches over broad file dumps.

## 2. Context Packages

| Package | When | What to read |
|---|---|---|
| `startup` | Orientation | `AGENTS.md` + task routing entry |
| `implementation` | Code/docs changes | startup + `<module>/module.yaml` + matching `specs/change-recipes/*.yaml` + touched files |
| `review` | Review/audit | startup + `AGENT_CODE_QUALITY_RULES.md` + changed files/PR diff |
| `control-plane` | Workflow/arch/spec changes | startup + `CODEX_WORKFLOW.md` + `AGENT_CODE_QUALITY_RULES.md` + changed specs/docs |

## 3. Split Thresholds

Split before implementation when: multiple primary modules, >5 files, >3 validation commands, unclear API/dependency/security/boundary impact, or non-atomic revert.

Task cards must state: Goal, Scope, Non-goals, Files, Tests, Docs Sync, Done Definition.

## 4. Preflight Ledger

The preflight template is in `AGENTS.md §4`. Use `Context package: control-plane` for workflow, architecture, and spec edits.

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

Keep full logs in the terminal, not in the conversational context.

## 6. Resume Discipline

After compaction: re-read the preflight ledger, `git status --short`, read changed files and relevant `start_with` entries, continue from the last incomplete validation step. If ledger is missing, return to `startup`.

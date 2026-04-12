# Card 0966: Control Plane Canonical HTTP Doc Resync

Priority: P1
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: control-plane docs
Depends On: 0963

## Goal

Resynchronize the repository control plane with Plumego's canonical HTTP
transport direction so agents, humans, and scaffolds stop receiving conflicting
guidance about response helpers and generated app layout.

## Problem

- `docs/CANONICAL_STYLE_GUIDE.md` still presents
  `internal/platform/httpjson/response.go` and
  `internal/platform/httperr/error.go` as part of the canonical app structure.
- That guidance now conflicts with `AGENTS.md` and the current contract-first
  direction where HTTP success and error writes should converge on
  `contract.WriteResponse` / `contract.WriteError` instead of generating a
  second transport helper family.
- Leaving the control plane split means future agent work can satisfy one source
  of truth while violating another, especially around scaffold output and app
  layout review.

## Scope

- Update canonical style and workflow docs to describe the contract-first HTTP
  response path.
- Remove or rewrite control-plane references that still teach
  `internal/platform/httpjson` / `internal/platform/httperr` as canonical
  default structure.
- Keep the docs aligned with the actual `reference/standard-service` and
  post-`0963` scaffold output.

## Non-Goals

- Do not change runtime code in this card.
- Do not redesign the broader reference app layout beyond the response/error
  helper guidance.
- Do not reopen unrelated doc sync already covered by done cards.

## Files

- `docs/CANONICAL_STYLE_GUIDE.md`
- `docs/CODEX_WORKFLOW.md` only if it still points to the old helper layout
- `README.md` / `README_CN.md` / `CLAUDE.md` only if they still teach the old
  HTTP helper packages

## Tests

```bash
rg -n 'internal/platform/httpjson|internal/platform/httperr|httpjson|httperr' docs README.md README_CN.md AGENTS.md CLAUDE.md -g '*.*'
go run ./internal/checks/agent-workflow
go run ./internal/checks/reference-layout
```

## Docs Sync

- this card is the docs sync

## Done Definition

- The control plane no longer teaches `internal/platform/httpjson` or
  `internal/platform/httperr` as canonical default HTTP helper packages.
- Style-guide and workflow docs align with `AGENTS.md`, `contract`, and the
  reference/scaffolded HTTP response path.
- Repo grep for the removed canonical-structure references is empty outside
  intentional historical notes.

## Outcome

- Removed the stale `internal/platform/httpjson` and
  `internal/platform/httperr` entries from the canonical application-structure
  example in `docs/CANONICAL_STYLE_GUIDE.md`.
- Tightened the style-guide guidance so generated and hand-written handlers are
  explicitly expected to write success and error responses through `contract`
  rather than app-local JSON/error helper families.
- Verified that the control plane no longer contains the stale helper-package
  guidance outside generator tests that intentionally assert their absence.

## Validation Run

```bash
rg -n 'internal/platform/httpjson|internal/platform/httperr|httpjson|httperr' docs README.md README_CN.md AGENTS.md CLAUDE.md -g '*.*'
go run ./internal/checks/agent-workflow
go run ./internal/checks/reference-layout
```

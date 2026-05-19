# Card 1497

Milestone: M-020
Recipe: specs/change-recipes/update-docs.yaml
Priority: P1
State: active
Primary Module: docs
Owned Files:
- `README.md`
- `README_CN.md`
Depends On: M-008

## Goal

Rewrite `README.md` with a 5-minute quick-start, a stdlib ServeMux vs. plumego feature comparison table, and a package overview table; mirror the structure in `README_CN.md`.

## Scope

- `README.md` — full rewrite preserving existing badge line at the top.
- `README_CN.md` — structural mirror of README.md with accurate Chinese translation.
- No Go source file changes.

## Non-goals

- Do not add doc comments to any `.go` files (that is Card 1498).
- Do not add benchmark results (that is M-011).
- Do not translate `x/*` package descriptions in this card.
- Do not change `AGENTS.md`, `CLAUDE.md`, or any spec file.

## Files

- `README.md`
- `README_CN.md`

## Acceptance Tests

(docs-only card — no failing test function)

## Tests

```bash
go run ./internal/checks/module-manifests
gofmt -l .
```

```text
Manual: README.md Quick Start section contains a complete, runnable Go example ≤30 lines.
Manual: README.md stdlib comparison table has ≥8 feature rows covering:
  - Basic routing
  - {param} path extraction
  - Route groups
  - Per-group middleware
  - Named routes + reverse URL
  - Route freeze (immutable after Prepare)
  - Structured error responses
  - Request ID context carriage
Manual: README.md Package overview table lists all 9 stable roots with one-line descriptions.
Manual: README_CN.md mirrors README.md section structure.
```

## Docs Sync

- `README.md` (primary target)
- `README_CN.md` (mirror)

## Validation

```bash
go run ./internal/checks/module-manifests
gofmt -l .
```

## Done Definition

- [ ] `README.md` has sections: Quick Start, Why plumego, stdlib comparison table, Package overview, Getting Help.
- [ ] Quick Start Go example is ≤30 lines, self-contained, and syntactically valid.
- [ ] Comparison table has ≥8 feature rows.
- [ ] Package overview table lists all 9 stable roots.
- [ ] `README_CN.md` mirrors structure with accurate translation.
- [ ] All Validation commands exit 0.
- [ ] `gofmt -l .` produces no output.

## Outcome

<!-- Agent fills after completion -->

# Plan for M-013: Migration Guides

Milestone: `M-013`
Objective: Publish four migration guides in docs/migration/ covering from-gin,
from-echo, from-chi, and middleware-compat, each with a framework-concept
mapping table and runnable Go code snippets that give teams a clear,
code-anchored path to adopting plumego incrementally.
Constraints: no code changes to stable roots or x/*, no compatibility shims
or dual-framework wrappers, all code snippets reference current API surface
only, docs/ADOPTION_PATH.md updated to reference the new section.
Affected Modules: docs.

## Phase Map

- Phase 1: Orient — read canonical style guide and existing adoption path doc to
  understand tone and link conventions.
- Phase 2: Implement (parallel) — write all four guides concurrently since they
  are independent documents.
- Phase 3: Validate and Ship — update ADOPTION_PATH.md, verify code snippet API
  accuracy, run acceptance criteria, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1530 | Write docs/migration/from-gin.md | docs | `docs/migration/from-gin.md` | M-008 | file exists, gofmt -l . |
| 1531 | Write docs/migration/from-echo.md | docs | `docs/migration/from-echo.md` | M-008 | file exists, gofmt -l . |
| 1532 | Write docs/migration/from-chi.md | docs | `docs/migration/from-chi.md` | M-008 | file exists, gofmt -l . |
| 1533 | Write docs/migration/middleware-compat.md | docs | `docs/migration/middleware-compat.md` | M-008 | file exists, gofmt -l . |

## Dependency Edges

- Cards 1530, 1531, 1532, 1533 all depend on M-008 being merged (stable API surface fixed).
- Cards 1530, 1531, 1532, 1533 are independent of each other.

## Parallel Groups

- Group A (parallel): cards 1530, 1531, 1532, 1533 — independent documents, no file overlap.
- Group B (sequential after A): update docs/ADOPTION_PATH.md to reference docs/migration/.

## Risk Register

- Risk: code snippets in a guide reference a contract symbol that was renamed or removed.
  Mitigation: each guide author must read contract/module.yaml and the current exported
  API before writing snippets; the reference-layout check catches broken import paths.
- Risk: from-gin and from-echo guides inadvertently suggest running both frameworks
  simultaneously.
  Mitigation: each guide opens with an explicit "translation approach, not dual-runtime"
  callout; the non-goals section in this plan prohibits shims.

## Verification Strategy

- Card-level checks: file exists and is non-empty after each card; no gofmt issues
  in any embedded Go code blocks.
- Milestone-level checks: run dependency-rules, module-manifests, and reference-layout;
  confirm docs/ADOPTION_PATH.md links resolve.
- API accuracy: verify each guide's code snippets use only symbols present in the
  current contract, router, and middleware packages.

## Exit Condition

- all four guide files exist in docs/migration/
- docs/ADOPTION_PATH.md references docs/migration/ under the migration entry point
- all code snippets reference current API surface symbols only
- verify report shows pass
- milestone acceptance criteria ready for PR packaging

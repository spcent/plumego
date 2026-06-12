# Card 1532

Milestone: M-025
Recipe: (none — docs registration)
Context Package: control-plane
Priority: P1
State: done
Primary Module: docs
Owned Files: use-cases/AGENTS.md, AGENTS.md, CLAUDE.md, milestone artifacts
Depends On: 1531

## Goal

mini-saas-api is registered everywhere agents look: the use-cases app table,
the root AGENTS.md go.mod-rationale list, and the CLAUDE.md snapshot; the
M-025 milestone artifacts reflect completed implementation.

## Scope

use-cases/AGENTS.md app table; root AGENTS.md §2 rationale; CLAUDE.md
snapshot; M-025.md checkboxes; verify-M-025.md; ROADMAP/STATUS rows;
cards 1525–1532 moved to done/.

## Non-goals

No website-sync (no generated doc sources changed: root README, docs/modules,
task-routing, roadmap untouched).

## Files

use-cases/AGENTS.md
AGENTS.md
CLAUDE.md
tasks/milestones/active/M-025-mini-saas-api/M-025.md
tasks/milestones/active/M-025-mini-saas-api/verify-M-025.md

## Acceptance Tests

(docs-only) go run ./internal/checks/agent-workflow exits 0.

## Tests

All six baseline boundary checks + make validate-diff (docs_only profile).

## Docs Sync

This card is the docs sync.

## Validation

go run ./internal/checks/agent-workflow
make validate-diff

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated.

## Outcome

mini-saas-api registered in use-cases/AGENTS.md table, root AGENTS.md §2
(zero-external-deps rationale), and CLAUDE.md snapshot. M-025 phases marked
complete; verify-M-025.md records gate output; ROADMAP/STATUS rows updated;
cards 1525–1532 archived to done/.

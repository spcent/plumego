# Card 0777

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: active
Primary Module: contract
Owned Files:
- tasks/cards/active/0777-contract-stable-residual-gate.md
- tasks/cards/done/0777-contract-stable-residual-gate.md
Depends On:
- 0776

Goal:
Run and record the stable gate after residual contract hardening.

Scope:
- Run focused contract tests and vet.
- Run affected external module tests for trace usage changes.
- Run repo-wide tests, vet, and boundary/manifest/workflow checks.
- Archive the card with exact validation results.

Non-goals:
- Do not make unrelated runtime changes unless a gate fails.
- Do not alter other active queues.

Files:
- tasks/cards/active/0777-contract-stable-residual-gate.md
- tasks/cards/done/0777-contract-stable-residual-gate.md

Tests:
- go test -race -timeout 60s ./contract/...
- go test -timeout 20s ./...
- go vet ./...

Docs Sync:
- Not required unless a gate failure requires behavior or docs changes.

Done Definition:
- Final validation is recorded in the archived card.
- Worktree is clean after the final commit.

Outcome:


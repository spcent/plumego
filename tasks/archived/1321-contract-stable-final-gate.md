# Card 1321

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: done
Primary Module: contract
Owned Files:
- tasks/cards/done/1321-contract-stable-final-gate.md
- tasks/cards/done/1321-contract-stable-final-gate.md
Depends On:
- 0770

Goal:
Run and record the final stable-readiness gate after the contract hardening queue is complete.

Scope:
- Run contract race tests and normal tests.
- Run contract vet plus repo boundary/manifest/workflow checks.
- Run repo-wide test and vet if earlier cards touched cross-module behavior.
- Record exact validation results in the card before archiving it.

Non-goals:
- Do not make new runtime changes in this card unless a gate fails.
- Do not expand unrelated active card queues.

Files:
- tasks/cards/done/1321-contract-stable-final-gate.md
- tasks/cards/done/1321-contract-stable-final-gate.md

Tests:
- go test -race -timeout 60s ./contract/...
- go test -timeout 20s ./...
- go vet ./...

Docs Sync:
- Not required unless gate failure requires behavior or docs changes.

Done Definition:
- Final gates are recorded in the archived card.
- Worktree is clean after the final commit.

Outcome:
- Ran the final contract stable gate after cards 0765-0770.
- Boundary, manifest, workflow, reference-layout, formatting, race, vet, and repo-wide tests passed.
- An initial parallel `go test -timeout 20s ./...` run timed out in the contract conformance cold scan while race/vet/checks were running concurrently; standalone reruns passed at both 60s and 20s.

Validation:
- go test -race -timeout 60s ./contract/...
- go test -timeout 60s ./...
- go test -timeout 20s ./...
- go vet ./...
- go run ./internal/checks/dependency-rules
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
- go run ./internal/checks/reference-layout
- gofmt -l .

# Card 1401

Milestone: M-004
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: done
Primary Module: contract
Owned Files:
- tasks/milestones/active/M-004.md
- tasks/cards/active/1401-stable-root-cleanup-gate-evidence.md
Depends On:
- 1394
- 1376
- 1377
- 1378
- 1379
- 1380
- 1395
- 1396
- 1397
- 1398
- 1399
- 1400

Goal:
Record final stable-root cleanup validation evidence and surface any remaining v1 blockers.

Scope:
- Run stable-root boundary checks and targeted stable-root race/test/vet gates.
- Confirm `gofmt -l .` is empty.
- Record exact commands and pass/fail outcomes in the completed card or milestone verify artifact.
- If failures reveal new cleanup work, create bounded follow-up cards instead of expanding this card.

Non-goals:
- Do not tag a release.
- Do not run extension promotion work.
- Do not modify runtime code except to fix validation failures found by this card.

Files:
- tasks/milestones/active/M-004.md
- tasks/cards/active/1401-stable-root-cleanup-gate-evidence.md

Tests:
- go run ./internal/checks/dependency-rules
- go test -race -timeout 60s ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics
- go vet ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics

Docs Sync:
- Record remaining blockers if validation fails.

Done Definition:
- Stable-root cleanup gates pass or blockers are recorded as active cards.
- The milestone has enough evidence to decide whether stable roots are v1-cleanup ready.
- No untracked stable-root cleanup item remains implicit.

Outcome:
- Ran the M-004 stable-root cleanup acceptance criteria after completing cards
  1394 through 1400 and contract hardening cards 1376 through 1380.
- No blockers found.
- No runtime code was changed in this evidence card.

Validation:
- go run ./internal/checks/dependency-rules
- go run ./internal/checks/agent-workflow
- go run ./internal/checks/module-manifests
- go run ./internal/checks/reference-layout
- go run ./internal/checks/deprecation-inventory -strict
- go test -race -timeout 60s ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics
- go test -timeout 20s ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics
- go vet ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics
- gofmt -l .

# Card 0764

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: contract
Owned Files:
- tasks/cards/done/0764-contract-release-gate-evidence.md
Depends On:
- 0763

Goal:
Record final validation evidence for the contract stable hardening pass.

Scope:
- Run targeted contract race/test/vet gates.
- Run boundary, manifest, workflow, and reference layout checks.
- Run repo-wide test/vet when code-bearing changes in earlier cards require it.
- Record the exact commands and outcomes in the completed card.

Non-goals:
- Do not change runtime code.
- Do not tag or promote a release.
- Do not run unrelated websocket active cards.

Files:
- tasks/cards/done/0764-contract-release-gate-evidence.md

Tests:
- go test -race -timeout 60s ./contract/...
- go test -timeout 20s ./...
- go vet ./...

Docs Sync:
- None unless validation reveals mismatch.

Done Definition:
- Required stable gates pass or blockers are recorded.
- The card is moved to done with exact validation evidence.

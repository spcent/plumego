# Card 0752

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: contract
Owned Files:
- docs/modules/contract/README.md
Depends On:
- 0751

Goal:
Record the stable readiness gate expectations for `contract` after the final hardening pass.

Scope:
- Add a release-readiness checklist to the contract module docs.
- Run targeted contract checks, race tests, and repo-level boundary/manifest checks.
- Keep runtime behavior unchanged.

Non-goals:
- Do not promote or tag a release.
- Do not modify contract runtime code.
- Do not run unrelated websocket active cards.

Files:
- docs/modules/contract/README.md

Tests:
- go test -race -timeout 60s ./contract/...
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- This card is docs and verification evidence only.

Done Definition:
- Contract stable readiness checklist is documented.
- Required checks pass.

Outcome:


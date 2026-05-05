# Card 0772

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P2
State: active
Primary Module: x/data
Owned Files:
- docs/modules/x-data/README.md
- tasks/cards/active/0772-x-data-stable-readiness-fourth-gate.md
Depends On:
- 0771-x-data-sharding-config-doc-and-manifest-sync

Goal:
Run the fourth x/data stable-readiness gate after cards 0762-0771 and record the remaining go/no-go state.

Scope:
- Run x/data tests, race tests, vet, and boundary/manifest checks.
- Update docs with implemented fixes and any remaining stable blockers.
- Keep x/data experimental unless stable criteria are fully met.

Non-goals:
- Do not mark stable without repo-wide evidence and explicit no-go closure.
- Do not widen scope to unrelated active websocket cards.
- Do not modify runtime behavior in this gate card.

Files:
- docs/modules/x-data/README.md
- tasks/cards/active/0772-x-data-stable-readiness-fourth-gate.md

Tests:
- go test -timeout 20s ./x/data/...
- go test -race -timeout 60s ./x/data/...
- go vet ./x/data/...

Docs Sync:
- Update x/data docs with the fourth gate result and remaining blockers.

Done Definition:
- Validation commands are recorded.
- Remaining NO-GO items are explicit, or x/data stable readiness is justified.
- Completed card is archived to done.


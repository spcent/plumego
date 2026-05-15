# Card 1378

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/active_cards_regression_test.go
- docs/modules/contract/README.md
Depends On:
- 0761

Goal:
Freeze the misleading `EnableBodyCache` compatibility semantics.

Scope:
- Add regression coverage that `EnableBodyCache=false` still reads and retains body bytes inside `Ctx`, but does not restore `R.Body`.
- Document that the name means response-body restoration/cross-reader cache, not zero-copy/no-memory mode.
- Keep runtime behavior unchanged.

Non-goals:
- Do not rename `EnableBodyCache`.
- Do not change `BindJSON` to stream decode.
- Do not expand `Ctx` APIs.

Files:
- contract/active_cards_regression_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update BindJSON cache semantics.

Done Definition:
- The compatibility behavior is executable and documented.
- Target checks pass.

Outcome:
- Confirmed this active card is a stale duplicate of completed card
  `tasks/cards/done/1243-contract-bindjson-cache-semantics.md`.
- Confirmed current `contract` no longer exposes `BindJSON`, `EnableBodyCache`,
  `BindOptions`, or `Ctx`, so no runtime follow-up is in scope for M-004.
- No runtime change was required.

Validation:
- rg -n --glob '*.go' 'BindJSON|EnableBodyCache|type Ctx|BindOptions' contract
- go test -timeout 20s ./contract/...
- go vet ./contract/...

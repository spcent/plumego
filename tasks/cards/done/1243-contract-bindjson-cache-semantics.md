# Card 1243

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
- Added regression coverage proving `EnableBodyCache=false` still leaves body bytes retained inside `Ctx`.
- Confirmed the same mode does not restore `R.Body` for later readers.
- Documented that the flag is not a zero-copy/no-memory mode.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

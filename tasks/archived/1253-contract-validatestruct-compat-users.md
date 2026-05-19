# Card 1253

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- docs/modules/contract/README.md
Depends On:
- 0762

Goal:
Finalize the stable decision for current external production `ValidateStruct` users.

Scope:
- Record the exact v1 compatibility users and rationale.
- State that new complex validation should be module-owned.
- Keep the current allowlist as the executable guard.

Non-goals:
- Do not migrate existing users in this card.
- Do not expand validator behavior.
- Do not change conformance logic unless documentation reveals a mismatch.

Files:
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- This card is docs sync.

Done Definition:
- Existing external users are formally accepted as v1 compatibility users.
- New usage policy is clear.
- Target checks pass.

Outcome:
- Documented the exact v1 compatibility users for external production `ValidateStruct` usage.
- Recorded the rationale for retaining the current users rather than migrating late in the stable pass.
- Stated the rule for any future production caller: prefer module-owned validation, or update conformance allowlist and docs in the same change.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

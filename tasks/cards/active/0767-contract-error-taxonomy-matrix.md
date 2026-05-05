# Card 0767

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md
Depends On:
- 0766

Goal:
Make the precise typed error taxonomy and coarse status/category helpers visibly consistent and intentionally different where documented.

Scope:
- Add focused tests for `CategoryForStatus`, `HTTPStatusFromCategory`, `codeForStatus`, and `ErrorType.Meta` interaction.
- Document that `CategoryForStatus` and `HTTPStatusFromCategory` are coarse compatibility helpers.
- Preserve the existing 422 and timeout behavior unless a test exposes an unintended mismatch.

Non-goals:
- Do not change public taxonomy values without a breaking-change card.
- Do not add extension-specific categories.
- Do not move taxonomy ownership outside `contract`.

Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update taxonomy guidance and helper caveats.

Done Definition:
- Typed and coarse taxonomy paths are covered by an explicit stable matrix.
- Documented intentional mismatches are guarded by tests.
- Target checks pass.

Outcome:


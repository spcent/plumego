# Card 0766

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- docs/modules/contract/README.md
- contract/module.yaml
- contract/conformance_test.go
Depends On:
- 0765

Goal:
Record the v1 decision for exported `APIError` as compatibility surface guarded by conformance, not by type-system closure.

Scope:
- Make the `APIError` compatibility decision explicit in module and docs.
- Clarify that non-test external struct literals remain forbidden by conformance.
- Ensure conformance coverage continues to reject external non-test literals.

Non-goals:
- Do not hide `APIError` fields in v1.
- Do not introduce a replacement error type.
- Do not change error response JSON shape.

Files:
- docs/modules/contract/README.md
- contract/module.yaml
- contract/conformance_test.go

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update contract public surface decision text.

Done Definition:
- The exported struct tradeoff is explicitly frozen for v1.
- Builder-first usage remains the only allowed external non-test construction path.
- Target checks pass.

Outcome:
- Recorded the v1 decision that `APIError` remains an exported compatibility struct, while non-test external construction must stay builder-first.
- Added module/docs language that this is enforced by conformance rather than type-system closure.
- Added an enforcement-point comment to the APIError literal conformance test.

Validation:
- go test -timeout 60s ./contract/...
- go vet ./contract/...

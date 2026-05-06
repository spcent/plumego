# Card 0776

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/errors_test.go
- docs/modules/contract/README.md
Depends On:
- 0775

Goal:
Freeze the `APIError.Details` unsupported-value passthrough boundary with focused tests.

Scope:
- Add a regression test for a typed string-key map containing pointer/struct values.
- Clarify that unsupported values can cause the enclosing typed container to remain compatibility passthrough.

Non-goals:
- Do not reject unsupported values in v1.
- Do not change the error envelope shape.
- Do not add non-stdlib dependencies.

Files:
- contract/errors_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update details passthrough matrix.

Done Definition:
- The all-or-passthrough edge case is tested and documented.
- Target checks pass.

Outcome:


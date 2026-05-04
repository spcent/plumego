# Card 0743

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- contract/freeze_test.go
- docs/modules/contract/README.md
Depends On:
- 0742

Goal:
Reduce drift from direct `APIError` literals by making builder-first construction and typed metadata conformance executable.

Scope:
- Add focused tests for invalid typed/severity normalization.
- Add tests that document direct literal repair behavior as compatibility only.
- Clarify builder-first construction guidance for external callers.
- Keep exported API unchanged.

Non-goals:
- Do not hide or remove `APIError`.
- Do not add new public symbols.
- Do not scan or rewrite every external caller in this card.

Files:
- contract/errors.go
- contract/errors_test.go
- contract/freeze_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update contract docs if construction guidance changes.

Done Definition:
- Typed error normalization behavior is covered by explicit regression tests.
- Invalid literal repair is documented as compatibility behavior.
- Contract tests and vet pass.

Outcome:


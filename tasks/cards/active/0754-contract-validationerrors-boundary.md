# Card 0754

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: contract
Owned Files:
- contract/validation_external_test.go
- docs/modules/contract/README.md
Depends On:
- 0753

Goal:
Clarify that `ValidationErrors` is a read-only external consumption surface, not an externally constructible error model.

Scope:
- Add external-package tests proving callers can consume `ValidationErrors` through `errors.As`, `Error()`, and `Errors()`.
- Verify `Errors()` remains defensive and cannot mutate internal field errors.
- Document the construction boundary and supported extension pattern.

Non-goals:
- Do not export `ValidationErrors` fields.
- Do not add constructors for external callers.
- Do not change validation response shape.

Files:
- contract/validation_external_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update validation guidance in the contract module README.

Done Definition:
- External consumption behavior is covered by tests.
- Docs explain why construction stays package-owned and how extensions should wrap or translate their own validation errors.

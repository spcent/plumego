# Card 0748

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/conformance_test.go
- docs/modules/contract/README.md
Depends On:
- 0747

Goal:
Make the builder-first `APIError` construction rule executable by detecting external non-test `contract.APIError{}` literals.

Scope:
- Add a contract-level conformance test that scans repository Go files for external `contract.APIError{}` composite literals.
- Allow package-internal contract literals and tests.
- Keep `APIError` exported for compatibility.
- Clarify the conformance expectation in contract docs.

Non-goals:
- Do not hide or remove `APIError`.
- Do not rewrite unrelated extension error builders.
- Do not add a new standalone check binary.

Files:
- contract/conformance_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update contract docs to mention the conformance test.

Done Definition:
- External non-test `contract.APIError{}` literals fail contract tests.
- Builder-first construction remains documented as the supported path.
- Contract tests and vet pass.

Outcome:
- Added a contract conformance test that scans non-test Go files outside `contract` and rejects external `contract.APIError{}` composite literals.
- Kept package-internal and test literals available for compatibility and regression coverage.
- Documented that `go test ./contract` enforces builder-first construction for non-test external callers.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

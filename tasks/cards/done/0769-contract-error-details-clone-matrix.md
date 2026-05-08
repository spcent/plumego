# Card 0769

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md
Depends On:
- 0768

Goal:
Turn `APIError.Details` clone behavior into a stable support matrix instead of an implicit implementation detail.

Scope:
- Add tests for common typed JSON-like maps and slices.
- Expand cloning where it is safe and stdlib-only.
- Document unsupported passthrough behavior clearly.

Non-goals:
- Do not reject existing unsupported detail values in v1.
- Do not change the error envelope shape.
- Do not add reflection-heavy validation framework behavior outside cloning.

Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update details clone support matrix.

Done Definition:
- JSON-like details have stable clone tests.
- Unsupported values remain compatibility passthrough and are documented.
- Target checks pass.

Outcome:
- Expanded `APIError.Details` cloning to safely handle typed JSON-like maps, slices, arrays, and scalar containers through stdlib reflection.
- Preserved unsupported values such as pointers and structs as compatibility passthrough.
- Added tests for typed map/slice clone isolation and unsupported passthrough behavior.
- Documented the clone support matrix.

Validation:
- go test -timeout 60s ./contract/...
- go vet ./contract/...

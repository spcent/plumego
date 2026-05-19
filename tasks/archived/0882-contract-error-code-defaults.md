# Card 0882

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md
Depends On:
- 0730

Goal:
Make normalized fallback error codes canonical and machine-safe instead of deriving codes from HTTP status text.

Scope:
- Replace the `normalizeAPIError` fallback that currently uses `http.StatusText(status)` as a code.
- Add or update focused tests for incomplete `APIError` normalization.
- Keep existing explicit caller-supplied codes unchanged.
- Document the fallback contract in the contract module README if it is not already clear.

Non-goals:
- Do not change exported error code constants beyond fallback selection.
- Do not change response envelope shape.
- Do not alter extension-owned error codes.

Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Update `docs/modules/contract/README.md` only for the implemented fallback semantics.

Done Definition:
- Missing `APIError.Code` values normalize to stable `Code*` constants rather than title-cased HTTP reason phrases.
- Explicit codes are preserved.
- Targeted tests, vet, and manifest checks pass.

Outcome:
- Added canonical status-to-code fallback mapping for incomplete `APIError` values.
- Updated default-code tests to assert stable `Code*` constants rather than `http.StatusText` values.
- Added coverage that explicit caller-supplied codes are preserved.
- Documented that missing codes normalize to machine-safe `Code*` constants.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

# Card 0738

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/error_codes.go
- contract/errors_test.go
- contract/module.yaml
- docs/modules/contract/README.md
Depends On:
- 0737

Goal:
Add canonical gateway-timeout error metadata so upstream timeout responses do not misuse request-timeout semantics.

Scope:
- Add a transport-owned `Code*` constant for gateway timeout.
- Add an `ErrorType` and metadata entry for HTTP 504 gateway timeout.
- Update status-to-code fallback coverage for HTTP 504.
- Sync public surface docs and module manifest.

Non-goals:
- Do not change extension-specific timeout handling yet.
- Do not add gateway protocol behavior to `contract`.
- Do not change response envelope shape.

Files:
- contract/errors.go
- contract/error_codes.go
- contract/errors_test.go
- contract/module.yaml
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Sync `contract/module.yaml` and `docs/modules/contract/README.md` with the new public error type/code.

Done Definition:
- HTTP 504 has canonical type/code/status metadata.
- Missing code fallback for 504 uses the new machine code.
- Targeted tests, vet, and manifest checks pass.

Outcome:

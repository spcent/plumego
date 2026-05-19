# Card 0847

Milestone: v1
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md
- contract/module.yaml
Depends On:
- 0727

Goal:
Make `ErrorType.Meta()` return a clear exported metadata type instead of an unnamed public shape.

Scope:
- Export the error type metadata type with a stable name.
- Keep the existing metadata fields and behavior intact.
- Add or adjust tests proving callers can use the returned metadata directly.
- Update module docs/public entrypoint inventory.

Non-goals:
- Do not change existing `ErrorType` string values.
- Do not change status/code/category mappings.
- Do not add specific error constructor families.

Files:
- contract/errors.go
- contract/errors_test.go
- contract/module.yaml
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Document the exported metadata type in `contract/module.yaml` and `docs/modules/contract/README.md`.

Done Definition:
- `ErrorType.Meta()` exposes an exported, nameable metadata type.
- Existing error builder behavior remains unchanged.
- Targeted tests, vet, and manifest checks pass.

Outcome:
- Exported `ErrorTypeMeta` as the nameable metadata value returned by `ErrorType.Meta()`.
- Preserved the existing `Status`, `Code`, and `Category` fields and all `ErrorType` mappings.
- Added regression coverage proving callers can assign `TypeNotFound.Meta()` to `ErrorTypeMeta`.
- Updated contract public surface documentation and module manifest.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

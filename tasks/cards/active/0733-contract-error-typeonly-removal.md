# Card 0733

Milestone: v1
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/bind_helpers.go
- contract/errors_test.go
- contract/freeze_test.go
- x/ai/streaming/handler.go
- docs/modules/contract/README.md
- contract/module.yaml
Depends On:
- 0732

Goal:
Remove `ErrorBuilder.TypeOnly` so callers cannot attach an error type without its canonical status, category, and code metadata.

Scope:
- Enumerate all Go callers of `TypeOnly(` before editing.
- Replace remaining callers with `Type(...)` when canonical metadata is desired or explicit `Status/Category/Code` fields when not.
- Delete the `TypeOnly` method and tests that assert arbitrary mismatched type metadata.
- Update docs and module manifest public API listings.

Non-goals:
- Do not remove `ErrorBuilder.Type`.
- Do not add another bypass method for type-only assignment.
- Do not widen bind-helper behavior beyond preserving current response status/code/category.

Files:
- contract/errors.go
- contract/bind_helpers.go
- contract/errors_test.go
- contract/freeze_test.go
- x/ai/streaming/handler.go
- docs/modules/contract/README.md
- contract/module.yaml

Tests:
- rg -n --glob '*.go' 'TypeOnly\\(' .
- go test -timeout 20s ./contract/... ./x/ai/streaming/...
- go build ./...

Docs Sync:
- Sync `contract/module.yaml` and `docs/modules/contract/README.md` with the removed public method.

Done Definition:
- `ErrorBuilder.TypeOnly` is removed.
- No Go callers remain.
- Binding and streaming tests pass with explicit or canonical error metadata.

Outcome:

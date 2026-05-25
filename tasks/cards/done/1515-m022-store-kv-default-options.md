# Card 1515

Milestone: M-022
Recipe: specs/change-recipes/docs-and-config.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: store
Owned Files:
- `store/kv/options.go`
- `store/kv/kv_test.go`
- `docs/modules/store/README.md`
Depends On: none

Goal:
- Make `store/kv` defaults discoverable through a public helper without
  renaming `Options` or weakening the explicit `DataDir` requirement.

Scope:
- Add a public `DefaultOptions(dataDir string) Options` helper.
- Add focused tests for the helper behavior.
- Document the helper in the store module README.

Non-goals:
- Do not rename `Options` to `Config`.
- Do not make `DataDir` implicit or optional.
- Do not change runtime default values beyond exposing them through a helper.

Files:
- `store/kv/options.go`
- `store/kv/kv_test.go`
- `docs/modules/store/README.md`

Acceptance Tests:
- `go test -timeout 20s ./store/kv`

Tests:
- `go test -timeout 20s ./store/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- `docs/modules/store/README.md`

Validation:
- `go test -timeout 20s ./store/...`
- `go run ./internal/checks/module-manifests`
- `gofmt -l .`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Added `store/kv.DefaultOptions(dataDir)` as the public way to start from the
  embedded KV defaults without hiding the required `DataDir`.
- Added focused test coverage for the helper and updated the store module docs
  to point callers at it.
- Corrected the README drift so the documented `store/kv` defaults now match the
  implemented values (100000 entries, 200 MiB) instead of the stale smaller
  numbers.
- Validation:
  - `go test -timeout 20s ./store/...`
  - `go run ./internal/checks/module-manifests`
  - `gofmt -l .`

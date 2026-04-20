# Card 2010

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/store
Owned Files:
- `reference/workerfleet/internal/platform/store/interfaces.go`
- `reference/workerfleet/internal/platform/store/types.go`
- `reference/workerfleet/internal/platform/store/memory/*`
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/internal/app/service_test.go`
Depends On:
- `tasks/cards/active/2009-workerfleet-submodule-bootstrap.md`
Blocked By:
- workerfleet MongoDB design review approval

Goal:
- Split the app-local store interfaces from the in-memory implementation so `reference/workerfleet/internal/app` no longer depends on `*MemoryStore` directly.
- Create a stable app-local repository seam for a later MongoDB backend without changing handler or domain semantics.

Scope:
- Extract interface definitions and shared filter/query types into app-local store abstractions.
- Move the in-memory backend under `reference/workerfleet/internal/platform/store/memory`.
- Update app service construction to depend on interface combinations instead of a concrete memory type.

Non-goals:
- Do not add MongoDB code.
- Do not change HTTP request or response contracts.
- Do not change worker status rules or alert logic.

Files:
- `reference/workerfleet/internal/platform/store/interfaces.go`
- `reference/workerfleet/internal/platform/store/types.go`
- `reference/workerfleet/internal/platform/store/memory/*.go`
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/internal/app/service_test.go`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/store/...`
- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go test ./...`

Docs Sync:
- Update `reference/workerfleet/docs/storage.md` if the app-local persistence abstraction layout changes in a way future cards must rely on.

Done Definition:
- `app.Service` depends on app-local store interfaces rather than `*MemoryStore`.
- The in-memory implementation remains the default backend and preserves current behavior.
- Existing query and ingest tests continue to pass without HTTP contract changes.
- Later MongoDB cards can implement the same interfaces without reshaping app logic.

Outcome:
- Added app-local store interfaces and shared filter/query types under `reference/workerfleet/internal/platform/store`.
- Moved the in-memory backend to `reference/workerfleet/internal/platform/store/memory` and updated tests to cover the new package.
- Updated `reference/workerfleet/internal/app/service.go` to depend on the `store.QueryStore` interface instead of a concrete memory type.
- Updated storage docs to describe the store-interface versus memory-backend split.
- Validation run:
  - `cd reference/workerfleet && go test ./internal/platform/store/...`
  - `cd reference/workerfleet && go test ./internal/app/...`
  - `cd reference/workerfleet && go test ./...`

# Card 1407

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/rest
Owned Files:
- x/rest/resource_db.go
- x/rest/resource_db_repo.go
- x/rest/resource_db_routes.go
- x/rest/resource_db_test.go
- x/rest/routes_test.go
Depends On:
- 1406

Goal:
- Split `x/rest` repository-backed resource wiring into focused files without changing beta behavior.

Scope:
- Move repository adapter helpers into `resource_db_repo.go`.
- Move DB-backed route registration helpers into `resource_db_routes.go`.
- Preserve query parsing, pagination, error mapping, and route registration behavior.
- Keep legacy sort-query compatibility expectations as currently tested.

Non-goals:
- Do not change resource public interfaces.
- Do not change route naming or HTTP method registration.
- Do not alter `x/rest/versioning`.

Files:
- x/rest/resource_db.go
- x/rest/resource_db_repo.go
- x/rest/resource_db_routes.go
- x/rest/resource_db_test.go
- x/rest/routes_test.go

Tests:
- go test -timeout 20s ./x/rest/...
- go vet ./x/rest/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless examples or comments move in a user-visible way.

Done Definition:
- DB resource wiring has clear repository and route file ownership.
- Existing REST resource and route tests pass.
- No public API or route behavior change is introduced.

Outcome:
- Completed on May 15, 2026.
- Split SQL builder and `BaseRepository` implementation from `x/rest/resource_db.go` into `x/rest/resource_db_repo.go`.
- Split `DBResourceController` HTTP handlers into `x/rest/resource_db_routes.go`.
- Kept `x/rest/resource_db.go` focused on the public repository interface, DB controller shape, and common model field sets.
- Preserved route behavior, query parsing, pagination, and DB error mapping.
- Validation passed:
  - `go test -timeout 20s ./x/rest/...`
  - `go vet ./x/rest/...`
  - `go run ./internal/checks/dependency-rules`

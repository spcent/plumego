# Card 1417

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/tenant/config
Owned Files:
- x/tenant/config/manager.go
- x/tenant/config/manager_legacy.go
- x/tenant/config/manager_queries.go
- x/tenant/config/manager_test.go
Depends On:
- 1416

Goal:
- Split tenant config manager persistence and legacy column migration paths.

Scope:
- Move SQL query assembly and scan helpers into `manager_queries.go`.
- Move legacy quota column migration helpers into `manager_legacy.go`.
- Preserve migration behavior from per-minute legacy columns into quota limits.
- Keep tenant config public API unchanged.

Non-goals:
- Do not change quota semantics.
- Do not change database schema.
- Do not promote `x/tenant` maturity in this card.

Files:
- x/tenant/config/manager.go
- x/tenant/config/manager_legacy.go
- x/tenant/config/manager_queries.go
- x/tenant/config/manager_test.go

Tests:
- go test -timeout 20s ./x/tenant/config
- go vet ./x/tenant/config
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless legacy migration behavior is documented differently.

Done Definition:
- Tenant config SQL and legacy migration code have separate file ownership.
- Existing config manager tests pass.
- No tenant config API or schema behavior changes.

Outcome:
- Completed on 2026-05-15.
- Moved tenant config SELECT/UPSERT SQL, scan helpers, row conversion, and JSON parsing into `x/tenant/config/manager_queries.go`.
- Moved legacy per-minute quota column migration into `x/tenant/config/manager_legacy.go`.
- Kept tenant config public API and schema behavior unchanged.
- Validation:
  - `go test -timeout 20s ./x/tenant/config`
  - `go vet ./x/tenant/config`
  - `go run ./internal/checks/dependency-rules`

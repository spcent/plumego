# Card 0319: Store Stable Tenant Artifact Eviction

Priority: P0
State: done
Primary Module: store

## Goal

Remove tenant-specific artifacts from stable `store/db` and bring its package-level API guidance back in sync with the actual stable SQL helper surface.

## Problem

- `store/db/migrations/001_create_tenants_table.up.sql` and `.down.sql` define tenant quota and allowed-model metadata tables inside the stable store root.
- This violates the stable store boundary declared in:
  - `AGENTS.md`
  - `store/module.yaml`
  - `docs/modules/store/README.md`
- These migrations are clearly tenant-aware application policy, not transport-agnostic persistence primitives.
- `store/db/sql.go` also contains a stale package comment/example that still references a nonexistent `db.Connect(...)` API and non-existent config fields such as `Host`, `Port`, `Username`, and `Password`, even though the real stable API is `Open(Config{Driver, DSN, ...})`.
- The result is boundary drift plus API guidance drift in the same stable package.

## Scope

- Remove tenant-specific migration assets from stable `store/db`.
- If the migrations are still needed, relocate them to the owning tenant package in the same change or document the exact follow-up landing zone before merge.
- Rewrite the `store/db` package comment/example so it matches the real stable API (`DefaultConfig`, `Open`, `OpenWith`, `Config{Driver, DSN, ...}`).
- Re-run a repository search to confirm no tenant-specific migration artifacts remain under stable `store/db`.

## Non-Goals

- Do not redesign the SQL helper APIs themselves.
- Do not add new provider-specific migration tooling to stable `store`.
- Do not widen the cleanup into topology or observability work.

## Files

- `store/db/migrations/001_create_tenants_table.up.sql`
- `store/db/migrations/001_create_tenants_table.down.sql`
- `store/db/sql.go`
- `docs/modules/store/README.md`

## Tests

- `go test -timeout 20s ./store/... ./x/tenant/...`
- `go test -race -timeout 60s ./store/... ./x/tenant/...`
- `go run ./internal/checks/dependency-rules`

## Docs Sync

- Keep the `store/db` package comment and `docs/modules/store/README.md` aligned with the final landing zone for tenant-aware migration assets.

## Done Definition

- Stable `store/db` no longer contains tenant-specific migration assets.
- `store/db` comments and examples describe the actual exported API.
- A post-change search under `store/db` is free of tenant-specific schema ownership.
- Store and tenant package tests pass.

## Outcome

- Removed the tenant schema migration files from stable `store/db`.
- Added tenant-owned migration assets under `x/tenant/config/migrations/`.
- Updated the `store/db` package comment to document the real `DefaultConfig` + `Open` API instead of the stale `Connect` example.
- Clarified in store and tenant module docs that tenant configuration schema ownership lives under `x/tenant/config`.

## Validation Run

```bash
gofmt -w store/db/sql.go x/tenant/config/doc.go
rg -n 'create_tenants_table|quota_requests_per_minute|allowed_models|allowed_tools|allowed_methods|allowed_paths' store/db x/tenant/config
go test -timeout 20s ./store/... ./x/tenant/...
go run ./internal/checks/dependency-rules
```

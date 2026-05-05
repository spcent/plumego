# Card 0766

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/types.go
- x/data/file/metadata.go
- x/data/file/metadata_test.go
- docs/modules/x-data/README.md
Depends On:
- 0765-x-data-sharding-cross-shard-queryrow-contract

Goal:
Make metadata listing semantics match the tenant-scoped metadata contract.

Scope:
- Require tenant id for tenant-facing `List`.
- Add an explicit admin/global listing path if needed by existing callers.
- Add tests for empty tenant list rejection and tenant-filtered results.

Non-goals:
- Do not change the metadata table schema.
- Do not change store/file interfaces.
- Do not add HTTP behavior.

Files:
- x/data/file/types.go
- x/data/file/metadata.go
- x/data/file/metadata_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Update x/data docs to distinguish tenant-facing and admin metadata listing.

Done Definition:
- Tenant-facing `List` cannot omit tenant id.
- Admin/global listing is explicit if retained.
- Tests and docs cover the contract.


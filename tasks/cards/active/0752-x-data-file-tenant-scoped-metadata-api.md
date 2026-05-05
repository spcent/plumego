# Card 0752

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/types.go
- x/data/file/metadata.go
- x/data/file/metadata_test.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md
Depends On:
- 0751-x-data-stable-readiness-second-gate

Goal:
Make file metadata reads and mutations tenant-scoped so callers cannot access or mutate another tenant's metadata by global id or path.

Scope:
- Add tenant-aware metadata lookup and mutation methods.
- Update database queries to include tenant_id for Get, GetByPath, Delete, and UpdateAccessTime.
- Preserve explicit admin/global access only if it is named as such and not used by tenant storage paths.
- Add focused tests for tenant-scoped SQL predicates.

Non-goals:
- Do not change store/file interfaces.
- Do not add HTTP behavior.
- Do not change the file metadata schema.

Files:
- x/data/file/types.go
- x/data/file/metadata.go
- x/data/file/metadata_test.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Update x/data docs to state that tenant-facing metadata access is tenant-scoped.

Done Definition:
- Tenant-facing metadata access requires tenant id.
- Cross-tenant metadata id/path access returns not found.
- Tests and docs are updated.

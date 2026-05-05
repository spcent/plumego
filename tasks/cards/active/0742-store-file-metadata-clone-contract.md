# Card 0742

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: active
Primary Module: store
Owned Files:
- store/file/types.go
- store/file/coverage_test.go
- docs/modules/store/README.md

Goal:
Make store/file metadata cloning detach common nested mutable values.

Scope:
- Deep-clone metadata maps and common slice/map value shapes.
- Preserve scalar values unchanged.
- Add tests proving nested map/slice mutations do not affect clones.
- Sync store docs with metadata clone behavior.

Non-goals:
- Do not introduce reflection-heavy generic copying.
- Do not validate backend-specific metadata schemas.
- Do not add tenant-aware metadata behavior.

Files:
- store/file/types.go
- store/file/coverage_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/file
- go vet ./store/file
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update store module docs for metadata clone semantics.

Done Definition:
- Nested metadata maps/slices are detached for supported common shapes.
- Existing file contract tests still pass.
- Targeted tests, vet, and dependency checks pass.

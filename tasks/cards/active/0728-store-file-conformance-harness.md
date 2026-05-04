# Card 0728: Store File Conformance Harness

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/file/file.go
- store/file/TESTING.md
- x/data/file/local_test.go
- x/data/file/s3_test.go
Depends On:
- 0727

Goal:
Turn `store/file.Storage` semantics from loose comments into executable first-party conformance coverage.

Scope:
- Define deterministic stable expectations for missing delete, invalid path, negative limit, list ordering, and copy overwrite behavior.
- Add focused first-party backend tests for local and S3 storage where the stable behavior is implementable without external services.
- Keep concrete backend logic in `x/data/file`.

Non-goals:
- Do not add new stable file backend implementations.
- Do not add signed URL, metadata-manager, tenant path policy, or HTTP transport behavior to `store/file`.
- Do not add new non-stdlib dependencies.

Files:
- store/file/file.go
- store/file/TESTING.md
- x/data/file/local_test.go
- x/data/file/s3_test.go

Tests:
- go test -timeout 20s ./store/file ./x/data/file
- go test -race -timeout 60s ./store/file ./x/data/file
- go vet ./store/file ./x/data/file

Docs Sync:
- Required for file contract wording.

Done Definition:
- Stable file semantics are stated as deterministic expectations.
- Local and S3 tests cover the stable expectations that apply to them.
- Focused tests and vet pass.

Outcome:

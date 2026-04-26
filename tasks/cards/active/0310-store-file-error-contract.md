# Card 0310

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: store
Owned Files:
- store/file/errors.go
- store/file/coverage_test.go
Depends On:
- 0309-store-db-error-chain-guardrails

Goal:
Tighten `store/file.Error` boundary behavior and coverage for incomplete error values.

Scope:
- Make `(*Error).Error` nil-receiver safe.
- Make `(*Error).Unwrap` nil-receiver safe.
- Avoid formatting an absent underlying error as `<nil>`.
- Add focused tests for nil receivers and missing underlying causes.

Non-goals:
- Do not add file backend implementations, signed URLs, metadata managers, path helper policy, or HTTP behavior.
- Do not change the `Storage` interface or shared file value types.

Files:
- store/file/errors.go
- store/file/coverage_test.go

Tests:
- go test -timeout 20s ./store/file
- go test -race -timeout 60s ./store/file
- go vet ./store/file

Docs Sync:
- Not required; this is internal error-shape hardening.

Done Definition:
- File operation errors are stable for nil and partially populated values.
- Existing `errors.Is` behavior still works.
- Store/file remains a pure contract/type/error package.

Outcome:

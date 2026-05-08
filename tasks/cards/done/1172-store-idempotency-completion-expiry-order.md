# Card 1172

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- docs/modules/store/README.md

Goal:
Ensure expired idempotency records are classified as expired before duplicate completion.

Scope:
- Update `ValidateCompletionAt` classification order.
- Add a test for completed-but-expired records.
- Keep request-hash mismatch semantics unchanged.
- Sync store docs if wording needs clarification.

Non-goals:
- Do not change the `Store` interface.
- Do not implement concrete providers.
- Do not add HTTP idempotency behavior.

Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/idempotency
- go vet ./store/idempotency
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required if completion classification wording changes.

Done Definition:
- Expired records return `ErrExpired` even when status is completed.
- Duplicate completed records that are not expired still return `ErrAlreadyCompleted`.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- `ValidateCompletionAt` now classifies same-hash expired records as `ErrExpired` before duplicate completion replay checks.
- Non-expired completed records still return `ErrAlreadyCompleted`.
- Store docs now state that expired records are rejected before replay or duplicate-completion decisions.

Validation:
- `go test -timeout 20s ./store/idempotency`
- `go vet ./store/idempotency`
- `go run ./internal/checks/dependency-rules`

# Card 0752

Milestone:
Recipe: specs/change-recipes/feature.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- docs/modules/store/README.md

Goal:
Make request-hash mismatch and duplicate completion semantics explicit without breaking the existing `Store` interface.

Scope:
- Add stable sentinel errors for request-hash mismatch and already-completed records.
- Add an optional extended interface for hash-aware completion.
- Add helper validation for completing a record against an expected request hash.
- Sync store docs.

Non-goals:
- Do not change the existing `Store` interface method signatures.
- Do not implement concrete SQL/KV providers in this card.
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
- Required for stable contract semantics.

Done Definition:
- Stable store/idempotency exposes request-hash mismatch and duplicate completion contract helpers.
- Existing `Store` interface remains source-compatible.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Added `ErrRequestMismatch` and `ErrAlreadyCompleted`.
- Added optional `HashAwareStore` completion extension without changing `Store`.
- Added `ValidateCompletion` and `ValidateCompletionAt` helpers for mismatch, duplicate, expired, and invalid-record classification.
- Synced store docs.

Validation:
- `go test -timeout 20s ./store/idempotency`
- `go vet ./store/idempotency`
- `go run ./internal/checks/dependency-rules`

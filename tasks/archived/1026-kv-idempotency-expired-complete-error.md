# Card 1026

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/kv.go
- x/data/idempotency/kv_test.go
Depends On:

Goal:
Align KV idempotency Complete expired-record behavior with the stable missing-record contract.

Scope:
- Return ErrNotFound when Complete observes an expired record.
- Preserve cleanup of expired KV state.
- Add regression coverage for the race-like post-Get expiry branch.

Non-goals:
- Do not change PutIfAbsent or Get semantics.
- Do not change stable store/idempotency interfaces.

Files:
- x/data/idempotency/kv.go
- x/data/idempotency/kv_test.go

Tests:
- go test -timeout 20s ./x/data/idempotency
- go test -timeout 20s ./store/idempotency

Docs Sync:
- None unless idempotency provider docs require explicit error mapping.

Done Definition:
- KV and SQL providers both treat expired Complete as not found.
- Targeted tests pass.

Outcome:
KV Complete now returns ErrNotFound instead of ErrExpired when a record expires between the initial read and completion write. Added a controlled-clock regression that exercises the post-Get expiry branch and verifies cleanup.

Validation:
- go test -timeout 20s ./x/data/idempotency
- go test -timeout 20s ./store/idempotency

# Card 1169

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/kv.go
- x/data/idempotency/kv_test.go
Depends On:

Goal:
Make the KV-backed idempotency PutIfAbsent claim atomic for concurrent callers.

Scope:
- Prevent two concurrent PutIfAbsent calls for the same key from both returning inserted=true.
- Preserve expired-record reclamation and existing ErrInvalidKey/ErrExpired semantics.
- Add a focused concurrency regression test.

Non-goals:
- Do not change the stable store/kv public API.
- Do not add external dependencies.

Files:
- x/data/idempotency/kv.go
- x/data/idempotency/kv_test.go

Tests:
- go test ./x/data/idempotency

Docs Sync:
- Not required unless public behavior or docs change.

Done Definition:
- Concurrent PutIfAbsent for the same key produces exactly one successful claim.
- Existing KV idempotency tests still pass.

Outcome:
- Added a KVStore-level mutex around PutIfAbsent so the existing-record check
  and Set execute as one adapter-local critical section.
- Added TestKVStorePutIfAbsentConcurrentClaim to require exactly one successful
  claim across concurrent callers.
- Validated with:
  - go test -timeout 20s ./x/data/idempotency
  - go test -race -timeout 60s ./x/data/idempotency
  - go vet ./x/data/idempotency

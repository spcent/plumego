# Card 0764

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files: x/data/idempotency/kv.go, x/data/idempotency/kv_test.go
Depends On:

Goal:

Make KV idempotency Complete/Delete/PutIfAbsent serialize same-key state transitions inside the adapter so Complete cannot resurrect a record that Delete removed between Get and Set.

Scope:

- Hold the KVStore mutation lock across Complete read/validate/update/write.
- Hold the same lock across Delete.
- Preserve current public API and backend error mapping.
- Add focused regression coverage for mutation serialization.

Non-goals:

- Adding a stable CAS API to store/kv.
- Solving cross-wrapper concurrency for separate KVStore instances sharing the same underlying store.
- Changing SQL idempotency semantics.

Files:

- x/data/idempotency/kv.go
- x/data/idempotency/kv_test.go

Tests:

- go test -race -timeout 60s ./x/data/idempotency
- go test -timeout 20s ./x/data/idempotency
- go vet ./x/data/idempotency

Docs Sync:

- Not required; this preserves the existing idempotency contract.

Done Definition:

- Complete and Delete share the same adapter mutation lock as PutIfAbsent.
- A regression test would fail if Complete/Delete bypass the lock.
- Module tests and vet pass.

Outcome:

- Complete and Delete now hold KVStore's mutation lock across their full read/write sequences.
- Added regression coverage that fails if Complete or Delete bypasses the mutation lock.
- Validation passed:
  - go test -race -timeout 60s ./x/data/idempotency
  - go test -timeout 20s ./x/data/idempotency
  - go vet ./x/data/idempotency

# Card 0771

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/cache
Owned Files: x/cache/redis/redis.go, x/cache/redis/redis_test.go, docs/modules/x-cache/README.md
Depends On:

Goal:

Make Redis optional capability discovery explicit so the base Adapter does not advertise counter/appender support unless the wrapped client supports those operations.

Scope:

- Move Incr/Decr/Append optional methods off the base Adapter.
- Add explicit Redis adapter wrappers or constructors for counter/appender-capable clients.
- Update internal tests and docs for the new discovery shape.
- Run the symbol-change protocol for removed or moved exported methods.

Non-goals:

- Changing store/cache optional interfaces.
- Adding Redis client implementations.
- Changing distributed cache replication semantics.

Files:

- x/cache/redis/redis.go
- x/cache/redis/redis_test.go
- docs/modules/x-cache/README.md

Tests:

- rg -n --glob '*.go' 'Incr|Append|NewAdapter|CounterCache|AppenderCache' .
- go build ./...
- go test -race -timeout 60s ./x/cache/...
- go test -timeout 20s ./x/cache/...
- go vet ./x/cache/...

Docs Sync:

- Update x/cache docs if constructor or capability discovery semantics change.

Done Definition:

- Base Redis Adapter no longer satisfies optional counter/appender interfaces by method set alone.
- Capability-aware wrappers or constructors expose optional support only when valid.
- Symbol-change search/build/tests pass.

Outcome:

- Base redis.Adapter now exposes only store/cache.Cache methods.
- Added CounterAdapter, AppenderAdapter, and AtomicAdapter constructors that require matching Redis client capabilities.
- Updated Redis tests and x/cache docs for explicit capability discovery.
- Validation passed:
  - rg -n --glob '*.go' 'NewAdapter|\.Incr\(|\.Decr\(|\.Append\(|CounterCache|AppenderCache' .
  - go build ./...
  - go test -race -timeout 60s ./x/cache/...
  - go test -timeout 20s ./x/cache/...
  - go vet ./x/cache/...

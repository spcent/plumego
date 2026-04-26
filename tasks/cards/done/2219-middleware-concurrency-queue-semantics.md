# Card 2219

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/concurrencylimit/concurrency_limit.go; middleware/concurrencylimit/concurrency_limit_test.go; tasks/cards/done/2219-middleware-concurrency-queue-semantics.md
Depends On:

Goal:
Align concurrency limit queue behavior with its documented meaning: `queueDepth` is the number of waiting requests, not active plus waiting requests.

Scope:
- Preserve `maxConcurrent` as the worker semaphore limit.
- Interpret positive `queueDepth` as allowed waiting requests in addition to active requests.
- Keep `queueDepth == 0` as fail-fast once workers are full.
- Add focused tests for one queued waiter, queue overflow, and queue timeout.

Non-goals:
- Do not introduce a new config type or rename the existing constructor.
- Do not add priority queues, fairness policy, or tenant-aware limits.
- Do not change unrelated middleware.

Files:
- `middleware/concurrencylimit/concurrency_limit.go`
- `middleware/concurrencylimit/concurrency_limit_test.go`

Tests:
- `go test -timeout 20s ./middleware/concurrencylimit`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this corrects existing documented semantics.

Done Definition:
- Tests prove `queueDepth=1` allows one waiting request behind one active request.
- Queue overflow and queue timeout still return canonical 503 errors.
- Targeted middleware tests and vet pass.

Outcome:
- Changed the queue gate capacity to `maxConcurrent + queueDepth` so active workers do not consume all waiting capacity.
- Preserved `queueDepth=0` as fail-fast while workers are full.
- Added tests for queued success, fail-fast, and queued timeout.
- Validation run: `go test -timeout 20s ./middleware/concurrencylimit`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.

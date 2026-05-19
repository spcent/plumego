# Card 1435

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: x/data
Owned Files:
- `x/data/sharding/logging_test.go`
Depends On:
- 1430 gate failure

Goal:
- Fix the release-gate false positive in sharding log redaction coverage.

Scope:
- Keep `LoggingRouter.LogShardResolution` behavior unchanged.
- Replace the broad raw-output substring assertion with JSON field/value
  assertions so timestamps or logger metadata cannot trip the shard-key leak
  check.

Non-goals:
- Do not change sharding runtime behavior.
- Do not widen x/data public APIs.
- Do not alter logging shape beyond test assertions.

Files:
- `x/data/sharding/logging_test.go`

Tests:
- `go test -race -timeout 60s ./x/data/sharding -run TestLoggingRouter_LogShardResolution -count=1`
- `go test -race -timeout 60s ./x/data/sharding`
- `go test -timeout 20s ./x/data/sharding`

Docs Sync:
- Not required. This is a test hardening fix for existing documented redaction
  behavior.

Done Definition:
- The focused race test passes.
- The full `x/data/sharding` test package passes with and without race.
- The release gate can resume at card 1430.

Outcome:
- Added JSON field/value assertions for shard-key redaction instead of scanning
  the entire log line for `123`.

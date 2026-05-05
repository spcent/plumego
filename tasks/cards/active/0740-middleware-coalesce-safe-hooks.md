# Card 0740

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md
Depends On:
- 0739-middleware-coalesce-waiter-cancel

Goal:
Ensure coalesce instrumentation hooks cannot prevent canonical responses or
surface post-response panics.

Scope:
- Wrap `OnError` and `OnCoalesced` invocations with internal recover guards.
- Preserve existing hook timing and arguments.
- Add regression tests for hook panics on error and successful coalesced replay.

Non-goals:
- Do not make hooks asynchronous.
- Do not change hook count semantics.
- Do not add hook error reporting.

Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/coalesce
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Hook panics are recovered internally.
- Upstream error responses are still written when `OnError` panics.
- Successful coalesced replay does not panic when `OnCoalesced` panics.

Outcome:


# Card 0715

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P0
State: active
Primary Module: middleware
Owned Files:
- middleware/timeout/timeout.go
- middleware/timeout/timeout_test.go
- docs/modules/middleware/README.md
Depends On:
- 0714-middleware-coalesce-stable-contract

Goal:
Make timeout middleware semantics unambiguous for stable users.

Scope:
- Correct comments that imply oversized responses bypass buffering.
- Clarify naming and docs around the replay threshold.
- Add tests proving timeout does not stop downstream work that ignores context.
- Add tests for buffer overflow and post-timeout write behavior.

Non-goals:
- Do not introduce goroutine cancellation beyond request context cancellation.
- Do not add new dependencies.
- Do not change public type names unless unavoidable.

Files:
- middleware/timeout/timeout.go
- middleware/timeout/timeout_test.go
- docs/modules/middleware/README.md

Tests:
- go test ./middleware/timeout
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Timeout comments and docs match implementation behavior.
- Users can tell the middleware buffers and replays rather than streams.
- Ignored-context and overflow behavior are covered by tests.

Outcome:


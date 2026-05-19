# Card 0795

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files:
- middleware/timeout/timeout.go
- middleware/timeout/timeout_test.go
Depends On:

Goal:
Make timeout middleware panic-safe when downstream handlers run in the timeout worker goroutine.

Scope:
- Capture panics from the downstream timeout worker goroutine.
- Report worker completion, panic, and normal return back to the parent goroutine.
- Re-panic on the parent middleware call stack when the worker panics before timeout wins, so outer recovery can apply the canonical recovery path.
- Add regression tests for panic before write and panic after partial write.

Non-goals:
- Do not make timeout recover panics itself.
- Do not change timeout response shape.
- Do not change public API.

Files:
- middleware/timeout/timeout.go
- middleware/timeout/timeout_test.go

Tests:
- go test ./middleware/timeout
- go test ./middleware/...

Docs Sync:
- None unless behavior wording needs clarification.

Done Definition:
- Downstream panics in the timeout worker do not escape as unhandled goroutine panics.
- Outer recovery can observe the re-panicked value when the panic wins before timeout.
- Timeout package and middleware-wide tests pass.

Outcome:
- Added a buffered timeout worker result channel that captures worker panics.
- Re-panics worker panics on the parent middleware call stack when the worker completes before timeout.
- Added regression tests for direct re-panic, outer recovery handling, and panic after buffered partial writes.
- Validation:
  - `go test ./middleware/timeout`
  - `go test ./middleware/...`

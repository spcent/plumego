# Card 0749

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/debug/debug_errors.go
- middleware/debug/debug_errors_test.go
Depends On:
- 0748-middleware-cors-strict-wildcard

Goal:
Prevent debug error replacement from preserving stale entity headers that no
longer describe the replacement JSON response.

Scope:
- Remove stale `Content-Length` before writing a replacement structured debug
  error.
- Add a regression test where downstream plain-text error sets a mismatched
  `Content-Length`.

Non-goals:
- Do not change debug skip rules.
- Do not change contract.WriteError behavior globally.
- Do not make debug production-facing.

Files:
- middleware/debug/debug_errors.go
- middleware/debug/debug_errors_test.go

Tests:
- go test -timeout 20s ./middleware/debug
- go test -timeout 20s ./middleware/...

Docs Sync:
- none unless behavior wording needs clarification

Done Definition:
- Replacement JSON responses do not retain stale Content-Length.
- Existing debug capture and passthrough tests pass.
- Middleware-wide tests pass.

Outcome:
- Removed stale `Content-Length` before writing replacement structured debug
  JSON responses.
- Added a regression test for plain-text error replacement with mismatched
  downstream `Content-Length`.

Validation:
- go test -timeout 20s ./middleware/debug
- go test -timeout 20s ./middleware/...

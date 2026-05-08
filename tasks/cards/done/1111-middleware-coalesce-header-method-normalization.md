# Card 1111

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md
Depends On:
- 0749-middleware-debug-content-length

Goal:
Make coalesce replay and option normalization consistent with stable response
contracts.

Scope:
- Freeze replay headers at first `WriteHeader` so waiters cannot see leader
  post-commit header mutations.
- Normalize configured methods with trim + uppercase + blank filtering.
- Add regression tests for header freeze and lowercase method configuration.
- Update coalesce docs.

Non-goals:
- Do not change default key hashing.
- Do not change high-risk/GA coalesce maturity classification.
- Do not add streaming support.

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
- Waiter replay headers match the leader committed header state.
- Method config accepts common lowercase/whitespace input.
- Middleware-wide tests pass.

Outcome:
- Normalized configured coalesce methods with trim, uppercase, and blank entry
  filtering.
- Froze captured replay headers when the leader first calls `WriteHeader`, so
  waiters do not see post-commit leader header mutations.
- Added regression tests for lowercase method config and committed header
  replay.
- Updated coalesce contract docs.

Validation:
- go test -timeout 20s ./middleware/coalesce
- go test -timeout 20s ./middleware/...

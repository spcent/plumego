# Card 0661

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/nethttp
Owned Files: internal/nethttp/client.go, internal/nethttp/metrics.go, internal/nethttp/client_extended_test.go
Depends On: 0660

Goal:
Make optional nethttp components tolerate nil entries without panicking.

Scope:
- Skip nil retry policies inside `CompositeRetryPolicy`.
- Make `MetricsMiddleware(nil)` a transparent pass-through middleware.
- Add focused tests for both behaviors.

Non-goals:
- Do not change retry behavior for non-nil policies.
- Do not change metrics collection semantics when a recorder is present.

Files:
- internal/nethttp/client.go
- internal/nethttp/metrics.go
- internal/nethttp/client_extended_test.go

Tests:
- go test -timeout 20s ./internal/nethttp
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; defensive internal behavior only.

Done Definition:
- Nil composite retry policies and nil metrics recorders do not panic.
- Focused and internal package validation pass.

Outcome:
Completed. Composite retry policies now skip nil children, and `MetricsMiddleware(nil)` is a transparent pass-through.

Validation:
- go test -timeout 20s ./internal/nethttp
- go test -timeout 20s ./internal/...
- go vet ./internal/...

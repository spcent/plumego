# Card 0667

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/nethttp
Owned Files: internal/nethttp/client.go, internal/nethttp/httpclient_test.go
Depends On: 0666

Goal:
Remove brittle timeout detection based on matching the word "timeout" in error strings.

Scope:
- Treat `context.DeadlineExceeded` and `net.Error` timeout errors as retryable timeouts.
- Ensure arbitrary errors whose text contains "timeout" are not classified as timeout errors.
- Preserve existing retry behavior for real timeout errors.

Non-goals:
- Do not redesign retry policies.
- Do not change status-code retry behavior or backoff behavior.

Files:
- internal/nethttp/client.go
- internal/nethttp/httpclient_test.go

Tests:
- go test -timeout 20s ./internal/nethttp
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal retry classification only.

Done Definition:
- Timeout retry classification is type/error-chain based rather than substring based.
- Tests cover false-positive string errors and true timeout errors.

Outcome:


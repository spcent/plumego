# Card 0640

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: internal/nethttp
Owned Files: internal/nethttp/client.go, internal/nethttp/httpclient_test.go
Depends On: tasks/cards/done/0639-internal-httputil-sanitizer-protocol-case.md

Goal:
Make retry backoff respect request context cancellation.

Scope:
- Replace unconditional retry `time.Sleep` with context-aware waiting.
- Return the request context error when cancellation happens during backoff.
- Add focused coverage that cancels after the first retryable response.

Non-goals:
- Do not change retry policy interfaces.
- Do not change jitter calculation.
- Do not change high-level response body ownership semantics.

Files:
- internal/nethttp/client.go
- internal/nethttp/httpclient_test.go

Tests:
- go test ./internal/nethttp
- go test ./internal/...

Docs Sync:
- None; this aligns implementation with context expectations.

Done Definition:
- Canceled requests do not wait for the full retry backoff.
- Retry response bodies are still closed before returning.

Outcome:
- Added context-aware retry backoff waiting.
- Added cancellation coverage for a retryable response before the next attempt.
- Validation: `go test ./internal/nethttp`; `go test ./internal/...`.

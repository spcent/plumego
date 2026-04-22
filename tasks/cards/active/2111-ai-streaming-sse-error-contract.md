# Card 2111: AI Streaming SSE Error Contract

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: high
State: active
Primary Module: x/ai
Owned Files:
- x/ai/sse/sse.go
- x/ai/sse/sse_test.go
- x/ai/streaming/handler.go
- x/ai/streaming/handler_test.go
- docs/modules/x-ai/README.md
Depends On: none

Goal:
- Align AI streaming and SSE HTTP errors with the canonical structured error contract.
- Avoid leaking raw stream-creation, invalid-request, or provider errors through HTTP error messages and SSE `error` events by default.

Scope:
- Audit `x/ai/sse` and `x/ai/streaming` HTTP handlers for raw `err.Error()` and type-only errors.
- Add stable codes for method, invalid JSON/request, and stream creation failures.
- Decide a safe default SSE error event payload and test it.
- Preserve streaming APIs, provider interfaces, and explicit handler composition.

Non-goals:
- Do not change provider, session, tool, or cache public APIs.
- Do not introduce WebSocket, webhook, tenant, or core bootstrap behavior into `x/ai`.
- Do not change successful stream chunk format.

Files:
- `x/ai/sse/sse.go`: normalize SSE handler HTTP and event error paths.
- `x/ai/sse/sse_test.go`: assert safe SSE failure payloads.
- `x/ai/streaming/handler.go`: normalize streaming handler error codes and messages.
- `x/ai/streaming/handler_test.go`: cover method, invalid JSON, and stream creation failures.
- `docs/modules/x-ai/README.md`: document streaming/SSE error behavior if behavior changes.

Tests:
- `go test -race -timeout 60s ./x/ai/sse ./x/ai/streaming`
- `go test -timeout 20s ./x/ai/sse ./x/ai/streaming`
- `go vet ./x/ai/sse ./x/ai/streaming`

Docs Sync:
- Required if HTTP error codes, public messages, or SSE error-event payloads change.

Done Definition:
- Default AI streaming/SSE handlers do not expose raw internal error strings to clients.
- All handler error responses use explicit stable codes.
- Focused tests prove safe behavior for invalid request and stream creation failures.
- The three listed validation commands pass.

Outcome:


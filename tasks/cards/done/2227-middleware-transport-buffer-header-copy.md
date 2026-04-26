# Card 2227

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files: middleware/internal/transport/response_buffer.go; middleware/internal/transport/transport_test.go; tasks/cards/done/2227-middleware-transport-buffer-header-copy.md
Depends On: 2226

Goal:
Make buffered response replay deterministic by replacing buffered headers instead of appending duplicate values onto an already-populated destination.

Scope:
- Copy buffered headers to the destination using replacement semantics.
- Preserve multi-value headers from the buffered response.
- Keep `EnsureNoSniff` behavior.
- Add tests for replacing existing destination headers and preserving multi-value buffered headers.

Non-goals:
- Do not change the public contract package.
- Do not refactor all response recorder helpers.
- Do not alter gzip/coalesce behavior beyond using the fixed transport primitive.

Files:
- `middleware/internal/transport/response_buffer.go`
- `middleware/internal/transport/transport_test.go`

Tests:
- `go test -timeout 20s ./middleware/internal/transport`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; internal helper semantics become less surprising.

Done Definition:
- Buffered replay replaces stale destination header values.
- Buffered multi-value headers are preserved.
- Targeted middleware tests and vet pass.

Outcome:
- Replaced append-style header copy in `BufferedResponse.WriteTo` with replacement semantics.
- Preserved buffered multi-value headers via cloned slices.
- Added tests for stale destination header replacement and multi-value header preservation.
- Validation run: `go test -timeout 20s ./middleware/internal/transport`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.

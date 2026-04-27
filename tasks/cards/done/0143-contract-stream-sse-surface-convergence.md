# Card 0143

Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/context_stream.go`
- `contract/context_stream_test.go`
- `contract/context_test.go`
- `docs/modules/contract/README.md`

Goal:
- Keep one streaming/SSE API in `contract` by removing the older helper family
  that duplicates `Ctx.Stream(...)`.

Problem:
- `Ctx.Stream(StreamConfig)` is already the general streaming entrypoint and
  internally covers JSON, text, SSE, generators, channels, and readers.
- `RespondWithSSE`, `IsSSESupported`, `SetSSEHeaders`, and deprecated `WriteSSE`
  keep a second public SSE helper family alive alongside the generalized stream
  API.
- Repo-internal usage already routes through `Stream` internals rather than the
  old public helper family, so the extra surface mainly adds confusion.

Scope:
- Decide the single public SSE/streaming contract for `contract`.
- Remove duplicate public SSE helpers that are no longer needed.
- Keep low-level `NewSSEWriter` only if it still serves a clear stdlib-shaped
  purpose alongside the chosen canonical path.

Non-goals:
- Do not redesign SSE wire format or retry behavior in this card.
- Do not add protocol-specific streaming helpers beyond the chosen canonical
  surface.

Files:
- `contract/context_stream.go`
- `contract/context_stream_test.go`
- `contract/context_test.go`
- `docs/modules/contract/README.md`

Tests:
- Add or update tests to cover the chosen canonical SSE/streaming path after the
  duplicate public helpers are removed.
- `go test -race -timeout 60s ./contract/...`

Docs Sync:
- Update contract docs so streaming/SSE is documented through one API surface.

Done Definition:
- `contract` exposes one canonical streaming/SSE API.
- Duplicate public SSE helpers are removed or explicitly quarantined.
- Contract docs describe one preferred streaming entrypoint.

Outcome:
- `Ctx.Stream(StreamConfig{...})` is now the only high-level streaming/SSE API in `contract`.
- Removed duplicate public SSE helpers: `RespondWithSSE`, `IsSSESupported`, `SetSSEHeaders`, and `WriteSSE`.
- Kept `NewSSEWriter(...)` as the low-level stdlib-shaped SSE writer.
- Contract docs now describe one canonical stream/SSE entrypoint.

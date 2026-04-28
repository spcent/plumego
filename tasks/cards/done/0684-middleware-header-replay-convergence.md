# Card 0684

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/internal/transport/http.go
- middleware/internal/transport/transport_test.go
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md
Depends On: 0683

Goal:
Converge buffered/captured response header replay so middleware replaces stale destination values while preserving multi-value response headers.

Scope:
- Add an internal transport header-copy helper with replacement semantics.
- Use it from gzip header flush and coalesced response replay.
- Preserve multi-value headers such as `Set-Cookie`.
- Keep `X-Coalesced` and nosniff behavior intact.
- Sync middleware module docs for the internal `AddVary` and header-copy helper.

Non-goals:
- Do not change compression eligibility or coalescing request matching.
- Do not expose a new public middleware API.
- Do not refactor unrelated debug or timeout behavior.

Files:
- `middleware/internal/transport/http.go`
- `middleware/internal/transport/transport_test.go`
- `middleware/compression/gzip.go`
- `middleware/compression/gzip_test.go`
- `middleware/coalesce/coalesce.go`
- `middleware/coalesce/coalesce_test.go`
- `docs/modules/middleware/README.md`

Tests:
- `go test -timeout 20s ./middleware/internal/transport ./middleware/compression ./middleware/coalesce`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- Required; internal helper list changes.

Done Definition:
- Gzip replay does not append duplicate destination header values.
- Coalesced replay replaces stale destination header values and preserves multi-value captured headers.
- Targeted middleware tests and vet pass.

Outcome:
- Added `internal/transport.CopyHeaders` with replacement semantics and cloned multi-value header slices.
- Updated gzip header flush and coalesced response replay to replace stale destination values instead of appending duplicates.
- Preserved `Set-Cookie` multi-value replay and `X-Coalesced` response marking.
- Synced the middleware module primer with `AddVary` and `CopyHeaders`.
- Validation run: `go test -timeout 20s ./middleware/internal/transport ./middleware/compression ./middleware/coalesce`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.

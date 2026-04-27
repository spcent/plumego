# Card 0204

Priority: P1
State: done
Primary Module: middleware
Owned Files:
- `middleware/accesslog/accesslog.go`
- `middleware/tracing/tracing.go`
- `middleware/httpmetrics/http_metrics.go`
- `middleware/internal/observability/helpers.go`
- `middleware/internal/transport/http.go`
Depends On:
- `0203-observability-id-contract-convergence.md`

Goal:
- Collapse the duplicated request-observability setup in stable middleware into one shared preparation path with one recorder/helper implementation.
- Collapse it all the way to one canonical implementation, not a pair of equivalent helpers with different call sites.

Problem:
- `accesslog`, `tracing`, and `httpmetrics` each repeat parts of the same work: ensure request ID, set response headers, wrap a recorder, derive route/path labels, and finalize request stats.
- `middleware/internal/observability` and `middleware/internal/transport` both define transport helpers such as `ClientIP` and `ResponseRecorder`, but with different capabilities and state models.
- This duplication makes small behavior fixes expensive and increases drift risk in status capture, byte counting, header propagation, and span finalization.

Scope:
- Introduce one shared middleware-internal request preparation path for stable observability middleware.
- Reduce recorder/helper duplication so access logging, tracing, and HTTP metrics all read from the same transport snapshot rules.
- Keep the public middleware packages narrow; shared logic should stay internal and transport-only.
- Remove the losing internal helper path in the same change rather than leaving parallel implementations.

Non-goals:
- Do not add new observability exports or adapters.
- Do not merge access logging, tracing, and HTTP metrics into one public middleware package.
- Do not pull feature-level observability concerns into `core`.
- Do not keep duplicate recorder or transport helper implementations for compatibility.

Files:
- `middleware/accesslog/accesslog.go`
- `middleware/tracing/tracing.go`
- `middleware/httpmetrics/http_metrics.go`
- `middleware/internal/observability/helpers.go`
- `middleware/internal/transport/http.go`

Tests:
- `go test -timeout 20s ./middleware/...`
- `go test -race -timeout 60s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- Update the middleware primer only if the canonical composition order or shared helper expectations become materially different.

Done Definition:
- Stable observability middleware uses one shared request preparation and recorder path.
- There is no second internal `ResponseRecorder` or `ClientIP` implementation serving the same stable middleware pipeline.
- Access log, tracing, and HTTP metrics tests all assert the same header/status/byte semantics.
- The removed helper path has zero remaining references after the change.

Outcome:
- Completed.
- Converged stable request observability middleware on shared preparation helpers under `middleware/internal/observability`, including request-id preparation, response recording, and request metric derivation.
- Removed duplicate request-observability plumbing paths so access logging, tracing, and HTTP metrics use one canonical transport preparation flow.

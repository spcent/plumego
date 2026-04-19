# Card 0016

Priority: P2
State: active
Primary Module: x/gateway
Owned Files:
  - x/gateway/transform/transform.go
  - x/websocket/security.go

Depends On: 0007 (safe to reference once internal/nethttp package name is fixed)

Goal:
Two extension packages have accumulated small but concrete technical debt, grouped into one card for focused resolution:

**Problem 1: x/gateway/transform private responseRecorder duplicates existing utility**

`x/gateway/transform/transform.go:144-169` defines a private `responseRecorder` struct with
`Header()`, `WriteHeader()`, and `Write()` methods to capture the response body.

The same module `x/gateway` already uses `nethttp.NewResponseRecorder(w)` (i.e.,
`internal/nethttp.ResponseRecorder`) in `cache/http_cache.go:227` for the same purpose.

Two parallel paths exist within the same module, and `internal/nethttp.ResponseRecorder` is
more complete (includes write-through, EnsureNoSniff security header, etc.).
transform.go should reuse the existing utility.

**STATUS: CANNOT BE DONE — write-through vs. capture-only semantics mismatch**

After verification, this change cannot be made: `internal/nethttp.ResponseRecorder` uses
write-through semantics — it writes to both the capture buffer AND the underlying
ResponseWriter simultaneously. The transform middleware requires capture-only semantics: it
intercepts the handler's full response, applies transformations, then writes the MODIFIED
response to the actual writer. Using write-through would cause the untransformed response to
reach the client before the transform code runs. The private `responseRecorder` in
transform.go is intentionally kept as-is.

**Problem 2: x/websocket SecurityMetrics retains deprecated fields**

The `SecurityMetrics` struct in `x/websocket/security.go` contains 3 fields whose comments
explicitly state "no longer tracked here":
- `InvalidWebSocketKeys` (recommended: use `Hub.Metrics().InvalidWSKeys`)
- `BroadcastQueueFull` (recommended: use `Hub.Metrics().BroadcastDropped`)
- `RejectedConnections` (recommended: use `Hub.Metrics().SecurityRejections`)

These 3 fields are always zero-valued and mislead callers into thinking they are active.
Per the style guide "Deprecated symbols must be removed in the same PR", they should be deleted.

**STATUS: DONE.** `InvalidWebSocketKeys`, `BroadcastQueueFull`, and `RejectedConnections`
fields have been removed from `x/websocket/security.go`.

Scope:
- **transform.go**: No change — the private `responseRecorder` is intentionally retained
  because `internal/nethttp.ResponseRecorder` uses write-through semantics, which would send
  the untransformed response to the client before transformations are applied. Capture-only
  semantics are required here. This is documented and the divergence is intentional.
- **security.go**: DONE. Removed `InvalidWebSocketKeys`, `BroadcastQueueFull`, and
  `RejectedConnections` fields. Confirmed no external reads or assignments existed
  (fields were only defined in security.go).

Non-goals:
- Do not change the public API of the transform package (exported function signatures unchanged)
- Do not modify the remaining fields of SecurityMetrics (InvalidJWTSecrets etc. remain valid)
- Do not modify the return type of Hub.Metrics()

Files:
  - x/gateway/transform/transform.go (no change — responseRecorder intentionally kept)
  - x/websocket/security.go (3 deprecated fields removed — DONE)
  - x/websocket/security_test.go (field assertions updated if present — DONE)

Tests:
  - go build ./x/gateway/...
  - go test ./x/gateway/...
  - go build ./x/websocket/...
  - go test ./x/websocket/...

Docs Sync: —

Done Definition:
- `grep -n "InvalidWebSocketKeys\|BroadcastQueueFull\|RejectedConnections" x/websocket/security.go` returns empty (DONE)
- Problem 1 (transform.go responseRecorder): intentionally kept — write-through vs. capture-only semantics mismatch prevents using internal/nethttp.ResponseRecorder. No grep requirement.
- All modified packages compile and tests pass

Outcome:
Problem 2 (websocket deprecated fields) is complete. Problem 1 (transform.go) is closed as
intentionally not done: the private responseRecorder cannot be replaced by
internal/nethttp.ResponseRecorder due to incompatible write semantics. The divergence is
documented above and in this card for future reference.

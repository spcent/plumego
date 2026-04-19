# Card 0009

Priority: P2
State: done
Primary Module: x/ops
Owned Files:
  - x/ops/healthhttp/handlers.go
  - x/ops/healthhttp/runtime.go
  - reference/standard-service/internal/handler/health.go

Depends On: —

Goal:
There are three intertwined sources of confusion within the `x/ops` subdomain:

1. **WriteResponse vs WriteJSON split**: `x/ops/ops.go` uses `contract.WriteResponse` (with a
   `{data, meta, request_id}` envelope), while all handlers in `x/ops/healthhttp/` consistently
   use `contract.WriteJSON` (no envelope). Both approaches coexist within the same `x/ops`
   package with no comments explaining the rationale. Bypassing the envelope for health check
   endpoints is reasonable (Kubernetes probes do not expect an envelope), but this needs to be
   made explicit and consistently applied.

2. **HealthHandler / DebugHealthHandler naming confusion**:
   - `HealthHandler(manager, debug bool)` → when debug=false delegates to `DetailedHandler`;
     when debug=true additionally attaches RuntimeInfo.
   - `DebugHealthHandler(manager, debug bool)` → when debug=false returns 404 directly;
     when debug=true returns a diagnostics map.
   The `debug bool` parameter has completely different semantics in the two functions
   ("add fields" vs "control whether to expose"), yet they share similar names — very easy to confuse.

3. **reference/standard-service hand-rolls its own HealthHandler**: The reference implementation
   `reference/standard-service/internal/handler/health.go` does not use `x/ops/healthhttp` —
   it hand-writes simple live/ready handlers using the `WriteResponse` envelope format
   (opposite of healthhttp's WriteJSON), making it unsuitable as a canonical example.

Scope:
- **Document the WriteResponse/WriteJSON layering**: Add a one-line comment at the top of
  `x/ops/ops.go` and in the `x/ops/healthhttp` package doc explaining:
  - `x/ops/healthhttp`: health endpoints bypass the envelope, use `WriteJSON` (probe-compatible format)
  - `x/ops/ops.go`: ops API endpoints follow the standard envelope, use `WriteResponse`
- **Rename DebugHealthHandler**: Rename `DebugHealthHandler(manager, debug bool)` to
  `DiagnosticsHandler(manager Manager, enabled bool)`, where the parameter name `enabled`
  more accurately expresses the intent; also rename the `debug` parameter in
  `HealthHandler(manager, debug bool)` to `includeRuntime bool` to eliminate ambiguity
- **Update reference/standard-service**: Replace the hand-written `HealthHandler` in
  `reference/standard-service` with calls to `x/ops/healthhttp.LiveHandler()` and
  `x/ops/healthhttp.ReadinessHandler(manager)` (if a manager is already available),
  or explicitly note (via comment) that the reference implementation is intentionally minimal

Non-goals:
- Do not unify WriteJSON and WriteResponse into a single function
- Do not change the actual behavior or response body structure of any handler
- Do not introduce a health manager dependency into the reference (if too complex, just add a comment)
- Do not modify any logic related to Kubernetes probe compatibility

Files:
  - x/ops/healthhttp/handlers.go (comment + HealthHandler parameter rename)
  - x/ops/healthhttp/runtime.go (rename DebugHealthHandler to DiagnosticsHandler)
  - x/ops/ops.go (comment explaining WriteResponse usage)
  - reference/standard-service/internal/handler/health.go (update or add explanatory comment)
  - reference/standard-service/internal/app/routes.go (update if handler signature changes)

Tests:
  - go test ./x/ops/...
  - go build ./reference/standard-service/...
  - grep -rn "DebugHealthHandler" . --include="*.go" should return no results (or only post-rename references)

Docs Sync:
  - If x/ops/healthhttp module.yaml doc_paths includes API docs, sync those as well

Done Definition:
- `DebugHealthHandler` no longer exists; it is now `DiagnosticsHandler`
- The bool parameter in `HealthHandler` is named `includeRuntime`
- `x/ops/ops.go` and `x/ops/healthhttp` each have comments explaining their respective write strategy
- `go build ./...` and `go test ./x/ops/...` both pass
- The reference/standard-service health handler is aligned with healthhttp usage, or has a comment explaining the difference

Outcome:
- Renamed `DebugHealthHandler` to `DiagnosticsHandler`
- Renamed the `debug bool` parameter in `HealthHandler` to `includeRuntime bool`
- Updated all internal callers and tests to reflect the new names

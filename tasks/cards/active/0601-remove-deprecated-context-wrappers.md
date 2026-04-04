# Card 0601

Priority: P0

Goal:
- Remove the two deprecated context-accessor wrapper functions that remain in
  `contract` after cards 0503 and 0509 addressed the underlying key types.

Problem:
- `ContextWithPrincipal` (auth.go:48-52) is a one-line wrapper around
  `WithPrincipal` marked `// Deprecated: Use WithPrincipal instead.`
- `ContextWithTraceContext` (trace.go:47-51) is a one-line wrapper around
  `WithTraceContext` marked `// Deprecated: Use WithTraceContext instead.`
- Style guide §16.5 requires deprecated symbols to be removed in the same PR
  that replaces their last caller. Dead wrappers must not remain.
- Card 0503 removed `TraceIDKey` / `TraceContextKey` *types* but left the
  `ContextWithTraceContext` function wrapper behind.

Scope:
- Search the entire codebase for callers of `ContextWithPrincipal` and
  `ContextWithTraceContext`.
- Migrate each call site to the canonical `WithPrincipal` / `WithTraceContext`.
- Delete both deprecated functions.

Non-goals:
- Do not change any other function signatures or data in context.
- Do not change `WithPrincipal` or `WithTraceContext` implementations.

Files:
- `contract/auth.go`
- `contract/trace.go`
- Any file in `middleware/` or elsewhere calling the deprecated functions

Tests:
- `go build ./...`
- `go test ./contract/... ./middleware/...`
- `go vet ./...`

Done Definition:
- `ContextWithPrincipal` does not exist anywhere in the codebase.
- `ContextWithTraceContext` does not exist anywhere in the codebase.
- All callers use `WithPrincipal` / `WithTraceContext` directly.
- All tests pass.

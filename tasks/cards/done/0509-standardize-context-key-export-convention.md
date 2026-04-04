# Card 0509

Priority: P2

Goal:
- Establish a consistent convention for context-key types in the `contract`
  package: all package-internal keys unexported, all keys intended for external
  middleware use exported via accessor functions rather than direct type export.

Problem:
- Current keys mix exported and unexported types with no documented convention:
  ```go
  // Exported (accessible to all importers)
  type ParamsContextKey struct{}
  type TraceIDKey struct{}        ← legacy, targeted for removal in card 0503
  type TraceContextKey struct{}   ← legacy, targeted for removal in card 0503
  type RequestContextKey struct{}

  // Unexported (internal only)
  type principalContextKey struct{}
  type traceContextKey struct{}
  ```
- Some keys use a package-level variable (`principalContextKeyVar`) while
  others are used as zero-value struct literals — two different idioms with no
  stated reason for the difference.

Scope:
- Unexport all context key types (`ParamsContextKey`, `RequestContextKey`)
  after confirming no external package imports them directly.
- Provide exported accessor functions (`ParamsFromContext`,
  `RequestContextFrom`) that hide the key type from callers.
- Standardise the zero-value struct literal idiom (remove `*ContextKeyVar`
  package-level variables in favour of `type fooKey struct{}` with inline
  `ctx.Value(fooKey{})`).
- Add a comment block in `context_core.go` documenting the convention for
  future keys.

Note: complete card 0503 first, as it removes two of the exported keys.

Non-goals:
- Do not change the data stored in context, only the key mechanism.
- Do not touch auth `Principal` context handling beyond key style alignment.

Files:
- `contract/context_core.go`
- `contract/auth.go`
- Any middleware that currently imports `contract.ParamsContextKey` or
  `contract.RequestContextKey` directly

Tests:
- `go test ./contract/... ./middleware/...`
- `go build ./...`

Done Definition:
- No exported context-key *types* remain in `contract` (keys are unexported;
  values are accessed through exported functions).
- All context-key variables use the zero-value struct literal idiom.
- A convention comment is present in `context_core.go`.
- All tests pass.

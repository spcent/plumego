# Card 0073

Priority: P1

Goal:
- Move `Session`, `SessionStore`, `SessionValidator`, `RefreshManager`, and
  `SessionManager` interfaces out of `contract/auth.go` and into `x/tenant`,
  so `contract` retains only the transport-level auth primitives it should own.

Problem:
- `contract/auth.go` contains 194 lines of session lifecycle management:
  `Session` struct, `SessionStore`, `SessionValidator`, `RefreshManager`,
  `SessionManager` interfaces. These are domain-specific and violate the
  boundary rule in `CANONICAL_STYLE_GUIDE.md §16.1` and `AGENTS.md §6`.
- `contract` should export only the primitives needed at the HTTP transport
  layer: `Principal`, `Authenticator` interface, `WithPrincipal`,
  `PrincipalFromContext`.

Scope:
- Move to `x/tenant/session/` (or `x/security/session/`):
  - `Session` struct (lines ~43-138 in auth.go)
  - `SessionStore` interface
  - `SessionValidator` interface
  - `RefreshManager` interface
  - `SessionManager` interface
- Keep in `contract/auth.go` (stripped):
  - `Principal` struct
  - `Authenticator` interface (if defined)
  - `WithPrincipal(ctx, *Principal) context.Context`
  - `PrincipalFromContext(ctx) *Principal`
- Update callers of moved types to import from the new location.

Non-goals:
- Do not change how `Principal` is stored in or retrieved from context.
- Do not change JWT or token verification logic in `security/jwt`.

Files:
- `contract/auth.go` (strip to primitives)
- `x/tenant/session/` or `x/security/session/` (new package)
- Any file importing moved interfaces

Tests:
- `go test ./contract/...`
- `go test ./x/tenant/...`
- `go vet ./...`

Done Definition:
- `contract/auth.go` ≤ 60 lines, contains only `Principal` and context accessors.
- Session lifecycle interfaces live in `x/tenant` or `x/security`.
- All tests pass.

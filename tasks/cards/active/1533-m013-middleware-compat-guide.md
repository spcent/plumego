# Card 1533

Milestone: M-013
Recipe: specs/change-recipes/update-docs.yaml
Priority: P2
State: active
Primary Module: docs
Owned Files:
- `docs/migration/middleware-compat.md`
- `docs/ADOPTION_PATH.md`

Goal:
- Write docs/migration/middleware-compat.md explaining how to wrap third-party
  middleware written in gorilla/handlers, negroni, or alice styles into
  plumego's func(http.Handler)http.Handler type.
- Update docs/ADOPTION_PATH.md to reference the docs/migration/ directory.

Scope:
- Write docs/migration/middleware-compat.md with:
  - gorilla/handlers: most handlers already return http.Handler; wrap with
    func(next http.Handler) http.Handler literal.
  - negroni style (negroni.Handler with ServeHTTP(w, r, next)): show adapter
    function converting to func(http.Handler)http.Handler.
  - alice style (alice.Constructor = func(http.Handler)http.Handler): directly
    compatible; use as-is in middleware.NewChain.
  - net/http middleware pattern (func(http.Handler)http.Handler): directly
    compatible.
  - Warning: middleware that calls next multiple times violates plumego contract;
    show how to detect and fix.
  - Four complete, compilable code snippets for each wrapping pattern.
- Update docs/ADOPTION_PATH.md to add a "Migrating from another framework"
  section pointing to docs/migration/.

Non-goals:
- Do not introduce a compatibility layer in the codebase.
- Do not document framework-specific middleware packages (e.g., specific
  gorilla/csrf or gorilla/sessions packages).

Files:
- `docs/migration/middleware-compat.md`
- `docs/ADOPTION_PATH.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `gofmt -l .`

Docs Sync:
- docs/ADOPTION_PATH.md is updated in this card.

Done Definition:
- docs/migration/middleware-compat.md exists with four wrapping pattern snippets.
- docs/ADOPTION_PATH.md references docs/migration/.
- All four guides (from-gin, from-echo, from-chi, middleware-compat) exist in
  docs/migration/.
- `gofmt -l .` outputs nothing.

Outcome:
-

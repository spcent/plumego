# Card 1531

Milestone: M-013
Recipe: specs/change-recipes/update-docs.yaml
Priority: P2
State: active
Primary Module: docs
Owned Files:
- `docs/migration/from-echo.md`

Goal:
- Write docs/migration/from-echo.md giving Echo users a code-anchored migration
  path to plumego with emphasis on the Context abstraction difference.

Scope:
- Write docs/migration/from-echo.md with:
  - Concept mapping table: echo.Echo → core.App, echo.Group → router.Group,
    echo.Context → (http.ResponseWriter, *http.Request), echo.HandlerFunc →
    func(http.ResponseWriter, *http.Request), echo.HTTPError → contract.APIError,
    c.JSON → contract.WriteResponse, c.Param → contract.RequestParamFromContext,
    c.Bind → json.NewDecoder(r.Body).Decode + x/validate.Bind[T],
    echo.MiddlewareFunc → func(http.Handler)http.Handler.
  - Three pattern conversion snippets: handler with JSON response, middleware,
    error handling.
  - Key difference callout: Echo's c.Next() pattern does not exist; middleware
    calls next.ServeHTTP(w, r) instead.
  - Incompatibilities section: no template rendering, no binder registry,
    no validator hook, no SkipperMiddleware.
  - Links to docs/CANONICAL_STYLE_GUIDE.md and reference/standard-service.

Non-goals:
- Do not introduce compatibility shims.
- Do not document Echo-specific features like WebSocket upgrader (use x/websocket).
- Do not reference planned features as implemented.

Files:
- `docs/migration/from-echo.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `gofmt -l .`

Docs Sync:
- docs/ADOPTION_PATH.md updated in card 1533 (done after all four guides exist).

Done Definition:
- docs/migration/from-echo.md exists with concept mapping table, three code
  snippets, and incompatibilities section.
- All code snippets reference only currently exported plumego symbols.

Outcome:
-

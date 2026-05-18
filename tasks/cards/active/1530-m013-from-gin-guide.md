# Card 1530

Milestone: M-013
Recipe: specs/change-recipes/update-docs.yaml
Priority: P2
State: active
Primary Module: docs
Owned Files:
- `docs/migration/from-gin.md`

Goal:
- Write docs/migration/from-gin.md giving Gin users a code-anchored migration
  path to plumego covering the three most impactful translation points.

Scope:
- Create docs/migration/ directory if it does not exist.
- Write docs/migration/from-gin.md with:
  - Concept mapping table: gin.Engine → core.App, gin.RouterGroup → router.Group,
    gin.Context → (*http.Request, http.ResponseWriter), gin.HandlerFunc →
    func(http.ResponseWriter, *http.Request), gin.H → contract.WriteResponse,
    gin.AbortWithError → contract.WriteError + return, c.Param → contract.RequestParamFromContext,
    c.ShouldBindJSON → json.NewDecoder(r.Body).Decode
  - Three pattern conversion snippets (handler, route group, JSON response) as
    compilable Go code blocks referencing current plumego API symbols.
  - Middleware translation: gin.MiddlewareFunc wrapping pattern to
    func(http.Handler)http.Handler.
  - Explicit incompatibilities section: no c.Next(), no context value bag,
    no automatic 404/405 JSON, no panic recovery by default (use middleware/recovery).
  - Links to docs/CANONICAL_STYLE_GUIDE.md and reference/standard-service.

Non-goals:
- Do not introduce compatibility shims.
- Do not document Gin features that have no plumego equivalent.
- Do not reference planned features as if implemented.

Files:
- `docs/migration/from-gin.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `gofmt -l .`

Docs Sync:
- docs/ADOPTION_PATH.md: add reference to docs/migration/ after this card is done.

Done Definition:
- docs/migration/from-gin.md exists with concept mapping table, three code
  snippets using current API, and explicit incompatibilities list.
- All code snippets reference only currently exported plumego symbols.
- `gofmt -l .` outputs nothing.

Outcome:
-

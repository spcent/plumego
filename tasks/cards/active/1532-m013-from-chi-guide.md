# Card 1532

Milestone: M-013
Recipe: specs/change-recipes/update-docs.yaml
Priority: P2
State: active
Primary Module: docs
Owned Files:
- `docs/migration/from-chi.md`

Goal:
- Write docs/migration/from-chi.md focused on the minimal delta between chi
  and plumego, since both are net/http compatible, highlighting convention and
  tooling differences rather than API rewrites.

Scope:
- Write docs/migration/from-chi.md with:
  - Compatibility statement: chi middleware (func(http.Handler)http.Handler)
    works in plumego chains without modification.
  - Concept mapping table: chi.NewRouter → router.NewRouter + core.New,
    chi.Route → router.AddRoute, r.With → middleware.NewChain,
    chi.URLParam → contract.RequestParamFromContext,
    chi.RouteContext → contract.RequestContextFromContext,
    chi.Use → app.Use.
  - Route naming difference: plumego uses router.WithRouteName option;
    chi uses named routes via patterns.
  - Group prefix stacking: chi r.Route vs plumego router.Group — equivalent,
    different syntax.
  - What to keep: all existing chi-compatible middleware, httptest patterns,
    and net/http handler funcs are unchanged.
  - What to change: swap chi.NewRouter for plumego routing + core.App wiring;
    use contract helpers instead of ad-hoc JSON writes.
  - Links to reference/standard-service and docs/CANONICAL_STYLE_GUIDE.md.

Non-goals:
- Do not introduce compatibility shims.
- Do not document chi-specific features that have no plumego equivalent.

Files:
- `docs/migration/from-chi.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `gofmt -l .`

Docs Sync:
- docs/ADOPTION_PATH.md updated in card 1533.

Done Definition:
- docs/migration/from-chi.md exists with mapping table, minimal-delta focus,
  and middleware compatibility statement.
- No plumego symbols referenced that do not exist in the current API.

Outcome:
-

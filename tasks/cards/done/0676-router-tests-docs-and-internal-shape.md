# Card 0676

Milestone:
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/dispatch.go, router/router_test.go, router/router_contract_test.go, docs/modules/router/README.md
Depends On: 0674, 0675

Goal:
Reduce router test drift and align module documentation with the fixed canonical behaviors.

Scope:
- Remove or reshape duplicate tests that assert the same ANY-route and group behaviors in multiple files.
- Rename misleading internal dispatch helpers if they imply middleware ownership where router only attaches route context.
- Document the canonical route path and static prefix normalization guarantees now covered by tests.

Findings:
- `TestMethodNotAllowed` in `router_test.go` duplicates ANY-route fallback behavior instead of testing method-not-allowed behavior.
- Several tests repeat the same context-param assertions with slightly different names, making regressions harder to localize.
- `applyMiddlewareAndServe` is misleading inside `router`; this module does not own middleware chaining.
- The router primer documents intent but not the path normalization guarantees expected from `AddRoute` and static mounts.

Non-goals:
- Do not alter public API names.
- Do not broaden router into middleware ownership.
- Do not remove behavioral coverage; consolidate only when equivalent coverage remains.

Files:
- router/dispatch.go
- router/router_test.go
- router/router_contract_test.go
- docs/modules/router/README.md

Tests:
- go test -race -timeout 60s ./router/...
- go test -timeout 20s ./router/...
- go vet ./router/...

Docs Sync:
Update `docs/modules/router/README.md` for implemented route/static normalization behavior only.

Done Definition:
- Router tests stay focused and non-duplicative around ANY, context params, and group registration behavior.
- Internal helper names match router responsibilities.
- Router docs describe the fixed canonical behavior without introducing unimplemented features.

Outcome:
Done. Removed a duplicate ANY-route test and an unused helper, renamed the
internal dispatch helper so it describes route-context attachment rather than
middleware ownership, and documented the implemented route/static normalization
guarantees in the router primer.

Validation:
- go test -race -timeout 60s ./router/...
- go test -timeout 20s ./router/...
- go vet ./router/...

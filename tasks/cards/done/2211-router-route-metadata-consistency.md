# Card 2211: Router Route Metadata Consistency

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: router
Owned Files:
- `router/metadata.go`
- `router/router_contract_test.go`
- `router/reverse_routing_group_test.go`
Depends On: 2210

Goal:
Make route metadata snapshots and reverse URL generation behave consistently
for root, grouped root, missing params, and route-name replacement.

Problem:
Route metadata is stored in multiple maps and named routes overwrite silently.
`URL` returns partially substituted URLs when required params are missing, while
wildcards use a different empty-value behavior. The current behavior is uneven
and not obviously documented by tests.

Scope:
- Keep named-route replacement behavior explicit through tests.
- Make missing named route params produce a deterministic empty URL result.
- Keep grouped root URL generation stable.
- Keep `NamedRoutes` returning deep copies.

Non-goals:
- Do not remove `URLMust` or other exported symbols.
- Do not introduce a new reverse-routing API.
- Do not change route registration syntax.

Files:
- `router/metadata.go`
- `router/router_contract_test.go`
- `router/reverse_routing_group_test.go`

Tests:
- `go test -race -timeout 60s ./router/...`
- `go test -timeout 20s ./router/...`
- `go vet ./router/...`

Docs Sync:
No docs change required unless behavior change proves user-visible beyond
invalid/missing reverse-route input.

Done Definition:
- Missing required reverse-route params are consistently rejected by `URL`.
- Existing grouped root and same-name replacement behavior remains covered.
- Route metadata context remains unchanged for normal dispatch.

Outcome:
- Made `URL` return an empty string when any required param or wildcard value is
  missing, instead of producing partially substituted URLs.
- Kept grouped root and same-name replacement behavior unchanged.
- Added explicit deep-copy coverage for `NamedRoutes`.

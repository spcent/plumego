# Card 0499: Router Registration Path Contracts

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: router
Owned Files:
- `router/registration.go`
- `router/router_contract_test.go`
Depends On: none

Goal:
Make route registration reject ambiguous or malformed path patterns before they
enter the trie.

Problem:
`AddRoute` currently accepts empty parameter names, empty wildcard names, and
wildcards in non-terminal path positions. These shapes are not useful route
structure, make reverse routing and param extraction incomplete, and can create
surprising matches that are hard to review.

Scope:
- Reject parameter segments with no name.
- Reject wildcard segments with no name.
- Reject wildcard segments unless they are the final segment.
- Keep existing duplicate and conflict behavior unchanged.

Non-goals:
- Do not change exported symbols.
- Do not change matching precedence for valid routes.
- Do not introduce a new route-pattern parser package.

Files:
- `router/registration.go`
- `router/router_contract_test.go`

Tests:
- `go test -race -timeout 60s ./router/...`
- `go test -timeout 20s ./router/...`
- `go vet ./router/...`

Docs Sync:
No docs change required; this tightens invalid input handling inside existing
router responsibilities.

Done Definition:
- Invalid anonymous param and wildcard patterns return registration errors.
- Non-terminal wildcard patterns return registration errors.
- Existing valid static, param, wildcard, group, and reverse routes still pass.

Outcome:
- Added route-segment validation during `AddRoute`.
- Anonymous params, anonymous wildcards, and non-terminal wildcards now return
  registration errors before trie mutation.
- Added focused malformed-pattern coverage in `router_contract_test.go`.

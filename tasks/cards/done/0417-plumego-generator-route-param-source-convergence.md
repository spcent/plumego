# Card 0417: Plumego Generator Route Param Source Convergence
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: high
State: done
Primary Module: cmd/plumego
Owned Files:
- cmd/plumego/internal/codegen/codegen.go
- cmd/plumego/internal/codegen/codegen_test.go
- cmd/plumego/internal/scaffold/scaffold.go
- cmd/plumego/internal/scaffold/scaffold_test.go
Depends On: none

Goal:
Make generated handlers use Plumego's canonical route-parameter source instead
of `http.Request.PathValue`. Generated CRUD and scaffold handlers currently
emit `r.PathValue("id")`, but the repository router stores route params in
`contract.RequestContext`.

Scope:
- Update generated handler templates in codegen and scaffold output to read
  `id` from `contract.RequestContextFromContext(r.Context()).Params`.
- Add or update tests that assert generated source does not contain
  `PathValue("id")`.
- Preserve existing generated error response shape and `contract.WriteResponse`
  usage.
- Avoid adding a generated dependency on the `router` package.

Non-goals:
- Do not redesign generated service interfaces or domain layout.
- Do not change generated route paths.
- Do not add external dependencies or change `go.mod`.

Tests:
- go test -timeout 20s ./cmd/plumego/internal/codegen ./cmd/plumego/internal/scaffold
- go vet ./cmd/plumego/internal/codegen ./cmd/plumego/internal/scaffold
- rg -n 'PathValue\(' cmd/plumego/internal/codegen cmd/plumego/internal/scaffold -g '*.go'

Docs Sync:
No docs update is required unless generator docs explicitly mention
`http.Request.PathValue`.

Done Definition:
- Generated handlers no longer emit `PathValue("id")`.
- Codegen and scaffold tests cover the canonical parameter source.
- The listed command tests pass; the final `rg` has no production-template
  matches.

Outcome:
Updated codegen and scaffold handler templates to read route params from
`contract.RequestContextFromContext(r.Context()).Params["id"]`. Added template
tests that require the canonical lookup and reject `PathValue(` in generated
handler content.

Validation:
- `go test -timeout 20s ./internal/codegen ./internal/scaffold` from `cmd/plumego`
- `go vet ./internal/codegen ./internal/scaffold` from `cmd/plumego`
- `rg -n 'PathValue\(' internal/codegen/codegen.go internal/scaffold/scaffold.go`
  from `cmd/plumego` returned no matches.

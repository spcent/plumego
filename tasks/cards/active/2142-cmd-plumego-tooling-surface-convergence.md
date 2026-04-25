# Card 2142: cmd/plumego Tooling Surface Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/codegen/codegen.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/output/formatter.go`
- `cmd/plumego/commands/cli_e2e_test.go`
Depends On: none

Goal:
Reduce drift between scaffolded app code, generated handler code, devserver
handlers, and CLI JSON output so the tooling teaches the same canonical response
and handler style as the reference app.

Problem:
`cmd/plumego` has several independently maintained tooling surfaces that emit
or serve HTTP-like behavior: scaffold templates, codegen templates, devserver
JSON handlers, and CLI output formatting. They mostly use `contract.WriteResponse`
and `contract.WriteError`, but the response examples, ad hoc maps, and template
assertions are spread across large files. `scaffold.go` and `dashboard.go` are
large enough that future style changes are likely to be missed in one surface.

Scope:
- Audit scaffold/codegen/devserver output for canonical handler signature,
  `json.NewDecoder(r.Body).Decode`, `contract.WriteResponse`, and
  `contract.WriteError` consistency.
- Extract only tiny private helpers or constants that remove real duplication
  across generated status/health/example payloads.
- Add or update tests so scaffold and codegen templates assert the same canonical
  response and error paths.
- Keep CLI output envelope behavior separate from HTTP `contract.Response`, but
  document that distinction in code comments or tests.

Non-goals:
- Do not redesign the CLI command framework.
- Do not change generated project layout away from `reference/standard-service`.
- Do not add dependencies.
- Do not touch stable public APIs.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/codegen/codegen.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/output/formatter.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test -timeout 20s ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
Update `cmd/plumego/DEV_SERVER.md` only if devserver response behavior changes.

Done Definition:
- Scaffold and codegen templates teach the same canonical handler/error style.
- Devserver JSON handlers do not introduce a competing HTTP response shape.
- Tests pin the intended distinction between CLI output envelopes and HTTP
  `contract.Response`.
- The listed validation commands pass.

Outcome:

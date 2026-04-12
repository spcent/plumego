# Card 0963: Cmd Plumego Generated HTTP Surface Convergence

Priority: P1
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: cmd/plumego
Depends On: —

## Goal

Bring `cmd/plumego`'s scaffold and codegen output back to Plumego's canonical
HTTP response style so the tool no longer generates fresh projects and handlers
 that immediately drift from `contract.WriteResponse`, `contract.WriteError`,
and the reference app conventions.

## Problem

- `cmd/plumego/internal/scaffold/scaffold.go` still generates handlers that use
  `http.Error`, raw `json.NewEncoder(w).Encode(...)`, and `"encoding error"`
  fallbacks instead of the repo's canonical contract helpers.
- The same scaffold also emits bespoke `internal/platform/httpjson` and
  `internal/platform/httperr` helper packages, including a local
  `ErrorResponse` type, which recreates a second transport contract inside
  generated services.
- `cmd/plumego/internal/codegen/codegen.go` generates CRUD handlers with the
  same hand-rolled `http.Error` + manual JSON style, so tool-generated code and
  scaffold-generated code teach two variants of the same non-canonical pattern.
- Because these files are developer tooling surfaces, the drift propagates into
  every generated app instead of remaining local to one package.

## Scope

- Converge scaffold-generated handlers onto `contract.WriteResponse` /
  `contract.WriteError` and remove the `"encoding error"` fallback teaching.
- Stop generating `internal/platform/httpjson` and `internal/platform/httperr`
  when the same behavior should come from `contract`.
- Converge codegen-generated CRUD handlers onto the same canonical response and
  error path as the scaffold.
- Update scaffold/codegen tests so generated output asserts the new canonical
  transport vocabulary.

## Non-Goals

- Do not redesign scaffolded application layout beyond response and error write
  policy.
- Do not change CLI flag shape, command UX, or route topology.
- Do not widen `contract` just to support tool-only helper families.

## Files

- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/internal/codegen/codegen.go`
- `cmd/plumego/internal/codegen/*_test.go`
- any scaffold/codegen golden assertions that encode the old helper path

## Tests

```bash
rg -n 'encoding error|internal/platform/httpjson|internal/platform/httperr|http\.Error\(|json\.NewEncoder\(w\)\.Encode' cmd/plumego/internal/scaffold cmd/plumego/internal/codegen -g '!**/*_test.go'
go test -timeout 20s ./cmd/plumego/internal/scaffold ./cmd/plumego/internal/codegen ./cmd/plumego/commands
go vet ./cmd/plumego/internal/scaffold ./cmd/plumego/internal/codegen ./cmd/plumego/commands
```

## Docs Sync

- only if scaffold or codegen docs explicitly describe the generated response
  helper layout

## Done Definition

- `cmd/plumego` no longer generates bespoke JSON/error helper packages that
  compete with `contract`.
- Scaffold-generated and codegen-generated handlers share one canonical success
  and error write path.
- No generated template still teaches `"encoding error"` fallback or raw
  `http.Error` for normal JSON API handlers.
- Focused scaffold/codegen tests pass with the new generated output.

## Outcome

- Removed `internal/platform/httpjson` and `internal/platform/httperr` from the
  microservice scaffold file list so new projects stop generating a second
  helper family.
- Converged scaffolded health/API/user/metrics handlers and the inline minimal
  health route onto `contract.WriteResponse` / `contract.WriteError`.
- Converged codegen-generated CRUD handlers onto the same contract-based
  success and error path, while keeping JSON decode explicit.
- Added focused generator tests that fail on legacy helper packages,
  `http.Error`, manual JSON encoding, and raw JSON content-type writes.

## Validation Run

```bash
rg -n 'encoding error|internal/platform/httpjson|internal/platform/httperr|http\.Error\(|json\.NewEncoder\(w\)\.Encode|w\.Header\(\)\.Set\("Content-Type", "application/json"\)' cmd/plumego/internal/scaffold cmd/plumego/internal/codegen -g '!**/*_test.go'
cd cmd/plumego && go test -timeout 20s ./internal/scaffold ./internal/codegen ./commands
cd cmd/plumego && go vet ./internal/scaffold ./internal/codegen ./commands
```

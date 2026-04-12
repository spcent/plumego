# Card 0976: X/Devtools PubsubDebug and Cmd/Devserver Handler Migration

Priority: P3
State: active
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/devtools/pubsubdebug, cmd/plumego/internal/devserver
Depends On: 0975

## Goal

Migrate the final two packages still using the local `adaptCtx`/`*contract.Ctx`
handler pattern — `x/devtools/pubsubdebug` and `cmd/plumego/internal/devserver`
— to the canonical `func(w http.ResponseWriter, r *http.Request)` shape, then
delete the per-package adapter files/functions. This completes the full-codebase
retirement of the `adaptCtx` anti-pattern started in cards 0974–0975.

## Problem

### x/devtools/pubsubdebug

`x/devtools/pubsubdebug/ctx_adapter.go` is an isolated file whose sole purpose
is to define a local `adaptCtx` function. The two consumers in this package
(`component.go` line 44 and `configure.go` line 38) pass anonymous
`func(ctx *contract.Ctx)` closures through it, even though the handler bodies
are minimal enough to be inlined as standard `http.HandlerFunc` closures.

### cmd/plumego/internal/devserver/dashboard.go

The devserver dashboard defines a local `adaptCtx` closure at line 154 and uses
it to register 17 routes (lines 162–207). All 14 handler methods on `*Dashboard`
(plus `handleDeps` in `deps.go`, `parsePprofRequest`/`handlePprofRaw` in
`pprof_proxy.go`, and `handleConfigEditGet`/`handleConfigEditSave` in
`config_edit.go`) accept `*contract.Ctx` instead of `(w, r)`. This is the
largest single concentration of the pattern in the codebase.

Because `cmd/devserver` is developer tooling rather than a shipped library, the
migration is lower-risk than x/rest but still important to eliminate the
teaching surface of the old pattern.

## Scope

### x/devtools/pubsubdebug

- Rewrite the two anonymous handler closures in `component.go` and
  `configure.go` as `http.HandlerFunc` closures accepting `(w, r)`.
- Delete `ctx_adapter.go`.
- Remove the `contract` import if no longer needed after the migration.

### cmd/plumego/internal/devserver

- Change all handler methods on `*Dashboard` (in `dashboard.go`, `deps.go`,
  `pprof_proxy.go`, `config_edit.go`) to accept
  `(w http.ResponseWriter, r *http.Request)`.
- Update `parsePprofRequest` helper to accept `r *http.Request` instead of
  `ctx *contract.Ctx`.
- Remove the local `adaptCtx` closure from `RegisterRoutes` (or equivalent);
  register handlers directly as `http.HandlerFunc`.
- Replace all `ctx.W`, `ctx.R`, `ctx.PathParam()`, `ctx.QueryParam()`,
  `ctx.BodyBytes()`, `ctx.RequestID()`, `ctx.RequestHeaders()` with equivalent
  `w`, `r`, and standard-library calls.
- Prefix every `contract.WriteError(...)` call in migrated handlers with `_ =`.

## Non-Goals

- Do not change devserver route paths, UI asset embedding, or build/restart
  logic.
- Do not change `x/devtools/devtools.go` in the same pass (that file's bare
  WriteError call is already captured by card 0971).
- Do not expose the devserver's internal handlers as a public API.

## Files

- `x/devtools/pubsubdebug/ctx_adapter.go` (delete)
- `x/devtools/pubsubdebug/component.go`
- `x/devtools/pubsubdebug/configure.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/deps.go`
- `cmd/plumego/internal/devserver/pprof_proxy.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `cmd/plumego/internal/devserver/*_test.go` (if any)

## Tests

```bash
rg -n 'adaptCtx\|contract\.Ctx' x/devtools/pubsubdebug cmd/plumego/internal/devserver -g '*.go'
go test -timeout 20s ./x/devtools/... ./cmd/plumego/...
go vet ./x/devtools/... ./cmd/plumego/...
```

After all adaptCtx migration cards (0974–0976) are complete, run the full
codebase check:

```bash
rg -rn 'adaptCtx\|func.*ctx \*contract\.Ctx' --include='*.go' .
go test -timeout 30s ./...
go vet ./...
```

## Docs Sync

- `docs/modules/x-devtools/README.md` if it shows pubsubdebug route
  registration or handler shape.

## Done Definition

- `rg 'adaptCtx\|contract\.Ctx' x/devtools/pubsubdebug cmd/plumego/internal/devserver -g '*.go'`
  returns no results.
- `x/devtools/pubsubdebug/ctx_adapter.go` no longer exists.
- All devtools and devserver tests pass; `go vet` is clean.
- Full-codebase `rg 'adaptCtx' --include='*.go' .` returns no results.

## Outcome


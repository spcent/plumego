# Card 0963: Reference Apps, Scaffold, and Codegen Response-Writing Canonicalization

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: active
Primary Module: reference/standard-service
Owned Files: reference/*, cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/internal/codegen/codegen.go, middleware/debug/debug_errors.go
Depends On:

## Goal

Eliminate `http.Error` from JSON API handlers in all reference apps, scaffold templates,
and codegen output, and replace the awkward config-merge pattern in `DebugErrors` with a
direct assignment. These files are teaching surfaces — incorrect patterns in them become
the first thing new contributors copy.

## Problem

### 1. `http.Error` fallback after `contract.WriteResponse` in reference apps

All five reference apps contain this pattern:

```go
if err := contract.WriteResponse(w, r, http.StatusOK, resp, nil); err != nil {
    http.Error(w, "encoding error", http.StatusInternalServerError)
}
```

The fallback is both wrong and unreachable in practice:
- `contract.WriteResponse` writes the Content-Type and status headers *before* encoding.
  If encoding fails later, the 200 is already on the wire — the `http.Error` call has no
  effect on the wire format.
- Even when encoding succeeds, the plain-text `http.Error` body would violate the rule
  "Structured errors only — no ad hoc `http.Error` in JSON APIs" (Style Guide §10).

Canonical pattern: `_ = contract.WriteResponse(...)` — discard the error because nothing
useful can be done once headers are sent.

Affected files:
- `reference/standard-service/internal/handler/api.go` — lines 36–38, 61–63
- `reference/standard-service/internal/handler/health.go` — lines 22–29, 39–46
- `reference/with-gateway/internal/app/routes.go` — line 18
- `reference/with-messaging/internal/app/routes.go` — line 18
- `reference/with-messaging/internal/handler/messaging.go` — line 68
- `reference/with-webhook/internal/app/routes.go` — line 18
- `reference/with-websocket/internal/app/routes.go` — line 18

### 2. `http.Error` and raw `json.NewEncoder` in scaffold/codegen templates

`cmd/plumego/internal/scaffold/scaffold.go` generates handler source code that contains
`http.Error(w, "...", status)` and raw `json.NewEncoder(w).Encode(...)` for errors and
responses respectively.  `cmd/plumego/internal/codegen/codegen.go` does the same.
New projects created with `plumego new` inherit these patterns verbatim.

The generated templates should model the canonical approach:
- Errors → `contract.WriteError(w, r, contract.NewErrorBuilder()...Build())`
- Success responses → `contract.WriteResponse(w, r, status, data, nil)`

Affected files:
- `cmd/plumego/internal/scaffold/scaffold.go` — multiple instances
- `cmd/plumego/internal/codegen/codegen.go` — multiple instances

### 3. Redundant config-merge in `DebugErrors` (middleware/debug)

`DebugErrors` initialises `cfg` from `DefaultDebugErrorConfig()` and then applies the
caller's config field-by-field using conditional guards:

```go
cfg := DefaultDebugErrorConfig()
if config.NotFoundHint != "" { cfg.NotFoundHint = config.NotFoundHint }
if config.IncludeBody     { cfg.IncludeBody = true }
if !config.IncludeRequest { cfg.IncludeRequest = false }
if !config.IncludeQuery   { cfg.IncludeQuery   = false }
```

The guards produce surprising semantics: `DebugErrors(DebugErrorConfig{})` (zero value)
disables `IncludeRequest` and `IncludeQuery` even though `DefaultDebugErrorConfig()` sets
them to `true`. Callers are expected to start from `DefaultDebugErrorConfig()` anyway (as
the doc comment says), so the merge is unnecessary. The body should simply be:

```go
cfg := config
```

Also: `contract.WriteError(w, r, ...)` on line 129 of `debug_errors.go` does not discard
the return value, inconsistent with the canonical `_ = contract.WriteError(...)` pattern.

## Scope

- Replace every `if err := contract.WriteResponse(...); err != nil { http.Error(...) }` in
  the reference apps with `_ = contract.WriteResponse(...)`.
- Update scaffold.go generated handler templates to use `contract.WriteError` and
  `contract.WriteResponse` instead of `http.Error` and raw `json.NewEncoder`.
- Update codegen.go in the same pass.
- Simplify `DebugErrors` config initialisation to `cfg := config`.
- Add `_ =` to the bare `contract.WriteError(...)` call on line 129 of `debug_errors.go`.

## Non-Goals

- Do not change any JSON response envelope shape.
- Do not change middleware/debug error detection logic or the `shouldReplaceError` predicate.
- Do not refactor scaffold template structure beyond the error/response lines.
- Do not touch `x/ops/healthhttp`, `x/pubsub/prometheus`, or other packages that use
  `http.Error` for non-JSON (CSV, Prometheus text) responses — those are correct as-is.

## Files

- `reference/standard-service/internal/handler/api.go`
- `reference/standard-service/internal/handler/health.go`
- `reference/with-gateway/internal/app/routes.go`
- `reference/with-messaging/internal/app/routes.go`
- `reference/with-messaging/internal/handler/messaging.go`
- `reference/with-webhook/internal/app/routes.go`
- `reference/with-websocket/internal/app/routes.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/codegen/codegen.go`
- `middleware/debug/debug_errors.go`

## Tests

```bash
go test -timeout 20s ./reference/... ./middleware/... ./middleware/debug/...
go vet ./reference/... ./middleware/...
(cd cmd/plumego && go test -timeout 20s ./...)
```

## Docs Sync

None required — the style guide already describes the canonical pattern.

## Done Definition

- No `http.Error` call remains in reference handler files or in the handler templates
  embedded in scaffold.go and codegen.go.
- `DebugErrors` body opens with `cfg := config` instead of `cfg := DefaultDebugErrorConfig()`.
- `contract.WriteError` call in `debug_errors.go:129` uses `_ =`.
- All targeted tests and vet pass.

## Outcome


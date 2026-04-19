# Card 0985: Unify Error-Writing Path in Timeout Middleware

Priority: P2
State: active
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: middleware/timeout

## Goal

`middleware/timeout/timeout.go` violates the one-canonical-error-path principle by using two
different error-writing helpers inside the same handler's `select` block:

- **line 96**: `_ = contract.WriteError(w, r, contract.NewErrorBuilder()...)` — direct call for
  the buffer-overflow case
- **line 104**: `mw.WriteTransportError(w, r, http.StatusGatewayTimeout, ...)` — shorthand for
  the timeout case

Every other middleware that writes transport errors (`coalesce`, `concurrencylimit`,
`bodylimit`, `ratelimit`) uses only `mw.WriteTransportError`.  The overflow path in `timeout`
is the single outlier.  It makes the code harder to read and breaks grep-ability of the
transport error surface.

## Scope

Rewrite the overflow branch to use `mw.WriteTransportError` with an appropriate status,
code, message, and category.  The overflow case is an internal buffer exhaustion — status
`500`, code `contract.CodeInternalError`, category `contract.CategoryServer`.

```go
// before (line 96)
_ = contract.WriteError(w, r, contract.NewErrorBuilder().
    Type(contract.TypeInternal).
    Message("response exceeded buffer limit").
    Build())

// after
mw.WriteTransportError(w, r, http.StatusInternalServerError, contract.CodeInternalError,
    "response exceeded buffer limit", contract.CategoryServer, nil)
```

Also remove the now-unused direct `contract` import alias if it becomes unreferenced after
this change (keep the `mw` alias).

## Non-goals

- Do not change the timeout-case error response (line 104) — it is already correct.
- Do not touch `timeout.go:179` (the secondary `WriteError` in `timeoutResponseWriter.WriteError`
  method) — it writes to an internal buffer recorder, not to the live `ResponseWriter`, so
  `WriteTransportError` is not appropriate there.

## Files

- `middleware/timeout/timeout.go` (lines 93-104 — the select block overflow arm)

## Tests

```bash
go test -timeout 20s ./middleware/timeout/...
go test -timeout 20s ./middleware/...
go vet ./middleware/...
```

Confirm no `contract.WriteError` call remains in the select block after the change:

```bash
grep -n "contract\.WriteError" middleware/timeout/timeout.go
```

## Done Definition

- `middleware/timeout/timeout.go` select block uses `mw.WriteTransportError` for both arms.
- No `contract.WriteError` call remains inside the `Timeout` handler func.
- `go test ./middleware/timeout/...` passes.
- `go vet ./middleware/...` clean.

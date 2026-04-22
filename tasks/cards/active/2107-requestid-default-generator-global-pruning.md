# Card 2107: Prune Mutable Request-ID Default Generator

Priority: P1
State: active
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: middleware
Owned Files:
- middleware/requestid/generator.go
- middleware/requestid/helpers.go
- middleware/requestid/request_id.go
- middleware/requestid/request_id_generation_test.go
- docs/modules/middleware/README.md
Depends On:

## Goal

`middleware/requestid` currently exposes a package-level mutable generator:

```go
var DefaultGenerator = NewRequestIDGenerator()
```

This is convenient, but it is also an exported mutable global in a stable
transport middleware package.  The canonical style prefers explicit
constructor/configuration paths and avoids hidden global state.  The package
already supports explicit injection through `WithGenerator`, so the exported
global is not needed as the public customization path.

## Scope

- Follow the exported-symbol change protocol for `DefaultGenerator`.
- Replace the exported mutable variable with an unexported package default or a
  function that cannot be reassigned by callers.
- Keep `NewRequestID()` and `WithGenerator(...)` as the public generation paths.
- Ensure concurrent generation remains race-free and monotonic enough for the
  existing tests.
- Update tests to assert callers cannot mutate package default state through an
  exported variable.

## Non-goals

- Do not change the request ID wire format or `contract.RequestIDHeader`.
- Do not remove `NewRequestIDGenerator`.
- Do not change middleware route/header behavior.
- Do not add dependencies.

## Files

- `middleware/requestid/generator.go`
- `middleware/requestid/helpers.go`
- `middleware/requestid/request_id.go`
- `middleware/requestid/request_id_generation_test.go`
- `docs/modules/middleware/README.md`

## Tests

```bash
rg -n --glob '*.go' 'DefaultGenerator' .
go test -race -timeout 60s ./middleware/...
go vet ./middleware/...
```

## Docs Sync

Update `docs/modules/middleware/README.md` if it mentions request-id generation
customization.  The documented customization path should be `WithGenerator`.

## Done Definition

- No exported mutable `DefaultGenerator` remains.
- `NewRequestID()` still works and remains race-free.
- `WithGenerator` remains the explicit middleware customization path.
- Any old references to `DefaultGenerator` are removed or migrated.
- The listed validation commands pass.

## Outcome


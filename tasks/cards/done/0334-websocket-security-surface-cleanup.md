# Card 0334: WebSocket Security Surface Cleanup

Priority: P1
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/websocket
Depends On: —

## Goal

Tighten the `x/websocket` security surface so authentication and password
handling stop relying on hidden package-global logging and stale deprecated
wrappers, and so related APIs behave consistently across simple and secure auth
implementations.

## Problem

- `x/websocket/security.go` still exports deprecated `GetSecurityMetrics` and
  `ResetSecurityMetrics`, but repo-wide grep shows no live callers; they are
  dead compatibility shims now returning empty or no-op behavior.
- `x/websocket/security.go` and `x/websocket/auth.go` call `log.Printf`
  directly for weak-secret warnings, weak-password warnings, JWT verification
  failures, and password-hash failures, which injects package-global logging
  into a library module.
- `SimpleRoomAuth.SetRoomPassword` logs a hashing failure and returns `void`,
  while `SecureRoomAuth.SetRoomPassword` returns an error. The API asymmetry
  makes failure handling incomplete and non-obvious.
- The module manifest says auth flow and secret handling must stay explicit, but
  the current surface still contains silent/no-op or log-only escape hatches.

## Scope

- Remove deprecated global metrics wrappers after verifying zero remaining
  callers.
- Replace package-global logging in auth/security code with an explicit policy:
  return errors, use injected logging, or keep warnings fully caller-controlled.
- Resolve the `SetRoomPassword` asymmetry between `SimpleRoomAuth` and
  `SecureRoomAuth` without leaving silent hash-failure behavior behind.
- Update tests to cover the chosen failure path and removal of stale wrappers.

## Non-Goals

- Do not redesign websocket hub protocol, room semantics, or broadcast flow.
- Do not weaken JWT, password-strength, or origin-check behavior.
- Do not move websocket behavior into stable roots.

## Expected Files

- `x/websocket/auth.go`
- `x/websocket/security.go`
- `x/websocket/*security*_test.go`
- `x/websocket/ws_test.go`
- `x/websocket/websocket_extended_test.go`

## Validation

```bash
rg -n 'GetSecurityMetrics|ResetSecurityMetrics|log\.Printf\(|SetRoomPassword\(' x/websocket -g '!**/*_test.go'
go test -timeout 20s ./x/websocket/...
go vet ./x/websocket/...
```

## Docs Sync

- `docs/modules/x-websocket/README.md` only if the public auth API or supported
  metrics surface changes

## Done Definition

- Deprecated global metrics wrappers are removed with zero residual references.
- WebSocket auth/security code no longer depends on package-global `log.Printf`
  side effects.
- Password-setting failure behavior is explicit and consistent across the simple
  and secure auth surfaces.
- Focused websocket tests cover the finalized security API.

## Outcome

- Removed deprecated global websocket security metric wrappers after migrating the module fully onto instance-scoped metrics.
- Changed `SimpleRoomAuth.SetRoomPassword` to return `error`, aligning it with `SecureRoomAuth` and eliminating silent hash-failure behavior.
- Replaced websocket auth/security package-global logging with caller-provided optional logger hooks in `SecurityConfig`.
- Added focused tests for explicit logger behavior and the new password-setting contract.

## Validation Run

```bash
rg -n 'GetSecurityMetrics|ResetSecurityMetrics|log\.Printf\(|SetRoomPassword\(' x/websocket -g '!**/*_test.go'
go test -timeout 20s ./x/websocket/...
go vet ./x/websocket/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

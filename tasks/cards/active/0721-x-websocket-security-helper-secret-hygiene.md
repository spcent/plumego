# 0721 - x/websocket security helper secret hygiene

Status: active
Priority: P2
Primary module: `x/websocket`

## Problem

`ValidateSecurityConfig` converts JWT secrets to strings for weak-pattern
warnings, creating avoidable immutable copies. `NewSecureRoomAuth` also needs a
clear secret clone/ownership contract.

## Scope

- Replace string-based weak-pattern checks with byte-slice checks.
- Clone caller-provided secrets when storing them.
- Document and test secret ownership semantics.
- Keep timing-safe comparisons for secret checks.

## Out of Scope

- Full JWT policy support.
- External secret manager integration.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`


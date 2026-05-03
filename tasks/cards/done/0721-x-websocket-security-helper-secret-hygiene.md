# 0721 - x/websocket security helper secret hygiene

Status: done
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

## Outcome

- Replaced string-based weak JWT secret pattern checks with byte-slice checks.
- Cloned effective JWT secrets before storing them in `SecureRoomAuth` config
  and token auth state.
- Added regression coverage for caller secret mutation.
- Updated module primer notes.

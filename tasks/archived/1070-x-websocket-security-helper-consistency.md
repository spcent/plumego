# 1070 - x/websocket security helper consistency

Status: done
Priority: P1
State: done
Primary module: `x/websocket`

## Goal

Remove remaining confusing security-helper defaults before API freeze.

## Scope

- Align `SecurityConfig.EnforcePasswordStrength` documentation and runtime
  default.
- Validate room names in built-in room-auth password setters.
- Clarify `SimpleRoomAuth` versus `SecureRoomAuth` intended use.
- Add tests for default strong-password enforcement and invalid room names.

## Non-goals

- Full OIDC/JWT policy.
- Replacing the password package.
- Changing room password transport.

## Files

- `x/websocket/auth.go`
- `x/websocket/security.go`
- `x/websocket/security_test.go`
- `docs/modules/x-websocket/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document the safe default for `SecureRoomAuth` and the basic nature of
`SimpleRoomAuth`.

## Done Definition

- Built-in room auth rejects invalid room names.
- `SecureRoomAuth` enforces password strength by default.
- Validation passes.

## Outcome

- `SimpleRoomAuth.SetRoomPassword` and `SecureRoomAuth.SetRoomPassword` now
  reject invalid room names.
- `SecureRoomAuth` now enforces room password strength by default.
- `SecurityConfig.AllowWeakRoomPasswords` provides the explicit opt-out path
  for development or migration use.
- Module docs and API snapshot were refreshed.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go build ./...`

# Card 1353

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/security.go`
- `x/websocket/security_test.go`
- `x/websocket/auth.go`
- `x/websocket/module.yaml`

Goal:
- Make security helper defaults and validation semantics match their comments.

Scope:
- Ensure `NewSecureRoomAuth` enforces room password strength by default.
- Keep weak JWT pattern checks advisory unless an explicit hard-fail option exists.
- Avoid retaining unnecessary secret copies in stored helper config where possible.
- Clarify simple helpers as lightweight helpers, not full identity policy.

Non-goals:
- Do not implement full JWT/OIDC policy.
- Do not change the password package.

Files:
- `x/websocket/security.go`
- `x/websocket/security_test.go`
- `x/websocket/auth.go`
- `x/websocket/module.yaml`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for security default behavior changes.

Done Definition:
- Zero-value `SecurityConfig` no longer silently accepts weak room passwords through `NewSecureRoomAuth`.
- Tests cover default enforcement and explicit relaxed mode.

Outcome:
- Made `NewSecureRoomAuth` enforce room password strength by default.
- Added explicit `AllowWeakRoomPasswords` for relaxed room-password policy.
- Avoided retaining `JWTSecret` in the stored `SecurityConfig` after token auth
  construction.
- Clarified `SimpleRoomAuth` as a lightweight helper with open rooms when no
  room password is configured.
- Verified with `go test -timeout 20s ./x/websocket/...`, `go vet
  ./x/websocket/...`, and `go run ./internal/checks/module-manifests`.

# Card 1397

Milestone: v1-cleanup-phase-4
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/ipc
Owned Files:
- x/ipc/ipc.go
- x/ipc/framing.go
- x/ipc/ipc_test.go
Depends On:
- 1396

Goal:
- Reduce `x/ipc` edit radius by extracting message framing helpers from the large package file.

Scope:
- Move frame read/write helpers and related private constants/types into `framing.go`.
- Preserve existing exported types, functions, errors, and package documentation.
- Keep tests focused on existing read/write behavior, timeout behavior, and error wrapping.

Non-goals:
- Do not change the wire format.
- Do not change client/server public interfaces.
- Do not touch platform-specific Unix or Windows files.
- Do not add reconnect behavior changes.

Files:
- x/ipc/ipc.go
- x/ipc/framing.go
- x/ipc/ipc_test.go

Tests:
- go test -timeout 20s ./x/ipc
- go vet ./x/ipc
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless package docs become inaccurate after extraction.

Done Definition:
- Framing logic is isolated in `framing.go`.
- Public behavior and tests remain unchanged.
- `x/ipc` is easier to edit without widening future changes.

Outcome:
- Extracted `FramedClient`, frame constants, framed read/write methods, and `timeoutReader` into `x/ipc/framing.go`.
- Left protocol encoding, frame size limits, timeout behavior, and delegated client methods unchanged.
- Validation:
  - `go test -timeout 20s ./x/ipc`
  - `go vet ./x/ipc`
  - `go run ./internal/checks/dependency-rules`

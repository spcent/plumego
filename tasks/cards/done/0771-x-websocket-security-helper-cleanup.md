# Card 0771

Milestone: M-003
Recipe: specs/change-recipes/security-hardening.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/security.go`
- `x/websocket/validation.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0764, 0765, 0769

Goal:
- Remove or tighten security helpers that do not belong in the stable WebSocket transport contract.

Problem:
Some helpers create confusing security promises: weak-secret warning converts secrets to strings, `SanitizeForLogging` claims log-injection protection while preserving newlines/tabs, and `ContainsDangerousPatterns` is a heuristic XSS/SQL scanner inside a transport package.

Scope:
- Remove secret-to-string weak-pattern checks or replace them with byte-safe explicit validation.
- Clone secrets stored in auth/security structs.
- Align message validation and log sanitization semantics.
- Delete or internalize `ContainsDangerousPatterns` unless it is explicitly accepted as stable API.
- Update tests and docs for the reduced helper surface.

Non-goals:
- Do not build a generic input-security framework.
- Do not add business-layer XSS or SQL validation to websocket transport.
- Do not leave deprecated wrappers behind.

Files:
- `x/websocket/security.go`
- `x/websocket/validation.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for security helper removals and logging semantics.

Done Definition:
- No helper creates unnecessary immutable string copies of secrets.
- Log sanitization documentation matches behavior.
- Heuristic XSS/SQL detection is not part of the stable transport API unless explicitly documented.

Outcome:
- Replaced secret weak-pattern checks with byte-safe matching and kept warning
  logs caller-provided.
- Cloned secrets stored in `SecureRoomAuth` configuration and covered caller
  mutation with tests.
- Tightened log sanitization to replace all control characters, including
  newlines and tabs, and documented that websocket validation is transport
  scoped rather than XSS/SQL scanning.
- Verified with `go test -timeout 20s ./x/websocket/...`, `go vet
  ./x/websocket/...`, `go build ./...`, and `go run
  ./internal/checks/module-manifests`.

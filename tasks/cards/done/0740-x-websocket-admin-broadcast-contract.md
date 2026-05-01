# Card 0740

Milestone:
Recipe: specs/change-recipes/security-hardening.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`
Depends On: 0739

Goal:
- Make the admin broadcast endpoint explicitly configured, separately authorized, and disabled by default.

Problem:
The default config enables `/_admin/broadcast`, and the endpoint authenticates with the same raw `Secret` used by other WebSocket concerns. This couples client JWT signing material to admin operations and leaves a sensitive route enabled unless applications remember to turn it off.

Scope:
- Change `DefaultWebSocketConfig` so `BroadcastEnabled` is false.
- Split broadcast authorization from client auth using `BroadcastSecret`, `BroadcastAuthorizer`, or an equivalent explicit policy hook.
- Use timing-safe comparison for static broadcast secrets.
- Remove any dependency on the general `Secret` for admin broadcast authorization.
- Keep broadcast body limits and JSON validation fail-closed.
- Remove implicit environment reads from default constructors if they blur security ownership; prefer an explicit env-loading helper or caller-owned wiring.

Non-goals:
- Do not add RBAC, persistence, or audit storage.
- Do not change Hub fanout algorithms except where needed for authorization tests.
- Do not promote maturity state in this card.

Files:
- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for admin route defaults, broadcast secret separation, env-loading behavior, and migration notes.

Done Definition:
- Admin broadcast is disabled unless explicitly enabled.
- Enabling broadcast without an authorizer or dedicated secret fails visibly.
- The JWT/client secret cannot authorize admin broadcast.
- Static broadcast secret checks use timing-safe comparison.
- Tests cover disabled route, missing admin auth, wrong admin auth, valid admin auth, and oversized body.

Outcome:
- Made admin broadcast disabled by default.
- Added dedicated `BroadcastSecret` and `BroadcastAuthorizer` configuration.
- Rejected broadcast enablement without dedicated auth and rejected reuse of the JWT `Secret`.
- Removed environment reads from `DefaultWebSocketConfig`; callers now pass secrets explicitly.
- Updated docs and focused tests for missing, wrong, valid, JWT-secret, authorizer, disabled-route, and oversized-body paths.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`

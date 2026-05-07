# Card 0722

Milestone: cmd stable hardening
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: cmd/plumego/internal/devserver
Owned Files: cmd/plumego/internal/devserver/dashboard.go, cmd/plumego/internal/devserver/dashboard_info_test.go, cmd/plumego/commands/dev.go, cmd/plumego/commands/dev_test.go, cmd/plumego/DEV_SERVER.md
Depends On: 0721

Goal:
Finish local dashboard safety hardening for browser-triggerable action and event endpoints.

Scope:
- Apply dashboard token validation to WebSocket upgrades when a token is configured.
- Replace permissive dashboard CORS defaults with explicit local-origin behavior.
- Avoid manual wildcard CORS on raw pprof downloads.
- Document loopback behavior, token behavior, and remote binding expectations.

Non-goals:
- Do not build multi-user authentication.
- Do not make the dashboard production-supported.
- Do not remove local no-token ergonomics for loopback unless required by tests.

Files:
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`
- `cmd/plumego/commands/dev.go`
- `cmd/plumego/commands/dev_test.go`
- `cmd/plumego/DEV_SERVER.md`

Tests:
- `go test ./commands ./internal/devserver`
- `go test -race ./internal/devserver`
- `go build .`

Docs Sync:
- `cmd/plumego/DEV_SERVER.md`

Done Definition:
- Token-protected dashboard action and WebSocket paths reject missing or invalid tokens.
- Dashboard CORS does not allow arbitrary origins by default.
- Docs describe implemented dashboard security behavior only.

Outcome:
- Replaced dashboard default wildcard CORS with explicit dashboard-origin
  allowlists and token-aware allowed headers.
- Routed `/ws` through dashboard token validation before WebSocket handshake
  when a token is configured.
- Removed the manual wildcard CORS header from raw pprof downloads.
- Documented WebSocket token and dashboard CORS behavior.
- Validation Run:
  - `go test ./commands ./internal/devserver`
  - `go test -race ./internal/devserver`
  - `go build .`

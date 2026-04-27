# Card 0589

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P2
State: done
Primary Module: cmd/plumego
Owned Files:
- cmd/plumego/internal/scaffold/scaffold.go
- cmd/plumego/internal/scaffold/scaffold_test.go
- cmd/plumego/README.md
- docs/getting-started.md
- tasks/cards/active/README.md
Depends On: 2293

Goal:
Turn gateway and realtime scaffold profiles into runnable scenario templates.

Scope:
- Generate runnable minimal gateway and realtime routes.
- Keep default scaffold stable-root-only.
- Add focused scaffold expectations.

Non-goals:
- Do not add external discovery or broker dependencies.
- Do not install secrets.
- Do not change `rest-api` or `tenant-api` behavior.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/README.md`
- `docs/getting-started.md`
- `tasks/cards/active/README.md`

Tests:
- `cd cmd/plumego && go test -timeout 20s ./internal/scaffold/...`
- `cd cmd/plumego && go test -timeout 20s ./commands/...`
- `scripts/check-spec tasks/cards/done/0589-runnable-scenario-scaffold-edge-realtime.md`

Docs Sync:
- Required because scaffold output changes.

Done Definition:
- `gateway` and `realtime` generated projects expose runnable scenario routes.
- Profiles remain explicit and opt-in.

Outcome:
- Upgraded the `gateway` scaffold profile with a runnable `/edge` loopback
  proxy route using `x/gateway`.
- Upgraded the `realtime` scaffold profile with a runnable
  `/realtime/metrics` route backed by an `x/websocket` hub.
- Updated scaffold tests and docs to reflect runnable gateway/realtime scenario
  routes while keeping profiles opt-in.

Validations:
- `cd cmd/plumego && go test -timeout 20s ./internal/scaffold/...`
- `cd cmd/plumego && go test -timeout 20s ./commands/...`
- `scripts/check-spec tasks/cards/done/0589-runnable-scenario-scaffold-edge-realtime.md`

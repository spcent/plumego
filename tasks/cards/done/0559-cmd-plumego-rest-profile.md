# Card 0559

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: done
Primary Module: cmd/plumego
Owned Files:
- cmd/plumego/internal/scaffold/scaffold.go
- cmd/plumego/internal/scaffold/scaffold_test.go
- cmd/plumego/commands/new.go
- cmd/plumego/README.md
Depends On: 2268

Goal:
Add or harden a REST API scaffold profile that starts from canonical bootstrap and adds explicit resource-route wiring.

Scope:
- Define the generated REST profile in the CLI scaffold layer.
- Keep stable-root bootstrap identical to the canonical template.
- Add only the minimal `x/rest` example wiring needed to demonstrate resource APIs.
- Cover generated output with parseability and no-bare-TODO tests.

Non-goals:
- Do not add tenant, gateway, AI, or realtime profiles in this card.
- Do not make `x/rest` part of the default template.
- Do not change `x/rest` APIs.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/commands/new.go`
- `cmd/plumego/README.md`

Tests:
- `go test -timeout 20s ./cmd/plumego/internal/scaffold/...`
- `go test -timeout 20s ./cmd/plumego/commands/...`
- `go test -timeout 20s ./x/rest/...`

Docs Sync:
- Required in CLI README once the profile exists.

Done Definition:
- `plumego new --template api` or its documented equivalent generates a canonical app plus explicit REST wiring.
- The generated code is parseable, TODO-free, and does not imply `x/rest` owns bootstrap.

Outcome:
Completed. Reworked the `api` scaffold template to start from the canonical app
layout and add a minimal `x/rest` users resource profile with explicit route
wiring in `internal/app/routes.go`. Added scaffold tests covering the canonical
bootstrap, absence of legacy `internal/httpapp` output, and parseable REST
resource files. Updated CLI docs for the implemented template behavior.

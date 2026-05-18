# Card 1560

Milestone: M-016
Recipe: specs/change-recipes/add-package.yaml
Priority: P2
State: done
Primary Module: reference/with-events
Owned Files:
- `reference/with-events/main.go`
- `reference/with-events/go.mod`
- `reference/with-events/internal/config/config.go`
- `reference/with-events/internal/app/app.go`
- `reference/with-events/internal/app/routes.go`

Goal:
- Scaffold the reference/with-events application with core.App wiring, an
  in-process messaging service, and the route group structure that subsequent
  cards (1561–1563) will populate.

Scope:
- Create reference/with-events/go.mod with module
  github.com/spcent/plumego/reference/with-events; require stable roots and
  x/messaging.
- Create main.go following the reference/standard-service pattern: load config,
  build dependencies, call app.New(), app.RegisterRoutes(), then
  app.Start(ctx).
- Create internal/config/config.go: Config struct with Addr string,
  LogLevel string, WebhookTargetURL string (optional, for card 1563).
- Create internal/app/app.go: App struct holding core.App, Logger, and
  messaging.Service; constructor New(cfg Config, deps Deps).
- Create internal/app/routes.go: RegisterRoutes(app *App) function with
  placeholder route groups for /orders, /scheduler, /webhook — populated
  by later cards.
- Confirm `go build ./...` succeeds with empty route handlers.

Non-goals:
- Do not implement business logic in this card (that is cards 1561–1563).
- Do not add any database dependency.
- Do not add routes beyond placeholder stubs.

Files:
- `reference/with-events/main.go`
- `reference/with-events/go.mod`
- `reference/with-events/internal/config/config.go`
- `reference/with-events/internal/app/app.go`
- `reference/with-events/internal/app/routes.go`

Tests:
- `go build ./reference/with-events/...`
- `go vet ./reference/with-events/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- none at this card; README.md added in a later card.

Done Definition:
- `go build ./reference/with-events/...` exits 0.
- App struct wires core.App and messaging.Service via constructor.
- Route groups for /orders, /scheduler, /webhook are registered (may be empty).
- `go run ./internal/checks/reference-layout` exits 0.

Outcome:
- Added `reference/with-events` as a separate scenario module with local
  replace wiring to the Plumego checkout.
- Implemented config loading, main bootstrap, `App` construction with
  `core.App`, logger, in-process messaging broker, and `messaging.Service`.
- Added placeholder `/orders`, `/scheduler`, and `/webhook` route groups for
  later M-016 cards.
- Validation passed with `reference/with-events` build, `reference/with-events`
  vet, reference-layout, agent-workflow, and `git diff --check`.

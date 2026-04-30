# Getting Started

Plumego is a standard-library-first HTTP toolkit for Go.

Use this page for the smallest runnable example. For the canonical application
layout, read `reference/standard-service` immediately after this page.

## Requirements

- Go 1.24 or later
- no external runtime dependencies in the main module

Install:

```bash
go get github.com/spcent/plumego@latest
```

## Smallest Runnable Example

Create `main.go`:

```go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	plog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
)

func main() {
	ctx := context.Background()
	cfg := core.DefaultConfig()
	cfg.Addr = ":8080"
	app := core.New(cfg, core.AppDependencies{Logger: plog.NewLogger()})

	if err := app.Use(
		requestid.Middleware(),
		recovery.Recovery(app.Logger()),
	); err != nil {
		log.Fatalf("register middleware: %v", err)
	}

	if err := app.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		if err := contract.WriteResponse(w, r, http.StatusOK, map[string]string{
			"message": "hello",
		}, nil); err != nil {
			http.Error(w, "write response", http.StatusInternalServerError)
		}
	}); err != nil {
		log.Fatalf("register route: %v", err)
	}

	if err := app.Prepare(); err != nil {
		log.Fatalf("prepare server: %v", err)
	}
	srv, err := app.Server()
	if err != nil {
		log.Fatalf("get server: %v", err)
	}
	defer app.Shutdown(ctx)

	log.Println("server started at :8080")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped: %v", err)
	}
}
```

Run it:

```bash
go run main.go
```

Open `http://localhost:8080/hello`.

## Basic Routing

Plumego uses the standard `net/http` handler shape.

```go
app.Get("/users", func(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("list users"))
})

app.Post("/users", func(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("create user"))
})
```

For path parameters, use `router.Param`:

```go
import "github.com/spcent/plumego/router"

app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	_, _ = w.Write([]byte(id))
})
```

## Choose Your Next Module

After the smallest example works, keep the app layout from
`reference/standard-service` and add only the capability family you need:

| Goal | First module |
| --- | --- |
| Standard JSON API with explicit handlers | stable roots: `core`, `router`, `contract`, `middleware` |
| Reusable CRUD/resource conventions | `x/rest` |
| Tenant resolution, policy, quota, and isolation | `x/tenant` |
| Reverse proxy, rewrite, balancing, and edge transport | `x/gateway` |
| WebSocket transport | `x/websocket` |
| Messaging workflows | `x/messaging` |
| AI providers, sessions, streaming, and tools | `x/ai/provider`, `x/ai/session`, `x/ai/streaming`, `x/ai/tool` |
| Observability export or protected diagnostics | `x/observability`, `x/ops`, or `x/devtools` depending on the surface |

Do not start a new application layout from an `x/*` package. Treat extensions as
explicit additions to the canonical app wiring.

For generated project scaffolds, keep `canonical` as the default template and
select a scenario profile only when you want explicit optional capability
wiring. Baseline templates are `canonical`, `minimal`, `api`, `fullstack`, and
`microservice`. Scenario templates are `rest-api`, `tenant-api`, `gateway`,
`realtime`, `ai-service`, and `ops-service`; each is accepted by
`plumego new --template`.

The `canonical` scaffold follows the same runtime wiring as
`reference/standard-service`, but it is shaped as a generated project: the
entrypoint is `cmd/app/main.go`, project-local files are generated, and
reference-only tests are not copied.

```bash
plumego new myapi --template rest-api
plumego new tenantapi --template tenant-api
plumego new edge --template gateway
plumego new realtime --template realtime
plumego new ai --template ai-service
plumego new ops --template ops-service
```

## Canonical Next Reads

Read these next, in order:

1. `reference/standard-service`
2. `docs/README.md`
3. `docs/CANONICAL_STYLE_GUIDE.md`
4. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
5. `specs/task-routing.yaml`
6. the target module's `module.yaml`

## Notes

- `reference/standard-service` is the canonical application layout.
- `docs/README.md` is the entrypoint for the human-readable docs surface.
- `x/*` packages are optional capability families, not the default bootstrap path.
- Prefer explicit route wiring and explicit middleware registration.
- Prefer `contract.WriteError` and `contract.WriteResponse` for JSON APIs.
- When the example stops being enough, copy structure from `reference/standard-service` instead of inventing a new bootstrap shape.

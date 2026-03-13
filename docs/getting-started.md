# Getting Started

Plumego is a lightweight HTTP toolkit for Go, built entirely on the standard library.

## Installation

```bash
go get github.com/spcent/plumego@latest
```

Requires Go 1.24 or later. Plumego has zero external dependencies.

## Hello World

Create a file called `main.go`:

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/spcent/plumego/core"
)

func main() {
	app := core.New(core.WithAddr(":8080"))

	if err := app.AddRoute(http.MethodGet, "/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})); err != nil {
		log.Fatal(err)
	}

	if err := app.Prepare(); err != nil {
		log.Fatal(err)
	}
	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
	srv, err := app.Server()
	if err != nil {
		log.Fatal(err)
	}
	defer app.Shutdown(context.Background())

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

Run it:

```bash
go run main.go
```

Visit `http://localhost:8080/hello` to see the response.

## Basic Routing

Plumego supports all standard HTTP methods. Handlers use the standard `net/http` signature.

```go
app.Get("/users", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("list users"))
})

app.Post("/users", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create user"))
})

app.Delete("/users/:id", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("deleted"))
})
```

### Path Parameters

Use `:name` in route paths to capture parameters:

```go
import "github.com/spcent/plumego/router"

app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":   id,
		"name": "Alice",
	})
})

app.Put("/users/:id", func(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"updated": "true",
	})
})
```

### Route Groups

Group related routes under a shared prefix using the router directly:

```go
api := app.Router().Group("/api/v1")

api.Get("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"status":"ok"}`))
}))

api.Get("/version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"version":"1.0.0"}`))
}))
```

## Running with Middleware

Register middleware explicitly before `Prepare()`:

```go
import (
	"log"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/recovery"
)

app := core.New(
	core.WithAddr(":8080"),
	core.WithLogger(plumelog.NewGLogger()),
)
if err := app.Use(
	requestid.Middleware(),
	recovery.Recovery(app.Logger()),
); err != nil {
	log.Fatalf("register middleware: %v", err)
}
```

## Next Steps

- Browse `reference/standard-service/` for a full-featured application
- See `examples/` for focused examples on routing, middleware, caching, and more
- Read the source at `core/`, `router/`, and `contract/` for API details

## Canonical Note

Plumego's application route registration uses the standard `net/http` handler shape as the canonical and only supported style for `core.App` APIs.

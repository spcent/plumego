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
	"log"
	"net/http"

	"github.com/spcent/plumego"
)

func main() {
	app := plumego.New(plumego.WithAddr(":8080"))

	app.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	if err := app.Boot(); err != nil {
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

Use `:name` in route paths to capture parameters. Context-aware handlers (`GetCtx`, `PostCtx`, etc.) provide a `*plumego.Context` with helpers for parameter access and JSON responses:

```go
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
	id, _ := ctx.Param("id")
	ctx.JSON(http.StatusOK, map[string]string{
		"id":   id,
		"name": "Alice",
	})
})

app.PutCtx("/users/:id", func(ctx *plumego.Context) {
	id, _ := ctx.Param("id")
	ctx.JSON(http.StatusOK, map[string]string{
		"id":      id,
		"updated": "true",
	})
})
```

### Route Groups

Group related routes under a shared prefix using the router directly:

```go
api := app.Router().Group("/api/v1")

api.GetFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"status":"ok"}`))
})

api.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"version":"1.0.0"}`))
})
```

## Running with Middleware

Enable common middleware using functional options:

```go
app := plumego.New(
	plumego.WithAddr(":8080"),
	plumego.WithRecommendedMiddleware(), // RequestID + Logging + Recovery
)
```

## Next Steps

- Browse `examples/reference/` for a full-featured application
- See `examples/` for focused examples on routing, middleware, caching, and more
- Read the source at `core/`, `router/`, and `contract/` for API details

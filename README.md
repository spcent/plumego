# Minimal Go HTTP server with routing, middleware, and graceful shutdown

## Overview

A lightweight HTTP server skeleton built with the Go standard library.
It includes dynamic routing, middleware, graceful shutdown, and environment-based configuration — ideal for quickly building simple APIs or learning HTTP server development.

## Features

* **Dynamic Routing**: Supports path parameters like `/hello/:name`
* **Middleware System**: Built-in Logging and Auth middlewares, with support for custom extensions
* **Graceful Shutdown**: Cleans up connections within 5 seconds after receiving an interrupt signal
* **Logging**: Add glog logging library for structured logging
* **Env Configuration**: Supports `.env` files (e.g., log level, etc.)
* **Test Coverage**: Includes tests for routing, middleware, 404 handling, and more
* **Developer Toolchain**: Comes with a `Makefile` and `dev.sh` script for easy build/run/test workflows
* **Static Frontend Hosting**: Serve built Node/Next.js bundles directly from the Go server (flag: `-frontend-dir`, env: `FRONTEND_DIR`)

## Getting Started

### Requirements

* Go 1.18+

### Build & Run

```bash
# Using Makefile
make run   # Build and start the server (default port: 8080)

# Or using the dev.sh script
./dev.sh run
```

### Custom Port

```bash
./plumego -addr :9090  # Start on port 9090
```

### Serve a built Node/Next.js frontend

Point the server at a production build directory (for example `next export` output in `./web/out`):

```bash
./plumego -frontend-dir ./web/out
# or
FRONTEND_DIR=./web/out ./plumego
```

The `-frontend-dir` flag has the highest priority and overrides `FRONTEND_DIR`.

All assets under the directory are served with SPA-style fallback to `index.html`, making it suitable for exported Next.js/Vite/React builds.

To ship a single self-contained binary, copy your built frontend (the same `frontend-dir` contents) into `pkg/frontend/embedded/` before building:

```bash
cp -r ./web/out/* pkg/frontend/embedded/
go build ./...
./plumego  # mounts embedded assets when no frontend dir/env is provided
```

## Routing

Register parameterized routes with `app.(Get|Post|Delete|Put)` (see `main.go` for examples):

```go
app.Get("/hello/:name", func(w http.ResponseWriter, r *http.Request) {
    params := router.ParamsFromContext(r.Context())
    // Access params["name"]
    fmt.Fprintf(w, `{"message":"Hello, %s!"}`, params["name"])
})

app.Get("/users/:id/posts/:postID", func(w http.ResponseWriter, r *http.Request) {
    params := router.ParamsFromContext(r.Context())
    // Access params["id"] and params["postID"]
    fmt.Fprintf(w, `{"message":"User %s, Post %s"}`, params["id"], params["postID"])
})
```

## Route Testing
After starting the service, test the routes using curl:
```bash
curl http://localhost:8080/ping # pong
curl http://localhost:8080/hello # {"message":"Hello, World!"}
curl -H "X-Token: secret" http://localhost:8080/hello/Alice # {"message":"Hello, Alice!"}
```

## Middleware

* **LoggingMiddleware**: Logs request duration (`[time] METHOD PATH (duration)`)
* **AuthMiddleware**: Validates `X-Token` header
* **CorsMiddleware**: Allow cross-domain requests

Combine middlewares (see `main.go` for usage):

```go
app.Use(app.Logging(), app.Auth())
```

## Testing

```bash
make test       # Run all tests
make coverage   # Generate coverage report (outputs to coverage.html)
```

## Usage

* **[English](docs/en/usage.md)**: Comprehensive guide with examples
* **[中文](docs/cn/usage.md)**: 中文文档，包含详细的使用说明和示例

## License
MIT
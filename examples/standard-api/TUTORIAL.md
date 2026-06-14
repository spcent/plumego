# Standard API Example Tutorial

This is the canonical example of a production-ready Plumego service. It shows how to structure a real service with proper logging, configuration, graceful shutdown, and a complete CRUD API.

## Running the example

```bash
cd examples/standard-api
go run . -config env.example
```

You should see logs like:
```
2026/06/14 21:00:00 config loaded: addr=:8080
2026/06/14 21:00:00 server started
```

## Testing the API

In another terminal, create a post:

```bash
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{"title":"Hello","content":"World"}'
```

List posts:

```bash
curl http://localhost:8080/posts
```

Get one post:

```bash
curl http://localhost:8080/posts/1
```

Update a post:

```bash
curl -X PUT http://localhost:8080/posts/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated","content":"Content"}'
```

Delete a post:

```bash
curl -X DELETE http://localhost:8080/posts/1
```

Check health:

```bash
curl http://localhost:8080/health
```

## What you'll learn from this code

### Structure

The example follows this layout:

```
cmd/api/main.go          # Entry point: load config, start server
internal/config/         # Config loading and validation
internal/app/            # HTTP wiring: routes, middleware, lifecycle
internal/handler/        # Route handlers organized by domain
internal/domain/         # Business logic (not shown here, but where it goes)
```

### Key files to read

1. **`main.go`** — Entry point. Shows how to:
   - Load configuration
   - Construct the app with `core.New()`
   - Attach middleware (logging, recovery, request ID)
   - Call `app.Prepare()` to freeze routes
   - Start the server with graceful shutdown

2. **`internal/app/routes.go`** — Route registration. Shows:
   - `POST /posts` — create a new post
   - `GET /posts` — list all posts
   - `GET /posts/{id}` — get one post
   - `PUT /posts/{id}` — update a post
   - `DELETE /posts/{id}` — delete a post
   - Error handling with `contract.WriteError()`

3. **`internal/handler/write.go`** — Handler patterns. Shows:
   - How to extract path params: `contract.Param(r, "id")`
   - How to decode JSON: `contract.BindJSON(r, &data)`
   - How to send responses: `contract.WriteResponse(w, r, status, data, nil)`
   - How to send errors: `contract.WriteError(w, r, errors.New("not found"))`

4. **`internal/config/config.go`** — Configuration. Shows:
   - Loading from files or environment
   - Validation on startup
   - Defaults for sensible behaviors

### Patterns to copy

**Error handling:**
```go
if err != nil {
    return contract.WriteError(w, r, err)
}
```

**JSON responses:**
```go
contract.WriteResponse(w, r, http.StatusOK, map[string]interface{}{
    "posts": posts,
}, nil)
```

**Request binding:**
```go
var req struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}
if err := contract.BindJSON(r, &req); err != nil {
    return contract.WriteError(w, r, err)
}
```

**Path params:**
```go
id := contract.Param(r, "id")
```

## Customizing for your service

1. Copy this directory to your project
2. Rename the module in `go.mod`
3. Change routes in `internal/app/routes.go`
4. Add handlers in `internal/handler/`
5. Add business logic in `internal/domain/`
6. Update configuration in `internal/config/`

## Next steps

Once you understand this example:

1. Read `docs/start/production-checklist.md` to understand what else a production service needs (metrics, health checks, tracing, etc.)
2. Add one extension family (e.g., `x/observability` for Prometheus metrics)
3. Read `docs/reference/canonical-style-guide.md` for deeper patterns
4. Check `docs/modules/` for details on any package you want to learn more about

## Questions?

- How do I add a database? See `docs/modules/store/` and pick a store (in-memory, PostgreSQL, etc.)
- How do I add authentication? See `docs/modules/security/` for auth adapters.
- How do I add metrics? See `docs/modules/x/observability/` for Prometheus/OTel integration.
- How do I structure a larger service? See `docs/reference/canonical-style-guide.md` §3 for multi-domain layout.

---

This example is meant to be copied, not just read. Treat it as a template and adapt it to your needs.

# Hello World Example

The simplest Plumego service: run it, hit one endpoint, see JSON response.

## Run it

```bash
cd examples/hello
go run main.go
```

You should see:
```
server started at :8080
```

## Test it

In another terminal:

```bash
curl http://localhost:8080/ping
```

Response:
```json
{"message":"pong"}
```

## What's happening

- `plumego.New()` creates an app with default config (`:8080`, 30s timeouts, HTTP/2)
- `app.Get("/ping", ...)` registers a route
- `contract.WriteResponse(w, r, 200, data, nil)` sends a structured JSON response
- `http.ListenAndServe(":8080", app)` starts the server (app is a standard `http.Handler`)

## Next steps

Once this works:

1. Add more routes (`app.Post()`, `app.Delete()`, etc.)
2. Copy `reference/standard-service` for a production layout
3. Read `docs/start/adoption-path.md` for the full learning path

## Customizing the address or port

Edit `main.go`:

```go
cfg := plumego.DefaultConfig()
cfg.Addr = ":9090"
app := plumego.NewWithConfig(cfg)
```

Or set an environment variable (with custom code in your config module).

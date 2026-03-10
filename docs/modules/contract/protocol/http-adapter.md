# HTTP Adapter Pattern

This page shows how protocol adapters work at the HTTP boundary.

## Minimal Adapter Skeleton

```go
type MyAdapter struct{}

func (a *MyAdapter) Name() string { return "my-protocol" }

func (a *MyAdapter) Handles(req *protocol.HTTPRequest) bool {
    return strings.HasPrefix(req.URL, "/rpc")
}

func (a *MyAdapter) Transform(ctx context.Context, req *protocol.HTTPRequest) (protocol.Request, error) {
    // parse HTTP request -> protocol request
    return myRequest{...}, nil
}

func (a *MyAdapter) Execute(ctx context.Context, req protocol.Request) (protocol.Response, error) {
    // call backend
    return myResponse{...}, nil
}

func (a *MyAdapter) Encode(ctx context.Context, resp protocol.Response, w protocol.ResponseWriter) error {
    w.WriteHeader(resp.StatusCode())
    _, err := io.Copy(responseBodyWriter{w}, resp.Body())
    return err
}
```

## Register Adapter

```go
registry := protocol.NewRegistry()
registry.Register(&MyAdapter{})

if err := app.Use(protomw.Middleware(registry)); err != nil {
    log.Fatal(err)
}
```

## Pass-through Behavior

If no adapter handles a request, middleware forwards request to the next HTTP handler in chain.

## Recommended Boundaries

- Keep transport concerns (headers/body/status mapping) in adapter.
- Keep business logic in service layer, not adapter.
- Return deterministic errors; avoid panic paths.

# Protocol Contract

> Package: `github.com/spcent/plumego/contract/protocol`

`contract/protocol` defines adapter contracts so Plumego can gateway gRPC/GraphQL/custom protocols without adding hard dependencies to core.

## Core Interfaces

- `ProtocolAdapter`
- `Request`
- `Response`
- `ResponseWriter`
- `Registry`

## Gateway Wiring

```go
registry := protocol.NewRegistry()
registry.Register(grpcAdapter)
registry.Register(graphqlAdapter)

if err := app.Use(protomw.Middleware(registry)); err != nil {
    log.Fatal(err)
}
```

Requests unmatched by adapters pass through to normal HTTP routes.

## Adapter Lifecycle

1. `Handles(req)` chooses whether adapter takes this request.
2. `Transform(ctx, req)` maps HTTP to protocol request.
3. `Execute(ctx, req)` runs protocol call.
4. `Encode(ctx, resp, writer)` maps protocol response to HTTP response.

## Built-in Helper Types

- `protocol.GRPCRequest` / `protocol.GRPCResponse`
- `protocol.GraphQLRequest` / `protocol.GraphQLResponse`
- `protocol.GRPCErrorCode(...)`

These helpers are optional, but useful for consistency.

## Error Mapping

`middleware/protocol` emits structured transport errors when transform/execute/encode fails.
Keep adapter errors explicit and avoid leaking backend internals.

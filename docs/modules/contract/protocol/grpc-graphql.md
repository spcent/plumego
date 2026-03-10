# gRPC and GraphQL Adapter Notes

`contract/protocol` includes helper contracts for gRPC and GraphQL adapters without importing those ecosystems into core.

## gRPC Adapter

Implement `protocol.GRPCAdapter`:

```go
type MyGRPCAdapter struct{}

func (a *MyGRPCAdapter) Name() string { return "grpc" }
func (a *MyGRPCAdapter) Service() string { return "user.v1.UserService" }
func (a *MyGRPCAdapter) Methods() []string { return []string{"GetUser", "ListUsers"} }

// plus Handles/Transform/Execute/Encode from ProtocolAdapter
```

Use `protocol.GRPCRequest` / `protocol.GRPCResponse` helpers and `protocol.GRPCErrorCode(code)` for HTTP status mapping.

## GraphQL Adapter

Implement `protocol.GraphQLAdapter`:

```go
type MyGraphQLAdapter struct{}

func (a *MyGraphQLAdapter) Name() string { return "graphql" }
func (a *MyGraphQLAdapter) Schema() string { return schemaSDL }
func (a *MyGraphQLAdapter) Validate(query string) error { return nil }

// plus Handles/Transform/Execute/Encode from ProtocolAdapter
```

Use `protocol.GraphQLRequest` / `protocol.GraphQLResponse` helpers for operation, variables, and error metadata.

## Wiring Example

```go
registry := protocol.NewRegistry()
registry.Register(grpcAdapter)
registry.Register(graphqlAdapter)

if err := app.Use(protomw.Middleware(registry)); err != nil {
    log.Fatal(err)
}
```

## Production Guidance

- Keep adapter matching rules explicit (`Handles`).
- Enforce auth/rate-limit at HTTP middleware layer before adapter execution.
- Normalize adapter errors to stable API responses.

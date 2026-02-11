# Protocol Adapters

> **Package**: `github.com/spcent/plumego/contract/protocol`

Protocol adapters allow Plumego to handle multiple protocols (HTTP, gRPC, GraphQL) using the same handler logic.

---

## Overview

Protocol adapters provide:
- **Unified Interface**: Same handler for different protocols
- **Request Translation**: Convert protocol-specific requests to Plumego context
- **Response Translation**: Convert Plumego responses back to protocol format
- **Middleware Compatibility**: Use same middleware across protocols

---

## Supported Protocols

| Protocol | Status | Adapter |
|----------|--------|---------|
| **HTTP/REST** | ✅ Native | Built-in |
| **gRPC** | ✅ Available | `protocol.GRPCAdapter` |
| **GraphQL** | ✅ Available | `protocol.GraphQLAdapter` |
| **WebSocket** | ✅ Native | Built-in |

---

## Quick Start

### HTTP (Default)

```go
// No adapter needed - native support
app.GetCtx("/users", func(ctx *plumego.Context) {
    ctx.JSON(200, users)
})
```

### gRPC

```go
import "github.com/spcent/plumego/contract/protocol"

// Create adapter
grpcAdapter := protocol.NewGRPCAdapter()

// In gRPC service
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // Convert to Plumego context
    pctx := grpcAdapter.ToContext(req)

    // Use standard handler
    handler(pctx)

    // Convert response
    return grpcAdapter.FromContext(pctx)
}
```

### GraphQL

```go
import "github.com/spcent/plumego/contract/protocol"

gqlAdapter := protocol.NewGraphQLAdapter()

// In GraphQL resolver
func (r *Resolver) User(ctx context.Context, args struct{ ID string }) (*User, error) {
    pctx := gqlAdapter.ToContext(ctx, args)
    handler(pctx)
    return gqlAdapter.FromContext(pctx)
}
```

---

## Documentation

- **[HTTP Adapter](http-adapter.md)** - HTTP/REST protocol (default)
- **[gRPC/GraphQL](grpc-graphql.md)** - gRPC and GraphQL adapters

---

## Use Cases

### 1. Multi-Protocol API

Expose same business logic via multiple protocols:

```go
// Business logic
func getUser(ctx *plumego.Context) {
    id := ctx.Param("id")
    user := loadUser(id)
    ctx.JSON(200, user)
}

// HTTP endpoint
app.GetCtx("/users/:id", getUser)

// gRPC service
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    pctx := grpcAdapter.ToContext(req)
    getUser(pctx)
    return grpcAdapter.FromContext(pctx)
}
```

### 2. Protocol Migration

Gradually migrate from REST to gRPC:

```go
// Same handler for both
handler := func(ctx *plumego.Context) {
    // Business logic
}

// REST (legacy)
app.GetCtx("/api/users", handler)

// gRPC (new)
grpcService.GetUsers = grpcAdapter.Wrap(handler)
```

---

**Next**: [HTTP Adapter](http-adapter.md) | [gRPC/GraphQL](grpc-graphql.md)

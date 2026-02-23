# gRPC and GraphQL Adapters

> **Package**: `github.com/spcent/plumego/contract/protocol`

Protocol adapters for gRPC and GraphQL allow reusing Plumego handlers across multiple protocols.

---

## Table of Contents

- [gRPC Adapter](#grpc-adapter)
- [GraphQL Adapter](#graphql-adapter)
- [Use Cases](#use-cases)

---

## gRPC Adapter

### Overview

The gRPC adapter converts between gRPC requests/responses and Plumego context.

### Setup

```go
import "github.com/spcent/plumego/contract/protocol"

// Create adapter
grpcAdapter := protocol.NewGRPCAdapter()
```

### Basic Usage

```go
// Protobuf definition
// message GetUserRequest {
//   string id = 1;
// }
// message User {
//   string id = 1;
//   string name = 2;
// }

// gRPC service implementation
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // Convert gRPC request to Plumego context
    pctx := grpcAdapter.ToContext(req)

    // Use standard Plumego handler
    user := getUserHandler(pctx)

    // Convert back to gRPC response
    return &pb.User{
        Id:   user.ID,
        Name: user.Name,
    }, nil
}
```

### With Middleware

```go
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    pctx := grpcAdapter.ToContext(req)

    // Apply middleware
    handler := authMiddleware(
        http.HandlerFunc(getUserHandler),
    )

    handler.ServeHTTP(pctx.Writer, pctx.Request)

    return grpcAdapter.FromContext(pctx)
}
```

### Parameter Mapping

```go
// gRPC request fields map to context parameters
req := &pb.GetUserRequest{
    Id: "123",
}

pctx := grpcAdapter.ToContext(req)
id := pctx.Param("id") // "123"
```

---

## GraphQL Adapter

### Overview

The GraphQL adapter converts GraphQL context and arguments to Plumego context.

### Setup

```go
import "github.com/spcent/plumego/contract/protocol"

gqlAdapter := protocol.NewGraphQLAdapter()
```

### Basic Usage

```go
// GraphQL schema
// type Query {
//   user(id: ID!): User
// }

// Resolver
func (r *Resolver) User(ctx context.Context, args struct{ ID string }) (*User, error) {
    // Convert to Plumego context
    pctx := gqlAdapter.ToContext(ctx, args)

    // Use standard handler
    user := getUserHandler(pctx)

    return user, nil
}
```

### With Arguments

```go
type UserArgs struct {
    ID   string
    Name string
}

func (r *Resolver) User(ctx context.Context, args UserArgs) (*User, error) {
    pctx := gqlAdapter.ToContext(ctx, args)

    // Access arguments
    id := pctx.Param("ID")
    name := pctx.Query("Name")

    // Use handler
    return getUserHandler(pctx)
}
```

---

## Use Cases

### 1. Multi-Protocol API

```go
// Shared business logic
func getUser(ctx *plumego.Context) *User {
    id := ctx.Param("id")
    return loadUserFromDB(id)
}

// HTTP endpoint
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    user := getUser(ctx)
    ctx.JSON(200, user)
})

// gRPC service
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    pctx := grpcAdapter.ToContext(req)
    user := getUser(pctx)
    return &pb.User{Id: user.ID, Name: user.Name}, nil
}

// GraphQL resolver
func (r *Resolver) User(ctx context.Context, args struct{ ID string }) (*User, error) {
    pctx := gqlAdapter.ToContext(ctx, args)
    return getUser(pctx), nil
}
```

### 2. Protocol Migration

Gradually migrate from REST to gRPC without rewriting business logic:

```go
// Phase 1: REST only
app.GetCtx("/api/users/:id", getUserHandler)

// Phase 2: Add gRPC (same handler)
grpcService.GetUser = grpcAdapter.Wrap(getUserHandler)

// Phase 3: Deprecate REST, gRPC only
// Remove REST endpoint, keep gRPC
```

### 3. API Gateway

Use Plumego as a protocol gateway:

```go
// Gateway receives REST, forwards as gRPC
app.GetCtx("/api/users/:id", func(ctx *plumego.Context) {
    // Convert to gRPC request
    req := &pb.GetUserRequest{Id: ctx.Param("id")}

    // Call backend gRPC service
    resp, err := grpcClient.GetUser(context.Background(), req)
    if err != nil {
        ctx.Error(500, err.Error())
        return
    }

    // Return as REST response
    ctx.JSON(200, resp)
})
```

---

## Limitations

### gRPC Adapter

- Streaming RPCs not yet supported
- Metadata mapping is simplified
- Error codes need manual translation

### GraphQL Adapter

- Subscriptions require custom implementation
- N+1 query problem must be handled separately
- Dataloader pattern recommended for batching

---

## Best Practices

### ✅ Do

1. **Reuse Business Logic**
   ```go
   // ✅ Shared handler
   func getUser(ctx *plumego.Context) *User {
       // Business logic
   }

   // Use in both HTTP and gRPC
   ```

2. **Handle Protocol-Specific Errors**
   ```go
   // ✅ Convert errors appropriately
   if err != nil {
       if isGRPC {
           return status.Error(codes.NotFound, err.Error())
       }
       ctx.Error(404, err.Error())
   }
   ```

### ❌ Don't

1. **Don't Mix Protocol Logic**
   ```go
   // ❌ Protocol-specific code in handler
   func handler(ctx *plumego.Context) {
       if isGRPC {
           // gRPC-specific logic
       } else {
           // HTTP-specific logic
       }
   }

   // ✅ Keep protocol-agnostic
   func handler(ctx *plumego.Context) {
       // Pure business logic
   }
   ```

---

**Next**: [Contract Overview](../README.md)

# x/rpc

`x/rpc` is the optional RPC transport family for Plumego applications. It keeps
the Plumego-facing surface dependency-free by exposing lifecycle, client-pool,
interceptor, and HTTP-adapter contracts for caller-owned transports.

Start with `x/rpc/server` when adding RPC service hosting. Application service
definitions, concrete transport adapters, and route or bootstrap wiring belong
in reference apps or user applications, not in stable roots.

Use `x/rpc/client` for target-keyed connection reuse and transport-neutral unary
interceptors. Use `x/rpc/gateway` when a service needs to mount caller-owned RPC
handlers as ordinary `net/http` handlers.

Concrete dependencies such as gRPC, Connect, or generated clients should live in
application code or out-of-tree adapters.

Validation:

```bash
go test -timeout 20s ./x/rpc/...
```

# x/rpc

`x/rpc` is the optional RPC transport family for Plumego applications. It keeps
gRPC dependencies in a separate module so stable roots and the main module stay
standard-library focused.

Start with `x/rpc/server` when adding gRPC service hosting. Application service
definitions and route or bootstrap wiring belong in reference apps or user
applications, not in stable roots.

Use `x/rpc/client` for target-keyed connection reuse and unary interceptors.
Use `x/rpc/gateway` when a service needs to mount HTTP routes over a gRPC
backend through grpc-gateway's runtime mux.

`reference/with-rpc` shows the pieces wired together with bufconn so the example
can run without binding a real gRPC port.

Validation:

```bash
cd x/rpc
go test -timeout 20s ./...
```

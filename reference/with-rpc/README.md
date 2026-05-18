# with-rpc Scenario Reference

`reference/with-rpc` shows how to host a gRPC service with `x/rpc/server` and
mount an HTTP endpoint through `x/rpc/gateway`.

The example uses `bufconn`, so it can run without binding a real gRPC network
port:

```bash
cd reference/with-rpc
go run ./...
```

The HTTP side is wired through `core.App`; the sample request is executed with
`httptest` and then both the HTTP app and gRPC server are shut down through
context-aware lifecycle calls.

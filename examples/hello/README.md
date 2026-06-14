# Hello World Example

Minimal Plumego "hello world" service demonstrating:

- Creating an app with `core.New()`
- Registering routes with `app.Get()`
- Using standard `http.HandlerFunc` handlers
- Returning JSON responses with `contract.WriteResponse()`
- Structured logging

## Run

```bash
go build -o hello .
./hello
```

Then in another terminal:

```bash
curl http://localhost:8080/ping
# {"message":"pong"}

curl http://localhost:8080/echo/alice
# {"hello":"alice"}
```

## Key APIs

- **`core.New(config, dependencies)`** — Create the app
- **`app.Get(path, handler)`** — Register a GET route
- **`contract.WriteResponse(w, r, statusCode, data, err)`** — Write a JSON response
- **`logger.InfoCtx(ctx, msg, fields)`** — Log with context

All handlers use standard `http.HandlerFunc` signature: `func(http.ResponseWriter, *http.Request)`

## Next Steps

For a more complete example with storage and middleware, see [`reference/standard-service`](../../reference/standard-service).

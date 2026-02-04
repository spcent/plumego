# SMS Gateway Example

This folder contains reference material for building an SMS gateway using Plumego.
It is not a runnable service yet; the code is illustrative.

## What's Included

- `feature.md`: product and system design notes
- `internal/message/`: status model, reason normalization, transition hooks
- `internal/message/example_handler.go`: wiring idempotency + state transitions in a handler

## Handler Example Wiring

The example handler lives at:

- `examples/sms-gateway/internal/message/example_handler.go`

It demonstrates:

- Idempotency key validation (`Idempotency-Key` header)
- Request hash comparison to prevent replay with different payloads
- Status transition from `accepted` → `queued`
- Optional transition hooks for audit logging

### Minimal usage sketch

```go
idem := idempotency.NewKVStore(kv, idempotency.DefaultKVConfig())
repo := newMessageRepo(db) // your implementation

app.PostCtx("/v1/messages", message.ExampleSendHandler(idem, repo))
```

### Notes

- The example uses `contract.Ctx` (`ctx.W`, `ctx.R`).
- Replace the `Repository` implementation with your persistence layer.
- For production, use the SQL idempotency store and apply the migrations in `docs/migrations/`.

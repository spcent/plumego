# SMS Gateway Example

This folder contains reference material for building an SMS gateway using Plumego.
It is not a runnable service yet; the code is illustrative.

## What's Included

- `feature.md`: product and system design notes
- `internal/message/`: status model, reason normalization, transition hooks
- `internal/message/example_handler.go`: wiring idempotency + state transitions in a handler (with route policy selection)
- `internal/routing/`: tenant route policy selection with cache support
- `internal/pipeline/`: queue payload + processor for provider sending
- `main.go`: end-to-end demo (enqueue → route policy → provider send)

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
queue := mq.NewTaskQueue(store.NewMemory(store.MemConfig{}))

routePolicyStore := plumego.NewInMemoryRoutePolicyStore()
_ = routePolicyStore.SetRoutePolicy(context.Background(), plumego.TenantRoutePolicy{
    TenantID: "tenant-id",
    Strategy: "weighted",
    Payload:  []byte(`{"providers":[{"provider":"a","weight":70},{"provider":"b","weight":30}]}`),
})
routePolicyCache := plumego.NewInMemoryRoutePolicyCache(1000, 5*time.Minute)
routePolicyProvider := plumego.NewCachedRoutePolicyProvider(routePolicyStore, routePolicyCache)
router := &routing.PolicyRouter{Provider: routePolicyProvider}

app.PostCtx("/v1/messages", message.ExampleSendHandler(idem, repo, router, queue))
```

## End-to-End Demo

Run the demo to see: HTTP → enqueue → route policy → provider send simulation.

```bash
go run ./examples/sms-gateway
```

By default it binds to `127.0.0.1:8089`. Override with:

```bash
SMS_GATEWAY_ADDR=127.0.0.1:9090 go run ./examples/sms-gateway
```

### Optional SQL message store

To persist message state/attempts/DLQ reasons to SQL, set:

```bash
SMS_GATEWAY_MESSAGE_DRIVER=postgres \
SMS_GATEWAY_MESSAGE_DSN="postgres://user:pass@localhost:5432/sms?sslmode=disable" \
go run ./examples/sms-gateway
```

Or for MySQL:

```bash
SMS_GATEWAY_MESSAGE_DRIVER=mysql \
SMS_GATEWAY_MESSAGE_DSN="user:pass@tcp(127.0.0.1:3306)/sms?parseTime=true" \
go run ./examples/sms-gateway
```

Apply the migrations in `docs/migrations/sms_gateway_messages_*.sql` first.
You can override the table name with `SMS_GATEWAY_MESSAGE_TABLE`.
You will also need a Go SQL driver (e.g. `pgx`, `pq`, or `mysql`) in your build.

The service stays running until you stop it. Try:

- Success:
  `curl -X POST http://127.0.0.1:8089/v1/messages -H 'X-Tenant-ID: tenant-1' -H 'Idempotency-Key: demo-1' -d '{"to":"+10000000001","body":"hello"}'`
- Failure → retry → DLQ:
  `curl -X POST http://127.0.0.1:8089/v1/messages -H 'X-Tenant-ID: tenant-2' -H 'Idempotency-Key: demo-2' -d '{"to":"+10000000002","body":"fail","max_attempts":2}'`
- Receipt update (emit receipt delay metric):
  `curl -X POST http://127.0.0.1:8089/v1/receipts -H 'Content-Type: application/json' -d '{"message_id":"<id-from-send>","delivered_at":"2026-02-05T12:00:00Z"}'`

Run the test-only pipeline check:

```bash
go test ./examples/sms-gateway -run TestSMSGatewayPipeline
```

### SQL integration test (optional)

Run the SQL-backed persistence test:

```bash
SMS_GATEWAY_SQL_IT=1 \
SMS_GATEWAY_MESSAGE_DRIVER=postgres \
SMS_GATEWAY_MESSAGE_DSN="postgres://user:pass@localhost:5432/sms?sslmode=disable" \
go test ./examples/sms-gateway -run TestSMSGatewaySQLPersistence
```

You will need a Go SQL driver (e.g. `pgx`, `pq`, or `mysql`) in your build.

### Notes

- The example uses `contract.Ctx` (`ctx.W`, `ctx.R`).
- Replace the `Repository` implementation with your persistence layer.
- For production, use the SQL idempotency store and apply the migrations in `docs/migrations/`.
- The demo emits basic HTTP/DB/MQ metrics and trace IDs via `metrics.OpenTelemetryTracer`.
- SMS gateway-specific metrics (queue depth, send latency, provider results, retries, status distribution) are reported through `metrics/smsgateway`.

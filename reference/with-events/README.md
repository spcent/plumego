# With Events Reference

`reference/with-events` shows how to wire an HTTP app to asynchronous order
events with in-process pubsub, idempotent consumption, delayed retry, and
optional webhook delivery.

```text
HTTP -> handler -> publisher -> pubsub -> consumer -> success: done
                                             |
                                             +-> failure: scheduler -> retry -> publisher
                                             |
                                             +-> webhook sender -> external target
```

## Run

```bash
go run ./...
```

Set `WEBHOOK_TARGET_URL` to deliver consumed order events to an external HTTP
target:

```bash
WEBHOOK_TARGET_URL=http://localhost:9000/webhooks go run ./...
```

The server defaults to `:8083`. Override it with `APP_ADDR` or `--addr`.

## Patterns

The handler accepts an order request and publishes an `orders.created` event.
The publisher uses an in-memory idempotency store as a small outbox-like guard:
duplicate event IDs are ignored before they enter pubsub. A production outbox
would persist these records with the order transaction and flush them to the
broker after commit.

The consumer subscribes to `orders.created` and tracks processed IDs so delivery
is at-least-once from the broker while handler side effects remain idempotent.
If downstream processing fails, the scheduler can enqueue a delayed retry that
publishes the event again after exponential backoff.

Webhook delivery is optional. When `WEBHOOK_TARGET_URL` is empty, the sender is
a no-op. When it is set, the sender posts the JSON order event to the target and
retries retryable failures up to three attempts with 1s and 2s backoff between
attempts.

The app uses `messaging.NewInProcBroker` for local development. To swap in a
real broker, provide an implementation of the same publish/subscribe interfaces
used by the order publisher and consumer, then wire it in `internal/app.New`
instead of the in-process broker.

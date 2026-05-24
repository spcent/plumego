# Workerfleet Notifiers

The workerfleet notifier layer delivers alert records to:

- Feishu webhook endpoints
- generic JSON webhooks

Delivery rules:

- firing alerts enqueue delivery jobs as they are created
- resolved alerts enqueue recovery notification jobs
- one outbox job is stored per alert and sink type
- delivery repairs missing jobs from persisted alert records before it claims work, so transient enqueue failures do not permanently suppress notification attempts
- delivery claims due jobs from the outbox and sends them to the configured sink
- transient failures retry with bounded backoff
- permanent failures are recorded on the outbox job and are not retried
- notifier errors must not include header secrets or raw remote response bodies
- delivery is at-least-once from the app perspective; exactly-once delivery is not claimed

Runtime configuration:

- `WORKERFLEET_NOTIFICATION_ENABLED=true` enables alert notification dispatch.
- Enabling notification dispatch requires at least one configured notifier URL:
  `WORKERFLEET_FEISHU_WEBHOOK_URL` or `WORKERFLEET_WEBHOOK_URL`.
- `WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT` controls per-alert delivery timeout.
- `WORKERFLEET_FEISHU_WEBHOOK_URL` configures the Feishu webhook sink.
- `WORKERFLEET_WEBHOOK_URL` configures the generic JSON webhook sink.
- `WORKERFLEET_WEBHOOK_HEADERS` configures generic webhook headers as comma-separated `key=value` entries.

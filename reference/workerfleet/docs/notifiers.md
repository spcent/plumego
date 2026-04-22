# Workerfleet Notifiers

The workerfleet notifier layer delivers alert records to:

- Feishu webhook endpoints
- generic JSON webhooks

Delivery rules:

- firing alerts are sent as they are created
- resolved alerts are sent as recovery notifications
- dispatcher fan-out stops on the first delivery error
- notifier errors must not include header secrets

Runtime configuration:

- `WORKERFLEET_NOTIFICATION_ENABLED=true` enables alert notification dispatch.
- `WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT` controls per-alert delivery timeout.
- `WORKERFLEET_FEISHU_WEBHOOK_URL` configures the Feishu webhook sink.
- `WORKERFLEET_WEBHOOK_URL` configures the generic JSON webhook sink.
- `WORKERFLEET_WEBHOOK_HEADERS` configures generic webhook headers as comma-separated `key=value` entries.

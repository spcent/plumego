# Workerfleet Notifiers

The workerfleet notifier layer delivers alert records to:

- Feishu webhook endpoints
- generic JSON webhooks

Delivery rules:

- firing alerts are sent as they are created
- resolved alerts are sent as recovery notifications
- dispatcher fan-out stops on the first delivery error
- notifier errors must not include header secrets

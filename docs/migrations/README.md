# Migrations Index

This folder contains SQL snippets for optional Plumego modules and examples.
Apply only the ones you need.

## Message Queue (Durable Tasks)

- `mq_task_queue_postgres.sql`
- `mq_task_queue_mysql.sql`

## Idempotency

- `idempotency_postgres.sql`
- `idempotency_mysql.sql`

## SMS Gateway Example

- `sms_gateway_messages_postgres.sql`
- `sms_gateway_messages_mysql.sql`

### Backfill note: `sent_at`

The SMS gateway schema includes a nullable `sent_at` column. For existing rows, you can backfill:

```sql
UPDATE sms_messages
SET sent_at = updated_at
WHERE sent_at IS NULL AND status = 'sent';
```

If you track a more accurate send timestamp elsewhere, prefer that value instead.

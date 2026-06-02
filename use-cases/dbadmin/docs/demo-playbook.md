# dbadmin Demo Playbook

This playbook shows the core value of dbadmin as a local-first internal data
workbench: one UI for multiple databases, with backend-enforced safety
guardrails.

## Start The Demo Stack

```bash
cd use-cases/dbadmin
docker compose -f docker-compose.dev.yml up -d --wait
go run . -addr=127.0.0.1:8080 -data-dir=/tmp/dbadmin-demo
```

Open `http://127.0.0.1:8080` and sign in with `admin / admin`.

## Add Connections

Use **Manage Connections** to add these profiles:

| Driver | Name | Connection details |
|---|---|---|
| MySQL | Demo MySQL | host `localhost`, port `13306`, database `testdb`, user `dbadmin`, password `dbadminpass` |
| Redis | Demo Redis | host `localhost`, port `16379`, password `redispass`, DB `0` |
| MongoDB | Demo MongoDB | URI `mongodb://dbadmin:dbadminpass@localhost:37017/testdb?authSource=admin` |
| Elasticsearch | Demo Elasticsearch | node `http://localhost:19200` |

For each connection, enable **Read-only connection** once when demonstrating
guardrails, then disable it when demonstrating writes.

## Demo Script

1. Open **Demo MySQL**.
   - Browse `users`, `orders`, and `user_order_summary`.
   - Run `SELECT username, balance FROM users ORDER BY balance DESC`.
   - Run `SELECT * FROM users` without a limit and show automatic result
     truncation.
   - Try `DELETE FROM users` and show the confirmation requirement.
   - Enable read-only mode and show that writes are rejected by the backend.

2. Open **Demo Redis**.
   - Browse `db0` with SCAN-backed key listing.
   - Run safe commands such as `PING` and `GET demo:string`.
   - Open the Redis history tab and show the server-side command record.
   - Try `KEYS *` and show that the backend blocks it.
   - Use batch preview before batch delete.

3. Open **Demo MongoDB**.
   - Browse `users`, `products`, and `orders`.
   - Query `{"age":{"$gt":30}}` with a limit.
   - Run an aggregation pipeline:

   ```json
   [
     {"$match":{"isActive":true}},
     {"$group":{"_id":"$preferences.theme","count":{"$sum":1}}}
   ]
   ```

   - Show schema sampling, stats, and ObjectId parsing.

4. Open **Demo Elasticsearch**.
   - Browse indices and mappings.
   - Run a DSL search:

   ```json
   {"query":{"match_all":{}},"size":50}
   ```

   - Show document lookup and delete confirmation behavior.

## Seed Optional Redis And Elasticsearch Data

The compose stack seeds MySQL and MongoDB automatically. Use these commands to
add quick Redis and Elasticsearch sample data:

```bash
docker exec dbadmin-redis redis-cli -a redispass SET demo:string "hello from dbadmin"
docker exec dbadmin-redis redis-cli -a redispass HSET demo:user name alice role admin
docker exec dbadmin-redis redis-cli -a redispass LPUSH demo:events login query export

curl -s -X PUT http://localhost:19200/demo-products -H 'Content-Type: application/json' -d '{
  "mappings":{"properties":{"name":{"type":"text"},"category":{"type":"keyword"},"price":{"type":"double"}}}
}'
curl -s -X POST http://localhost:19200/demo-products/_doc/1 -H 'Content-Type: application/json' -d '{"name":"Laptop Pro","category":"electronics","price":1299.99}'
curl -s -X POST http://localhost:19200/demo-products/_doc/2 -H 'Content-Type: application/json' -d '{"name":"Coffee Beans","category":"food","price":24.99}'
curl -s -X POST http://localhost:19200/demo-products/_refresh
```

## What To Emphasize

- Data and credentials stay local.
- One workbench spans relational, key-value, document, and search databases.
- Destructive operations require backend confirmation.
- Read-only mode is enforced server-side.
- Large values are previewed instead of blindly streamed to the browser.
- Long-running operations are bounded by configurable timeouts.
- SQL queries can be cancelled from the UI when cancellation is enabled.

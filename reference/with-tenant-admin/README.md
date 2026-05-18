# with-tenant-admin

`reference/with-tenant-admin` is a self-contained reference application for a
multi-tenant administrative plane. It demonstrates tenant lifecycle operations,
quota administration, usage recording, and fail-closed admin authentication
using `net/http`, `core.App`, `contract.WriteResponse`, and `x/tenant`
primitives.

## Run

```bash
ADMIN_TOKEN=secret go run ./...
```

The server listens on `:8086` by default. Override it with `APP_ADDR`.

Every admin endpoint requires:

```http
X-Admin-Token: secret
```

Missing or invalid tokens return `401`.

## API

Create a tenant:

```bash
curl -s -X POST http://localhost:8086/admin/tenants \
  -H 'X-Admin-Token: secret' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Acme"}'
```

Read, suspend, and delete a tenant:

```bash
curl -s http://localhost:8086/admin/tenants/{tenantID} \
  -H 'X-Admin-Token: secret'

curl -s -X POST http://localhost:8086/admin/tenants/{tenantID}/suspend \
  -H 'X-Admin-Token: secret'

curl -i -X DELETE http://localhost:8086/admin/tenants/{tenantID} \
  -H 'X-Admin-Token: secret'
```

Manage per-minute request quota:

```bash
curl -s -X PUT http://localhost:8086/admin/quota/{tenantID} \
  -H 'X-Admin-Token: secret' \
  -H 'Content-Type: application/json' \
  -d '{"limit":1000}'

curl -s http://localhost:8086/admin/quota/{tenantID} \
  -H 'X-Admin-Token: secret'

curl -s -X POST http://localhost:8086/admin/quota/{tenantID}/reset \
  -H 'X-Admin-Token: secret'
```

Record and read usage:

```bash
curl -s -X POST http://localhost:8086/admin/usage/{tenantID} \
  -H 'X-Admin-Token: secret' \
  -H 'Content-Type: application/json' \
  -d '{"resource":"api_requests","count":42}'

curl -s http://localhost:8086/admin/usage/{tenantID} \
  -H 'X-Admin-Token: secret'
```

Responses use the standard `contract` envelope. Successful reads return
`{"data":...}`; errors return the canonical structured error payload.

## Adapter Replacement

The reference app uses in-memory adapters so it runs without infrastructure:

- `internal/tenant/admin.InMemoryStore`
- `x/tenant/core.InMemoryConfigManager`
- `x/tenant/core.InMemoryQuotaStore`
- `internal/usage.InMemoryUsageStore`

Production applications should replace these with database-backed
implementations and inject them through `app.Deps`. Keep the same handler
interfaces: tenant lookup needs `Get`, usage stores need `Record` and `Report`,
and quota administration needs a config manager plus quota store.

## Auth Pattern

`internal/auth.RequireAdminToken` hashes the configured admin token and the
request token, then compares fixed-size digests with `hmac.Equal`. The
middleware fails closed before any handler code runs and never logs or echoes
the supplied token.

# with-tenant Scenario Reference

`reference/with-tenant` is a non-canonical scenario reference.

It shows how to add `x/tenant` resolution, policy, quota, and rate limiting to a
small API while keeping Plumego route registration explicit.

`x/tenant` remains experimental until beta promotion evidence is complete. Use
this scenario reference as wiring guidance, not as a compatibility claim.

## What It Demonstrates

- tenant resolution from `X-Tenant-ID`
- policy enforcement through `X-Model`
- fixed-window quota and token-bucket rate limiting
- app-local route registration using a route-level middleware chain

## Routes

- `GET /api/models`

Example request:

```bash
curl -H 'X-Tenant-ID: tenant-a' -H 'X-Model: gpt-4o' http://localhost:8085/api/models
```

## Run

```bash
cd reference/with-tenant
go run .
```

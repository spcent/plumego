# Production Checklist — with-rest

Run through this list in addition to `reference/standard-service/PRODUCTION_CHECKLIST.md`
before deploying a REST resource service.

---

## REST resource-specific

- [ ] **Input validation is configured**.
  This demo uses no validator. In production, call `controller.WithValidator(v)`
  with a validator that checks required fields, string lengths, and enum values
  before any write is committed.

- [ ] **Repository is backed by a persistent store**.
  `user.Repository` is in-memory and loses all data on restart. Replace it with
  a database-backed implementation of `rest.Repository[T]` before deploying.

- [ ] **Write routes are authenticated**.
  `Create`, `Update`, and `Delete` routes are currently unprotected. In
  production, wrap write routes with an auth middleware or add handler-level
  guards before registering them.

- [ ] **Pagination limits are set**.
  `x/rest` supports `rest.QueryParams` with page size and offset. Without a
  configured limit, `Index` may return unbounded result sets from large
  collections.

- [ ] **ID generation is collision-safe**.
  The demo uses a sequential counter (`u_1`, `u_2`, …). In production, use
  UUIDs or a database-assigned primary key to avoid collisions under concurrent
  writes or after restart.

- [ ] **Error shapes match your API contract**.
  `x/rest` uses `contract.WriteError` for error responses. Verify that your
  client SDK or API gateway expects the standard Plumego error envelope.

---

## Inherited from standard-service

All items from `reference/standard-service/PRODUCTION_CHECKLIST.md` apply:
transport timeouts, body size limits, security headers, TLS, observability,
signal handling.

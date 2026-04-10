# Card 0865: Store DB Timeout Policy Convergence

Priority: P2
State: active
Primary Module: store

## Goal

Make `store/db` helpers explicit and context-driven. Stable DB helpers should not discover timeout policy through optional config interfaces at call time.

## Problem

`store/db/sql.go` still applies query and transaction timeouts implicitly by probing for a `GetConfig()` method:

- `ConfigurableDB`
- `getConfig`
- `withQueryTimeout`
- `withTransactionTimeout`

That means helper behavior depends on hidden optional interface detection instead of the caller-provided `context.Context`. The stable store layer should provide DB primitives, not implicit runtime policy.

There is also a duplicate unreachable return in `WithTransaction` after a failed transaction begin, which should be fixed while pruning this helper surface.

## Scope

- Remove optional `GetConfig()` probing from query and transaction helpers.
- Make query and transaction helpers rely only on the `context.Context` passed by the caller.
- Keep `Config` and `Open` connection pool setup if they remain useful as stable DB primitives.
- Fix the duplicate unreachable return in `WithTransaction`.
- Update tests that assert implicit timeout behavior to use explicit context timeouts instead.
- Update docs to state that callers own operation deadlines through context.

## Non-Goals

- Do not remove stdlib `database/sql` compatibility.
- Do not add a new DB abstraction layer.
- Do not add retry, circuit-breaker, tenant, or sharding behavior to stable `store/db`.
- Do not change connection pool defaults unless tests prove they are coupled to the removed behavior.

## Expected Files

- `store/db/sql.go`
- `store/db/sql_test.go`
- `docs/modules/store/README.md`
- `store/module.yaml`
- affected `x/data/**` callers/tests if any

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./store/db ./x/data/...
go test -race -timeout 60s ./store/db
go vet ./store/db ./x/data/...
```

Then run the required repo-wide gates before committing.

## Done Definition

- Query and transaction helpers no longer use optional config introspection.
- Operation deadlines are controlled by caller-provided contexts.
- The duplicate unreachable return is removed.
- Focused tests cover explicit context timeout behavior.
- Focused gates and repo-wide gates pass.

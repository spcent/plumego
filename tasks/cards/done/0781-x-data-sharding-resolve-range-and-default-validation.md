---
id: 0781
title: x/data sharding range resolvability and default validation
status: done
priority: P1
primary_module: x/data/sharding
---

## Goal

Make sharding resolvability match the resolver's supported range-query behavior
and make `ShardingRule.Validate` reject the same invalid default shard values as
`SetDefaultShard`.

## Scope

- Fix range operators such as `>=`, `<=`, `!=`, and `<>` in parsed WHERE
  conditions.
- Ensure `Router.QueryContext` treats shard-key range predicates as resolvable
  and reaches `ResolveMultiple`.
- Reject `DefaultShard < -1` in rule validation.
- Add focused parser, resolver, router, and rule tests.

## Non-goals

- Introduce a full SQL parser.
- Expand sharding SQL support beyond the documented single-table subset.
- Change cross-shard policy names or defaults.

## Files

- `x/data/sharding/parser.go`
- `x/data/sharding/resolver.go`
- `x/data/sharding/router.go`
- `x/data/sharding/rule.go`
- `x/data/sharding/*_test.go`

## Tests

- `go test -timeout 20s ./x/data/sharding`
- `go test -race -timeout 60s ./x/data/sharding`
- `go vet ./x/data/sharding`

## Docs Sync

No user-facing docs change expected unless behavior wording needs correction.

## Done Definition

- Range predicates on the shard key are classified as resolvable.
- Router range-query tests prove the query no longer falls into unresolved
  fallback paths.
- `Register` rejects direct rules with `DefaultShard < -1`.

## Validation

- `go test -timeout 20s ./x/data/sharding`
- `go test -race -timeout 60s ./x/data/sharding`
- `go vet ./x/data/sharding`

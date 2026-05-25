# middleware/conformance

This directory is the shared test-only conformance suite for stable
`middleware/*` packages.

It intentionally contains only `_test.go` files. It is not an implementation
package and should not grow runtime code, constructors, or public transport
helpers.

## Purpose

Use this suite to keep cross-cutting middleware contracts aligned across stable
packages when the same behavior must hold in more than one place.

Current coverage includes:

- canonical JSON error envelope expectations
- panic propagation through high-risk wrappers
- response-writer optional interface preservation such as `Unwrap`, `Flush`,
  and `Hijack`
- timeout post-deadline write behavior
- partial gzip panic finalization and other regression fixtures
- runtime invariants shared by access logging, auth, rate limiting, recovery,
  tracing, and metrics middleware

## When to add coverage here

Add a test here when:

- the invariant is shared by multiple stable middleware packages
- the behavior is transport-level rather than package-private implementation
- duplicating the same assertion in many package-local tests would create drift

Keep package-specific behavior in the owning package's tests when the contract
is not shared.

## How to run it

Run the full middleware suite:

```bash
go test -timeout 20s ./middleware/...
```

Or target only this directory while iterating:

```bash
go test -timeout 20s ./middleware/conformance/...
```

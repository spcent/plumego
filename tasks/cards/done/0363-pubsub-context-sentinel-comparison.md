# Card 0363: x/pubsub Context Sentinel Comparison Anti-pattern

Priority: P3
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/pubsub

## Goal

`x/pubsub/pubsub.go` compares a caller-supplied `ctx` against `context.Background()` and
`context.TODO()` by pointer identity to decide whether to wire up cancellation:

```go
// lines 341 and 419 (identical pattern)
needsCtxWatch := ctx != nil && ctx != context.Background() && ctx != context.TODO()
```

This pattern has two problems:

1. **Fragile assumption**: it relies on `context.Background()` and `context.TODO()` being
   distinguishable singleton values.  That happens to be true in current Go stdlib, but it is
   not part of the `context` package's public contract.

2. **Wrong heuristic**: the actual intent is "does this context carry a Done channel we should
   monitor?"  The stdlib provides `ctx.Done()` for exactly that question.  A context with no
   Done channel (like `context.Background()` and `context.TODO()`) returns `nil` from
   `Done()`.  Custom contexts that wrap background but add no cancellation also return `nil`.
   The sentinel comparison misses those cases.

## Fix

Replace both occurrences with the `ctx.Done() != nil` idiom:

```go
// before
needsCtxWatch := ctx != nil && ctx != context.Background() && ctx != context.TODO()

// after
needsCtxWatch := ctx != nil && ctx.Done() != nil
```

This is semantically correct: only contexts that can be cancelled carry a non-nil `Done`
channel.  `context.Background()` and `context.TODO()` both return `nil` from `Done()`.

## Scope

Two identical lines in `x/pubsub/pubsub.go`: the `subscribe` function (line ~341) and the
`subscribeMQTT` function (line ~419).  Replace both.

## Non-goals

- Do not change the cancellation logic that follows (`context.WithCancel`, `cancel` call, etc.)
  — only the detection heuristic changes.
- Do not add new tests beyond verifying existing subscribe tests still pass.

## Files

- `x/pubsub/pubsub.go` (two lines)

## Tests

```bash
go test -timeout 20s -race ./x/pubsub/...
go vet ./x/pubsub/...
```

Verify that subscribing with `context.Background()` still works (no spurious cancellation),
and subscribing with a cancellable context still auto-unsubscribes when the context is done.

## Done Definition

- Neither `context.Background()` nor `context.TODO()` appears in the `needsCtxWatch`
  expression in `pubsub.go`.
- Both replaced with `ctx.Done() != nil`.
- `go test -race ./x/pubsub/...` passes.

## Outcome

Completed. Both occurrences of `ctx != context.Background() && ctx != context.TODO()`
in `x/pubsub/pubsub.go` (lines 341 and 419) replaced with `ctx.Done() != nil`.
`go test -race ./x/pubsub/...` passes.

# Card 0734

Priority: P2

Goal:
- Clarify and fix the retry semantics in `streamFromGenWithRetry` so that
  `maxRetries` means "total retries", not "consecutive retries between
  successes", and document the chosen policy.

Problem:

`context_stream.go:446-471`:
```go
func streamFromGenWithRetry[T any](..., maxRetries int, ...) error {
    retries := 0
    for {
        item, err := gen()
        if err != nil {
            if retries < maxRetries {
                retries++
                continue
            }
            return err
        }
        retries = 0   // ← reset after each success
        // ...
    }
}
```

`retries` is reset to 0 after every successful call to `gen`. This means
`maxRetries` is actually "maximum consecutive failures", not "total retries
for the entire stream". A generator that alternates success/failure (S, F, S,
F, …) with `maxRetries=1` would:

- Call 1: success → retries=0
- Call 2: failure → retries=1 (≤ maxRetries), retry
- Call 3: success → **retries=0 again**
- Call 4: failure → retries=1, retry
- Call 5: success → retries=0 again
- … indefinitely

The function effectively never terminates for intermittently-failing generators,
regardless of `maxRetries`. The parameter name and doc comment imply a total
budget.

Two fixes:

**Option A (preferred): total retry budget**
Remove `retries = 0`. The counter accumulates across the whole stream and
`maxRetries` is consumed once per failure regardless of intervening successes.
A generator that produces 100 items and fails 3 times against `maxRetries=2`
would fail at the third error.

**Option B: document "consecutive" semantics**
Add a clear doc comment: "maxRetries is the maximum number of *consecutive*
failures; a successful item resets the counter." This is honest but the
semantics are surprising and the current name is misleading.

Option A is preferred: generators should have a fixed fault-tolerance budget,
not an infinitely renewable one.

Non-goals:
- Do not add per-item retry state tracking.
- Do not change the `sleepWithContext` delay behaviour.

Files:
- `contract/context_stream.go`

Tests:
- Add a test: a generator that alternates success/failure must terminate
  after `maxRetries` total failures with Option A.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `maxRetries` counts total failures across the stream (Option A), or the
  "consecutive" semantics are clearly documented (Option B).
- The chosen policy is stated in the function's doc comment.
- Tests verify the retry limit is enforced.
- All tests pass.

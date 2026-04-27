# Card 0247

Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/observability_policy.go`
Depends On: —

Goal:
- Remove `DefaultObservabilityPolicy()`, which is a no-argument alias for `NewObservabilityPolicy()` and adds nothing.

Problem:
`DefaultObservabilityPolicy()` (line 24–26 in `observability_policy.go`) is defined as:

```go
func DefaultObservabilityPolicy() ObservabilityPolicy {
    return NewObservabilityPolicy()
}
```

It calls `NewObservabilityPolicy()` with no arguments, producing an identical result.
The companion doc comment even directs callers to `NewObservabilityPolicy` for customisation,
making `DefaultObservabilityPolicy` a confusing dead-end in the API surface.
Having two names for the same zero-argument construction creates unnecessary cognitive load:
callers wonder whether the two functions differ, what "default" means here, and which one to reach for.

Scope:
- Delete `DefaultObservabilityPolicy()` from `contract/observability_policy.go`.
- Replace all call sites with `NewObservabilityPolicy()`.
- Grep for callers before editing: `grep -rn 'DefaultObservabilityPolicy' . --include='*.go'`

Non-goals:
- Do not change `NewObservabilityPolicy` signature or behaviour.
- Do not add new configuration knobs to `ObservabilityPolicy`.

Files:
- `contract/observability_policy.go`
- Any caller files found by grep

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `DefaultObservabilityPolicy` is gone from the codebase.
- `grep -rn 'DefaultObservabilityPolicy' . --include='*.go'` returns empty.
- All tests pass.

Outcome:
- `DefaultObservabilityPolicy()` deleted from `contract/observability_policy.go`.
- All 15 call sites replaced with `NewObservabilityPolicy()` via bulk sed.
- `grep -rn 'DefaultObservabilityPolicy' . --include='*.go'` returns empty.
- `go test -timeout 20s ./...` passes.

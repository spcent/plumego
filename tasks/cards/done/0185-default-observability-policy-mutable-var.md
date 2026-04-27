# Card 0185

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/observability_policy.go`
Depends On: —

Goal:
- Replace the mutable `DefaultObservabilityPolicy` package-level variable with an
  immutable function following the same pattern as `DefaultConfig()`.

Problem:
`observability_policy.go` exports a mutable package-level variable:

```go
var DefaultObservabilityPolicy = NewObservabilityPolicy()
```

Any caller can reassign it:
```go
contract.DefaultObservabilityPolicy = contract.NewObservabilityPolicy("mykey")
```

This is a global mutable state footgun and is inconsistent with how the package
exposes the other default — `DefaultConfig()` (card 0608) was changed to a
function that returns a copy precisely to prevent this pattern.

`ObservabilityPolicy` is a value type (not a pointer) so callers already receive
a copy when they call methods on the zero value. The only thing the package-level
variable enables is reassignment, which is undesirable.

Current callers reference `contract.DefaultObservabilityPolicy.SomeMethod(...)`.
Changing to `contract.DefaultObservabilityPolicy()` is a one-line change per call
site and is straightforward.

Scope:
- Replace `var DefaultObservabilityPolicy = NewObservabilityPolicy()` with
  `func DefaultObservabilityPolicy() ObservabilityPolicy { return NewObservabilityPolicy() }`.
- Update all call sites in `middleware/` that use `contract.DefaultObservabilityPolicy`.
- The function name deliberately matches the old variable name to minimise diff;
  callers add `()` and nothing else changes.

Non-goals:
- No change to `ObservabilityPolicy` internals.
- No change to `NewObservabilityPolicy`.
- No new configuration parameters.

Files:
- `contract/observability_policy.go`
- `middleware/accesslog/accesslog.go`
- `middleware/recovery/recover.go`
- `middleware/limits/limit.go`
- `middleware/requestid/request_id.go`
- `middleware/ratelimit/limiter.go`
- `middleware/ratelimit/abuse_guard.go`
- `middleware/internal/observability/helpers.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/... ./middleware/...`
- `go vet ./...`
- After the change, `grep -rn 'DefaultObservabilityPolicy[^(]' . --include='*.go'`
  must return only the function declaration itself, not any bare variable accesses.

Docs Sync: —

Done Definition:
- `DefaultObservabilityPolicy` is a function, not a variable.
- No call site accesses it as a value (no bare `DefaultObservabilityPolicy.Method`).
- All tests pass.

Outcome:

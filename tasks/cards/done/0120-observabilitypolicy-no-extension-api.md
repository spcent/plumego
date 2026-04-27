# Card 0120

Priority: P3

Goal:
- Add a `WithSensitiveKeys` constructor option (or a builder method) to
  `ObservabilityPolicy` so callers can add application-specific sensitive keys
  without replacing the entire default policy.

Problem:

`observability_policy.go:29-39`:
```go
func NewObservabilityPolicy() ObservabilityPolicy {
    return ObservabilityPolicy{
        mask: "***",
        sensitiveKeys: map[string]struct{}{
            "token":     {},
            "secret":    {},
            "signature": {},
            "password":  {},
        },
    }
}
```

The default policy redacts four key patterns. Applications commonly need to
add domain-specific sensitive keys such as `"api_key"`, `"ssn"`,
`"credit_card"`, `"phone"`, or `"dob"`. Currently there is no way to extend
the policy:

- `ObservabilityPolicy.sensitiveKeys` is unexported — callers cannot set it.
- `NewObservabilityPolicy()` takes no parameters.
- `DefaultObservabilityPolicy` is a package-level var initialized from
  `NewObservabilityPolicy()` — callers who reassign it replace all defaults.

A caller who wants to add `"api_key"` must either:
1. Replace `DefaultObservabilityPolicy` with a hand-crafted value (losing
   the default keys), or
2. Fork `NewObservabilityPolicy()` with a hardcoded extra key.

Fix: Add a variadic option to `NewObservabilityPolicy`:

```go
func NewObservabilityPolicy(extraSensitiveKeys ...string) ObservabilityPolicy {
    keys := map[string]struct{}{
        "token": {}, "secret": {}, "signature": {}, "password": {},
    }
    for _, k := range extraSensitiveKeys {
        keys[strings.ToLower(k)] = struct{}{}
    }
    return ObservabilityPolicy{mask: "***", sensitiveKeys: keys}
}
```

Callers can then extend the defaults:
```go
contract.DefaultObservabilityPolicy = contract.NewObservabilityPolicy("api_key", "ssn")
```

Alternatively, add a fluent method:
```go
func (p ObservabilityPolicy) WithSensitiveKeys(keys ...string) ObservabilityPolicy
```

The variadic option on the constructor is preferred: it is simpler and keeps
`ObservabilityPolicy` as a value type.

Non-goals:
- Do not add key removal capability in this card.
- Do not change the default set of four sensitive keys.
- Do not change the `"***"` mask value.

Files:
- `contract/observability_policy.go`

Tests:
- Add a test: `NewObservabilityPolicy("api_key")` redacts `"api_key"` fields.
- Add a test: default keys (`"password"`, etc.) still redacted when extra keys
  are provided.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `NewObservabilityPolicy` accepts optional extra sensitive key names.
- The four default keys are always included.
- `DefaultObservabilityPolicy` is unchanged (still uses no-arg constructor).
- All tests pass.

Outcome:
- Completed in the 2026-04-05 contract cleanup batch.
- Verified as part of the shared contract/task-card completion pass.

Validation Run:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `go build ./...`
- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`

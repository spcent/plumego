# Card 0132

Priority: P2
State: done
Primary Module: contract
Owned Files: contract/error_wrap.go

Goal:
- Fix the params merge in `WrapError` so outer params fill missing keys only
  (matching the inner-wins rule for operation/module), and update the doc
  comment to accurately describe the merge semantics for all three fields.

Problem:

`error_wrap.go:64-94`:
```go
if wrapped, ok := err.(*WrappedErrorWithContext); ok {
    var mergedParams map[string]any
    ...
    for k, v := range wrapped.Context.Params {
        mergedParams[k] = v          // inner params written first
    }
    for k, v := range params {
        mergedParams[k] = v          // outer params written second ← OVERRIDES inner
    }

    op := wrapped.Context.Operation
    if op == "" { op = operation }   // inner wins for operation
    mod := wrapped.Context.Module
    if mod == "" { mod = module }    // inner wins for module
```

The doc comment says:
> "If err is already a *WrappedErrorWithContext, the inner operation/module take
> precedence and outer calls only fill fields that were previously empty."

This correctly describes operation/module. But it says nothing about params,
where the **opposite rule applies**: outer values OVERRIDE inner values for
any conflicting key.

Concrete impact: if the original error was wrapped at the database layer with
`params["entity_id"] = "42"`, and a service layer re-wraps it with
`params["entity_id"] = "99"` (perhaps a different ID from the outer scope),
the database-layer value `"42"` is silently lost. The operator investigating
the log sees the outer value, not the original site value.

This contradicts the stated design intent. The doc says inner takes precedence —
callers who read it expect inner params to win, not outer params.

Fix — change the merge loop to match the operation/module rule (inner wins):
```go
for k, v := range params {
    if _, exists := mergedParams[k]; !exists {
        mergedParams[k] = v    // outer fills only missing keys
    }
}
```

And update the doc comment:
```go
// WrapError wraps an existing error with additional context.
// If err is already a *WrappedErrorWithContext, inner fields take precedence:
//   - Operation and Module: inner value is used if non-empty; outer is the fallback.
//   - Params: inner keys win; outer keys are added only if not already present.
// This preserves the original failure-site context. The intermediate wrapper is
// flattened — the returned error's Unwrap() chain is: result → inner.Err.
```

The last sentence also documents the "flattening" behavior (the intermediate
`*WrappedErrorWithContext` is not preserved in the chain), which is currently
undocumented.

Non-goals:
- Do not change the flattening behavior (replacing chain with flat structure).
- Do not add a new wrapping mode that preserves the full chain.
- Do not change `WrapErrorf` or `NewWrappedError`.

Files:
- `contract/error_wrap.go`

Tests:
- Extend `TestWrapErrorPreservesInnerContextPrecedence` (or add a sibling)
  to cover the conflicting-key case:
  ```go
  func TestWrapErrorInnerParamWinsOnConflict(t *testing.T) {
      inner := WrapError(errors.New("base"), "op", "mod",
          map[string]any{"entity_id": "inner-value", "inner_only": true})
      outer := WrapError(inner, "op2", "mod2",
          map[string]any{"entity_id": "outer-value", "outer_only": true})

      details := GetErrorDetails(outer)
      params := details["params"].(map[string]any)
      if params["entity_id"] != "inner-value" {
          t.Errorf("inner param should win on conflict, got %v", params["entity_id"])
      }
      if params["inner_only"] != true || params["outer_only"] != true {
          t.Errorf("both unique keys should be present, got %v", params)
      }
  }
  ```
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- Outer params no longer override inner params for conflicting keys.
- `WrapError` doc comment accurately describes merge semantics for all three
  fields plus the flattening behaviour.
- The new test passes; all existing tests pass.

Outcome:
- Completed by changing `WrapError` param merging to inner-wins semantics and
  documenting both the field precedence rules and the wrapper-flattening
  behavior.

Validation Run:
- `gofmt -w contract/error_wrap.go contract/active_cards_regression_test.go`
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

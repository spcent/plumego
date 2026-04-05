# Card 0730

Priority: P2

Goal:
- Extend `setFieldFromQuery` to handle non-string slice types (`[]int`, `[]int64`,
  `[]bool`, etc.) so that multi-value query parameters bound to typed slices
  are not silently ignored.

Problem:

`context_bind.go:261-264`:
```go
case reflect.Slice:
    if fv.Type().Elem().Kind() == reflect.String {
        fv.Set(reflect.ValueOf(vals))
    }
    // ← all other slice element types: silently do nothing, return nil
```

Only `[]string` fields are populated from multi-value query parameters. Any
other slice type (`[]int`, `[]int64`, `[]float64`, `[]bool`) is silently skipped:

```go
type Filter struct {
    IDs    []int    `query:"id"`     // ← silently stays nil
    Scores []float64 `query:"score"` // ← silently stays nil
    Flags  []bool   `query:"flag"`   // ← silently stays nil
}
```

`c.BindQuery(&f)` returns nil (no error) for any URL, even
`?id=1&id=2&id=3`, leaving `IDs` as nil. This is the same silent data-loss
pattern as pointer fields (card 0724).

Fix: Extend the `reflect.Slice` case to parse each element of `vals` into the
slice's element type, using the same parsing logic as the scalar cases:

```go
case reflect.Slice:
    elem := fv.Type().Elem()
    result := reflect.MakeSlice(fv.Type(), 0, len(vals))
    for _, v := range vals {
        el := reflect.New(elem).Elem()
        if err := setFieldFromQuery(el, v, nil); err != nil {
            return err
        }
        result = reflect.Append(result, el)
    }
    fv.Set(result)
```

This reuses `setFieldFromQuery` recursively for each element, keeping the
parsing logic in one place.

Non-goals:
- Do not handle slices-of-structs or slices-of-slices.
- Do not change the `[]string` fast path (it can remain as a short-circuit
  for performance).
- Do not change `BindJSON`'s JSON array handling.

Files:
- `contract/context_bind.go`

Tests:
- Add tests:
  - `[]int` field: `?id=1&id=2&id=3` populates `[1, 2, 3]`.
  - `[]bool` field: `?flag=true&flag=false` populates `[true, false]`.
  - `[]int` field: invalid value `?id=notanint` returns a BindError.
  - `[]string` field: still works as before.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `setFieldFromQuery` handles `[]int`, `[]int64`, `[]float64`, `[]bool`, `[]uint`,
  and `[]uint64` slice fields.
- Absent parameters leave slices nil (existing behavior preserved).
- Invalid element values return a `BindError`.
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

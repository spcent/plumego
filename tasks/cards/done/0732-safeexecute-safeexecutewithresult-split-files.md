# Card 0732

Priority: P3

Goal:
- Co-locate `SafeExecute` and `SafeExecuteWithResult` in the same file so that
  the two halves of the same API are not split across `error_utils.go` and
  `error_wrap.go`.

Problem:

`SafeExecute` lives in `contract/error_utils.go:136-148`:
```go
func SafeExecute(fn func() error, operation, module string, params map[string]any) (err error)
```

`SafeExecuteWithResult` lives in `contract/error_wrap.go:229-244`:
```go
func SafeExecuteWithResult[T any](fn func() (T, error), operation, module string, params map[string]any) (result T, err error)
```

These two functions form a natural pair: one for functions that return only an
error, one for functions that return a value and an error. They have identical
signatures except for the generic return type, identical panic-recovery logic,
and are intended to be used the same way.

A developer looking for `SafeExecuteWithResult` would naturally look in the
same file as `SafeExecute`, and vice versa. The split forces readers to know
which file each variant lives in.

The reason for the split is likely that `SafeExecuteWithResult` uses Go generics
(which may have been added later) and was added to `error_wrap.go` where other
generic helpers live, while `SafeExecute` was written earlier in `error_utils.go`.

Fix: Move `SafeExecuteWithResult` from `error_wrap.go` to `error_utils.go`,
placing it immediately after `SafeExecute`. Both functions already import the
same dependencies.

Non-goals:
- Do not change either function's signature or behavior.
- Do not merge the two functions.
- Do not rename either function.

Files:
- `contract/error_wrap.go` (remove `SafeExecuteWithResult`)
- `contract/error_utils.go` (add `SafeExecuteWithResult`)

Tests:
- `go build ./...` (no behavior change, just a move)
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `SafeExecute` and `SafeExecuteWithResult` are in the same file (`error_utils.go`).
- `error_wrap.go` no longer contains `SafeExecuteWithResult`.
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

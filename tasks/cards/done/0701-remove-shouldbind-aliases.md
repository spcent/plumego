# Card 0701

Priority: P2

Goal:
- Remove `ShouldBindJSON` and `ShouldBindQuery` from `contract/context_bind.go`.
  Both are pure pass-throughs with no behavioral distinction from their base
  methods, and their doc comments make a false semantic claim.

Problem:

`context_bind.go:138-140`:
```go
func (c *Ctx) ShouldBindJSON(dst any) error {
    return c.BindJSON(dst)
}
```
`context_bind.go:151-153`:
```go
func (c *Ctx) ShouldBindQuery(dst any) error {
    return c.BindQuery(dst)
}
```

The doc comments say "Should methods return errors without writing a response."
`BindJSON` and `BindQuery` also return errors without writing a response — they
are not write-response methods. The naming convention adds zero behavioral
distinction and expands the exported API with redundant symbols. Callers who
want to signal "I will handle the error" can just call `BindJSON` directly.

No caller should be able to `ShouldBindJSON` that couldn't equally use
`BindJSON`; they are byte-for-byte equivalent.

Scope:
- Remove `ShouldBindJSON` and `ShouldBindQuery`.
- Grep all callers before removing; migrate each to `BindJSON` / `BindQuery`.

Non-goals:
- Do not change `BindJSON` or `BindQuery` behavior.
- Do not add any new aliases.

Files:
- `contract/context_bind.go`
- All callers (run `grep -rn 'ShouldBind' . --include='*.go'` first)

Tests:
- `go test ./contract/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `ShouldBindJSON` and `ShouldBindQuery` do not exist in the package.
- All callers migrated to base methods.
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

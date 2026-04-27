# Card 0183

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/validation.go`
Depends On: —

Goal:
- Remove the unused `parseNumericString` helper from `validation.go`.

Problem:
`validation.go` defines a private helper at lines 299–313:

```go
func parseNumericString(s string) (float64, bool) { ... }
```

It is never called anywhere in the package or the repo. The `validateMin` and
`validateMax` functions work directly on `reflect.Value` and do not delegate to
this helper. Dead unexported code adds noise and can mislead readers into thinking
the function is part of the validation path.

Scope:
- Delete `parseNumericString` from `contract/validation.go`.
- Confirm no other file in the repo references it
  (`grep -rn 'parseNumericString' . --include='*.go'`).

Non-goals:
- No change to actual validation logic.
- No refactoring of `validateMin` / `validateMax`.

Files:
- `contract/validation.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `parseNumericString` does not exist in the codebase.
- All tests pass.

Outcome:

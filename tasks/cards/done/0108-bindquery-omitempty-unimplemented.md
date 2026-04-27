# Card 0108

Priority: P3

Goal:
- Implement the `omitempty` query tag option, or remove the TODO comment and
  document that tag options beyond the field name are not supported.

Problem:

`context_bind.go:206-207`:
```go
// Split tag to get the name (ignore options like omitempty for now)
name := strings.SplitN(tag, ",", 2)[0]
```

The comment "ignore options like omitempty for now" has been in the code since
the function was written. A caller who writes:

```go
type Query struct {
    Page int `query:"page,omitempty"`
}
```

gets the same behavior as `query:"page"`. The `omitempty` option is silently
discarded. No error, no warning, no documentation to tell callers that tag
options are unimplemented.

This matters because `omitempty` is the standard option in Go struct tags
(json, yaml, query packages all use it), and callers will naturally add it
expecting some effect.

Two options:

**Option A: Implement `omitempty`**
`omitempty` for query binding means: if the query parameter is absent or empty,
do not set the field (leave it at its zero value). This is already the current
behavior for all fields (the early-return in `setFieldFromQuery` handles absent
params). So `omitempty` is effectively always active. The tag option can be
parsed and accepted without any behavioral change, removing the misleading
TODO.

Implementation: parse and accept the option; since the behavior is already
omitempty-by-default, no logic changes are needed. Document this in the function
comment.

**Option B: Return an error for unrecognized options**
If a tag contains options other than known ones, return a `BindError` at bind
time to surface the misconfiguration.

**Option C: Remove the TODO comment, document no tag options supported**
Change the comment to: `// Tag options (e.g. "omitempty") are not supported;
only the field name is used.` This is honest and prevents future readers from
thinking it is a planned feature.

Option A is preferred as it requires minimal code and aligns `query` with the
behavior callers expect from `omitempty` in other Go tags.

Non-goals:
- Do not implement other tag options (`required`, `default`, etc.) in this card.
- Do not change `BindJSON`'s handling of struct tags.

Files:
- `contract/context_bind.go`

Tests:
- Add a test: `query:"page,omitempty"` must behave identically to `query:"page"`.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- The TODO comment is removed.
- `omitempty` is either accepted silently (Option A) or clearly documented as
  unsupported (Option C).
- Tests verify the chosen behavior.
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

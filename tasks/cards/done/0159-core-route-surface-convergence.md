# Card 0159

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/app.go`
- `core/routing.go`
- `core/routing_test.go`
- `x/devtools/devtools.go`
- `docs/CANONICAL_STYLE_GUIDE.md`
Depends On:

Goal:
- Converge `core` route registration and route inspection onto one canonical
  surface each.

Problem:
- `Handle(...)`, `HandleFunc(...)`, and `Any(...)` all represent the same ANY
  route registration concept.
- That leaves three overlapping public entrypoints for one behavior, while the
  docs already prefer explicit method wiring.
- `App.Print(...)` is also a formatted mirror of `Routes()`, and its only
  first-party caller is `x/devtools`.
- The kernel therefore exposes duplicate APIs for both route registration and
  route table inspection.

Scope:
- Pick one canonical ANY-route registration path and remove the overlapping
  helpers.
- Remove `App.Print(...)` and make first-party tooling render text from
  `Routes()` instead of asking `core` to own formatting.
- Update tests and the style guide to show the converged route surface.

Non-goals:
- Do not remove method-specific helpers such as `Get` or `Post`.
- Do not redesign router matching behavior.
- Do not move route introspection out of first-party tooling.

Files:
- `core/app.go`
- `core/routing.go`
- `core/routing_test.go`
- `x/devtools/devtools.go`
- `docs/CANONICAL_STYLE_GUIDE.md`

Tests:
- `go test -race -timeout 60s ./core/... ./x/devtools/...`
- `go test -timeout 20s ./...`
- `go vet ./...`

Docs Sync:
- Keep examples on one explicit route-registration path and one structured
  route-inspection path.

Done Definition:
- `core` no longer exposes overlapping ANY-route helpers.
- Route table formatting is not owned by the kernel.
- First-party docs and devtools use the converged route surface.

Outcome:
- Kept `Any(...)` as the single explicit ANY-route registration helper in
  `core` and removed the overlapping `Handle(...)` / `HandleFunc(...)`
  aliases.
- Removed `App.Print(...)` so route table formatting is no longer owned by the
  kernel.
- Updated devtools to render route text from `Routes()` and tightened the style
  guide toward one canonical route surface.

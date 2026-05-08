# Card 1276

Milestone: Router stable readiness
Recipe: specs/change-recipes/refactor.yaml
Priority: P3
State: done
Primary Module: router
Owned Files: router/metadata.go, router/router_advanced_test.go, tasks/cards/active/README.md
Depends On: 0764-router-static-prefix-validation-order

Goal:
Avoid holding router locks while writing to user-provided `io.Writer` values.

Scope:
- Make `Print` snapshot route data before writing to the writer.
- Preserve existing print output ordering and wildcard labels.
- Add regression coverage using a writer that calls back into router lifecycle
  methods while printing.

Non-goals:
- Changing `Print` output format.
- Adding structured output APIs.
- Changing `Routes` behavior.

Files:
- router/metadata.go
- router/router_advanced_test.go
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required.

Done Definition:
- `Print` does not perform user writer I/O while holding the router lock.
- Existing print tests keep passing.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Changed `Print` to snapshot route data under lock and write to the user
  writer after releasing the router lock.
- Preserved method/path ordering and wildcard labels.
- Added regression coverage with a writer that calls back into `Freeze`.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

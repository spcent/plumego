# Card 0656

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/nethttp
Owned Files: internal/nethttp/*.go
Depends On: 0655

Goal:
Align the `internal/nethttp` package declaration with its directory and import intent.

Scope:
- Rename package declarations in `internal/nethttp` from `http` to `nethttp`.
- Update package comments so documentation and examples are not confused with the standard-library `net/http` package.
- Verify existing import sites already use explicit aliases or compile cleanly after the package rename.

Non-goals:
- Do not change HTTP client behavior, retry policy, SSRF behavior, or metrics behavior.
- Do not rename the directory or exported symbols.
- Do not refactor gateway or messaging users beyond any compile-required import cleanup.

Files:
- internal/nethttp/client.go
- internal/nethttp/security.go
- internal/nethttp/response_recorder.go
- internal/nethttp/metrics.go
- internal/nethttp/*_test.go

Tests:
- go test -timeout 20s ./internal/nethttp
- go test -timeout 20s ./internal/... ./x/gateway ./x/messaging
- go vet ./internal/...

Docs Sync:
Not required; no user-facing behavior or API surface changes beyond internal package name clarity.

Done Definition:
- `rg -n "package http$" internal/nethttp --glob '*.go'` returns no matches.
- Internal nethttp tests and compile coverage for known users pass.

Outcome:
Completed. `internal/nethttp` now declares `package nethttp`, its package comment matches the directory, and known import users compile without behavioral changes.

Validation:
- rg -n "package http$|Package http" internal/nethttp --glob '*.go'
- go test -timeout 20s ./internal/nethttp
- go test -timeout 20s ./internal/... ./x/gateway ./x/messaging
- go vet ./internal/...

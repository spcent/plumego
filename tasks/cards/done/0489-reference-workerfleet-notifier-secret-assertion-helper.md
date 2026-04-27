# Card 0489: Reference Workerfleet Notifier Secret Assertion Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- `reference/workerfleet/internal/platform/notifier/notifier_test.go`
Depends On: none

Goal:
Name the notifier test assertion that errors must not leak secret headers.

Problem:
The notifier test uses an inline substring check for `super-secret-token`.
Because the assertion is security-sensitive, a named helper makes the intent
more explicit.

Scope:
- Add a local helper for asserting an error string does not contain a secret.
- Use it in the notifier failure test.

Non-goals:
- Do not change notifier behavior or public APIs.
- Do not add dependencies.

Files:
- `reference/workerfleet/internal/platform/notifier/notifier_test.go`

Tests:
- from `reference/workerfleet`: `go test -race -timeout 60s ./internal/platform/notifier/...`
- from `reference/workerfleet`: `go test -timeout 20s ./internal/platform/notifier/...`
- from `reference/workerfleet`: `go vet ./internal/platform/notifier/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Notifier secret leak assertion uses the named helper.
- The listed validation commands pass.

Outcome:

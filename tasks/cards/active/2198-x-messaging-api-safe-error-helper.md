# Card 2198: x/messaging API Safe Error Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/messaging
Owned Files:
- `x/messaging/api_test.go`
Depends On: none

Goal:
Consolidate safe public error-message assertions in messaging API tests.

Problem:
Messaging API error-path tests repeat the same check that response messages do
not expose raw `messaging:` error text.

Scope:
- Add a local helper for asserting public error messages do not expose internal
  prefixes.
- Use it in send and batch-send error tests.

Non-goals:
- Do not change messaging API behavior or public errors.
- Do not add dependencies.

Files:
- `x/messaging/api_test.go`

Tests:
- `go test -race -timeout 60s ./x/messaging/...`
- `go test -timeout 20s ./x/messaging/...`
- `go vet ./x/messaging/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Safe error-message checks use the shared helper.
- The listed validation commands pass.

Outcome:

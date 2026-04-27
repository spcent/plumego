# Card 0636

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: internal/nethttp
Owned Files: internal/nethttp/client.go, internal/nethttp/httpclient_test.go
Depends On: tasks/cards/done/0635-internal-config-watch-startup-events.md

Goal:
Ensure high-level `Client` helpers close response bodies when `doRequest`
returns both a response and an error.

Scope:
- Centralize the high-level request-and-read path used by Get/Post/Put/Patch/Delete.
- Close response bodies on error before returning from byte-returning helpers.
- Add coverage using a tracking response body from a custom transport.

Non-goals:
- Do not change `Client.Do`; callers still own raw response bodies.
- Do not change retry policy APIs or status-code classification.
- Do not add dependencies.

Files:
- internal/nethttp/client.go
- internal/nethttp/httpclient_test.go

Tests:
- go test ./internal/nethttp
- go test ./internal/...

Docs Sync:
- None; behavior becomes safer without API or documented semantic changes.

Done Definition:
- A final 5xx response returned with an error is closed by high-level helpers.
- Successful responses are still drained and closed by `readResponse`.
- Focused nethttp tests pass.

Outcome:
- Added `doAndRead` to centralize high-level request execution.
- Closed response bodies when high-level helpers receive both response and error.
- Added custom transport coverage for final server-error response body closure.
- Validation: `go test ./internal/nethttp`; `go test ./internal/...`.

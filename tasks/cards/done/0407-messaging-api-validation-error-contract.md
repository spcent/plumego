# Card 0407: Messaging API Validation Error Contract

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: high
State: done
Primary Module: x/messaging
Owned Files:
- x/messaging/api.go
- x/messaging/api_test.go
- docs/modules/x-messaging/README.md
Depends On: none

Goal:
- Make messaging send API validation failures use one stable, safe response contract.
- Remove direct `err.Error()` exposure from request validation branches in `HandleSend` and `HandleBatchSend`.

Scope:
- Audit `x/messaging/api.go` request decode and validation branches for single-send and batch-send handlers.
- Keep JSON decode behavior, service error mapping, delivery flow, and response body shape unchanged outside validation failures.
- Add focused handler tests for malformed requests that currently surface validator error text.
- Keep `x/messaging` as the app-facing messaging family entrypoint.

Non-goals:
- Do not change provider, scheduler, queue, pubsub, or webhook semantics.
- Do not introduce new response helper families outside the local API handler need.
- Do not make subordinate messaging packages compete with `x/messaging` as public discovery roots.

Files:
- `x/messaging/api.go`: route validation failures through a stable code and safe message.
- `x/messaging/api_test.go`: cover missing required fields for single and batch send requests.
- `docs/modules/x-messaging/README.md`: document validation response expectations if public behavior changes.

Tests:
- `go test -race -timeout 60s ./x/messaging/...`
- `go test -timeout 20s ./x/messaging/...`
- `go vet ./x/messaging/...`

Docs Sync:
- Required if public error codes, messages, or documented API examples change.

Done Definition:
- `HandleSend` and `HandleBatchSend` no longer expose raw validation error strings in HTTP response messages.
- Validation failures have explicit stable codes and safe messages.
- Focused tests cover both single and batch validation failure paths.
- The three listed validation commands pass.

Outcome:
- Added a local `invalidMessagingRequestError` helper for handler-level request validation failures.
- Replaced the remaining `Message(err.Error())` branches in `HandleSend` and `HandleBatchSend`.
- Added focused handler tests for single-send missing required fields and empty batch requests to assert safe public messages.
- Validation passed:
  - `go test -race -timeout 60s ./x/messaging/...`
  - `go test -timeout 20s ./x/messaging/...`
  - `go vet ./x/messaging/...`

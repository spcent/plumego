# Card 1516

Milestone: M-022
Recipe: specs/change-recipes/docs-and-config.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/messaging/mq
Owned Files:
- `x/messaging/mq/config.go`
- `x/messaging/mq/broker_routes.go`
- `docs/modules/x/messaging/README.md`
- `specs/deprecation-inventory.yaml`
Depends On: 1510

Goal:
- Turn the MQTT/AMQP protocol placeholders into explicitly deprecated,
  documented compatibility surface instead of silent-looking config options
  that only fail at validation time.

Scope:
- Add `Deprecated:` guidance to the unsupported config fields and protocol
  bridge stubs.
- Tighten the messaging docs so the unsupported surface and preferred path are
  obvious.
- Update the deprecation inventory decision text to record the removal
  condition more clearly.

Non-goals:
- Do not implement MQTT or AMQP support.
- Do not rename the existing config fields in this card.
- Do not change the current fail-closed validation behavior.

Files:
- `x/messaging/mq/config.go`
- `x/messaging/mq/broker_routes.go`
- `docs/modules/x/messaging/README.md`
- `specs/deprecation-inventory.yaml`

Acceptance Tests:
- `go test -timeout 20s ./x/messaging/mq/...`

Tests:
- `go test -timeout 20s ./x/messaging/mq/...`
- `go run ./internal/checks/deprecation-inventory -strict`

Docs Sync:
- `docs/modules/x/messaging/README.md`

Validation:
- `go test -timeout 20s ./x/messaging/mq/...`
- `go run ./internal/checks/deprecation-inventory -strict`
- `gofmt -l .`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Added explicit `Deprecated:` guidance to the unsupported MQTT/AMQP config
  fields and protocol bridge stubs so the public surface now advertises them as
  compatibility placeholders instead of looking like dormant features.
- Tightened `docs/modules/x/messaging/README.md` to state that the port fields
  have no effect, the flags always fail closed, and `x/messaging/pubsub` is the
  preferred direct path until a real protocol bridge lands.
- Updated the deprecation inventory decision text so these placeholders have a
  documented removal condition instead of an open-ended `status: keep`.
- Validation:
  - `go test -timeout 20s ./x/messaging/mq/...`
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `gofmt -l .`

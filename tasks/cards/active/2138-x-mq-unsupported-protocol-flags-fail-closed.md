# Card 2138: x/mq Unsupported Protocol Flags Fail Closed

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/mq
Owned Files:
- `x/mq/config.go`
- `x/mq/broker.go`
- `x/mq/errors.go`
- `x/mq/mq_test.go`
- `docs/modules/x-mq/README.md`
Depends On: none

Goal:
Make unsupported MQTT and AMQP protocol flags fail at configuration time instead
of allowing a broker to be constructed with options that only fail later through
stub `Start*Server` methods.

Problem:
`Config.Validate` currently accepts `EnableMQTT` and `EnableAMQP` as long as the
ports are valid, while `StartMQTTServer` and `StartAMQPServer` both return
`ErrNotImplemented`. Tests also preserve the idea that a broker may be built
with those flags set. This advertises a capability-shaped API that is not
implemented and creates a fail-late path for production configuration.

Scope:
- Change config validation so `EnableMQTT` and `EnableAMQP` return
  `ErrNotImplemented` or a clearly wrapped unsupported-protocol error until the
  protocols are actually implemented.
- Update tests that currently expect successful construction with those flags.
- Keep default config unchanged with both flags disabled.
- Keep `StartMQTTServer` and `StartAMQPServer` only as compatibility stubs if
  needed, but make docs and tests clear that callers must not enable those flags.

Non-goals:
- Do not implement MQTT or AMQP servers in this card.
- Do not change pubsub MQTT topic-pattern matching.
- Do not change queue persistence, ack, transaction, or DLQ behavior.
- Do not make `x/mq` an app-facing entrypoint over `x/messaging`.

Files:
- `x/mq/config.go`
- `x/mq/broker.go`
- `x/mq/errors.go`
- `x/mq/mq_test.go`
- `docs/modules/x-mq/README.md`

Tests:
- `go test -race -timeout 60s ./x/mq/...`
- `go test -timeout 20s ./x/mq/...`
- `go vet ./x/mq/...`

Docs Sync:
Update `docs/modules/x-mq/README.md` if it does not clearly state that MQTT and
AMQP server support is unavailable and must not be enabled.

Done Definition:
- `Config{EnableMQTT: true}` and `Config{EnableAMQP: true}` fail validation
  before broker construction succeeds.
- The unsupported protocol error path is covered with `errors.Is`.
- Existing in-process queue behavior and MQTT pattern subscription tests still
  pass.
- The listed validation commands pass.

Outcome:

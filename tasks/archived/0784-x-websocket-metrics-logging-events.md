# 0784 - x/websocket metrics logging events

Status: done
Priority: P2
State: done
Primary module: `x/websocket`

## Problem

Some fields are stored but unused (`Server.debug`, `Server.logger`,
`Hub.metrics`), metrics are not consistently gated by config, and
`SecurityEvent` is exported without a public consumption API.

## Scope

- Remove unused stored state or put it behind a real behavior.
- Make metric counters and configuration semantics explicit.
- Either provide a small security-event subscription/handler API or internalize
  `SecurityEvent`.
- Update tests and docs for the chosen contract.

## Out of Scope

- Prometheus/exporter integration.
- Cross-extension observability abstractions.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

## Outcome

- Removed unused stored `Server` debug/logger state and unused `Hub.metrics`.
- Removed the obsolete `EnableMetrics` switch; hub metrics are always collected.
- Added `HubConfig.SecurityEventHandler` as the public consumption path for
  `SecurityEvent`.
- Added focused event-handler coverage and updated docs.

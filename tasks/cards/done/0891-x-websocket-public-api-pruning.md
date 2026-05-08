# 0891 - x/websocket public API pruning

Status: done
Priority: P1
Primary module: `x/websocket`

## Problem

The public entrypoint list includes helper symbols that are not core transport
contracts, including queue DTOs and heuristic content-screening helpers. Stable
promotion requires a smaller, intentional API surface.

## Scope

- Audit exported symbols that should not be stable API.
- Remove or internalize non-core exported helpers such as queued DTOs and
  heuristic content checks where no external caller requires them.
- Update module manifest, primer docs, tests, and website references.
- Follow AGENTS.md symbol-change protocol for every removed exported symbol.

## Out of Scope

- Removing required stable transport types such as `Conn`, `Hub`, or
  `ServerConfig`.
- Release promotion.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Outcome

- Internalized the connection write queue DTO by renaming `Outbound` to package-private `outbound`.
- Removed `ContainsDangerousPatterns` and its heuristic XSS/SQL screening tests from the WebSocket transport package.
- Updated `x/websocket/module.yaml` so the public entrypoint inventory no longer lists those non-core helpers.
- Re-ran the symbol search for `Outbound` and `ContainsDangerousPatterns` under `x/websocket`; both old names are absent.

Completed validations:

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`

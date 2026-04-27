# Card 0353: Route Registration State Cleanup (routesOnce Removal)

Priority: P2
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/ops, x/websocket, x/devtools/pubsubdebug
Depends On: 0974, 0976

## Goal

Remove the `routesOnce sync.Once` guards from `x/ops`, `x/websocket`, and
`x/devtools/pubsubdebug` so that route registration in those packages becomes
explicit and idempotent — consistent with the cleanup already applied to
`x/webhook` in card 0969.

## Problem

Three packages still hide HTTP route registration behind `sync.Once`, causing
`RegisterRoutes` to silently no-op on the second call instead of surfacing
duplicate registration as a router error:

- `x/ops/ops.go` line 25: `routesOnce sync.Once` — the `RegisterRoutes` body
  is wrapped in `c.routesOnce.Do(...)`, masking repeat calls
- `x/websocket/websocket.go` line 69: `routesOnce sync.Once` — same pattern in
  `Server.RegisterRoutes`
- `x/devtools/pubsubdebug/component.go` line 17: `routesOnce sync.Once` — same
  pattern in `Component.RegisterRoutes`

Card 0969 established the canonical rule: duplicate registration should return a
router error, not be silently swallowed. The same fix should be applied
uniformly.

Note: `x/ops/ops.go` will be further modified by card 0974 (adaptCtx handler
migration); `x/devtools/pubsubdebug/component.go` will be further modified by
card 0976. Whichever card executes first should remove `routesOnce` as part of
its changes, and the dependent card should adapt accordingly. If card 0977
executes independently first, it removes only `routesOnce`; the handler shapes
are touched separately.

## Scope

- Remove the `routesOnce sync.Once` field from each struct.
- Unwrap the `routesOnce.Do(func() { ... })` closure in each `RegisterRoutes`
  so the body executes directly.
- Adjust error propagation if needed: if the `Do` closure set a captured `err`
  variable, replace with a direct `if err := ...; err != nil { return err }`
  chain.
- Remove the `sync` import from each file if it becomes unused.

## Non-Goals

- Do not change handler signatures or route paths.
- Do not address adaptCtx/contract.Ctx usage in the same pass (those are
  covered by cards 0974 and 0976).
- Do not add new routes or change auth wiring.

## Files

- `x/ops/ops.go`
- `x/websocket/websocket.go`
- `x/devtools/pubsubdebug/component.go`

## Tests

```bash
rg -n 'routesOnce' x/ops x/websocket x/devtools -g '*.go'
go test -timeout 20s ./x/ops/... ./x/websocket/... ./x/devtools/...
go vet ./x/ops/... ./x/websocket/... ./x/devtools/...
```

## Docs Sync

None required — this is an internal registration mechanism, not a public API
change.

## Done Definition

- `rg 'routesOnce' x/ops x/websocket x/devtools -g '*.go'` returns no results.
- Double-calling `RegisterRoutes` on each component now propagates the duplicate
  router error instead of silently returning nil.
- All targeted package tests pass; `go vet` is clean.

## Outcome


# x/messaging Maturity Evidence

Module: `x/messaging`

Owner: `messaging`

Current status: `experimental`

Candidate status: not selected

Evidence state: initial triage

## Current Coverage

- App-facing API coverage exists for service behavior, integration flows,
  failover, monitoring, rate limiting, receipts, templates, and SMS metrics.
- The primer identifies `x/messaging` as the family entrypoint before
  subordinate `x/mq`, `x/pubsub`, `x/scheduler`, and `x/webhook` primitives.
- HTTP response guidance is already aligned with `contract.WriteResponse` and
  safe error messages.

## Boundary State

- Primer: `docs/modules/x-messaging/README.md`
- Manifest: `x/messaging/module.yaml`
- Boundary state: app-facing entrypoint is documented; subordinate primitive
  ownership still needs release observation before family promotion.

## Why This Is Not A Beta Candidate Yet

`x/messaging` depends on sibling primitives whose public contracts still need
family-level evaluation. Promoting the root before `x/mq`, `x/pubsub`,
`x/scheduler`, and `x/webhook` settle would make the compatibility promise too
broad.

## Candidate Surface Inventory

| Surface | Package | Current decision | Why | Next blocker |
| --- | --- | --- | --- | --- |
| App-facing messaging service | `x/messaging` | Likely beta candidate after inventory | Canonical user entrypoint for send, receipt, monitoring, failover, and route wiring | Freeze service, provider, receipt, route, and response shapes separately from primitives |
| Queue primitive | `x/mq` | Experimental | Durable queue and worker coordination are subordinate but broad | Inventory task lifecycle, store contracts, DLQ replay, leases, and SQL/memory stores |
| Pub/sub primitive | `x/pubsub` | Experimental | Broker, pattern matching, backpressure, replay, hooks, and TTL behavior need separate evidence | Select a small broker API subset before any beta snapshot |
| Scheduler primitive | `x/scheduler` | Experimental | Job lifecycle, store, backpressure, worker, clock, and retry behavior remain broad | Freeze the constructor/options/store interfaces across release refs |
| Webhook primitive | `x/webhook` | Experimental | Inbound verification and outbound delivery have different security and lifecycle risks | Split inbound verifier surface from outbound delivery service before promotion |

`x/messaging` remains the recommended discovery entrypoint. Subordinate
primitive promotion work must not turn `x/mq`, `x/pubsub`, `x/scheduler`, or
`x/webhook` into competing app-facing family roots.

## Next Evidence Needed

- Explicit contract inventory for app-facing `x/messaging` entrypoints.
- Subordinate primitive inventories for queue, pub/sub, scheduler, and webhook
  behavior before selecting any primitive beta target.
- Exported API snapshots after the selected surface inventory is complete.
- Release-history evidence for the selected API surface.

## Current Decision

Keep `x/messaging` experimental.

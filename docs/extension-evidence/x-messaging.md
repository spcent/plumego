# x/messaging Maturity Evidence

Module: `x/messaging`

Owner: `messaging`

Current status: `experimental`

Candidate status: `not selected`

Evidence state: surface inventory

## Current Coverage

- App-facing API coverage exists for service behavior, integration flows,
  failover, monitoring, rate limiting, receipts, templates, and SMS metrics.
- The primer identifies `x/messaging` as the family entrypoint before
  subordinate `x/mq`, `x/pubsub`, `x/scheduler`, and `x/webhook` primitives.
- HTTP response guidance is already aligned with `contract.WriteResponse` and
  safe error messages.

## Primer And Boundary State

- Primer: `docs/modules/x-messaging/README.md`
- Manifest: `x/messaging/module.yaml`
- Boundary state: app-facing entrypoint is documented; subordinate primitive
  ownership still needs release observation before family promotion.

## Why No `beta` Candidate Is Selected Yet

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

## Required Release Evidence

Missing. No selected `x/messaging` surface has two consecutive minor release
refs with unchanged exported API.

Release refs:

- none recorded

## API Snapshot Evidence

Not recorded. The app-facing service is the leading `beta` candidate, but no
release-backed or current-head API snapshot pair is checked in for that surface
yet.

## Release Evidence

`specs/extension-beta-evidence.yaml` tracks
`x/messaging:app-facing-service` as a `surface_candidate` covering the
app-facing service package. It remains blocked on release history, API
snapshots, and owner sign-off. Queue, pub/sub, scheduler, and webhook
primitives are intentionally excluded from that surface.

Current state:

- Selected release candidate: `x/messaging:app-facing-service`
- API snapshot comparison: not recorded
- Release-history comparison: not recorded

## Owner Sign-Off

Missing. No selected `x/messaging` surface has owner sign-off recorded.

## Next Evidence Needed

- Explicit contract inventory for app-facing `x/messaging` entrypoints.
- Subordinate primitive inventories for queue, pub/sub, scheduler, and webhook
  behavior before selecting any primitive beta target.
- Exported API snapshots after the selected surface inventory is complete.
- Release-history evidence for the selected API surface.

## Blockers

- The app-facing service surface is not release-frozen yet.
- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Posture

Keep `x/messaging` experimental.

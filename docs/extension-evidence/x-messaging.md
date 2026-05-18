# x/messaging Maturity Evidence

Module: `x/messaging`

Owner: `messaging`

Current status: `beta`

Candidate status: `beta`

Evidence state: complete

## Current Coverage

- App-facing API coverage exists for service behavior, integration flows,
  failover, monitoring, rate limiting, receipts, templates, and SMS metrics.
- The primer identifies `x/messaging` as the family entrypoint before
  subordinate `x/messaging/mq`, `x/messaging/pubsub`, `x/messaging/scheduler`, and `x/messaging/webhook` primitives.
- HTTP response guidance is already aligned with `contract.WriteResponse` and
  safe error messages.

## Primer And Boundary State

- Primer: `docs/modules/x-messaging/README.md`
- Manifest: `x/messaging/module.yaml`
- Boundary state: app-facing entrypoint is documented; subordinate primitive
  ownership still needs release observation before family promotion.

## Why No `beta` Candidate Is Selected Yet

`x/messaging` depends on sibling primitives whose public contracts still need
family-level evaluation. Promoting the root before `x/messaging/mq`, `x/messaging/pubsub`,
`x/messaging/scheduler`, and `x/messaging/webhook` settle would make the compatibility promise too
broad.

## Candidate Surface Inventory

| Surface | Package | Current decision | Why | Next blocker |
| --- | --- | --- | --- | --- |
| App-facing messaging service | `x/messaging` | Beta at v1.1.0 | Canonical user entrypoint for send, receipt, monitoring, failover, and route wiring | Subordinate primitives remain experimental |
| Queue primitive | `x/messaging/mq` | Experimental | Durable queue and worker coordination are subordinate but broad | Inventory task lifecycle, store contracts, DLQ replay, leases, and SQL/memory stores |
| Pub/sub primitive | `x/messaging/pubsub` | Experimental | Broker, pattern matching, backpressure, replay, hooks, and TTL behavior need separate evidence | Select a small broker API subset before any beta snapshot |
| Scheduler primitive | `x/messaging/scheduler` | Experimental | Job lifecycle, store, backpressure, worker, clock, and retry behavior remain broad | Freeze the constructor/options/store interfaces across release refs |
| Webhook primitive | `x/messaging/webhook` | Experimental | Inbound verification and outbound delivery have different security and lifecycle risks | Split inbound verifier surface from outbound delivery service before promotion |

`x/messaging` remains the recommended discovery entrypoint. Subordinate
primitive promotion work must not turn `x/messaging/mq`, `x/messaging/pubsub`, `x/messaging/scheduler`, or
`x/messaging/webhook` into competing app-facing family roots.

## Required Release Evidence

Recorded. The `x/messaging:app-facing-service` surface has two consecutive
minor release refs with unchanged exported API.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target)
- `v1.1.0`

## API Snapshot Evidence

Release-backed API snapshots are recorded for the selected app-facing service
surface. The v1 baseline intake artifacts remain useful history, but the
v1.0.0 to v1.1.0 snapshots are the promotion evidence.

Snapshot refs:

- `docs/extension-evidence/snapshots/v1-baseline/x-messaging/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-messaging/head.snapshot`
- `docs/extension-evidence/snapshots/x-messaging/base.snapshot`
- `docs/extension-evidence/snapshots/x-messaging/head.snapshot`

## Release Evidence

`specs/extension-beta-evidence.yaml` tracks
`x/messaging:app-facing-service` as a `surface_candidate` covering the
app-facing service package. Queue, pub/sub, scheduler, and webhook primitives
are intentionally excluded from that surface.

Current state:

- Selected release candidate: `x/messaging:app-facing-service`
- API snapshot comparison: `v1.0.0` to `v1.1.0`, unchanged
- Release-history comparison: two release refs recorded

## Owner Sign-Off

Signed off by `messaging` for v1.1.0:

> I confirm that the `x/messaging` app-facing service surface meets the beta
> criteria in docs/EXTENSION_STABILITY_POLICY.md and accept the beta
> compatibility obligations for the documented public surface.

## Next Evidence Needed

- Subordinate primitive inventories for queue, pub/sub, scheduler, and webhook
  behavior before selecting any primitive beta target.

## Blockers

None for the app-facing `x/messaging` service surface. Subordinate primitives
remain experimental.

## Promotion Posture

Promoted app-facing `x/messaging` service surface to `beta` at v1.1.0. Keep
`x/messaging/mq`, `x/messaging/pubsub`, `x/messaging/scheduler`, and
`x/messaging/webhook` experimental.

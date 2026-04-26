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

## Next Evidence Needed

- Explicit contract inventory for app-facing `x/messaging` entrypoints.
- Subordinate primitive maturity notes for queue, pub/sub, scheduler, and
  webhook behavior.
- Exported API snapshots after the contract inventory is complete.
- Release-history evidence for the selected API surface.

## Current Decision

Keep `x/messaging` experimental.

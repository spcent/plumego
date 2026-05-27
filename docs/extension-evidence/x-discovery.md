# x/gateway/discovery Maturity Evidence

Module: `x/gateway/discovery`

Owner: `edge`

Current status: `experimental`

Candidate status: `not selected`

Evidence state: surface inventory

## Current Coverage

- Static backend tests cover fixed service lookup behavior.
- Consul backend tests use offline HTTP servers for health API behavior.
- Kubernetes backend tests cover Endpoints API discovery and configuration
  behavior without requiring a live cluster.
- etcd backend tests cover HTTP gateway registration, lookup, and health
  behavior.
- The primer documents backend selection and standard validation commands.

## Primer And Boundary State

- Primer: `docs/modules/x/gateway/README.md`
- Manifest: `x/gateway/discovery/module.yaml`
- Boundary state: discovery is a secondary capability root, not bootstrap or
  gateway-only policy.

## Why No `beta` Candidate Is Selected Yet

The backend set expanded recently and needs release observation. Kubernetes and
etcd behavior in particular should hold a stable constructor/config/API shape
before `x/gateway/discovery` receives a beta compatibility promise.

## Candidate Surface Inventory

| Surface | Package/file | Current decision | Why | Next blocker |
| --- | --- | --- | --- | --- |
| Core discovery contract | `x/gateway/discovery/discovery.go` | Likely beta candidate after inventory | Small common service, instance, watcher, and resolver contract | Freeze interface and error semantics across release refs |
| Static backend | `x/gateway/discovery/static.go` | Likely beta candidate after inventory | Deterministic fixed instance lookup with narrow behavior | Snapshot constructor/config behavior with the core contract |
| Consul backend | `x/gateway/discovery/consul.go` | Experimental | External HTTP API adapter behavior and health mapping need observation | Prove config, registration, watch, and health behavior across releases |
| Kubernetes backend | `x/gateway/discovery/kubernetes.go` | Experimental | Endpoint discovery and port selection behavior expanded recently | Freeze config, port selection, watch, and unsupported operation behavior |
| etcd backend | `x/gateway/discovery/etcd.go` | Experimental | HTTP gateway registration, lookup, watch, and health behavior need release history | Observe constructor/config and watch semantics across release refs |

Root `x/gateway/discovery` should only become a beta target after the common contract
and at least the static backend have matching release evidence. Consul,
Kubernetes, and etcd can remain experimental adapters behind that contract.

## Required Release Evidence

Partially recorded. The `x/gateway/discovery:core-static` surface has `v1.0.0` as its
first post-v1 release-ref intake point. It still needs a second release ref with
unchanged exported API.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target)

## API Snapshot Evidence

Partially recorded. The v1 baseline intake artifacts below are checked in. They
are first-release baselines only and do not clear `api_snapshot_missing` until
a second release-backed comparison is recorded.

Snapshot refs:

- `docs/extension-evidence/snapshots/v1-baseline/x-discovery/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-discovery/head.snapshot`

## Release Evidence

`specs/extension-beta-evidence.yaml` tracks `x/gateway/discovery:core-static` as a
`surface_candidate` covering the common contract and static backend package.
It remains blocked on release history, complete API snapshot evidence, and
owner sign-off, and does not imply beta status for Consul, Kubernetes, or etcd
adapters.

Current state:

- Selected release candidate: `x/gateway/discovery:core-static`
- API snapshot comparison: `v1.0.0` to `v1.0.0`, unchanged
- Release-history comparison: first ref only

## Owner Sign-Off

Missing. No selected `x/gateway/discovery` surface has owner sign-off recorded.

## Next Evidence Needed

- Second release-backed exported API snapshot for the core contract and static
  backend candidate.
- Backend-level contract inventories for Consul, Kubernetes, and etcd.
- Two consecutive minor release refs showing no exported API changes.
- Owner sign-off after backend API shape is observed across releases.

## Blockers

- The common contract plus static backend surface still lacks checked-in API
  snapshots.
- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Posture

Keep `x/gateway/discovery` experimental.

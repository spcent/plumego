# x/discovery Maturity Evidence

Module: `x/discovery`

Owner: `edge`

Current status: `experimental`

Candidate status: not selected

Evidence state: initial triage

## Current Coverage

- Static backend tests cover fixed service lookup behavior.
- Consul backend tests use offline HTTP servers for health API behavior.
- Kubernetes backend tests cover Endpoints API discovery and configuration
  behavior without requiring a live cluster.
- etcd backend tests cover HTTP gateway registration, lookup, and health
  behavior.
- The primer documents backend selection and standard validation commands.

## Boundary State

- Primer: `docs/modules/x-discovery/README.md`
- Manifest: `x/discovery/module.yaml`
- Boundary state: discovery is a secondary capability root, not bootstrap or
  gateway-only policy.

## Why This Is Not A Beta Candidate Yet

The backend set expanded recently and needs release observation. Kubernetes and
etcd behavior in particular should hold a stable constructor/config/API shape
before `x/discovery` receives a beta compatibility promise.

## Next Evidence Needed

- Exported API snapshot for `x/discovery`.
- Backend-level contract inventory for Static, Consul, Kubernetes, and etcd.
- Two consecutive minor release refs showing no exported API changes.
- Owner sign-off after backend API shape is observed across releases.

## Current Decision

Keep `x/discovery` experimental.

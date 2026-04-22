# x/discovery

## Purpose

`x/discovery` provides service discovery primitives and adapters for gateway and client integration.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is service lookup or resolver behavior
- the task is endpoint discovery for clients or gateways
- dynamic instance discovery is required

## Do not use this module for

- application bootstrap
- transport-specific proxy policy
- stable-root durable primitives before the design is proven

## First files to read

- `x/discovery/module.yaml`
- `specs/task-routing.yaml`
- `specs/extension-taxonomy.yaml`

## Main risks when changing this module

- hidden background state
- nondeterministic resolution behavior
- transport concerns leaking into discovery internals

## Canonical change shape

- keep discovery behavior explicit
- keep resolvers small and composable
- do not spread discovery concerns across stable roots

## Boundary with bootstrap

- `x/discovery` is a secondary capability root for resolver and discovery work, not an application bootstrap surface
- keep discovery concerns out of stable roots unless they become durable primitives
- keep service discovery wiring explicit in the owning application or extension

## Available backends

| Backend | Constructor | Notes |
|---|---|---|
| `Static` | `NewStatic(services)` | Fixed URL map; useful for tests and simple deployments |
| `Consul` | `NewConsul(address, ConsulConfig)` | HashiCorp Consul health API; supports long-poll watching |
| `Kubernetes` | `NewKubernetes(KubernetesConfig)` | Kubernetes Endpoints API; auto-detects in-cluster credentials |
| `Etcd` | `NewEtcd(endpoints, EtcdConfig)` | etcd v3 HTTP gateway; supports registration and health updates |

All backends implement the `Discovery` interface:
- `Resolve(ctx, serviceName)` — returns ready backend URLs
- `Watch(ctx, serviceName)` — delivers backend-list updates on a channel; polls at `PollInterval` for Kubernetes and etcd
- `Register` / `Deregister` / `Health` — supported by etcd; not supported by Static and Kubernetes (returns `ErrNotSupported`)

## Backend selection guidance

- **Static**: tests, local development, fixed infrastructure.
- **Consul**: existing Consul deployments; use `ConsulConfig.Tag` for canary or version filtering.
- **Kubernetes**: in-cluster or out-of-cluster K8s; leave `APIServerURL`, `BearerToken`, `CAFile`, and `Namespace` empty for in-cluster auto-configuration; set `PortName` when a service exposes multiple ports.
- **etcd**: direct etcd-backed registration; useful for non-Kubernetes infrastructure or when you need explicit health management. Store `serviceID` as `"serviceName/instanceID"` for `Deregister` and `Health`.

## Standard validation

- `go test -timeout 20s ./x/discovery/...`
- `go vet ./x/discovery/...`
- backend tests use `httptest.NewServer` for offline verification; no live Consul, K8s, or etcd required

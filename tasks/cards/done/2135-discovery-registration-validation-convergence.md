# Card 2135: Discovery Registration Validation Convergence
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: done
Primary Module: x/discovery
Owned Files:
- x/discovery/discovery.go
- x/discovery/consul.go
- x/discovery/etcd.go
- x/discovery/consul_test.go
- x/discovery/etcd_test.go
Depends On: none

Goal:
Make service registration validation consistent across discovery adapters.
`Etcd.Register` rejects missing instance ID and name with `ErrInvalidConfig`,
while `Consul.Register` builds and sends a registration payload without the same
front-door validation. That makes invalid instance behavior depend on the
chosen backend.

Scope:
- Add a shared private validation helper for registration-required `Instance`
  fields.
- Use the helper from Consul and Etcd registration paths.
- Preserve `ErrNotSupported` behavior for adapters that do not implement
  registration.
- Add Consul tests mirroring the existing Etcd missing-ID/missing-name cases.

Non-goals:
- Do not change Resolve, Watch, Health, or Deregister semantics.
- Do not add retry, backoff, or health-check behavior.
- Do not move discovery ownership into stable roots.

Files:
- `x/discovery/discovery.go`
- `x/discovery/consul.go`
- `x/discovery/etcd.go`
- `x/discovery/consul_test.go`
- `x/discovery/etcd_test.go`

Tests:
- go test -race -timeout 60s ./x/discovery/...
- go test -timeout 20s ./x/discovery/...
- go vet ./x/discovery/...

Docs Sync:
Update `docs/modules/x-discovery/README.md` only if it documents registration
validation behavior.

Done Definition:
- Consul and Etcd reject missing registration ID/name through the same validation
  helper and sentinel wrapping.
- Static/Kubernetes unsupported registration behavior is unchanged.
- x/discovery validation commands pass.
- No stable root imports x/discovery.

Outcome:
- Added a shared private registration validation helper for required
  `Instance.ID` and `Instance.Name` fields.
- Routed Consul and Etcd registration through the same helper and
  `ErrInvalidConfig` wrapping path.
- Added Consul missing-ID and missing-name tests and tightened Etcd assertions
  to verify the shared sentinel.
- Validation passed:
  - `go test -race -timeout 60s ./x/discovery/...`
  - `go test -timeout 20s ./x/discovery/...`
  - `go vet ./x/discovery/...`

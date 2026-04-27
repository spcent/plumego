# Card 0494: Discovery Consul Query Assertion Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: x/discovery
Owned Files:
- `x/discovery/consul_test.go`
Depends On: none

Goal:
Name repeated Consul raw query inclusion and omission assertions.

Problem:
The Consul resolver tests manually check raw query substrings for datacenter,
passing, and tag parameters. The same query assertion shape is repeated and the
negative IncludeUnhealthy check uses a separate ad hoc message.

Scope:
- Add local helpers for asserting query fragments are present or absent.
- Use them in Consul query parameter tests.

Non-goals:
- Do not change Consul request construction or resolver behavior.
- Do not add dependencies.

Files:
- `x/discovery/consul_test.go`

Tests:
- `go test -race -timeout 60s ./x/discovery/...`
- `go test -timeout 20s ./x/discovery/...`
- `go vet ./x/discovery/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Consul query assertions use named helpers.
- The listed validation commands pass.

Outcome:
- Added `assertQueryContains` and `assertQueryOmits` for Consul raw query checks.
- Validation passed for discovery race tests, normal tests, and vet.

# Card 0853

Priority: P1
State: active
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/policy.go`
- `x/tenant/core/route_policy_cached.go`
Depends On:

Goal:
- Unexport the dependency fields of `ConfigPolicyEvaluator` and `CachedRoutePolicyProvider` to match the encapsulation pattern used by every other implementation in the package.

Problem:
- `ConfigPolicyEvaluator` (`policy.go:47`) exposes its `Provider PolicyConfigProvider` as an exported field:
  ```go
  type ConfigPolicyEvaluator struct {
      Provider PolicyConfigProvider
  }
  ```
- `CachedRoutePolicyProvider` (`route_policy_cached.go:7`) exposes both its dependencies as exported fields:
  ```go
  type CachedRoutePolicyProvider struct {
      Provider RoutePolicyProvider
      Cache    RoutePolicyCache
  }
  ```
- Every other stateful type in the package keeps its dependencies unexported:
  - `FixedWindowQuotaManager` — unexported `provider`, `counters`, `lastEviction`
  - `WindowQuotaManager` — unexported `provider`, `store`
  - `SlidingWindowQuotaManager` — unexported `provider`, `windows`
  - `TokenBucketRateLimiter` — unexported `provider`, `buckets`
  - `InMemoryConfigManager` — unexported `mu`, `configs`
- The two outliers allow a caller to mutate or replace a dependency after construction:
  ```go
  e := NewConfigPolicyEvaluator(provider)
  e.Provider = nil  // silently breaks the evaluator
  ```
  This is a direct contradiction of the constructor-injection principle enforced everywhere else.
- Exported fields also invite struct-literal construction that bypasses the constructor, producing zero-value (`Provider: nil`) instances that `Evaluate` must then silently treat as "allow all" (`policy.go:59`).

Scope:
- In `ConfigPolicyEvaluator`: rename `Provider` → `provider` (unexport).
- In `CachedRoutePolicyProvider`: rename `Provider` → `provider` and `Cache` → `cache` (unexport both).
- Remove any struct-literal construction of these types outside their constructors.
- Constructors (`NewConfigPolicyEvaluator`, `NewCachedRoutePolicyProvider`) already exist and are the correct entry points — no new constructors needed.

Non-goals:
- Do not add getter methods for the unexported fields.
- Do not change the behavior of `Evaluate` or `RoutePolicy`.
- Do not touch other types.

Files:
- `x/tenant/core/policy.go`
- `x/tenant/core/policy_test.go`
- `x/tenant/core/route_policy_cached.go`
- `x/tenant/core/route_policy_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- `ConfigPolicyEvaluator` and `CachedRoutePolicyProvider` have no exported fields.
- No code constructs these types via struct literals outside their own files.
- All tests pass.

Outcome:
- Pending.

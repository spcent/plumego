# Card 0921

Priority: P3
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/config.go`
Depends On: —

Goal:
- Complete the ConfigManager adapter set so callers with a `ConfigManager` interface can
  obtain any of the three sub-providers (quota, policy, rate-limit) without needing the
  concrete `*InMemoryConfigManager` type.

---

## Problem

**File:** `x/tenant/core/config.go` lines 60–77

The package provides one adapter that wraps a `ConfigManager` interface as a specific
sub-provider:

```go
type RateLimitConfigProviderFromConfig struct {
    Manager ConfigManager
}
func (p *RateLimitConfigProviderFromConfig) RateLimitConfig(...) (RateLimitConfig, error)
```

No equivalent adapter exists for `QuotaConfigProvider` or `PolicyConfigProvider`.

This asymmetry means:

| Sub-provider interface | Wrappable via adapter? |
|------------------------|----------------------|
| `QuotaConfigProvider` | No — must use `*InMemoryConfigManager` directly |
| `PolicyConfigProvider` | No — must use `*InMemoryConfigManager` directly |
| `RateLimitConfigProvider` | Yes — `RateLimitConfigProviderFromConfig` |

A caller who implements a custom `ConfigManager` (e.g., a DB-backed or remote-config
loader) can wrap it for rate limits but must re-implement the same delegation logic for
quota and policy:

```go
// forced to hand-write this for quota and policy:
type myQuotaAdapter struct{ m ConfigManager }
func (a *myQuotaAdapter) QuotaConfig(ctx context.Context, tenantID string) (QuotaConfig, error) {
    cfg, err := a.m.GetTenantConfig(ctx, tenantID)
    if err != nil { return QuotaConfig{}, err }
    return cfg.Quota, nil
}
```

The inconsistency also raises a question for readers: was quota/policy intentionally left
without adapters, or was `RateLimitConfigProviderFromConfig` a one-off that was never
completed?

---

## Fix

Add the two missing adapter structs to match the existing pattern:

```go
// QuotaConfigProviderFromConfig wraps a ConfigManager as a QuotaConfigProvider.
type QuotaConfigProviderFromConfig struct {
    Manager ConfigManager
}

func (p *QuotaConfigProviderFromConfig) QuotaConfig(ctx context.Context, tenantID string) (QuotaConfig, error) {
    if p == nil || p.Manager == nil {
        return QuotaConfig{}, ErrTenantNotFound
    }
    cfg, err := p.Manager.GetTenantConfig(ctx, tenantID)
    if err != nil {
        return QuotaConfig{}, err
    }
    return cfg.Quota, nil
}

// PolicyConfigProviderFromConfig wraps a ConfigManager as a PolicyConfigProvider.
type PolicyConfigProviderFromConfig struct {
    Manager ConfigManager
}

func (p *PolicyConfigProviderFromConfig) PolicyConfig(ctx context.Context, tenantID string) (PolicyConfig, error) {
    if p == nil || p.Manager == nil {
        return PolicyConfig{}, ErrTenantNotFound
    }
    cfg, err := p.Manager.GetTenantConfig(ctx, tenantID)
    if err != nil {
        return PolicyConfig{}, err
    }
    return cfg.Policy, nil
}
```

Add nil-receiver guards consistent with `RateLimitConfigProviderFromConfig`.

---

Scope:
- Add `QuotaConfigProviderFromConfig` and `PolicyConfigProviderFromConfig` to `config.go`.
- No changes to existing types or interfaces.

Non-goals:
- Do not change `InMemoryConfigManager` methods.
- Do not change `ConfigManager` interface.
- Do not remove `RateLimitConfigProviderFromConfig`.

Files:
- `x/tenant/core/config.go`
- `x/tenant/core/config_test.go`

Tests:
- Test that `QuotaConfigProviderFromConfig` and `PolicyConfigProviderFromConfig` delegate
  correctly to the wrapped manager.
- Test nil-receiver behavior returns `ErrTenantNotFound`.
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync: —

Done Definition:
- `QuotaConfigProviderFromConfig` and `PolicyConfigProviderFromConfig` exist in `config.go`.
- All three adapter types have consistent nil-guard behavior.
- New adapter tests pass.

Outcome:

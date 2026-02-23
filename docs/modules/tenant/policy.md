# Policy Evaluation

> **Package**: `github.com/spcent/plumego/tenant` | **Feature**: Per-tenant policy enforcement

## Overview

Policy evaluation allows per-tenant business rules and feature flags. Policies determine whether a tenant can access specific features, perform certain actions, or meet business requirements.

## PolicyEvaluator Interface

```go
type PolicyEvaluator interface {
    EvaluatePolicy(ctx context.Context, tenantID, policy string, input map[string]interface{}) (PolicyDecision, error)
}

type PolicyDecision struct {
    TenantID   string
    PolicyName string
    Allowed    bool
    Reason     string
    Metadata   map[string]interface{}
}
```

## Setup

```go
configMgr := tenant.NewInMemoryConfigManager()
policyEval := tenant.NewConfigPolicyEvaluator(configMgr)

app := core.New(
    core.WithTenantConfigManager(configMgr),
    core.WithTenantMiddleware(core.TenantMiddlewareOptions{
        HeaderName:      "X-Tenant-ID",
        PolicyEvaluator: policyEval,
        Hooks: tenant.Hooks{
            OnPolicy: func(ctx context.Context, d tenant.PolicyDecision) {
                log.Warn("Policy denied",
                    "tenant", d.TenantID,
                    "policy", d.PolicyName,
                    "reason", d.Reason,
                )
            },
        },
    }),
)
```

## Feature Flags

### Checking Feature Access

```go
func handleWebhookEndpoint(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())

    // Check if tenant has webhook feature
    decision, err := policyEval.EvaluatePolicy(r.Context(), tenantID, "feature.webhooks", nil)
    if err != nil || !decision.Allowed {
        http.Error(w, "Webhooks not available on your plan. Upgrade to Pro.", http.StatusForbidden)
        return
    }

    // Handle webhook...
}
```

### Feature Configuration

```go
// Configure features per plan
planFeatures := map[string][]string{
    "free":       {"api"},
    "starter":    {"api", "webhooks"},
    "pro":        {"api", "webhooks", "analytics", "exports"},
    "enterprise": {"api", "webhooks", "analytics", "exports", "sso", "audit_logs", "custom_domain"},
}

configMgr.SetTenantConfig("tenant-a", tenant.Config{
    TenantID: "tenant-a",
    Policy: tenant.PolicyConfig{
        AllowedFeatures: planFeatures["pro"],
    },
    Metadata: map[string]string{"plan": "pro"},
})
```

## Geographic Policies

### Country-Based Access

```go
configMgr.SetTenantConfig("tenant-eu", tenant.Config{
    TenantID: "tenant-eu",
    Policy: tenant.PolicyConfig{
        AllowedCountries: []string{"DE", "FR", "GB", "NL", "SE"},
    },
})

func handleGeoRestrictedEndpoint(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())
    clientCountry := getClientCountry(r)

    decision, err := policyEval.EvaluatePolicy(r.Context(), tenantID, "geo.access", map[string]interface{}{
        "country": clientCountry,
    })

    if !decision.Allowed {
        http.Error(w, "Access not available in your region", http.StatusForbidden)
        return
    }
}
```

## IP Allowlisting

```go
configMgr.SetTenantConfig("tenant-enterprise", tenant.Config{
    Policy: tenant.PolicyConfig{
        AllowedIPs: []string{"10.0.0.0/8", "192.168.1.100"},
        BlockedIPs: []string{"1.2.3.4"},
    },
})

func handleIPRestrictedEndpoint(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())
    clientIP := getClientIP(r)

    decision, _ := policyEval.EvaluatePolicy(r.Context(), tenantID, "ip.access", map[string]interface{}{
        "ip": clientIP,
    })

    if !decision.Allowed {
        http.Error(w, "Access denied from this IP address", http.StatusForbidden)
        return
    }
}
```

## Custom Policies

### Custom Rule Engine

```go
type CustomPolicyEvaluator struct {
    configMgr tenant.ConfigManager
}

func (e *CustomPolicyEvaluator) EvaluatePolicy(ctx context.Context, tenantID, policy string, input map[string]interface{}) (tenant.PolicyDecision, error) {
    config, err := e.configMgr.GetTenantConfig(ctx, tenantID)
    if err != nil {
        return tenant.PolicyDecision{}, err
    }

    switch policy {
    case "feature.sso":
        allowed := config.Policy.CustomRules["sso_enabled"] == true
        return tenant.PolicyDecision{
            TenantID:   tenantID,
            PolicyName: policy,
            Allowed:    allowed,
            Reason:     map[bool]string{true: "SSO enabled", false: "SSO not enabled for this plan"}[allowed],
        }, nil

    case "export.max_rows":
        maxRows, _ := config.Policy.CustomRules["max_export_rows"].(int)
        requestedRows, _ := input["rows"].(int)
        allowed := requestedRows <= maxRows
        return tenant.PolicyDecision{
            TenantID:   tenantID,
            PolicyName: policy,
            Allowed:    allowed,
            Reason:     fmt.Sprintf("Max export rows: %d", maxRows),
        }, nil
    }

    // Unknown policy — deny by default
    return tenant.PolicyDecision{
        TenantID:   tenantID,
        PolicyName: policy,
        Allowed:    false,
        Reason:     "Unknown policy",
    }, nil
}
```

## Policy Middleware

Apply policies automatically via middleware:

```go
func requireFeature(feature string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenantID := tenant.TenantIDFromContext(r.Context())

            decision, err := policyEval.EvaluatePolicy(r.Context(), tenantID, "feature."+feature, nil)
            if err != nil || !decision.Allowed {
                http.Error(w, fmt.Sprintf("Feature '%s' not available on your plan", feature), http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Apply to routes
app.Get("/api/analytics", handleAnalytics, requireFeature("analytics"))
app.Post("/api/webhooks", handleWebhooks, requireFeature("webhooks"))
app.Get("/api/exports", handleExports, requireFeature("exports"))
```

## Best Practices

### 1. Fail Closed

```go
decision, err := policyEval.EvaluatePolicy(ctx, tenantID, policy, input)
if err != nil || !decision.Allowed {
    return ErrPolicyDenied // Deny on error
}
```

### 2. Log Policy Denials

```go
if !decision.Allowed {
    log.Warn("Policy denied",
        "tenant_id", tenantID,
        "policy", policy,
        "reason", decision.Reason,
    )
}
```

### 3. Return Helpful Upgrade Messages

```go
if !decision.Allowed {
    http.Error(w, fmt.Sprintf(
        "%s not available on %s plan. Upgrade to Pro at example.com/upgrade",
        feature, plan,
    ), http.StatusForbidden)
}
```

## Related Documentation

- [Tenant Overview](README.md) — Multi-tenancy overview
- [Configuration](configuration.md) — Policy configuration
- [Examples](examples.md) — Complete policy examples

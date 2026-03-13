# Tenant Policy

> **Packages**: `github.com/spcent/plumego/x/tenant/core`, `github.com/spcent/plumego/x/tenant/policy`  
> **Stability**: Experimental (v1)

Tenant policy controls allowed models/tools/methods/paths per tenant.

---

## Policy Model

```go
type PolicyConfig struct {
    AllowedModels  []string
    AllowedTools   []string
    AllowedMethods []string
    AllowedPaths   []string // exact or prefix with trailing *
}
```

Empty lists mean allow-all for that dimension.

---

## Evaluator

```go
policyEval := tenant.NewConfigPolicyEvaluator(configMgr)
```

Evaluation input:

```go
type PolicyRequest struct {
    Model  string
    Tool   string
    Method string
    Path   string
}
```

---

## Middleware Integration

```go
api := app.Router().Group("/api")

api.Use(tenantpolicy.Middleware(tenantpolicy.Options{
    Evaluator: policyEval,
    Hooks: tenant.Hooks{
        OnPolicy: func(ctx context.Context, d tenant.PolicyDecision) {
            if !d.Allowed {
                log.Printf("policy denied tenant=%s reason=%s", d.TenantID, d.Reason)
            }
        },
    },
}))
```

The middleware denies with `403` when policy evaluation rejects.

---

## Example Policy Configuration

```go
configMgr.SetTenantConfig(tenant.Config{
    TenantID: "acme",
    Policy: tenant.PolicyConfig{
        AllowedModels:  []string{"gpt-4o-mini"},
        AllowedTools:   []string{"search", "calculator"},
        AllowedMethods: []string{"GET", "POST"},
        AllowedPaths:   []string{"/api/*"},
    },
})
```

---

## Testing Focus

- model/tool allowed lists
- method/path denials
- backend-error behavior (fail closed)
- hook invocation for allowed/denied decisions

# Validator Module

> Legacy note: the historical public `validator/` root has been removed.

Validation in Plumego is now owned by the transport or application layer.

## Current Options

- explicit decode with `json.NewDecoder(r.Body).Decode(...)`
- `contract.Ctx.Bind(...)` when you already use `contract`
- `security/input` for transport-level input safety helpers
- app-local validation functions for business rules

## Recommended HTTP Pattern

```go
type CreateUserRequest struct {
    Email string `json:"email"`
    Name  string `json:"name"`
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        contract.WriteError(w, r, contract.APIError{
            Status: http.StatusBadRequest,
            Code: "invalid_json",
            Message: "invalid request body",
            Category: contract.CategoryClient,
        })
        return
    }

    if req.Email == "" || req.Name == "" {
        contract.WriteError(w, r, contract.APIError{
            Status: http.StatusBadRequest,
            Code: "invalid_input",
            Message: "email and name are required",
            Category: contract.CategoryClient,
        })
        return
    }

    _ = contract.WriteResponse(w, r, http.StatusCreated, map[string]string{
        "status": "ok",
    }, nil)
}
```

## Recommended Contract Pattern

```go
type CreateUserRequest struct {
    Email string `json:"email"`
    Name  string `json:"name"`
}

func handleCreateUser(ctx *contract.Ctx) {
    var req CreateUserRequest
    if err := ctx.BindJSON(&req); err != nil {
        _ = ctx.ErrorJSON(http.StatusBadRequest, "invalid_input", "invalid request body", nil)
        return
    }
    if req.Email == "" || req.Name == "" {
        _ = ctx.ErrorJSON(http.StatusBadRequest, "invalid_input", "email and name are required", nil)
        return
    }
}
```

## Guidance

- keep transport validation close to the handler
- keep business rules in application services
- fail closed on malformed input
- do not introduce a new shared validation root package

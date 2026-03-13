# Request Validation

> Legacy note: the historical public `validator/` root has been removed.

This page shows the current request-validation flow for Plumego applications.

## JSON Body

```go
type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        contract.WriteError(w, r, contract.APIError{
            Status: http.StatusBadRequest,
            Code: "invalid_json",
            Message: "invalid request body",
            Category: contract.CategoryClient,
        })
        return
    }

    if req.Email == "" {
        contract.WriteError(w, r, contract.APIError{
            Status: http.StatusBadRequest,
            Code: "email_required",
            Message: "email is required",
            Category: contract.CategoryClient,
        })
        return
    }
    if len(req.Password) < 8 {
        contract.WriteError(w, r, contract.APIError{
            Status: http.StatusBadRequest,
            Code: "password_too_short",
            Message: "password must be at least 8 characters",
            Category: contract.CategoryClient,
        })
        return
    }
}
```

## Query Parameters

```go
func handleSearch(w http.ResponseWriter, r *http.Request) {
    q := strings.TrimSpace(r.URL.Query().Get("q"))
    if len(q) < 2 {
        contract.WriteError(w, r, contract.APIError{
            Status: http.StatusBadRequest,
            Code: "query_too_short",
            Message: "q must be at least 2 characters",
            Category: contract.CategoryClient,
        })
        return
    }
}
```

## Security Helpers

Use `security/input` for reusable transport-level checks such as email, URL, or phone validation. Keep business-specific policies in app code.

## Cross-Field Rules

```go
func validatePasswordReset(req ResetPasswordRequest) error {
    if req.NewPassword != req.ConfirmPassword {
        return fmt.Errorf("password confirmation mismatch")
    }
    if req.NewPassword == req.OldPassword {
        return fmt.Errorf("new password must differ from old password")
    }
    return nil
}
```

## Guidance

- decode first
- validate required fields second
- apply cross-field or business rules third
- write one structured error response path

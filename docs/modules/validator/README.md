# Validator Module

> **Package Path**: `github.com/spcent/plumego/validator` | **Stability**: High | **Priority**: P1

## Overview

The `validator/` package provides request validation for Plumego applications. It validates incoming request data (JSON, form, query parameters) against defined rules, ensuring data integrity and security before processing.

**Key Features**:
- **Struct Tag Validation**: Declarative validation using struct tags
- **Built-in Rules**: Required, min/max, email, URL, regex, and more
- **Custom Validators**: Define application-specific validation logic
- **Error Messages**: Detailed, field-specific error messages
- **Type Safety**: Validates types, formats, and constraints
- **Internationalization**: Customizable error messages

## Quick Start

### Basic Validation

```go
import "github.com/spcent/plumego/validator"

type CreateUserRequest struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
    Age      int    `validate:"required,min=18,max=120"`
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest

    // Decode JSON
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate
    if err := validator.Validate(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Request is valid, proceed...
}
```

### With Contract Context

```go
import "github.com/spcent/plumego/contract"

func handleCreateUser(ctx *contract.Context) {
    var req CreateUserRequest

    // Bind and validate in one step
    if err := ctx.Bind(&req); err != nil {
        ctx.Error(http.StatusBadRequest, err)
        return
    }

    // Request is valid, proceed...
}
```

## Validation Rules

### Required

```go
type User struct {
    Email string `validate:"required"`
    Name  string `validate:"required"`
}
```

### String Length

```go
type User struct {
    Username string `validate:"required,min=3,max=20"`
    Password string `validate:"required,min=8"`
}
```

### Numeric Range

```go
type Product struct {
    Price    float64 `validate:"required,min=0.01"`
    Quantity int     `validate:"required,min=1,max=1000"`
    Age      int     `validate:"min=18,max=120"`
}
```

### Email

```go
type User struct {
    Email string `validate:"required,email"`
}
```

### URL

```go
type Webhook struct {
    URL string `validate:"required,url"`
}
```

### Regular Expression

```go
type User struct {
    Phone string `validate:"required,regex=^\\+[1-9]\\d{1,14}$"` // E.164 format
    Code  string `validate:"required,regex=^[A-Z]{3}$"`         // 3 uppercase letters
}
```

### Enum/OneOf

```go
type Order struct {
    Status string `validate:"required,oneof=pending processing shipped delivered"`
}
```

### Slice/Array

```go
type User struct {
    Tags []string `validate:"required,min=1,max=5,dive,min=2,max=20"`
    //                      ^^^^^^^^^^^^ slice validation
    //                                     ^^^^^^^^^^^^^^^^ item validation
}
```

### Nested Structs

```go
type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    Zip     string `validate:"required,regex=^\\d{5}$"`
}

type User struct {
    Name    string  `validate:"required"`
    Address Address `validate:"required"` // Validates nested struct
}
```

## Custom Validators

### Register Custom Validator

```go
// Define custom validator function
func validateUsername(value string) bool {
    // Username must:
    // - Start with letter
    // - Contain only alphanumeric and underscore
    // - Length 3-20
    match, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]{2,19}$`, value)
    return match
}

// Register validator
validator.RegisterValidator("username", validateUsername)

// Use in struct
type User struct {
    Username string `validate:"required,username"`
}
```

### Complex Custom Validator

```go
func validateCreditCard(value string) bool {
    // Remove spaces
    value = strings.ReplaceAll(value, " ", "")

    // Check length (13-19 digits)
    if len(value) < 13 || len(value) > 19 {
        return false
    }

    // Luhn algorithm
    sum := 0
    alternate := false

    for i := len(value) - 1; i >= 0; i-- {
        digit := int(value[i] - '0')

        if alternate {
            digit *= 2
            if digit > 9 {
                digit -= 9
            }
        }

        sum += digit
        alternate = !alternate
    }

    return sum%10 == 0
}

validator.RegisterValidator("creditcard", validateCreditCard)

// Usage
type Payment struct {
    CardNumber string `validate:"required,creditcard"`
}
```

## Error Handling

### Validation Errors

```go
err := validator.Validate(req)
if err != nil {
    // err contains field-specific errors
    // "email: must be a valid email address"
    // "age: must be at least 18"
}
```

### Structured Errors

```go
if err := validator.Validate(req); err != nil {
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        errors := make(map[string]string)
        for _, fieldErr := range validationErrors {
            errors[fieldErr.Field] = fieldErr.Message
        }

        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Validation failed",
            "fields": errors,
        })
        return
    }
}
```

**Response**:
```json
{
  "error": "Validation failed",
  "fields": {
    "email": "must be a valid email address",
    "age": "must be at least 18"
  }
}
```

## Validation Middleware

### Global Validation Middleware

```go
func validationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate before proceeding
        if err := validateRequest(r); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### Per-Route Validation

```go
func validate[T any](handler func(*contract.Context, T)) contract.HandlerFunc {
    return func(ctx *contract.Context) {
        var req T

        if err := ctx.Bind(&req); err != nil {
            ctx.Error(http.StatusBadRequest, err)
            return
        }

        handler(ctx, req)
    }
}

// Usage
app.Post("/users", validate[CreateUserRequest](handleCreateUser))
```

## Common Patterns

### Request Validation

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if err := validator.Validate(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Create user...
}
```

### Query Parameter Validation

```go
type ListUsersQuery struct {
    Page     int    `query:"page" validate:"min=1"`
    PageSize int    `query:"page_size" validate:"min=1,max=100"`
    SortBy   string `query:"sort_by" validate:"oneof=name email created_at"`
    Order    string `query:"order" validate:"oneof=asc desc"`
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
    var query ListUsersQuery

    // Parse query parameters
    query.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
    query.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
    query.SortBy = r.URL.Query().Get("sort_by")
    query.Order = r.URL.Query().Get("order")

    // Apply defaults
    if query.Page == 0 {
        query.Page = 1
    }
    if query.PageSize == 0 {
        query.PageSize = 20
    }

    // Validate
    if err := validator.Validate(query); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // List users...
}
```

### Conditional Validation

```go
type UpdateUserRequest struct {
    Email    *string `json:"email,omitempty" validate:"omitempty,email"`
    Password *string `json:"password,omitempty" validate:"omitempty,min=8"`
    Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
}
// Only validates fields that are present (non-nil)
```

## Testing

### Unit Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        request CreateUserRequest
        wantErr bool
    }{
        {
            name: "Valid request",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Password: "securepass123",
                Age:      25,
            },
            wantErr: false,
        },
        {
            name: "Invalid email",
            request: CreateUserRequest{
                Email:    "invalid-email",
                Password: "securepass123",
                Age:      25,
            },
            wantErr: true,
        },
        {
            name: "Password too short",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Password: "short",
                Age:      25,
            },
            wantErr: true,
        },
        {
            name: "Age below minimum",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Password: "securepass123",
                Age:      16,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Validate(tt.request)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Best Practices

### 1. Validate Early

```go
// ✅ Good: Validate immediately after parsing
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    return err
}
if err := validator.Validate(req); err != nil {
    return err
}

// ❌ Bad: Processing before validation
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    return err
}
processData(req.Email) // No validation!
```

### 2. Use Descriptive Error Messages

```go
// ✅ Good: Clear, actionable errors
"email: must be a valid email address"
"age: must be at least 18"

// ❌ Bad: Vague errors
"validation failed"
"invalid input"
```

### 3. Combine with Security

```go
// Validate AND sanitize
if err := validator.Validate(req); err != nil {
    return err
}

// Sanitize HTML input
req.Bio = html.EscapeString(req.Bio)
```

### 4. Set Reasonable Limits

```go
type User struct {
    Username string `validate:"required,min=3,max=20"`     // Not too short, not too long
    Bio      string `validate:"max=500"`                   // Limit text length
    Tags     []string `validate:"max=10,dive,max=20"`      // Limit array size
}
```

## Module Documentation

- **[Request Validation](request-validation.md)** — Complete validation guide with examples

## Related Documentation

- [Contract: Request](../contract/request.md) — Request binding
- [Security: Input](../security/input.md) — Input validation and sanitization
- [Middleware: Bind](../middleware/bind.md) — Request binding middleware

## Reference Implementation

See examples:
- `examples/reference/` — Request validation in production
- `examples/crud/` — CRUD API with validation

---

**Next Steps**:
1. Read [Request Validation](request-validation.md) for detailed validation patterns
2. Define validation rules for your API endpoints
3. Test validation with unit tests

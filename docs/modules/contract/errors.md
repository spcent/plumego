# Error Handling

> **Package**: `github.com/spcent/plumego/contract`

Plumego provides structured error handling with categories, codes, and detailed information for API responses.

---

## Table of Contents

- [Overview](#overview)
- [Error Structure](#error-structure)
- [Error Categories](#error-categories)
- [Creating Errors](#creating-errors)
- [Writing Errors](#writing-errors)
- [Predefined Errors](#predefined-errors)
- [Custom Errors](#custom-errors)
- [Best Practices](#best-practices)

---

## Overview

### Why Structured Errors?

Structured errors provide:
- **Consistency**: Uniform error format across API
- **Debugging**: Detailed information for troubleshooting
- **Client Handling**: Machine-readable error codes
- **Categorization**: Group errors by type
- **Localization**: Support for translated messages

### Error Format

```json
{
  "status": 404,
  "code": "USER_NOT_FOUND",
  "message": "User not found",
  "category": "client",
  "details": {
    "userId": "123"
  }
}
```

---

## Error Structure

### Error Type

```go
type Error struct {
    Status   int                    // HTTP status code
    Code     string                 // Machine-readable error code
    Message  string                 // Human-readable message
    Category ErrorCategory          // Error category
    Details  map[string]interface{} // Additional details
}
```

### Fields

- **Status**: HTTP status code (400, 404, 500, etc.)
- **Code**: Unique error code (e.g., "USER_NOT_FOUND")
- **Message**: Human-readable description
- **Category**: Error type (client, server, validation, etc.)
- **Details**: Additional context-specific information

---

## Error Categories

```go
type ErrorCategory string

const (
    CategoryClient         ErrorCategory = "client"         // 4xx errors
    CategoryServer         ErrorCategory = "server"         // 5xx errors
    CategoryBusiness       ErrorCategory = "business"       // Business logic
    CategoryValidation     ErrorCategory = "validation"     // Input validation
    CategoryAuthentication ErrorCategory = "authentication" // Auth failures
    CategoryAuthorization  ErrorCategory = "authorization"  // Permission denied
    CategoryRateLimit      ErrorCategory = "rate_limit"     // Rate limiting
    CategoryTimeout        ErrorCategory = "timeout"        // Timeouts
)
```

### Category Usage

```go
// Client errors (400-499)
err := contract.NewError(
    contract.WithStatus(http.StatusBadRequest),
    contract.WithCategory(contract.CategoryClient),
)

// Server errors (500-599)
err := contract.NewError(
    contract.WithStatus(http.StatusInternalServerError),
    contract.WithCategory(contract.CategoryServer),
)

// Validation errors
err := contract.NewError(
    contract.WithStatus(http.StatusBadRequest),
    contract.WithCategory(contract.CategoryValidation),
)
```

---

## Creating Errors

### Basic Error

```go
import "github.com/spcent/plumego/contract"

err := contract.NewError(
    contract.WithStatus(http.StatusNotFound),
    contract.WithCode("RESOURCE_NOT_FOUND"),
    contract.WithMessage("Resource not found"),
)
```

### Error with All Fields

```go
err := contract.NewError(
    contract.WithStatus(http.StatusBadRequest),
    contract.WithCode("INVALID_INPUT"),
    contract.WithMessage("Invalid user input"),
    contract.WithCategory(contract.CategoryValidation),
    contract.WithDetails(map[string]interface{}{
        "field": "email",
        "issue": "invalid format",
        "value": "not-an-email",
    }),
)
```

### Error Options

```go
// Status code
contract.WithStatus(statusCode int)

// Error code
contract.WithCode(code string)

// Message
contract.WithMessage(message string)

// Category
contract.WithCategory(category ErrorCategory)

// Details
contract.WithDetails(details map[string]interface{})
```

---

## Writing Errors

### WriteError

Send error response to client:

```go
import "github.com/spcent/plumego/contract"

app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    id := plumego.Param(r, "id")

    user, err := getUser(id)
    if err != nil {
        cerr := contract.NewError(
            contract.WithStatus(http.StatusNotFound),
            contract.WithCode("USER_NOT_FOUND"),
            contract.WithMessage("User not found"),
            contract.WithDetails(map[string]interface{}{
                "userId": id,
            }),
        )

        contract.WriteError(w, r, cerr)
        return
    }

    json.NewEncoder(w).Encode(user)
})
```

### With Context

```go
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")

    user, err := getUser(id)
    if err != nil {
        cerr := contract.NewError(
            contract.WithStatus(http.StatusNotFound),
            contract.WithCode("USER_NOT_FOUND"),
            contract.WithMessage("User not found"),
        )

        contract.WriteError(ctx.Writer, ctx.Request, cerr)
        return
    }

    ctx.JSON(http.StatusOK, user)
})
```

### Simple Error

```go
app.GetCtx("/protected", func(ctx *plumego.Context) {
    token := ctx.GetHeader("Authorization")
    if token == "" {
        ctx.Error(http.StatusUnauthorized, "Missing authorization token")
        return
    }

    // Continue...
})
```

---

## Predefined Errors

### Validation Error

```go
err := contract.NewValidationError(field string, reason string)

// Example
err := contract.NewValidationError("email", "invalid format")

// Result:
// Status: 400
// Code: "VALIDATION_ERROR"
// Message: "Validation failed"
// Category: "validation"
// Details: {"field": "email", "reason": "invalid format"}
```

### Not Found Error

```go
err := contract.NewNotFoundError(resource string)

// Example
err := contract.NewNotFoundError("User")

// Result:
// Status: 404
// Code: "NOT_FOUND"
// Message: "User not found"
// Category: "client"
```

### Unauthorized Error

```go
err := contract.NewUnauthorizedError(reason string)

// Example
err := contract.NewUnauthorizedError("Invalid token")

// Result:
// Status: 401
// Code: "UNAUTHORIZED"
// Message: "Unauthorized: Invalid token"
// Category: "authentication"
```

### Forbidden Error

```go
err := contract.NewForbiddenError(reason string)

// Example
err := contract.NewForbiddenError("Insufficient permissions")

// Result:
// Status: 403
// Code: "FORBIDDEN"
// Message: "Forbidden: Insufficient permissions"
// Category: "authorization"
```

### Internal Server Error

```go
err := contract.NewInternalServerError(message string)

// Example
err := contract.NewInternalServerError("Database connection failed")

// Result:
// Status: 500
// Code: "INTERNAL_SERVER_ERROR"
// Message: "Database connection failed"
// Category: "server"
```

### Rate Limit Error

```go
err := contract.NewRateLimitError(retryAfter int)

// Example
err := contract.NewRateLimitError(60) // Retry after 60 seconds

// Result:
// Status: 429
// Code: "RATE_LIMIT_EXCEEDED"
// Message: "Rate limit exceeded"
// Category: "rate_limit"
// Details: {"retryAfter": 60}
```

---

## Custom Errors

### Application-Specific Errors

```go
// User already exists
func NewUserExistsError(email string) *contract.Error {
    return contract.NewError(
        contract.WithStatus(http.StatusConflict),
        contract.WithCode("USER_EXISTS"),
        contract.WithMessage("User already exists"),
        contract.WithCategory(contract.CategoryBusiness),
        contract.WithDetails(map[string]interface{}{
            "email": email,
        }),
    )
}

// Usage
if userExists(email) {
    err := NewUserExistsError(email)
    contract.WriteError(w, r, err)
    return
}
```

### Domain Errors

```go
// Insufficient balance
func NewInsufficientBalanceError(balance, required float64) *contract.Error {
    return contract.NewError(
        contract.WithStatus(http.StatusPaymentRequired),
        contract.WithCode("INSUFFICIENT_BALANCE"),
        contract.WithMessage("Insufficient account balance"),
        contract.WithCategory(contract.CategoryBusiness),
        contract.WithDetails(map[string]interface{}{
            "balance":  balance,
            "required": required,
            "shortage": required - balance,
        }),
    )
}

// Order not cancellable
func NewOrderNotCancellableError(orderID string, status string) *contract.Error {
    return contract.NewError(
        contract.WithStatus(http.StatusBadRequest),
        contract.WithCode("ORDER_NOT_CANCELLABLE"),
        contract.WithMessage(fmt.Sprintf("Order cannot be cancelled in %s status", status)),
        contract.WithCategory(contract.CategoryBusiness),
        contract.WithDetails(map[string]interface{}{
            "orderId": orderID,
            "status":  status,
        }),
    )
}
```

---

## Complete Examples

### REST API with Error Handling

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/contract"
    "github.com/spcent/plumego"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var users = make(map[string]User)

func main() {
    app := core.New()

    // Get user with error handling
    app.GetCtx("/users/:id", func(ctx *plumego.Context) {
        id := ctx.Param("id")

        user, exists := users[id]
        if !exists {
            err := contract.NewNotFoundError("User")
            contract.WriteError(ctx.Writer, ctx.Request, err)
            return
        }

        ctx.JSON(http.StatusOK, user)
    })

    // Create user with validation
    app.PostCtx("/users", func(ctx *plumego.Context) {
        var user User

        if err := ctx.Bind(&user); err != nil {
            cerr := contract.NewValidationError("body", err.Error())
            contract.WriteError(ctx.Writer, ctx.Request, cerr)
            return
        }

        // Check if user exists
        for _, u := range users {
            if u.Email == user.Email {
                cerr := contract.NewError(
                    contract.WithStatus(http.StatusConflict),
                    contract.WithCode("USER_EXISTS"),
                    contract.WithMessage("User with this email already exists"),
                    contract.WithCategory(contract.CategoryBusiness),
                    contract.WithDetails(map[string]interface{}{
                        "email": user.Email,
                    }),
                )
                contract.WriteError(ctx.Writer, ctx.Request, cerr)
                return
            }
        }

        // Create user
        user.ID = generateID()
        users[user.ID] = user

        ctx.Status(http.StatusCreated).JSON(user)
    })

    // Delete with authorization
    app.DeleteCtx("/users/:id", func(ctx *plumego.Context) {
        // Check authorization
        token := ctx.GetHeader("Authorization")
        if token == "" {
            err := contract.NewUnauthorizedError("Missing token")
            contract.WriteError(ctx.Writer, ctx.Request, err)
            return
        }

        // Verify token
        userID, err := verifyToken(token)
        if err != nil {
            cerr := contract.NewUnauthorizedError("Invalid token")
            contract.WriteError(ctx.Writer, ctx.Request, cerr)
            return
        }

        id := ctx.Param("id")

        // Check if user can delete
        if userID != id && !isAdmin(userID) {
            cerr := contract.NewForbiddenError("Cannot delete other users")
            contract.WriteError(ctx.Writer, ctx.Request, cerr)
            return
        }

        // Check if user exists
        if _, exists := users[id]; !exists {
            cerr := contract.NewNotFoundError("User")
            contract.WriteError(ctx.Writer, ctx.Request, cerr)
            return
        }

        // Delete user
        delete(users, id)
        ctx.NoContent(http.StatusNoContent)
    })

    app.Boot()
}
```

### Error Middleware

```go
func errorRecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                // Log panic
                log.Printf("Panic recovered: %v", rec)

                // Send error response
                err := contract.NewInternalServerError("Internal server error")
                contract.WriteError(w, r, err)
            }
        }()

        next.ServeHTTP(w, r)
    })
}

// Usage
app := core.New()
app.Use(errorRecoveryMiddleware)
```

---

## Best Practices

### ✅ Do

1. **Use Structured Errors**
   ```go
   // ✅ Structured
   err := contract.NewNotFoundError("User")
   contract.WriteError(w, r, err)

   // ❌ Plain text
   http.Error(w, "Not found", 404)
   ```

2. **Provide Meaningful Codes**
   ```go
   // ✅ Descriptive
   contract.WithCode("USER_NOT_FOUND")
   contract.WithCode("INVALID_EMAIL_FORMAT")
   contract.WithCode("INSUFFICIENT_BALANCE")

   // ❌ Generic
   contract.WithCode("ERROR")
   contract.WithCode("BAD_REQUEST")
   ```

3. **Include Helpful Details**
   ```go
   // ✅ With context
   contract.WithDetails(map[string]interface{}{
       "field":    "email",
       "expected": "valid email format",
       "received": "not-an-email",
   })

   // ❌ No context
   contract.WithMessage("Invalid input")
   ```

4. **Use Appropriate Categories**
   ```go
   // ✅ Correct category
   contract.WithCategory(contract.CategoryValidation) // For invalid input
   contract.WithCategory(contract.CategoryAuthentication) // For auth
   contract.WithCategory(contract.CategoryBusiness) // For domain logic

   // ❌ Generic category
   contract.WithCategory(contract.CategoryClient) // For everything
   ```

### ❌ Don't

1. **Don't Expose Internal Errors**
   ```go
   // ❌ Exposing internal details
   err := contract.NewError(
       contract.WithMessage(dbError.Error()), // Database error exposed!
   )

   // ✅ Generic message
   err := contract.NewInternalServerError("Failed to process request")
   // Log the real error internally
   log.Printf("Database error: %v", dbError)
   ```

2. **Don't Use Wrong Status Codes**
   ```go
   // ❌ Wrong status
   contract.WithStatus(http.StatusOK) // For errors!

   // ✅ Appropriate status
   contract.WithStatus(http.StatusBadRequest)
   contract.WithStatus(http.StatusNotFound)
   contract.WithStatus(http.StatusInternalServerError)
   ```

3. **Don't Ignore Error Returns**
   ```go
   // ❌ Ignoring error
   contract.WriteError(w, r, err)
   ctx.JSON(200, data) // Still writes response!

   // ✅ Return after error
   contract.WriteError(w, r, err)
   return
   ```

---

## Error Response Format

### JSON Response

```json
{
  "status": 400,
  "code": "VALIDATION_ERROR",
  "message": "Invalid user input",
  "category": "validation",
  "details": {
    "field": "email",
    "issue": "invalid format",
    "value": "not-an-email"
  }
}
```

### Headers

```
HTTP/1.1 400 Bad Request
Content-Type: application/json
X-Request-ID: abc-123-def
```

---

## Testing Errors

```go
package main

import (
    "net/http/httptest"
    "testing"
    "github.com/spcent/plumego/contract"
)

func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name       string
        err        *contract.Error
        wantStatus int
        wantCode   string
    }{
        {
            name:       "Not Found",
            err:        contract.NewNotFoundError("User"),
            wantStatus: 404,
            wantCode:   "NOT_FOUND",
        },
        {
            name:       "Validation",
            err:        contract.NewValidationError("email", "invalid"),
            wantStatus: 400,
            wantCode:   "VALIDATION_ERROR",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := httptest.NewRecorder()
            r := httptest.NewRequest("GET", "/", nil)

            contract.WriteError(w, r, tt.err)

            if w.Code != tt.wantStatus {
                t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
            }

            // Verify response body contains error code
            body := w.Body.String()
            if !strings.Contains(body, tt.wantCode) {
                t.Errorf("Body missing code %s", tt.wantCode)
            }
        })
    }
}
```

---

## Next Steps

- **[Context](context.md)** - Request context
- **[Response](response.md)** - Response helpers
- **[Request](request.md)** - Request parsing

---

**Related**:
- [Contract Overview](README.md)
- [Validator Module](../validator/)
- [Middleware Module](../middleware/)

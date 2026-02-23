# Request Validation

> **Package**: `github.com/spcent/plumego/validator` | **Complete validation guide**

## Overview

This guide provides comprehensive examples and patterns for validating HTTP requests in Plumego applications. Request validation is the first line of defense against invalid or malicious input.

## Complete Examples

### Example 1: User Registration

```go
type RegisterRequest struct {
    Email           string `json:"email" validate:"required,email"`
    Password        string `json:"password" validate:"required,min=8,regex=^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
    FirstName       string `json:"first_name" validate:"required,min=2,max=50"`
    LastName        string `json:"last_name" validate:"required,min=2,max=50"`
    DateOfBirth     string `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
    PhoneNumber     string `json:"phone_number" validate:"omitempty,regex=^\\+[1-9]\\d{1,14}$"`
    AcceptTerms     bool   `json:"accept_terms" validate:"required,eq=true"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest

    // Decode JSON
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate
    if err := validator.Validate(req); err != nil {
        writeValidationError(w, err)
        return
    }

    // Additional business logic validation
    if exists, _ := db.EmailExists(req.Email); exists {
        http.Error(w, "Email already registered", http.StatusConflict)
        return
    }

    // Proceed with registration...
}
```

### Example 2: Product Creation

```go
type CreateProductRequest struct {
    Name        string   `json:"name" validate:"required,min=3,max=100"`
    Description string   `json:"description" validate:"required,min=10,max=1000"`
    Price       float64  `json:"price" validate:"required,min=0.01,max=999999.99"`
    Stock       int      `json:"stock" validate:"required,min=0,max=100000"`
    Category    string   `json:"category" validate:"required,oneof=electronics clothing food books"`
    Tags        []string `json:"tags" validate:"required,min=1,max=10,dive,min=2,max=30"`
    Images      []string `json:"images" validate:"required,min=1,max=5,dive,url"`
    SKU         string   `json:"sku" validate:"required,regex=^[A-Z]{3}-\\d{6}$"` // e.g., ABC-123456
    IsActive    bool     `json:"is_active" validate:"required"`
}

func handleCreateProduct(w http.ResponseWriter, r *http.Request) {
    var req CreateProductRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if err := validator.Validate(req); err != nil {
        writeValidationError(w, err)
        return
    }

    // Create product...
}
```

### Example 3: Search with Filters

```go
type SearchProductsRequest struct {
    Query       string   `query:"q" validate:"required,min=2,max=100"`
    Category    string   `query:"category" validate:"omitempty,oneof=electronics clothing food books"`
    MinPrice    float64  `query:"min_price" validate:"omitempty,min=0"`
    MaxPrice    float64  `query:"max_price" validate:"omitempty,min=0,gtfield=MinPrice"`
    InStock     bool     `query:"in_stock"`
    SortBy      string   `query:"sort_by" validate:"omitempty,oneof=price name created_at"`
    SortOrder   string   `query:"sort_order" validate:"omitempty,oneof=asc desc"`
    Page        int      `query:"page" validate:"omitempty,min=1"`
    PageSize    int      `query:"page_size" validate:"omitempty,min=1,max=100"`
    Tags        []string `query:"tags" validate:"omitempty,max=5,dive,min=2,max=30"`
}

func handleSearchProducts(w http.ResponseWriter, r *http.Request) {
    var req SearchProductsRequest

    // Parse query parameters
    req.Query = r.URL.Query().Get("q")
    req.Category = r.URL.Query().Get("category")
    req.MinPrice, _ = strconv.ParseFloat(r.URL.Query().Get("min_price"), 64)
    req.MaxPrice, _ = strconv.ParseFloat(r.URL.Query().Get("max_price"), 64)
    req.InStock, _ = strconv.ParseBool(r.URL.Query().Get("in_stock"))
    req.SortBy = r.URL.Query().Get("sort_by")
    req.SortOrder = r.URL.Query().Get("sort_order")
    req.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
    req.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))

    // Apply defaults
    if req.Page == 0 {
        req.Page = 1
    }
    if req.PageSize == 0 {
        req.PageSize = 20
    }
    if req.SortBy == "" {
        req.SortBy = "created_at"
    }
    if req.SortOrder == "" {
        req.SortOrder = "desc"
    }

    // Validate
    if err := validator.Validate(req); err != nil {
        writeValidationError(w, err)
        return
    }

    // Search products...
}
```

### Example 4: Update with Partial Fields

```go
type UpdateUserRequest struct {
    Email     *string `json:"email,omitempty" validate:"omitempty,email"`
    FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=2,max=50"`
    LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2,max=50"`
    Bio       *string `json:"bio,omitempty" validate:"omitempty,max=500"`
    Website   *string `json:"website,omitempty" validate:"omitempty,url"`
    Age       *int    `json:"age,omitempty" validate:"omitempty,min=18,max=120"`
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
    var req UpdateUserRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate only non-nil fields
    if err := validator.Validate(req); err != nil {
        writeValidationError(w, err)
        return
    }

    // At least one field must be provided
    if req.Email == nil && req.FirstName == nil && req.LastName == nil &&
        req.Bio == nil && req.Website == nil && req.Age == nil {
        http.Error(w, "At least one field must be provided", http.StatusBadRequest)
        return
    }

    // Update user with non-nil fields...
}
```

## Cross-Field Validation

### Password Confirmation

```go
type ChangePasswordRequest struct {
    OldPassword     string `json:"old_password" validate:"required,min=8"`
    NewPassword     string `json:"new_password" validate:"required,min=8,nefield=OldPassword"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
```

### Date Range

```go
type DateRangeRequest struct {
    StartDate string `json:"start_date" validate:"required,datetime=2006-01-02"`
    EndDate   string `json:"end_date" validate:"required,datetime=2006-01-02,gtfield=StartDate"`
}

// Custom validator for date range
func validateDateRange(req DateRangeRequest) error {
    start, _ := time.Parse("2006-01-02", req.StartDate)
    end, _ := time.Parse("2006-01-02", req.EndDate)

    if end.Sub(start) > 365*24*time.Hour {
        return errors.New("date range cannot exceed 1 year")
    }

    return nil
}
```

### Conditional Required

```go
type ShippingRequest struct {
    DeliveryType string  `json:"delivery_type" validate:"required,oneof=standard express pickup"`
    Address      *string `json:"address" validate:"required_if=DeliveryType standard DeliveryType express"`
    PostalCode   *string `json:"postal_code" validate:"required_if=DeliveryType standard DeliveryType express"`
    StoreID      *string `json:"store_id" validate:"required_if=DeliveryType pickup"`
}
```

## Custom Validation Functions

### Username Validator

```go
func validateUsername(value string) bool {
    // Must start with letter
    if len(value) == 0 || !unicode.IsLetter(rune(value[0])) {
        return false
    }

    // Only alphanumeric and underscore
    for _, char := range value {
        if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
            return false
        }
    }

    // Length 3-20
    return len(value) >= 3 && len(value) <= 20
}

validator.RegisterValidator("username", validateUsername)

// Usage
type User struct {
    Username string `validate:"required,username"`
}
```

### Slug Validator

```go
func validateSlug(value string) bool {
    // Lowercase letters, numbers, and hyphens only
    match, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, value)
    return match && len(value) >= 3 && len(value) <= 100
}

validator.RegisterValidator("slug", validateSlug)

// Usage
type Article struct {
    Slug string `validate:"required,slug"`
}
```

### Business-Specific Validators

```go
// Validate discount percentage
func validateDiscount(value float64) bool {
    return value >= 0 && value <= 100
}

// Validate shipping weight
func validateWeight(value float64) bool {
    return value > 0 && value <= 1000 // Max 1000kg
}

// Register validators
validator.RegisterValidator("discount", validateDiscount)
validator.RegisterValidator("weight", validateWeight)

// Usage
type Coupon struct {
    DiscountPercent float64 `validate:"required,discount"`
}

type Package struct {
    WeightKg float64 `validate:"required,weight"`
}
```

## Error Response Formatting

### Simple Error Response

```go
func writeValidationError(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": err.Error(),
    })
}
```

### Detailed Error Response

```go
func writeValidationError(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)

    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        fieldErrors := make(map[string]string)
        for _, fieldErr := range validationErrors {
            fieldErrors[fieldErr.Field] = formatErrorMessage(fieldErr)
        }

        json.NewEncoder(w).Encode(map[string]interface{}{
            "error":  "Validation failed",
            "fields": fieldErrors,
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": err.Error(),
    })
}

func formatErrorMessage(fieldErr validator.FieldError) string {
    switch fieldErr.Tag {
    case "required":
        return fmt.Sprintf("%s is required", fieldErr.Field)
    case "email":
        return "must be a valid email address"
    case "min":
        return fmt.Sprintf("must be at least %s characters", fieldErr.Param)
    case "max":
        return fmt.Sprintf("must be at most %s characters", fieldErr.Param)
    case "url":
        return "must be a valid URL"
    default:
        return fmt.Sprintf("validation failed: %s", fieldErr.Tag)
    }
}
```

**Response**:
```json
{
  "error": "Validation failed",
  "fields": {
    "email": "must be a valid email address",
    "password": "must be at least 8 characters",
    "age": "must be at least 18"
  }
}
```

## Testing Validation

### Table-Driven Tests

```go
func TestCreateUserValidation(t *testing.T) {
    tests := []struct {
        name    string
        request CreateUserRequest
        wantErr bool
        errMsg  string
    }{
        {
            name: "Valid request",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Password: "SecurePass123",
                Name:     "John Doe",
                Age:      25,
            },
            wantErr: false,
        },
        {
            name: "Missing email",
            request: CreateUserRequest{
                Password: "SecurePass123",
                Name:     "John Doe",
                Age:      25,
            },
            wantErr: true,
            errMsg:  "email is required",
        },
        {
            name: "Invalid email format",
            request: CreateUserRequest{
                Email:    "invalid-email",
                Password: "SecurePass123",
                Name:     "John Doe",
                Age:      25,
            },
            wantErr: true,
            errMsg:  "email must be valid",
        },
        {
            name: "Password too short",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Password: "short",
                Name:     "John Doe",
                Age:      25,
            },
            wantErr: true,
            errMsg:  "password must be at least 8 characters",
        },
        {
            name: "Age below minimum",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Password: "SecurePass123",
                Name:     "John Doe",
                Age:      16,
            },
            wantErr: true,
            errMsg:  "age must be at least 18",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Validate(tt.request)

            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("Expected error message containing %q, got %q", tt.errMsg, err.Error())
            }
        })
    }
}
```

## Best Practices

### 1. Validate at API Boundary

```go
// ✅ Good: Validate immediately after parsing
func handler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)

    if err := validator.Validate(req); err != nil {
        writeValidationError(w, err)
        return
    }

    // Process validated request
}
```

### 2. Separate Validation from Business Logic

```go
// ✅ Good: Clear separation
if err := validator.Validate(req); err != nil {
    return err
}

if err := businessValidation(req); err != nil {
    return err
}

// ❌ Bad: Mixed validation
if err := validator.Validate(req); err != nil {
    return err
}
if exists, _ := db.EmailExists(req.Email); exists {
    return errors.New("email exists")
}
```

### 3. Use Appropriate Error Codes

```go
// Validation error: 400 Bad Request
if err := validator.Validate(req); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

// Business logic error: 409 Conflict
if exists, _ := db.EmailExists(req.Email); exists {
    http.Error(w, "Email already exists", http.StatusConflict)
    return
}
```

## Related Documentation

- [Validator Overview](README.md) — Validator module overview
- [Contract: Request](../contract/request.md) — Request binding
- [Security: Input](../security/input.md) — Input sanitization

## Reference Implementation

See examples:
- `examples/reference/` — Complete validation examples
- `examples/crud/` — CRUD API with validation

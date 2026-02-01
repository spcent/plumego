package router

// This file provides usage examples for the enhanced CRUD framework.
// These examples demonstrate how to use the new features in real-world scenarios.

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/validator"
)

// ================================================
// Example 1: Basic Resource Controller
// ================================================

// User represents a user resource
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,minLength=2,maxLength=50"`
	Email string `json:"email" validate:"required,secureEmail"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name,omitempty" validate:"minLength=2,maxLength=50"`
	Email string `json:"email,omitempty" validate:"secureEmail"`
}

// UserController demonstrates a complete CRUD implementation
type UserController struct {
	*BaseContextResourceController
	validator *validator.Validator
}

// NewUserController creates a new user controller with enhanced features
func NewUserController() *UserController {
	ctrl := &UserController{
		BaseContextResourceController: NewBaseContextResourceController("user"),
		validator:                     validator.NewValidator(nil),
	}

	// Configure query builder with allowed fields
	ctrl.QueryBuilder.
		WithPageSize(20, 100).
		WithAllowedSorts("name", "email", "created_at").
		WithAllowedFilters("status", "role")

	// Configure custom hooks
	ctrl.Hooks = &UserHooks{}

	// Configure transformer
	ctrl.Transformer = &UserTransformer{}

	return ctrl
}

// IndexCtx lists users with pagination, filtering, and sorting
func (c *UserController) IndexCtx(ctx *contract.Ctx) {
	// Parse query parameters
	params := c.QueryBuilder.Parse(ctx.R)

	// Call before hook
	if err := c.Hooks.BeforeList(ctx.R.Context(), params); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Simulate database query
	users, total := c.fetchUsers(params)

	// Transform results
	transformedUsers, err := c.Transformer.TransformCollection(ctx.R.Context(), users)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Call after hook
	if err := c.Hooks.AfterList(ctx.R.Context(), params, transformedUsers); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Create pagination metadata
	pagination := NewPaginationMeta(params.Page, params.PageSize, total)

	// Return paginated response
	ctx.JSON(http.StatusOK, PaginatedResponse{
		Data:       transformedUsers,
		Pagination: pagination,
	})
}

// ShowCtx retrieves a single user by ID
func (c *UserController) ShowCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	// Simulate database query
	user := c.findUserByID(id)
	if user == nil {
		ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	// Transform result
	transformedUser, err := c.Transformer.Transform(ctx.R.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, transformedUser)
}

// CreateCtx creates a new user
func (c *UserController) CreateCtx(ctx *contract.Ctx) {
	var req CreateUserRequest

	// Bind and validate request
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := c.validator.Validate(req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": "Validation failed", "details": err})
		return
	}

	// Call before hook
	if err := c.Hooks.BeforeCreate(ctx.R.Context(), &req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Create user (simulated)
	user := c.createUser(&req)

	// Call after hook
	if err := c.Hooks.AfterCreate(ctx.R.Context(), user); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Transform result
	transformedUser, err := c.Transformer.Transform(ctx.R.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, transformedUser)
}

// UpdateCtx updates an existing user
func (c *UserController) UpdateCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	var req UpdateUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := c.validator.Validate(req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": "Validation failed", "details": err})
		return
	}

	// Call before hook
	if err := c.Hooks.BeforeUpdate(ctx.R.Context(), id, &req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Update user (simulated)
	user := c.updateUser(id, &req)
	if user == nil {
		ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	// Call after hook
	if err := c.Hooks.AfterUpdate(ctx.R.Context(), id, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Transform result
	transformedUser, err := c.Transformer.Transform(ctx.R.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, transformedUser)
}

// DeleteCtx deletes a user
func (c *UserController) DeleteCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	// Call before hook
	if err := c.Hooks.BeforeDelete(ctx.R.Context(), id); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Delete user (simulated)
	if !c.deleteUser(id) {
		ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	// Call after hook
	if err := c.Hooks.AfterDelete(ctx.R.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// BatchCreateCtx creates multiple users in a single request
func (c *UserController) BatchCreateCtx(ctx *contract.Ctx) {
	var requests []CreateUserRequest
	if err := ctx.BindJSON(&requests); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	processor := NewBatchProcessor(100)
	var items []any
	for i := range requests {
		items = append(items, &requests[i])
	}

	result := processor.Process(ctx.R.Context(), items, func(ctx context.Context, item any) error {
		req := item.(*CreateUserRequest)
		if err := c.validator.Validate(req); err != nil {
			return err
		}
		// Create user (simulated)
		_ = c.createUser(req)
		return nil
	})

	ctx.JSON(http.StatusOK, result)
}

// ================================================
// Example 2: Custom Hooks Implementation
// ================================================

// UserHooks implements custom lifecycle hooks for users
type UserHooks struct {
	NoOpResourceHooks
}

func (h *UserHooks) BeforeCreate(ctx context.Context, data any) error {
	// Example: validate business rules before creating
	req := data.(*CreateUserRequest)
	if req.Email == "admin@example.com" {
		return fmt.Errorf("cannot create user with reserved email")
	}
	return nil
}

func (h *UserHooks) AfterCreate(ctx context.Context, data any) error {
	// Example: send welcome email, log event, etc.
	user := data.(*User)
	fmt.Printf("User created: %s (%s)\n", user.Name, user.Email)
	return nil
}

func (h *UserHooks) BeforeDelete(ctx context.Context, id string) error {
	// Example: prevent deletion of admin users
	if id == "admin-id" {
		return fmt.Errorf("cannot delete admin user")
	}
	return nil
}

// ================================================
// Example 3: Custom Transformer Implementation
// ================================================

// UserTransformer transforms user entities for API responses
type UserTransformer struct{}

func (t *UserTransformer) Transform(ctx context.Context, resource any) (any, error) {
	user := resource.(*User)
	return map[string]any{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format(time.RFC3339),
		"updated_at": user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (t *UserTransformer) TransformCollection(ctx context.Context, resources any) (any, error) {
	users := resources.([]*User)
	result := make([]map[string]any, len(users))
	for i, user := range users {
		transformed, err := t.Transform(ctx, user)
		if err != nil {
			return nil, err
		}
		result[i] = transformed.(map[string]any)
	}
	return result, nil
}

// ================================================
// Simulated Database Operations (for example purposes)
// ================================================

func (c *UserController) fetchUsers(_ *QueryParams) ([]*User, int64) {
	// This would normally query a database
	users := []*User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "2", Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	return users, 2
}

func (c *UserController) findUserByID(id string) *User {
	// This would normally query a database
	if id == "1" {
		return &User{ID: "1", Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	}
	return nil
}

func (c *UserController) createUser(req *CreateUserRequest) *User {
	// This would normally insert into a database
	return &User{
		ID:        "new-id",
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *UserController) updateUser(id string, req *UpdateUserRequest) *User {
	// This would normally update a database record
	if id == "1" {
		user := &User{ID: id, Name: req.Name, Email: req.Email, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		return user
	}
	return nil
}

func (c *UserController) deleteUser(id string) bool {
	// This would normally delete from a database
	return id == "1"
}

// ================================================
// Example 4: Router Registration
// ================================================

/*
Example of how to register the UserController with a router:

```go
func RegisterUserRoutes(r *router.Router) {
	ctrl := NewUserController()

	// RESTful routes
	r.Get("/users", ctrl.IndexCtx)
	r.Get("/users/:id", ctrl.ShowCtx)
	r.Post("/users", ctrl.CreateCtx)
	r.Put("/users/:id", ctrl.UpdateCtx)
	r.Delete("/users/:id", ctrl.DeleteCtx)
	r.Patch("/users/:id", ctrl.PatchCtx)

	// Batch operations
	r.Post("/users/batch", ctrl.BatchCreateCtx)
	r.Delete("/users/batch", ctrl.BatchDeleteCtx)
}
```

Example API requests:

1. List users with pagination and filtering:
   GET /users?page=1&page_size=20&sort=-created_at&status=active

2. Search users:
   GET /users?search=alice&fields=name,email

3. Get a single user:
   GET /users/123

4. Create a user:
   POST /users
   {
     "name": "Charlie",
     "email": "charlie@example.com"
   }

5. Update a user:
   PUT /users/123
   {
     "name": "Charlie Updated"
   }

6. Delete a user:
   DELETE /users/123

7. Batch create users:
   POST /users/batch
   [
     {"name": "User 1", "email": "user1@example.com"},
     {"name": "User 2", "email": "user2@example.com"}
   ]
*/

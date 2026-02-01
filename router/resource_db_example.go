package router

// This file demonstrates how to use the Database CRUD framework
// to build a complete RESTful API with database integration.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/spcent/plumego/store/db"
)

// ================================================
// Example 1: User Management with Database
// ================================================

// UserModel represents a user in the database
type UserModel struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Status    string    `json:"status" db:"status"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserCreateRequest represents the request for creating a user
type UserCreateRequest struct {
	Name   string `json:"name" validate:"required,minLength=2,maxLength=50"`
	Email  string `json:"email" validate:"required,secureEmail"`
	Status string `json:"status" validate:"required,in=active,inactive"`
	Role   string `json:"role" validate:"required,in=admin,user,guest"`
}

// UserUpdateRequest represents the request for updating a user
type UserUpdateRequest struct {
	Name   string `json:"name,omitempty" validate:"minLength=2,maxLength=50"`
	Email  string `json:"email,omitempty" validate:"secureEmail"`
	Status string `json:"status,omitempty" validate:"in=active,inactive"`
	Role   string `json:"role,omitempty" validate:"in=admin,user,guest"`
}

// UserRepository provides database operations for users
type UserRepository struct {
	*BaseRepository[UserModel]
}

// NewUserRepository creates a new user repository
func NewUserRepository(database db.DB) *UserRepository {
	builder := NewSQLBuilder("users", "id").
		WithColumns("id", "name", "email", "status", "role", "created_at", "updated_at").
		WithScanFunc(scanUser).
		WithInsertFunc(userInsertValues).
		WithUpdateFunc(userUpdateValues)

	return &UserRepository{
		BaseRepository: NewBaseRepository[UserModel](database, builder),
	}
}

// scanUser scans a database row into a UserModel
func scanUser(rows *sql.Rows) (any, error) {
	var user UserModel
	err := rows.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Status,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// userInsertValues extracts values for INSERT statement
func userInsertValues(data any) ([]any, error) {
	user, ok := data.(*UserModel)
	if !ok {
		return nil, fmt.Errorf("expected *UserModel, got %T", data)
	}

	return []any{
		user.ID,
		user.Name,
		user.Email,
		user.Status,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	}, nil
}

// userUpdateValues extracts values for UPDATE statement
func userUpdateValues(data any) ([]any, error) {
	user, ok := data.(*UserModel)
	if !ok {
		return nil, fmt.Errorf("expected *UserModel, got %T", data)
	}

	return []any{
		user.Name,
		user.Email,
		user.Status,
		user.Role,
		user.UpdatedAt,
	}, nil
}

// Custom methods for UserRepository

// FindByEmail finds a user by email address
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*UserModel, error) {
	query := "SELECT id, name, email, status, role, created_at, updated_at FROM users WHERE email = ?"
	rows, err := db.QueryContext(ctx, r.db, query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	user, err := scanUser(rows)
	if err != nil {
		return nil, err
	}

	return user.(*UserModel), nil
}

// FindActiveUsers finds all active users
func (r *UserRepository) FindActiveUsers(ctx context.Context, params *QueryParams) ([]UserModel, int64, error) {
	// Modify params to filter by status
	if params.Filters == nil {
		params.Filters = make(map[string]string)
	}
	params.Filters["status"] = "active"

	return r.FindAll(ctx, params)
}

// ================================================
// Example 2: User Controller with Full CRUD
// ================================================

// UserDBController handles HTTP requests for user management
type UserDBController struct {
	*DBResourceController[UserModel]
	repo *UserRepository
}

// NewUserDBController creates a new user controller
func NewUserDBController(database db.DB) *UserDBController {
	repo := NewUserRepository(database)
	ctrl := NewDBResourceController[UserModel]("user", repo)

	// Configure query builder
	ctrl.QueryBuilder.
		WithPageSize(20, 100).
		WithAllowedSorts("name", "email", "created_at", "updated_at").
		WithAllowedFilters("status", "role")

	// Configure hooks
	ctrl.Hooks = &UserDBHooks{}

	// Configure transformer
	ctrl.Transformer = &UserDBTransformer{}

	return &UserDBController{
		DBResourceController: ctrl,
		repo:                 repo,
	}
}

// ================================================
// Example 3: Custom Hooks for Database Operations
// ================================================

// UserDBHooks provides lifecycle hooks for user operations
type UserDBHooks struct {
	NoOpResourceHooks
}

func (h *UserDBHooks) BeforeCreate(ctx context.Context, data any) error {
	user := data.(*UserModel)

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Generate ID if not set
	if user.ID == "" {
		user.ID = generateID() // Implement your ID generation logic
	}

	// Validate business rules
	if user.Email == "admin@example.com" && user.Role != "admin" {
		return fmt.Errorf("this email is reserved for admin users")
	}

	return nil
}

func (h *UserDBHooks) AfterCreate(ctx context.Context, data any) error {
	user := data.(*UserModel)
	// Send welcome email, emit events, etc.
	fmt.Printf("User created: %s (%s)\n", user.Name, user.Email)
	return nil
}

func (h *UserDBHooks) BeforeUpdate(ctx context.Context, id string, data any) error {
	user := data.(*UserModel)

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Validate business rules
	if user.Role == "admin" {
		// Additional validation for admin role changes
	}

	return nil
}

func (h *UserDBHooks) AfterUpdate(ctx context.Context, id string, data any) error {
	user := data.(*UserModel)
	// Invalidate cache, emit events, etc.
	fmt.Printf("User updated: %s (ID: %s)\n", user.Name, id)
	return nil
}

func (h *UserDBHooks) BeforeDelete(ctx context.Context, id string) error {
	// Prevent deletion of specific users
	if id == "admin-user-id" {
		return fmt.Errorf("cannot delete system admin user")
	}
	return nil
}

func (h *UserDBHooks) AfterDelete(ctx context.Context, id string) error {
	// Clean up related resources, emit events, etc.
	fmt.Printf("User deleted: %s\n", id)
	return nil
}

// ================================================
// Example 4: Custom Transformer for Database Models
// ================================================

// UserDBTransformer transforms UserModel for API responses
type UserDBTransformer struct{}

func (t *UserDBTransformer) Transform(ctx context.Context, resource any) (any, error) {
	user, ok := resource.(*UserModel)
	if !ok {
		return nil, fmt.Errorf("expected *UserModel, got %T", resource)
	}

	return map[string]any{
		"id":         user.ID,
		"name":       user.Name,
		"email":      maskEmail(user.Email),
		"status":     user.Status,
		"role":       user.Role,
		"created_at": user.CreatedAt.Format(time.RFC3339),
		"updated_at": user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (t *UserDBTransformer) TransformCollection(ctx context.Context, resources any) (any, error) {
	users, ok := resources.([]UserModel)
	if !ok {
		return nil, fmt.Errorf("expected []UserModel, got %T", resources)
	}

	result := make([]map[string]any, len(users))
	for i, user := range users {
		transformed, err := t.Transform(ctx, &user)
		if err != nil {
			return nil, err
		}
		result[i] = transformed.(map[string]any)
	}

	return result, nil
}

// ================================================
// Example 5: Product Management with Soft Delete
// ================================================

// ProductModel represents a product with soft delete support
type ProductModel struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Price       float64    `json:"price" db:"price"`
	Stock       int        `json:"stock" db:"stock"`
	CategoryID  string     `json:"category_id" db:"category_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ProductRepository with soft delete support
type ProductRepository struct {
	*BaseRepository[ProductModel]
}

// NewProductRepository creates a new product repository with soft delete
func NewProductRepository(database db.DB) *ProductRepository {
	builder := NewSQLBuilder("products", "id").
		WithColumns("id", "name", "description", "price", "stock", "category_id", "created_at", "updated_at", "deleted_at").
		WithScanFunc(scanProduct).
		WithInsertFunc(productInsertValues).
		WithUpdateFunc(productUpdateValues)

	repo := &ProductRepository{
		BaseRepository: NewBaseRepository[ProductModel](database, builder),
	}

	return repo
}

// SoftDelete implements soft delete for products
func (r *ProductRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	query := "UPDATE products SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL"
	result, err := db.ExecContext(ctx, r.db, query, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// FindAllActive returns only non-deleted products
func (r *ProductRepository) FindAllActive(ctx context.Context, params *QueryParams) ([]ProductModel, int64, error) {
	// Add filter for non-deleted products
	if params.Filters == nil {
		params.Filters = make(map[string]string)
	}
	// This is a simplified approach; in production, you'd modify the SQL builder
	// to handle NULL checks for deleted_at

	return r.FindAll(ctx, params)
}

func scanProduct(rows *sql.Rows) (any, error) {
	var product ProductModel
	err := rows.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func productInsertValues(data any) ([]any, error) {
	product, ok := data.(*ProductModel)
	if !ok {
		return nil, fmt.Errorf("expected *ProductModel, got %T", data)
	}

	return []any{
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		product.CreatedAt,
		product.UpdatedAt,
		product.DeletedAt,
	}, nil
}

func productUpdateValues(data any) ([]any, error) {
	product, ok := data.(*ProductModel)
	if !ok {
		return nil, fmt.Errorf("expected *ProductModel, got %T", data)
	}

	return []any{
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		product.UpdatedAt,
		product.DeletedAt,
	}, nil
}

// ================================================
// Example 6: Router Registration
// ================================================

/*
Example of how to register the database-backed controllers:

```go
import (
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/store/db"
)

func setupRoutes(r *router.Router, database db.DB) {
	// User routes
	userCtrl := router.NewUserDBController(database)
	r.Get("/users", userCtrl.IndexCtx)
	r.Get("/users/:id", userCtrl.ShowCtx)
	r.Post("/users", userCtrl.CreateCtx)
	r.Put("/users/:id", userCtrl.UpdateCtx)
	r.Delete("/users/:id", userCtrl.DeleteCtx)

	// Product routes
	productCtrl := router.NewProductDBController(database)
	r.Get("/products", productCtrl.IndexCtx)
	r.Get("/products/:id", productCtrl.ShowCtx)
	r.Post("/products", productCtrl.CreateCtx)
	r.Put("/products/:id", productCtrl.UpdateCtx)
	r.Delete("/products/:id", productCtrl.DeleteCtx)
}
```

Example database schema (SQLite):

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    role TEXT NOT NULL DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);

CREATE TABLE products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price REAL NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    category_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_deleted_at ON products(deleted_at);
```

Example API requests:

1. List users with pagination and filtering:
   GET /users?page=1&page_size=20&sort=-created_at&status=active&role=admin

2. Search users:
   GET /users?search=john&fields=id,name,email

3. Get a single user:
   GET /users/user-123

4. Create a user:
   POST /users
   {
     "name": "Alice Smith",
     "email": "alice@example.com",
     "status": "active",
     "role": "user"
   }

5. Update a user:
   PUT /users/user-123
   {
     "name": "Alice Johnson",
     "status": "inactive"
   }

6. Delete a user:
   DELETE /users/user-123

7. List products (non-deleted only):
   GET /products?page=1&page_size=50&sort=name

8. Create a product:
   POST /products
   {
     "name": "Laptop",
     "description": "High-performance laptop",
     "price": 1299.99,
     "stock": 50,
     "category_id": "electronics"
   }
*/

// ================================================
// Helper Functions
// ================================================

// generateID generates a unique ID for database records
// In production, use UUID or your preferred ID generation strategy
func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}

// maskEmail masks part of the email for privacy
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	local := parts[0]
	if len(local) <= 2 {
		return email
	}

	masked := local[0:1] + strings.Repeat("*", len(local)-2) + local[len(local)-1:]
	return masked + "@" + parts[1]
}

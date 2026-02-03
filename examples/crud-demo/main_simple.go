package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/router"
)

// User model
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InMemoryUserRepository
type InMemoryUserRepository struct {
	users map[string]*User
	mu    sync.RWMutex
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	repo := &InMemoryUserRepository{
		users: make(map[string]*User),
	}

	now := time.Now()
	repo.users = map[string]*User{
		"user-1": {ID: "user-1", Name: "Alice Johnson", Email: "alice@example.com", Status: "active", Role: "admin", CreatedAt: now, UpdatedAt: now},
		"user-2": {ID: "user-2", Name: "Bob Smith", Email: "bob@example.com", Status: "active", Role: "user", CreatedAt: now, UpdatedAt: now},
		"user-3": {ID: "user-3", Name: "Charlie Brown", Email: "charlie@example.com", Status: "active", Role: "user", CreatedAt: now, UpdatedAt: now},
		"user-4": {ID: "user-4", Name: "Diana Prince", Email: "diana@example.com", Status: "active", Role: "user", CreatedAt: now, UpdatedAt: now},
		"user-5": {ID: "user-5", Name: "Eve Davis", Email: "eve@example.com", Status: "inactive", Role: "user", CreatedAt: now, UpdatedAt: now},
	}

	return repo
}

func (r *InMemoryUserRepository) FindAll(ctx context.Context, params *router.QueryParams) ([]User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []User
	for _, user := range r.users {
		if len(params.Filters) > 0 {
			match := true
			if status, ok := params.Filters["status"]; ok && user.Status != status {
				match = false
			}
			if role, ok := params.Filters["role"]; ok && user.Role != role {
				match = false
			}
			if !match {
				continue
			}
		}

		if params.Search != "" {
			if !strings.Contains(strings.ToLower(user.Name), strings.ToLower(params.Search)) &&
				!strings.Contains(strings.ToLower(user.Email), strings.ToLower(params.Search)) {
				continue
			}
		}

		users = append(users, *user)
	}

	total := int64(len(users))
	start := params.Offset
	end := params.Offset + params.Limit
	if start > len(users) {
		return []User{}, total, nil
	}
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], total, nil
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepository) Create(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return fmt.Errorf("user already exists")
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) Update(ctx context.Context, id string, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return fmt.Errorf("user not found")
	}
	r.users[id] = user
	return nil
}

func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return fmt.Errorf("user not found")
	}
	delete(r.users, id)
	return nil
}

// UserHooks
type UserHooks struct {
	router.NoOpResourceHooks
}

func (h *UserHooks) BeforeCreate(ctx context.Context, data any) error {
	user := data.(*User)
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if user.ID == "" {
		user.ID = fmt.Sprintf("user-%d", now.UnixNano())
	}
	if user.Status == "" {
		user.Status = "active"
	}
	if user.Role == "" {
		user.Role = "user"
	}

	log.Printf("[✓] Creating user: %s (%s)", user.Name, user.Email)
	return nil
}

func (h *UserHooks) AfterCreate(ctx context.Context, data any) error {
	user := data.(*User)
	log.Printf("[✓] User created: %s (ID: %s)", user.Name, user.ID)
	return nil
}

func (h *UserHooks) BeforeUpdate(ctx context.Context, id string, data any) error {
	user := data.(*User)
	user.UpdatedAt = time.Now()
	log.Printf("[✓] Updating user: %s", id)
	return nil
}

func (h *UserHooks) AfterDelete(ctx context.Context, id string) error {
	log.Printf("[✓] User deleted: %s", id)
	return nil
}

// UserTransformer
type UserTransformer struct{}

func (t *UserTransformer) Transform(ctx context.Context, resource any) (any, error) {
	user := resource.(*User)
	return map[string]any{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"status":     user.Status,
		"role":       user.Role,
		"created_at": user.CreatedAt.Format(time.RFC3339),
		"updated_at": user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (t *UserTransformer) TransformCollection(ctx context.Context, resources any) (any, error) {
	users := resources.([]User)
	result := make([]map[string]any, len(users))
	for i, user := range users {
		transformed, _ := t.Transform(ctx, &user)
		result[i] = transformed.(map[string]any)
	}
	return result, nil
}

// UserController
type UserController struct {
	*router.BaseContextResourceController
	repo *InMemoryUserRepository
}

func NewUserController(repo *InMemoryUserRepository) *UserController {
	ctrl := &UserController{
		BaseContextResourceController: router.NewBaseContextResourceController("user"),
		repo:                          repo,
	}

	ctrl.QueryBuilder.WithPageSize(10, 50).
		WithAllowedSorts("name", "email", "created_at", "updated_at").
		WithAllowedFilters("status", "role")

	ctrl.Hooks = &UserHooks{}
	ctrl.Transformer = &UserTransformer{}

	return ctrl
}

func (c *UserController) IndexCtx(ctx *contract.Ctx) {
	params := c.QueryBuilder.Parse(ctx.R)
	results, total, err := c.repo.FindAll(ctx.R.Context(), params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	transformed, _ := c.Transformer.TransformCollection(ctx.R.Context(), results)
	pagination := router.NewPaginationMeta(params.Page, params.PageSize, total)

	ctx.JSON(http.StatusOK, router.PaginatedResponse{
		Data:       transformed,
		Pagination: pagination,
	})
}

func (c *UserController) ShowCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "ID required"})
		return
	}

	result, err := c.repo.FindByID(ctx.R.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
		return
	}

	transformed, _ := c.Transformer.Transform(ctx.R.Context(), result)
	ctx.JSON(http.StatusOK, transformed)
}

func (c *UserController) CreateCtx(ctx *contract.Ctx) {
	var data User
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	c.Hooks.BeforeCreate(ctx.R.Context(), &data)
	if err := c.repo.Create(ctx.R.Context(), &data); err != nil {
		ctx.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	c.Hooks.AfterCreate(ctx.R.Context(), &data)

	transformed, _ := c.Transformer.Transform(ctx.R.Context(), &data)
	ctx.JSON(http.StatusCreated, transformed)
}

func (c *UserController) UpdateCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	var data User
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	data.ID = id
	c.Hooks.BeforeUpdate(ctx.R.Context(), id, &data)
	if err := c.repo.Update(ctx.R.Context(), id, &data); err != nil {
		ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	c.Hooks.AfterUpdate(ctx.R.Context(), id, &data)

	transformed, _ := c.Transformer.Transform(ctx.R.Context(), &data)
	ctx.JSON(http.StatusOK, transformed)
}

func (c *UserController) DeleteCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	c.Hooks.BeforeDelete(ctx.R.Context(), id)
	if err := c.repo.Delete(ctx.R.Context(), id); err != nil {
		ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	c.Hooks.AfterDelete(ctx.R.Context(), id)
	ctx.W.WriteHeader(http.StatusNoContent)
}

func main() {
	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithRequestID(),
		core.WithLogging(),
		core.WithRecovery(),
		core.WithCORS(),
	)

	repo := NewInMemoryUserRepository()
	ctrl := NewUserController(repo)

	// Routes
	app.GetHandler("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h1>🚀 CRUD Demo</h1>
<p>Try: <code>curl http://localhost:8080/users</code></p></body></html>`)
	}))

	app.GetCtx("/users", ctrl.IndexCtx)
	app.GetCtx("/users/:id", ctrl.ShowCtx)
	app.PostCtx("/users", ctrl.CreateCtx)
	app.PutCtx("/users/:id", ctrl.UpdateCtx)
	app.DeleteCtx("/users/:id", ctrl.DeleteCtx)

	log.Println("🚀 Server: http://localhost:8080")
	log.Println("📚 GET /users - List users")
	log.Println("📚 POST /users - Create user")

	if err := app.Boot(); err != nil {
		log.Fatal(err)
	}
}

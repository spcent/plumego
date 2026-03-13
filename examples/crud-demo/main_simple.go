package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	plog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
)

// User represents a demo user resource.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type userPatch struct {
	Name   *string `json:"name,omitempty"`
	Email  *string `json:"email,omitempty"`
	Status *string `json:"status,omitempty"`
	Role   *string `json:"role,omitempty"`
}

// InMemoryUserRepository is a minimal thread-safe store for the demo.
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	now := time.Now()
	return &InMemoryUserRepository{users: map[string]User{
		"user-1": {ID: "user-1", Name: "Alice Johnson", Email: "alice@example.com", Status: "active", Role: "admin", CreatedAt: now, UpdatedAt: now},
		"user-2": {ID: "user-2", Name: "Bob Smith", Email: "bob@example.com", Status: "active", Role: "user", CreatedAt: now, UpdatedAt: now},
		"user-3": {ID: "user-3", Name: "Charlie Brown", Email: "charlie@example.com", Status: "active", Role: "user", CreatedAt: now, UpdatedAt: now},
		"user-4": {ID: "user-4", Name: "Diana Prince", Email: "diana@example.com", Status: "active", Role: "user", CreatedAt: now, UpdatedAt: now},
		"user-5": {ID: "user-5", Name: "Eve Davis", Email: "eve@example.com", Status: "inactive", Role: "user", CreatedAt: now, UpdatedAt: now},
	}}
}

func (r *InMemoryUserRepository) List(status, role, search string, page, pageSize int) ([]User, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]User, 0, len(r.users))
	needle := strings.ToLower(strings.TrimSpace(search))

	for _, u := range r.users {
		if status != "" && u.Status != status {
			continue
		}
		if role != "" && u.Role != role {
			continue
		}
		if needle != "" {
			if !strings.Contains(strings.ToLower(u.Name), needle) && !strings.Contains(strings.ToLower(u.Email), needle) {
				continue
			}
		}
		filtered = append(filtered, u)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ID < filtered[j].ID
	})

	total := len(filtered)
	start := (page - 1) * pageSize
	if start >= total {
		return []User{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	out := make([]User, end-start)
	copy(out, filtered[start:end])
	return out, total
}

func (r *InMemoryUserRepository) Get(id string) (User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	return u, ok
}

func (r *InMemoryUserRepository) Create(u User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[u.ID]; exists {
		return User{}, fmt.Errorf("user already exists")
	}
	r.users[u.ID] = u
	return u, nil
}

func (r *InMemoryUserRepository) Update(id string, patch userPatch) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.users[id]
	if !ok {
		return User{}, fmt.Errorf("user not found")
	}

	if patch.Name != nil {
		u.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.Email != nil {
		u.Email = strings.TrimSpace(*patch.Email)
	}
	if patch.Status != nil {
		u.Status = strings.TrimSpace(*patch.Status)
	}
	if patch.Role != nil {
		u.Role = strings.TrimSpace(*patch.Role)
	}
	u.UpdatedAt = time.Now()

	r.users[id] = u
	return u, nil
}

func (r *InMemoryUserRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return false
	}
	delete(r.users, id)
	return true
}

type userController struct {
	repo *InMemoryUserRepository
}

func newUserController(repo *InMemoryUserRepository) *userController {
	return &userController{repo: repo}
}

func parsePage(v string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func (c *userController) list(ctx *contract.Ctx) {
	page := parsePage(ctx.Query.Get("page"), 1)
	pageSize := parsePage(ctx.Query.Get("page_size"), 10)
	if pageSize > 100 {
		pageSize = 100
	}

	users, total := c.repo.List(
		strings.TrimSpace(ctx.Query.Get("status")),
		strings.TrimSpace(ctx.Query.Get("role")),
		ctx.Query.Get("search"),
		page,
		pageSize,
	)
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	_ = ctx.JSON(http.StatusOK, map[string]any{
		"data": users,
		"pagination": map[string]any{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (c *userController) show(ctx *contract.Ctx) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		_ = ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	u, found := c.repo.Get(id)
	if !found {
		_ = ctx.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	_ = ctx.JSON(http.StatusOK, u)
}

func (c *userController) create(ctx *contract.Ctx) {
	var req struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Status string `json:"status"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(ctx.R.Body).Decode(&req); err != nil {
		_ = ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	if req.Name == "" || req.Email == "" {
		_ = ctx.JSON(http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Role == "" {
		req.Role = "user"
	}

	now := time.Now()
	u := User{
		ID:        fmt.Sprintf("user-%d", now.UnixNano()),
		Name:      req.Name,
		Email:     req.Email,
		Status:    req.Status,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	created, err := c.repo.Create(u)
	if err != nil {
		_ = ctx.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	_ = ctx.JSON(http.StatusCreated, created)
}

func (c *userController) update(ctx *contract.Ctx) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		_ = ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	var patch userPatch
	if err := json.NewDecoder(ctx.R.Body).Decode(&patch); err != nil {
		_ = ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	updated, err := c.repo.Update(id, patch)
	if err != nil {
		_ = ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	_ = ctx.JSON(http.StatusOK, updated)
}

func (c *userController) delete(ctx *contract.Ctx) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		_ = ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	if !c.repo.Delete(id) {
		_ = ctx.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	ctx.W.WriteHeader(http.StatusNoContent)
}

func main() {
	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithLogger(plog.NewGLogger()),
	)
	if err := app.Use(
		requestid.Middleware(),
		mwtracing.Middleware(nil),
		httpmetrics.Middleware(nil),
		accesslog.Middleware(app.Logger()),
		recovery.Recovery(app.Logger()),
		cors.CORS,
	); err != nil {
		log.Fatal(err)
	}

	repo := NewInMemoryUserRepository()
	ctrl := newUserController(repo)

	app.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprint(w, `<html><body><h1>CRUD Demo</h1><p>Try: <code>curl http://localhost:8080/users</code></p></body></html>`)
	})

	adaptCtx := func(handler contract.CtxHandlerFunc) http.HandlerFunc {
		return contract.AdaptCtxHandler(handler, app.Logger()).ServeHTTP
	}

	app.Get("/users", adaptCtx(ctrl.list))
	app.Get("/users/:id", adaptCtx(ctrl.show))
	app.Post("/users", adaptCtx(ctrl.create))
	app.Put("/users/:id", adaptCtx(ctrl.update))
	app.Delete("/users/:id", adaptCtx(ctrl.delete))

	log.Println("server: http://localhost:8080")
	if err := app.Boot(); err != nil {
		log.Fatal(err)
	}
}

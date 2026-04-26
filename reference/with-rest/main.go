// Example: non-canonical
//
// This demo adds x/rest resource controllers to a service that keeps the
// standard Plumego bootstrap and explicit route registration style.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/x/rest"
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type userRepository struct {
	mu    sync.RWMutex
	users map[string]user
	next  int
}

func newUserRepository() *userRepository {
	return &userRepository{
		users: map[string]user{
			"u_1": {ID: "u_1", Name: "Ada"},
		},
		next: 2,
	}
}

func (r *userRepository) FindAll(_ context.Context, _ *rest.QueryParams) ([]user, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]user, 0, len(r.users))
	for _, item := range r.users {
		users = append(users, item)
	}
	return users, int64(len(users)), nil
}

func (r *userRepository) FindByID(_ context.Context, id string) (*user, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return &item, nil
}

func (r *userRepository) Create(_ context.Context, item *user) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if item.ID == "" {
		item.ID = fmt.Sprintf("u_%d", r.next)
		r.next++
	}
	r.users[item.ID] = *item
	return nil
}

func (r *userRepository) Update(_ context.Context, id string, item *user) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item.ID = id
	r.users[id] = *item
	return nil
}

func (r *userRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.users, id)
	return nil
}

func (r *userRepository) Count(_ context.Context, params *rest.QueryParams) (int64, error) {
	_, total, err := r.FindAll(context.Background(), params)
	return total, err
}

func (r *userRepository) Exists(_ context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.users[id]
	return ok, nil
}

func main() {
	cfg := core.DefaultConfig()
	cfg.Addr = envString("APP_ADDR", ":8084")

	app := core.New(cfg, core.AppDependencies{Logger: plumelog.NewLogger()})
	if err := app.Use(requestid.Middleware(), recovery.Recovery(app.Logger())); err != nil {
		log.Fatalf("register middleware: %v", err)
	}

	if err := app.Get("/api/hello", http.HandlerFunc(hello)); err != nil {
		log.Fatalf("register hello route: %v", err)
	}

	spec := rest.DefaultResourceSpec("users").WithPrefix("/api/users")
	users := rest.NewDBResource[user](spec, newUserRepository())
	if err := app.Get(spec.Prefix, http.HandlerFunc(users.Index)); err != nil {
		log.Fatalf("register user list: %v", err)
	}
	if err := app.Get(spec.Prefix+"/:id", http.HandlerFunc(users.Show)); err != nil {
		log.Fatalf("register user get: %v", err)
	}
	if err := app.Post(spec.Prefix, http.HandlerFunc(users.Create)); err != nil {
		log.Fatalf("register user create: %v", err)
	}

	if err := serve(app, cfg); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{
		"message": "hello from with-rest",
	}, nil)
}

func serve(app *core.App, cfg core.AppConfig) error {
	if err := app.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := app.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}
	defer app.Shutdown(context.Background())

	log.Printf("Starting with-rest demo on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

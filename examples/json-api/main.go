// Example: json-api
//
// Simple JSON API with basic CRUD operations. Demonstrates:
// - Request/response JSON marshaling
// - HTTP status codes
// - Error handling with contract errors
//
// Run:
//   go mod init example.com/json-api
//   go get github.com/spcent/plumego@latest
//   go run main.go
//
// Test:
//   curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"name":"alice","email":"alice@example.com"}'
//   curl http://localhost:8080/users
//   curl http://localhost:8080/users/1
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// User represents a simple user record.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// InMemoryStore is a simple thread-safe in-memory user store.
type InMemoryStore struct {
	mu    sync.RWMutex
	users map[int]User
	nextID int
}

func (s *InMemoryStore) Create(u User) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	u.ID = s.nextID
	s.nextID++
	s.users[u.ID] = u
	return u
}

func (s *InMemoryStore) Get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *InMemoryStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

func main() {
	logger := plumelog.NewLogger()
	cfg := core.DefaultConfig()
	cfg.Addr = ":8080"

	app := core.New(cfg, core.AppDependencies{Logger: logger})
	store := &InMemoryStore{
		users:  make(map[int]User),
		nextID: 1,
	}

	// POST /users — Create a user
	app.Post("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req User
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeBadRequest).
				Message("Invalid request body").
				Build())
			return
		}
		user := store.Create(req)
		_ = contract.WriteResponse(w, r, http.StatusCreated, user, nil)
	}))

	// GET /users — List all users
	app.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users := store.List()
		if users == nil {
			users = []User{}
		}
		_ = contract.WriteResponse(w, r, http.StatusOK, users, nil)
	}))

	// GET /users/:id — Get a user by ID
	app.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeBadRequest).
				Message("Invalid user ID").
				Build())
			return
		}

		user, ok := store.Get(id)
		if !ok {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeNotFound).
				Message("User not found").
				Build())
			return
		}
		_ = contract.WriteResponse(w, r, http.StatusOK, user, nil)
	}))

	if err := app.Prepare(); err != nil {
		log.Fatalf("Failed to prepare app: %v", err)
	}

	server, err := app.Server()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Printf("Server listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

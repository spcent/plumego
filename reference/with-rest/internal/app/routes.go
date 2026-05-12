package app

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// RegisterRoutes wires all HTTP routes for the with-rest demo.
func (a *App) RegisterRoutes() error {
	if err := a.Core.Get("/api/hello", http.HandlerFunc(hello)); err != nil {
		return err
	}

	// x/rest resource: list, get, create
	if err := a.Core.Get("/api/users", http.HandlerFunc(a.Users.Index)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/users/:id", http.HandlerFunc(a.Users.Show)); err != nil {
		return err
	}
	if err := a.Core.Post("/api/users", http.HandlerFunc(a.Users.Create)); err != nil {
		return err
	}
	if err := a.Core.Put("/api/users/:id", http.HandlerFunc(a.Users.Update)); err != nil {
		return err
	}
	if err := a.Core.Delete("/api/users/:id", http.HandlerFunc(a.Users.Delete)); err != nil {
		return err
	}

	return nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{
		"message": "hello from with-rest",
	}, nil)
}

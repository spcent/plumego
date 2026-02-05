package main

import (
	"encoding/json"
	stdlog "log"
	"net/http"

	logpkg "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/bind"
)

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required" mask:"true"`
}

func main() {
	logger := logpkg.NewGLogger()

	mux := http.NewServeMux()
	handler := bind.BindJSON[CreateUserRequest](bind.JSONOptions{
		Logger: logger,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := bind.FromRequest[CreateUserRequest](r)
		resp := map[string]any{
			"email": payload.Email,
			"ok":    true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	mux.Handle("/v1/users", handler)

	chain := middleware.NewChain()
	server := chain.Apply(mux)

	addr := ":8082"
	stdlog.Printf("bind example listening on %s", addr)
	stdlog.Fatal(http.ListenAndServe(addr, server))
}

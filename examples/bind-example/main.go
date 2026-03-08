package main

import (
	"encoding/json"
	"io"
	stdlog "log"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/validator"
)

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required" mask:"true"`
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/users", createUserHandler)

	chain := middleware.NewChain()
	server := chain.Apply(mux)

	addr := ":8082"
	stdlog.Printf("bind example listening on %s", addr)
	stdlog.Fatal(http.ListenAndServe(addr, server))
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		contract.WriteError(w, r, contract.BindErrorToAPIError(contract.ErrInvalidJSON))
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		contract.WriteError(w, r, contract.BindErrorToAPIError(contract.ErrUnexpectedExtraData))
		return
	}

	if err := validator.Validate(&req); err != nil {
		contract.WriteError(w, r, contract.BindErrorToAPIError(err))
		return
	}

	resp := map[string]any{
		"email": req.Email,
		"ok":    true,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

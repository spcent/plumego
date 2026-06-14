// Example: hello
//
// Minimal "hello world" Plumego service.
//
// Run:
//   go mod init example.com/hello
//   go get github.com/spcent/plumego@latest
//   go run main.go
//
// Test:
//   curl http://localhost:8080/ping
package main

import (
	"log"
	"net/http"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

func main() {
	// Create logger and app config
	logger := plumelog.NewLogger()
	cfg := core.DefaultConfig()
	cfg.Addr = ":8080"

	// Create app
	app := core.New(cfg, core.AppDependencies{Logger: logger})

	// Register routes using standard http.HandlerFunc
	app.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{"message": "pong"}, nil)
	}))

	app.Get("/echo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		logger.InfoCtx(r.Context(), "echo_request", plumelog.Fields{"name": name})
		_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{"hello": name}, nil)
	}))

	// Prepare app for serving
	if err := app.Prepare(); err != nil {
		log.Fatalf("Failed to prepare app: %v", err)
	}

	// Start server
	server, err := app.Server()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	log.Printf("Server listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

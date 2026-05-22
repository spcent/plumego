// Example: with-rpc
//
// This demo wires an in-process gRPC server to a plumego HTTP app using
// x/rpc/server (gRPC) and x/rpc/gateway (HTTP transcoding). Follows the
// canonical 4-step bootstrap: load config → build deps → register routes →
// start server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"with-rpc/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a, err := app.New(ctx)
	if err != nil {
		return err
	}
	if err := a.RegisterRoutes(); err != nil {
		return err
	}
	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare http app: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get http server: %w", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/hello", nil)
	srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		return fmt.Errorf("unexpected http status %d: %s", rec.Code, rec.Body.String())
	}
	fmt.Print(rec.Body.String())

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	if err := a.Core.Shutdown(stopCtx); err != nil {
		return fmt.Errorf("shutdown http app: %w", err)
	}
	return a.Stop(stopCtx)
}

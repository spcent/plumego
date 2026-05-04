package core

import (
	"net/http"
	"testing"
)

func TestNewAppStartsMutable(t *testing.T) {
	app := New(DefaultConfig(), AppDependencies{})

	if err := app.Use(func(next http.Handler) http.Handler { return next }); err != nil {
		t.Fatalf("expected new app to accept middleware registration, got %v", err)
	}
	if err := app.Get("/mutable", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})); err != nil {
		t.Fatalf("expected new app to accept route registration, got %v", err)
	}
}

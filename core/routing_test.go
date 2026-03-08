package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPostPutDeletePatch(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		register func(*App, string, http.HandlerFunc)
	}{
		{"GET", "/get", func(app *App, path string, h http.HandlerFunc) { app.Get(path, h) }},
		{"POST", "/post", func(app *App, path string, h http.HandlerFunc) { app.Post(path, h) }},
		{"PUT", "/put", func(app *App, path string, h http.HandlerFunc) { app.Put(path, h) }},
		{"DELETE", "/delete", func(app *App, path string, h http.HandlerFunc) { app.Delete(path, h) }},
		{"PATCH", "/patch", func(app *App, path string, h http.HandlerFunc) { app.Patch(path, h) }},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			app := New()
			called := false
			tt.register(app, tt.path, func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Method != tt.method {
					t.Errorf("expected method %s, got %s", tt.method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			app.ServeHTTP(rr, req)

			if !called {
				t.Error("handler was not called")
			}
			if rr.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rr.Code)
			}
		})
	}
}

func TestAny(t *testing.T) {
	app := New()

	called := false
	app.Any("/any", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		called = false
		req := httptest.NewRequest(method, "/any", nil)
		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		if !called {
			t.Errorf("handler was not called for method %s", method)
		}
		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 for %s, got %d", method, rr.Code)
		}
	}
}

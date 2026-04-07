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
		register func(*App, string, http.HandlerFunc) error
	}{
		{"GET", "/get", func(app *App, path string, h http.HandlerFunc) error { return app.Get(path, h) }},
		{"POST", "/post", func(app *App, path string, h http.HandlerFunc) error { return app.Post(path, h) }},
		{"PUT", "/put", func(app *App, path string, h http.HandlerFunc) error { return app.Put(path, h) }},
		{"DELETE", "/delete", func(app *App, path string, h http.HandlerFunc) error { return app.Delete(path, h) }},
		{"PATCH", "/patch", func(app *App, path string, h http.HandlerFunc) error { return app.Patch(path, h) }},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			app := newTestApp()
			called := false
			mustRegisterRoute(t, tt.register(app, tt.path, func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Method != tt.method {
					t.Errorf("expected method %s, got %s", tt.method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
			}))

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
	app := newTestApp()

	called := false
	mustRegisterRoute(t, app.Any("/any", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})))

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

func TestNamedRouteRegistration(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		route  string
	}{
		{
			name:   "get",
			method: http.MethodGet,
			path:   "/named/get/:id",
			route:  "named.get",
		},
		{
			name:   "post",
			method: http.MethodPost,
			path:   "/named/post/:id",
			route:  "named.post",
		},
		{
			name:   "put",
			method: http.MethodPut,
			path:   "/named/put/:id",
			route:  "named.put",
		},
		{
			name:   "delete",
			method: http.MethodDelete,
			path:   "/named/delete/:id",
			route:  "named.delete",
		},
		{
			name:   "patch",
			method: http.MethodPatch,
			path:   "/named/patch/:id",
			route:  "named.patch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp()
			called := false
			mustRegisterRoute(t, app.AddRouteWithName(tt.method, tt.path, tt.route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			})))

			url := app.URL(tt.route, "id", "42")
			req := httptest.NewRequest(tt.method, url, nil)
			rr := httptest.NewRecorder()
			app.ServeHTTP(rr, req)

			if !called {
				t.Fatalf("handler was not called for route %q", tt.route)
			}
			if rr.Code != http.StatusNoContent {
				t.Fatalf("expected status 204, got %d", rr.Code)
			}
			if url == "" {
				t.Fatalf("expected URL for route %q", tt.route)
			}
		})
	}
}

func TestAddRouteReturnsRegistrationErrors(t *testing.T) {
	app := newTestApp()

	if err := app.AddRoute(http.MethodGet, "/strict", nil); err == nil {
		t.Fatalf("expected nil-handler registration error")
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if err := app.AddRoute(http.MethodGet, "/strict", handler); err != nil {
		t.Fatalf("unexpected add route error: %v", err)
	}

	if err := app.AddRoute(http.MethodGet, "/strict", handler); err == nil {
		t.Fatalf("expected duplicate route error")
	}

	app.preparationState = PreparationStateServerPrepared
	if err := app.AddRoute(http.MethodGet, "/after-start", handler); err == nil {
		t.Fatalf("expected add route to fail after app started")
	}
}

func TestAddRouteWithNameRegistersURLAndReturnsErrors(t *testing.T) {
	app := newTestApp()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := app.AddRouteWithName(http.MethodGet, "/users/:id", "users.show", handler); err != nil {
		t.Fatalf("unexpected add named route error: %v", err)
	}

	if got := app.URL("users.show", "id", "42"); got != "/users/42" {
		t.Fatalf("named route URL = %q, want %q", got, "/users/42")
	}

	if err := app.AddRouteWithName(http.MethodGet, "/users/:id", "users.show", handler); err == nil {
		t.Fatalf("expected duplicate named route error")
	}
}

func TestMethodHelpersReturnRegistrationErrors(t *testing.T) {
	app := newTestApp()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if err := app.Get("/strict-get", handler); err != nil {
		t.Fatalf("unexpected get registration error: %v", err)
	}
	if err := app.Get("/strict-get", handler); err == nil {
		t.Fatalf("expected duplicate get registration error")
	}

	app.preparationState = PreparationStateServerPrepared
	if err := app.Post("/after-start", handler); err == nil {
		t.Fatalf("expected post registration to fail after app started")
	}
}

package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

func TestHandleFunc(t *testing.T) {
	app := New()

	called := false
	app.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	if !called {
		t.Error("handler was not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestHandle(t *testing.T) {
	app := New()

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	app.Handle("/test", handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	if !called {
		t.Error("handler was not called")
	}
}

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

func TestCtxHandlers(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		register func(*App, string, contract.CtxHandlerFunc)
		method   string
	}{
		{"GetCtx", "/get", func(app *App, path string, h contract.CtxHandlerFunc) { app.GetCtx(path, h) }, "GET"},
		{"PostCtx", "/post", func(app *App, path string, h contract.CtxHandlerFunc) { app.PostCtx(path, h) }, "POST"},
		{"PutCtx", "/put", func(app *App, path string, h contract.CtxHandlerFunc) { app.PutCtx(path, h) }, "PUT"},
		{"DeleteCtx", "/delete", func(app *App, path string, h contract.CtxHandlerFunc) { app.DeleteCtx(path, h) }, "DELETE"},
		{"PatchCtx", "/patch", func(app *App, path string, h contract.CtxHandlerFunc) { app.PatchCtx(path, h) }, "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()
			called := false
			tt.register(app, tt.path, func(ctx *contract.Ctx) {
				called = true
				if ctx == nil {
					t.Error("ctx is nil")
				}
				ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
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

func TestAnyCtx(t *testing.T) {
	app := New()

	called := false
	app.AnyCtx("/any", func(ctx *contract.Ctx) {
		called = true
		ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
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

func TestHandlerMethods(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		register func(*App, string, router.Handler)
		method   string
	}{
		{"GetHandler", "/get", func(app *App, path string, h router.Handler) { app.GetHandler(path, h) }, "GET"},
		{"PostHandler", "/post", func(app *App, path string, h router.Handler) { app.PostHandler(path, h) }, "POST"},
		{"PutHandler", "/put", func(app *App, path string, h router.Handler) { app.PutHandler(path, h) }, "PUT"},
		{"DeleteHandler", "/delete", func(app *App, path string, h router.Handler) { app.DeleteHandler(path, h) }, "DELETE"},
		{"PatchHandler", "/patch", func(app *App, path string, h router.Handler) { app.PatchHandler(path, h) }, "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()
			called := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})
			tt.register(app, tt.path, handler)

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

func TestAnyHandler(t *testing.T) {
	app := New()

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	app.AnyHandler("/any", handler)

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

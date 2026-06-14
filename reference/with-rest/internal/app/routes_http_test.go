package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"with-rest/internal/config"
)

// newHTTPTestApp builds a fully-wired app and returns it ready to serve requests.
// Each call returns a fresh in-memory store seeded with {id:"u_1", name:"Ada"}.
func newHTTPTestApp(t *testing.T) *App {
	t.Helper()
	cfg := config.Defaults()
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	return a
}

// do sends req through the app handler and returns the recorder.
func do(a *App, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	a.Core.ServeHTTP(rec, req)
	return rec
}

// TestUsersListSuccess verifies that GET /api/users returns 200 with the seeded
// user inside a paginated envelope.
func TestUsersListSuccess(t *testing.T) {
	a := newHTTPTestApp(t)
	rec := do(a, httptest.NewRequest(http.MethodGet, "/api/users", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	// contract.WriteResponse places the slice in "data" and pagination in "meta".
	var body struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
		Meta struct {
			Pagination struct {
				TotalItems int64 `json:"total_items"`
			} `json:"pagination"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v; raw = %s", err, rec.Body.String())
	}
	if body.Meta.Pagination.TotalItems < 1 {
		t.Fatalf("total_items = %d, want >= 1", body.Meta.Pagination.TotalItems)
	}
	found := false
	for _, u := range body.Data {
		if u.ID == "u_1" && u.Name == "Ada" {
			found = true
		}
	}
	if !found {
		t.Fatalf("seeded user u_1/Ada not found in list; data = %+v", body.Data)
	}
}

// TestUsersGetByIDSuccess verifies that GET /api/users/u_1 returns 200 with the
// correct user record.
func TestUsersGetByIDSuccess(t *testing.T) {
	a := newHTTPTestApp(t)
	rec := do(a, httptest.NewRequest(http.MethodGet, "/api/users/u_1", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v; raw = %s", err, rec.Body.String())
	}
	if body.Data.ID != "u_1" || body.Data.Name != "Ada" {
		t.Fatalf("got {id:%s name:%s}, want {id:u_1 name:Ada}", body.Data.ID, body.Data.Name)
	}
}

// TestUsersGetByIDNotFound verifies that GET /api/users/<unknown> returns 404.
func TestUsersGetByIDNotFound(t *testing.T) {
	a := newHTTPTestApp(t)
	rec := do(a, httptest.NewRequest(http.MethodGet, "/api/users/nonexistent", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v; raw = %s", err, rec.Body.String())
	}
	if body.Error.Type != "resource_not_found" {
		t.Fatalf("error.type = %q, want resource_not_found", body.Error.Type)
	}
}

// TestUsersCreateSuccess verifies that POST /api/users with a valid name returns
// 201 with the created user including an assigned ID.
func TestUsersCreateSuccess(t *testing.T) {
	a := newHTTPTestApp(t)
	body := strings.NewReader(`{"name":"Bob"}`)
	rec := do(a, httptest.NewRequest(http.MethodPost, "/api/users", body))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v; raw = %s", err, rec.Body.String())
	}
	if resp.Data.Name != "Bob" {
		t.Fatalf("name = %q, want Bob", resp.Data.Name)
	}
	if resp.Data.ID == "" {
		t.Fatal("created user has empty ID")
	}
}

// TestUsersCreateInvalidJSON verifies that POST /api/users with malformed JSON
// returns 400.
func TestUsersCreateInvalidJSON(t *testing.T) {
	a := newHTTPTestApp(t)
	rec := do(a, httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"name":`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

// TestUsersUpdateSuccess verifies that PUT /api/users/u_1 returns 200 with the
// updated user.
func TestUsersUpdateSuccess(t *testing.T) {
	a := newHTTPTestApp(t)
	body := strings.NewReader(`{"id":"u_1","name":"Ada Renamed"}`)
	rec := do(a, httptest.NewRequest(http.MethodPut, "/api/users/u_1", body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v; raw = %s", err, rec.Body.String())
	}
	if resp.Data.Name != "Ada Renamed" {
		t.Fatalf("name = %q, want Ada Renamed", resp.Data.Name)
	}
}

// TestUsersUpdateNotFound verifies that PUT /api/users/<unknown> returns 404.
func TestUsersUpdateNotFound(t *testing.T) {
	a := newHTTPTestApp(t)
	body := strings.NewReader(`{"name":"Ghost"}`)
	rec := do(a, httptest.NewRequest(http.MethodPut, "/api/users/nonexistent", body))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", rec.Code, rec.Body.String())
	}
}

// TestUsersDeleteSuccess verifies that DELETE /api/users/u_1 returns 204.
func TestUsersDeleteSuccess(t *testing.T) {
	a := newHTTPTestApp(t)
	rec := do(a, httptest.NewRequest(http.MethodDelete, "/api/users/u_1", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body = %s", rec.Code, rec.Body.String())
	}
}

// TestUsersDeleteNotFound verifies that DELETE /api/users/<unknown> returns 404.
func TestUsersDeleteNotFound(t *testing.T) {
	a := newHTTPTestApp(t)
	rec := do(a, httptest.NewRequest(http.MethodDelete, "/api/users/nonexistent", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", rec.Code, rec.Body.String())
	}
}

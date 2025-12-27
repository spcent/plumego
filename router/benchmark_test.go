package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

// createTestRouter builds a router with static and parameterized routes
func createTestRouter() *Router {
	r := NewRouter()

	// Static routes
	r.AddRoute(GET, "/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("about"))
	}))
	r.AddRoute(GET, "/contact", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("contact"))
	}))

	// Parameterized routes
	r.AddRoute(GET, "/hello/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, _ := contract.Param(r, "name")
		w.Write([]byte("Hello " + name))
	}))
	r.AddRoute(GET, "/users/:id/books/:bookId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		bookID, _ := contract.Param(r, "bookId")
		w.Write([]byte("User " + id + " Book " + bookID))
	}))

	return r
}

// BenchmarkStaticRoute tests a static route (/about)
func BenchmarkStaticRoute(b *testing.B) {
	router := createTestRouter()
	req := httptest.NewRequest("GET", "/about", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkParamRoute tests a route with one parameter (/hello/:name)
func BenchmarkParamRoute(b *testing.B) {
	router := createTestRouter()
	req := httptest.NewRequest("GET", "/hello/Alice", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkNestedParamRoute tests a route with multiple parameters (/users/:id/books/:bookId)
func BenchmarkNestedParamRoute(b *testing.B) {
	router := createTestRouter()
	req := httptest.NewRequest("GET", "/users/123/books/456", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

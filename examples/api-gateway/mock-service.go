package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// MockService is a simple HTTP service for testing the API gateway
func main() {
	port := "8081"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	serviceName := "mock-service"
	if len(os.Args) > 2 {
		serviceName = os.Args[2]
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"service": serviceName,
			"port":    port,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// List endpoint (GET)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": 1, "name": "Alice", "email": "alice@example.com"},
				{"id": 2, "name": "Bob", "email": "bob@example.com"},
			},
			"service": serviceName,
			"port":    port,
		})
	})

	// Get single user (GET)
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      123,
			"name":    "John Doe",
			"email":   "john@example.com",
			"service": serviceName,
			"port":    port,
		})
	})

	// Orders endpoint
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orders": []map[string]interface{}{
				{"id": 1, "product": "Laptop", "price": 999.99},
				{"id": 2, "product": "Mouse", "price": 29.99},
			},
			"service": serviceName,
			"port":    port,
		})
	})

	// Products endpoint
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"products": []map[string]interface{}{
				{"id": 1, "name": "Laptop", "price": 999.99},
				{"id": 2, "name": "Mouse", "price": 29.99},
				{"id": 3, "name": "Keyboard", "price": 79.99},
			},
			"service": serviceName,
			"port":    port,
		})
	})

	log.Printf("Starting %s on port %s", serviceName, port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/spcent/plumego"
	"github.com/spcent/plumego/store/db"
)

// APIHandler handles tenant-scoped business API requests
type APIHandler struct {
	db *db.TenantDB
}

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// RequestAnalytics represents request analytics data
type RequestAnalytics struct {
	TotalRequests  int            `json:"total_requests"`
	ByStatus       map[string]int `json:"by_status"`
	AvgDurationMS  float64        `json:"avg_duration_ms"`
	RecentRequests []RequestLog   `json:"recent_requests"`
}

// RequestLog represents a single request log entry
type RequestLog struct {
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	DurationMS int       `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListUsers retrieves all users for the current tenant
func (h *APIHandler) ListUsers(ctx *plumego.Context) {
	start := time.Now()
	defer h.logRequest(ctx, start)

	// Query is automatically filtered by tenant_id
	rows, err := h.db.QueryFromContext(
		ctx.R.Context(),
		"SELECT id, tenant_id, email, name, created_at FROM users WHERE 1=1",
	)
	if err != nil {
		log.Printf("Failed to query users: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve users",
		})
		return
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
			log.Printf("Failed to scan user: %v", err)
			continue
		}
		users = append(users, user)
	}

	ctx.JSON(http.StatusOK, map[string]any
		"users": users,
		"count": len(users),
	})
}

// CreateUser creates a new user for the current tenant
func (h *APIHandler) CreateUser(ctx *plumego.Context) {
	start := time.Now()
	defer h.logRequest(ctx, start)

	var req CreateUserRequest
	if err := json.NewDecoder(ctx.R.Body).Decode(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	if req.Email == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "email is required",
		})
		return
	}

	// Insert is automatically scoped to tenant_id from context
	result, err := h.db.ExecFromContext(
		ctx.R.Context(),
		"INSERT INTO users (email, name) VALUES (?, ?)",
		req.Email, req.Name,
	)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
		return
	}

	userID, _ := result.LastInsertId()

	user := User{
		ID:        int(userID),
		TenantID:  plumego.TenantIDFromContext(ctx.R.Context()),
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	log.Printf("Created user %d for tenant %s", userID, user.TenantID)

	ctx.JSON(http.StatusCreated, user)
}

// GetUser retrieves a specific user by ID (tenant-scoped)
func (h *APIHandler) GetUser(ctx *plumego.Context) {
	start := time.Now()
	defer h.logRequest(ctx, start)

	userID, _ := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
		return
	}

	// Query is automatically filtered by tenant_id
	var user User
	err := h.db.QueryRowFromContext(
		ctx.R.Context(),
		"SELECT id, tenant_id, email, name, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.TenantID, &user.Email, &user.Name, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "User not found",
			})
			return
		}
		log.Printf("Failed to get user: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve user",
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user (tenant-scoped)
func (h *APIHandler) DeleteUser(ctx *plumego.Context) {
	start := time.Now()
	defer h.logRequest(ctx, start)

	userID, _ := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
		return
	}

	// Delete is automatically scoped to tenant_id
	result, err := h.db.ExecFromContext(
		ctx.R.Context(),
		"DELETE FROM users WHERE id = ?",
		userID,
	)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete user",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
		return
	}

	tenantID := plumego.TenantIDFromContext(ctx.R.Context())
	log.Printf("Deleted user %s for tenant %s", userID, tenantID)

	ctx.JSON(http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

// GetRequestAnalytics retrieves analytics data for the current tenant
func (h *APIHandler) GetRequestAnalytics(ctx *plumego.Context) {
	start := time.Now()
	defer h.logRequest(ctx, start)

	tenantID := plumego.TenantIDFromContext(ctx.R.Context())

	// Get total requests
	var totalRequests int
	err := h.db.QueryRowFromContext(
		ctx.R.Context(),
		"SELECT COUNT(*) FROM api_requests WHERE 1=1",
	).Scan(&totalRequests)
	if err != nil {
		log.Printf("Failed to get total requests: %v", err)
		totalRequests = 0
	}

	// Get average duration
	var avgDuration sql.NullFloat64
	err = h.db.QueryRowFromContext(
		ctx.R.Context(),
		"SELECT AVG(duration_ms) FROM api_requests WHERE 1=1",
	).Scan(&avgDuration)
	if err != nil {
		log.Printf("Failed to get avg duration: %v", err)
	}

	// Get recent requests
	rows, err := h.db.QueryFromContext(
		ctx.R.Context(),
		`SELECT method, path, status_code, duration_ms, created_at
		 FROM api_requests
		 WHERE 1=1
		 ORDER BY created_at DESC
		 LIMIT 10`,
	)

	recentRequests := make([]RequestLog, 0)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var req RequestLog
			if err := rows.Scan(&req.Method, &req.Path, &req.StatusCode, &req.DurationMS, &req.CreatedAt); err == nil {
				recentRequests = append(recentRequests, req)
			}
		}
	}

	analytics := RequestAnalytics{
		TotalRequests:  totalRequests,
		AvgDurationMS:  avgDuration.Float64,
		RecentRequests: recentRequests,
		ByStatus:       make(map[string]int),
	}

	// Get status code distribution
	statusRows, err := h.db.QueryFromContext(
		ctx.R.Context(),
		"SELECT status_code, COUNT(*) FROM api_requests WHERE 1=1 GROUP BY status_code",
	)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var statusCode, count int
			if err := statusRows.Scan(&statusCode, &count); err == nil {
				analytics.ByStatus[strconv.Itoa(statusCode)] = count
			}
		}
	}

	log.Printf("Retrieved analytics for tenant %s: %d total requests", tenantID, totalRequests)

	ctx.JSON(http.StatusOK, analytics)
}

// logRequest logs the API request to the database
func (h *APIHandler) logRequest(ctx *plumego.Context, start time.Time) {
	duration := time.Since(start)
	tenantID := plumego.TenantIDFromContext(ctx.R.Context())

	// Log to database (use RawDB to bypass tenant filtering for logging)
	// Note: We use a simple status code of 200 for successful responses
	// In a production system, you would track the actual status code via middleware
	_, err := h.db.RawDB().Exec(
		`INSERT INTO api_requests (tenant_id, method, path, status_code, duration_ms)
		 VALUES (?, ?, ?, ?, ?)`,
		tenantID,
		ctx.R.Method,
		ctx.R.URL.Path,
		200, // Simplified - track via middleware in production
		duration.Milliseconds(),
	)
	if err != nil {
		log.Printf("Failed to log request: %v", err)
	}
}

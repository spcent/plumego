package websocket

import (
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
)

type routeRegistrar interface {
	AddRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error
}

// WebSocketConfig defines the configuration for WebSocket.
type WebSocketConfig struct {
	WorkerCount        int           // Number of worker goroutines
	JobQueueSize       int           // Size of the job queue
	SendQueueSize      int           // Size of the send queue per connection
	SendTimeout        time.Duration // Timeout for sending messages
	SendBehavior       SendBehavior  // Behavior when queue is full or timeout occurs
	Secret             []byte        // Secret key for JWT authentication
	WSRoutePath        string        // Path for WebSocket connection
	BroadcastPath      string        // Path for broadcasting messages
	BroadcastEnabled   bool          // Enable broadcast endpoint when true
	MaxConnections     int           // Maximum total connections (0 = unlimited)
	MaxRoomConnections int           // Maximum connections per room (0 = unlimited)
}

const (
	// DefaultSendQueueSize is the default buffer size for the WebSocket send queue.
	DefaultSendQueueSize = 256
)

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	secret := []byte(os.Getenv("WS_SECRET"))

	return WebSocketConfig{
		WorkerCount:        16,
		JobQueueSize:       4096,
		SendQueueSize:      DefaultSendQueueSize,
		SendTimeout:        200 * time.Millisecond,
		SendBehavior:       SendBlock,
		Secret:             secret,
		WSRoutePath:        "/ws",
		BroadcastPath:      "/_admin/broadcast",
		BroadcastEnabled:   true,
		MaxConnections:     0,
		MaxRoomConnections: 0,
	}
}

type Server struct {
	config WebSocketConfig
	debug  bool
	logger log.StructuredLogger
	hub    *Hub

	routesOnce sync.Once
}

const minWebSocketSecretLen = 32

func New(cfg WebSocketConfig, debug bool, logger log.StructuredLogger) (*Server, error) {
	if len(cfg.Secret) < minWebSocketSecretLen {
		return nil, fmt.Errorf(
			"websocket secret must be at least %d bytes (set the WS_SECRET environment variable or pass Secret via WebSocketConfig)",
			minWebSocketSecretLen,
		)
	}

	hub := NewHubWithConfig(HubConfig{
		WorkerCount:        cfg.WorkerCount,
		JobQueueSize:       cfg.JobQueueSize,
		MaxConnections:     cfg.MaxConnections,
		MaxRoomConnections: cfg.MaxRoomConnections,
	})

	return &Server{
		config: cfg,
		debug:  debug,
		logger: logger,
		hub:    hub,
	}, nil
}

func (c *Server) RegisterRoutes(r routeRegistrar) error {
	var regErr error
	c.routesOnce.Do(func() {
		wsAuth := NewSimpleRoomAuth(c.config.Secret)

		regErr = r.AddRoute(http.MethodGet, c.config.WSRoutePath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ServeWSWithAuth(w, r, c.hub, wsAuth, c.config.SendQueueSize,
				c.config.SendTimeout, c.config.SendBehavior)
		}))
		if regErr != nil {
			return
		}

		if c.config.BroadcastEnabled && c.config.BroadcastPath != "" {
			regErr = r.AddRoute(http.MethodPost, c.config.BroadcastPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					_ = contract.WriteError(w, r, contract.NewErrorBuilder().Status(http.StatusMethodNotAllowed).Code("method_not_allowed").Message("POST only").Category(contract.CategoryClient).Build())
					return
				}
				// Always require authentication for broadcast endpoint.
				// Debug mode should only affect logging verbosity, never security checks.
				const bearerPrefix = "bearer "
				rawAuth := r.Header.Get("Authorization")
				var provided []byte
				if strings.HasPrefix(strings.ToLower(rawAuth), bearerPrefix) {
					provided = []byte(strings.TrimSpace(rawAuth[len("Bearer "):]))
				}
				// Note: Query parameter secrets are no longer supported for security reasons.
				// Secrets in URLs can be leaked via server logs and Referer headers.

				if len(provided) == 0 || subtle.ConstantTimeCompare(provided, c.config.Secret) != 1 {
					_ = contract.WriteError(w, r, contract.NewErrorBuilder().
						Status(http.StatusUnauthorized).
						Category(contract.CategoryAuth).
						Type(contract.TypeUnauthorized).
						Code(contract.CodeUnauthorized).
						Message("unauthorized").
						Build())
					return
				}

				b, err := io.ReadAll(r.Body)
				if err != nil {
					_ = contract.WriteError(w, r, contract.NewErrorBuilder().Status(http.StatusInternalServerError).Code("read_body_failed").Message("Error reading request body").Build())
					return
				}

				// Optional ?room= parameter targets a specific room; omit for all-room broadcast.
				if room := r.URL.Query().Get("room"); room != "" {
					c.hub.BroadcastRoom(room, OpcodeText, b)
				} else {
					c.hub.BroadcastAll(OpcodeText, b)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
		}
	})
	return regErr
}

func (c *Server) Shutdown(ctx context.Context) error {
	if c.hub != nil {
		err := c.hub.Shutdown(ctx)
		c.hub = nil
		return err
	}
	return nil
}

func (c *Server) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"broadcastEnabled": c.config.BroadcastEnabled}}

	if c.hub == nil {
		status.Status = health.StatusUnhealthy
		status.Message = "hub not initialized"
	}

	return "websocket", status
}

// Hub exposes the underlying WebSocket hub for advanced usage.
func (c *Server) Hub() *Hub { return c.hub }

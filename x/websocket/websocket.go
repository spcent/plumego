package websocket

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
)

type routeRegistrar interface {
	AddRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error
}

// BroadcastAuthorizer authorizes admin broadcast requests.
type BroadcastAuthorizer func(*http.Request) bool

// WebSocketConfig defines the configuration for WebSocket.
type WebSocketConfig struct {
	WorkerCount          int           // Number of worker goroutines
	JobQueueSize         int           // Size of the job queue
	SendQueueSize        int           // Size of the send queue per connection
	SendTimeout          time.Duration // Timeout for sending messages
	SendBehavior         SendBehavior  // Behavior when queue is full or timeout occurs
	Secret               []byte        // Secret key for JWT authentication
	WSRoutePath          string        // Path for WebSocket connection
	BroadcastPath        string        // Path for broadcasting messages
	BroadcastEnabled     bool          // Enable broadcast endpoint when true
	BroadcastSecret      []byte        // Dedicated bearer secret for admin broadcast.
	BroadcastAuthorizer  BroadcastAuthorizer
	BroadcastMaxBytes    int64    // Maximum admin broadcast request body size. 0 means DefaultBroadcastMaxBytes.
	AllowedOrigins       []string // Browser origins allowed to connect. Empty rejects requests with Origin.
	AllowAllOrigins      bool     // Explicitly disable origin checks.
	AllowUnauthenticated bool     // Explicitly allow websocket connections without JWT.
	AllowQueryToken      bool     // Explicitly allow ?token= JWT transport for trusted clients.
	MaxConnections       int      // Maximum total connections (0 = unlimited)
	MaxRoomConnections   int      // Maximum connections per room (0 = unlimited)
}

const (
	// DefaultSendQueueSize is the default buffer size for the WebSocket send queue.
	DefaultSendQueueSize = 256
	// DefaultBroadcastMaxBytes is the default max body size for the admin broadcast endpoint.
	DefaultBroadcastMaxBytes = 1 << 20
)

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		WorkerCount:        16,
		JobQueueSize:       4096,
		SendQueueSize:      DefaultSendQueueSize,
		SendTimeout:        200 * time.Millisecond,
		SendBehavior:       SendBlock,
		WSRoutePath:        "/ws",
		BroadcastPath:      "/_admin/broadcast",
		BroadcastEnabled:   false,
		BroadcastMaxBytes:  DefaultBroadcastMaxBytes,
		MaxConnections:     0,
		MaxRoomConnections: 0,
	}
}

type Server struct {
	config WebSocketConfig
	debug  bool
	logger log.StructuredLogger
	hub    *Hub
}

const minWebSocketSecretLen = 32

func New(cfg WebSocketConfig, debug bool, logger log.StructuredLogger) (*Server, error) {
	if len(cfg.Secret) < minWebSocketSecretLen {
		return nil, fmt.Errorf(
			"websocket secret must be at least %d bytes (pass Secret via WebSocketConfig)",
			minWebSocketSecretLen,
		)
	}
	if cfg.BroadcastMaxBytes < 0 {
		return nil, fmt.Errorf("websocket broadcast max bytes cannot be negative")
	}
	if cfg.BroadcastMaxBytes == 0 {
		cfg.BroadcastMaxBytes = DefaultBroadcastMaxBytes
	}
	if cfg.BroadcastEnabled && cfg.BroadcastAuthorizer == nil {
		if len(cfg.BroadcastSecret) < minWebSocketSecretLen {
			return nil, fmt.Errorf("websocket broadcast secret must be at least %d bytes or BroadcastAuthorizer must be provided", minWebSocketSecretLen)
		}
		if string(cfg.BroadcastSecret) == string(cfg.Secret) {
			return nil, fmt.Errorf("websocket broadcast secret must be separate from websocket JWT secret")
		}
	}

	hub, err := NewHubWithConfigE(HubConfig{
		WorkerCount:        cfg.WorkerCount,
		JobQueueSize:       cfg.JobQueueSize,
		MaxConnections:     cfg.MaxConnections,
		MaxRoomConnections: cfg.MaxRoomConnections,
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		config: cfg,
		debug:  debug,
		logger: logger,
		hub:    hub,
	}, nil
}

func (c *Server) RegisterRoutes(r routeRegistrar) error {
	if c == nil {
		return fmt.Errorf("%w: websocket server is nil", ErrInvalidConfig)
	}
	if r == nil {
		return fmt.Errorf("%w: websocket route registrar is nil", ErrInvalidConfig)
	}
	if c.hub == nil {
		return ErrNilHub
	}
	if strings.TrimSpace(c.config.WSRoutePath) == "" {
		return fmt.Errorf("%w: websocket route path is empty", ErrInvalidConfig)
	}
	if c.config.BroadcastEnabled && strings.TrimSpace(c.config.BroadcastPath) == "" {
		return fmt.Errorf("%w: websocket broadcast path is empty", ErrInvalidConfig)
	}

	wsAuth, err := NewSimpleRoomAuth(c.config.Secret)
	if err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodGet, c.config.WSRoutePath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWSWithConfig(w, r, ServerConfig{
			Hub:                  c.hub,
			Auth:                 wsAuth,
			QueueSize:            c.config.SendQueueSize,
			SendTimeout:          c.config.SendTimeout,
			SendBehavior:         c.config.SendBehavior,
			AllowedOrigins:       c.config.AllowedOrigins,
			AllowAllOrigins:      c.config.AllowAllOrigins,
			AllowUnauthenticated: c.config.AllowUnauthenticated,
			AllowQueryToken:      c.config.AllowQueryToken,
		})
	})); err != nil {
		return err
	}

	if c.config.BroadcastEnabled && c.config.BroadcastPath != "" {
		if err := r.AddRoute(http.MethodPost, c.config.BroadcastPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !c.authorizeBroadcast(r) {
				_ = contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeUnauthorized).
					Code(contract.CodeUnauthorized).
					Message("unauthorized").
					Build())
				return
			}

			b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, c.config.BroadcastMaxBytes))
			if err != nil {
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					_ = contract.WriteError(w, r, contract.NewErrorBuilder().
						Type(contract.TypeOutOfRange).
						Status(http.StatusRequestEntityTooLarge).
						Code(codeWebSocketRequestTooLarge).
						Message("request body too large").
						Build())
					return
				}
				_ = contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeInternal).
					Code(codeWebSocketRequestReadFailure).
					Message("error reading request body").
					Build())
				return
			}

			// Optional ?room= parameter targets a specific room; omit for all-room broadcast.
			if room := r.URL.Query().Get("room"); room != "" {
				c.hub.BroadcastRoom(room, OpcodeText, b)
			} else {
				c.hub.BroadcastAll(OpcodeText, b)
			}
			w.WriteHeader(http.StatusNoContent)
		})); err != nil {
			return err
		}
	}

	return nil
}

func (c *Server) authorizeBroadcast(r *http.Request) bool {
	if c.config.BroadcastAuthorizer != nil {
		return c.config.BroadcastAuthorizer(r)
	}
	provided := bearerToken(r.Header.Get("Authorization"))
	if len(provided) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare(provided, c.config.BroadcastSecret) == 1
}

func bearerToken(rawAuth string) []byte {
	const bearerPrefix = "bearer "
	if !strings.HasPrefix(strings.ToLower(rawAuth), bearerPrefix) {
		return nil
	}
	token := strings.TrimSpace(rawAuth[len(bearerPrefix):])
	if token == "" {
		return nil
	}
	return []byte(token)
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

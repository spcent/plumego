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
	"github.com/spcent/plumego/router"
)

type routeRegistrar interface {
	AddRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error
}

// WebSocketConfig defines the configuration for WebSocket.
type WebSocketConfig struct {
	WorkerCount           int           // Number of worker goroutines
	JobQueueSize          int           // Size of the job queue
	SendQueueSize         int           // Size of the send queue per connection
	SendTimeout           time.Duration // Timeout for sending messages
	SendBehavior          SendBehavior  // Behavior when queue is full or timeout occurs
	Secret                []byte        // Secret key for JWT authentication
	RoomAuth              RoomAuthorizer
	TokenAuth             TokenAuthenticator
	AllowUnauthenticated  bool
	AllowQueryToken       bool
	AllowedOrigins        []string
	WSRoutePath           string // Path for WebSocket connection
	BroadcastPath         string // Path for broadcasting messages
	BroadcastEnabled      bool   // Enable broadcast endpoint when true
	BroadcastSecret       []byte // Secret token for admin broadcast endpoint
	BroadcastMaxBodyBytes int64  // Maximum admin broadcast body bytes (0 = default)
	MaxRoomRegistrations  int    // Maximum room registrations (0 = unlimited)
	MaxRoomConnections    int    // Maximum connections per room (0 = unlimited)
	OnMessage             MessageHandler
}

const (
	// DefaultSendQueueSize is the default buffer size for the WebSocket send queue.
	DefaultSendQueueSize = 256

	defaultBroadcastMaxBodyBytes = 1 << 20
)

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		WorkerCount:           16,
		JobQueueSize:          4096,
		SendQueueSize:         DefaultSendQueueSize,
		SendTimeout:           200 * time.Millisecond,
		SendBehavior:          SendBlock,
		WSRoutePath:           "/ws",
		BroadcastPath:         "/_admin/broadcast",
		BroadcastEnabled:      false,
		BroadcastMaxBodyBytes: defaultBroadcastMaxBodyBytes,
		MaxRoomRegistrations:  0,
		MaxRoomConnections:    0,
	}
}

type Server struct {
	config WebSocketConfig
	hub    *Hub
}

func New(cfg WebSocketConfig) (*Server, error) {
	if len(cfg.Secret) > 0 {
		if err := validateJWTSecret(cfg.Secret, minJWTSecretLength); err != nil {
			return nil, fmt.Errorf("%w (read WS_SECRET in application code and pass it via WebSocketConfig.Secret)", err)
		}
	}
	if cfg.TokenAuth == nil && !cfg.AllowUnauthenticated && len(cfg.Secret) == 0 {
		return nil, ErrNilTokenAuthorizer
	}
	if cfg.BroadcastEnabled && len(cfg.BroadcastSecret) == 0 {
		return nil, ErrEmptyBroadcastToken
	}
	if len(cfg.BroadcastSecret) > 0 && len(cfg.BroadcastSecret) < minJWTSecretLength {
		return nil, fmt.Errorf("%w: minimum %d bytes required", ErrEmptyBroadcastToken, minJWTSecretLength)
	}
	if cfg.BroadcastMaxBodyBytes < 0 {
		return nil, fmt.Errorf("%w: broadcast max body bytes cannot be negative", ErrInvalidConfig)
	}
	if cfg.BroadcastMaxBodyBytes == 0 {
		cfg.BroadcastMaxBodyBytes = defaultBroadcastMaxBodyBytes
	}

	hub, err := NewHubWithConfigE(HubConfig{
		WorkerCount:          cfg.WorkerCount,
		JobQueueSize:         cfg.JobQueueSize,
		MaxRoomRegistrations: cfg.MaxRoomRegistrations,
		MaxRoomConnections:   cfg.MaxRoomConnections,
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		config: cfg,
		hub:    hub,
	}, nil
}

func (c *Server) RegisterRoutes(r routeRegistrar) error {
	if r == nil {
		return ErrNilRegistrar
	}
	if c == nil || c.hub == nil {
		return ErrNilHub
	}
	if c.config.WSRoutePath == "" {
		return ErrEmptyRoutePath
	}

	roomAuth := c.config.RoomAuth
	if roomAuth == nil {
		roomAuth = NewSimpleRoomAuth()
	}
	tokenAuth := c.config.TokenAuth
	if tokenAuth == nil && len(c.config.Secret) > 0 {
		var err error
		tokenAuth, err = NewHS256TokenAuth(c.config.Secret)
		if err != nil {
			return err
		}
	}
	serverCfg := ServerConfig{
		Hub:                  c.hub,
		RoomAuth:             roomAuth,
		TokenAuth:            tokenAuth,
		AllowUnauthenticated: c.config.AllowUnauthenticated,
		AllowQueryToken:      c.config.AllowQueryToken,
		QueueSize:            c.config.SendQueueSize,
		SendTimeout:          c.config.SendTimeout,
		SendBehavior:         c.config.SendBehavior,
		AllowedOrigins:       c.config.AllowedOrigins,
		OnMessage:            c.config.OnMessage,
	}

	if err := r.AddRoute(http.MethodGet, c.config.WSRoutePath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serverCfg.OnMessage == nil {
			ServeRoomFanoutWS(w, r, serverCfg)
			return
		}
		ServeWSWithConfig(w, r, serverCfg)
	})); err != nil {
		return err
	}

	if c.config.BroadcastEnabled {
		if c.config.BroadcastPath == "" {
			return ErrEmptyRoutePath
		}
		if len(c.config.BroadcastSecret) == 0 {
			return ErrEmptyBroadcastToken
		}
		if err := r.AddRoute(http.MethodPost, c.config.BroadcastPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			if len(provided) == 0 || subtle.ConstantTimeCompare(provided, c.config.BroadcastSecret) != 1 {
				_ = contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeUnauthorized).
					Code(contract.CodeUnauthorized).
					Message("unauthorized").
					Build())
				return
			}

			body := http.MaxBytesReader(w, r.Body, c.config.BroadcastMaxBodyBytes)
			b, err := io.ReadAll(body)
			if err != nil {
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					_ = contract.WriteError(w, r, contract.NewErrorBuilder().
						Type(contract.TypeInvalidFormat).
						Status(http.StatusRequestEntityTooLarge).
						Code(contract.CodeRequestBodyTooLarge).
						Message("broadcast body too large").
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
	} else if c.hub.stopped.Load() {
		status.Status = health.StatusUnhealthy
		status.Message = "hub stopped"
	}

	return "websocket", status
}

// Hub exposes the underlying WebSocket hub for advanced usage.
func (c *Server) Hub() *Hub { return c.hub }

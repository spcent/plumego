package core

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

	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	ws "github.com/spcent/plumego/net/websocket"
	"github.com/spcent/plumego/router"
)

// WebSocketConfig defines the configuration for WebSocket.
type WebSocketConfig struct {
	WorkerCount        int             // Number of worker goroutines
	JobQueueSize       int             // Size of the job queue
	SendQueueSize      int             // Size of the send queue per connection
	SendTimeout        time.Duration   // Timeout for sending messages
	SendBehavior       ws.SendBehavior // Behavior when queue is full or timeout occurs
	Secret             []byte          // Secret key for JWT authentication
	WSRoutePath        string          // Path for WebSocket connection
	BroadcastPath      string          // Path for broadcasting messages
	BroadcastEnabled   bool            // Enable broadcast endpoint when true
	MaxConnections     int             // Maximum total connections (0 = unlimited)
	MaxRoomConnections int             // Maximum connections per room (0 = unlimited)
}

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	secret := []byte(os.Getenv("WS_SECRET"))

	return WebSocketConfig{
		WorkerCount:        16,
		JobQueueSize:       4096,
		SendQueueSize:      256,
		SendTimeout:        200 * time.Millisecond,
		SendBehavior:       ws.SendBlock,
		Secret:             secret,
		WSRoutePath:        "/ws",
		BroadcastPath:      "/_admin/broadcast",
		BroadcastEnabled:   true,
		MaxConnections:     0,
		MaxRoomConnections: 0,
	}
}

type webSocketComponent struct {
	BaseComponent
	config WebSocketConfig
	debug  bool
	logger log.StructuredLogger
	hub    *ws.Hub

	routesOnce sync.Once
}

const minWebSocketSecretLen = 32

func newWebSocketComponent(cfg WebSocketConfig, debug bool, logger log.StructuredLogger) (*webSocketComponent, error) {
	if len(cfg.Secret) < minWebSocketSecretLen {
		return nil, fmt.Errorf("websocket secret must be at least %d bytes", minWebSocketSecretLen)
	}

	hub := ws.NewHubWithConfig(ws.HubConfig{
		WorkerCount:        cfg.WorkerCount,
		JobQueueSize:       cfg.JobQueueSize,
		MaxConnections:     cfg.MaxConnections,
		MaxRoomConnections: cfg.MaxRoomConnections,
	})

	return &webSocketComponent{
		config: cfg,
		debug:  debug,
		logger: logger,
		hub:    hub,
	}, nil
}

func (c *webSocketComponent) RegisterRoutes(r *router.Router) {
	c.routesOnce.Do(func() {
		wsAuth := ws.NewSimpleRoomAuth(c.config.Secret)

		r.GetFunc(c.config.WSRoutePath, func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWSWithAuth(w, r, c.hub, wsAuth, c.config.SendQueueSize,
				c.config.SendTimeout, c.config.SendBehavior)
		})

		if c.config.BroadcastEnabled && c.config.BroadcastPath != "" {
			r.PostFunc(c.config.BroadcastPath, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					writeHTTPError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
					return
				}
				if !c.debug {
					const bearerPrefix = "bearer "
					rawAuth := r.Header.Get("Authorization")
					var provided []byte
					if strings.HasPrefix(strings.ToLower(rawAuth), bearerPrefix) {
						provided = []byte(strings.TrimSpace(rawAuth[len("Bearer "):]))
					} else if q := r.URL.Query().Get("secret"); q != "" {
						provided = []byte(q)
					}

					if len(provided) == 0 || subtle.ConstantTimeCompare(provided, c.config.Secret) != 1 {
						writeHTTPError(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
						return
					}
				}

				b, err := io.ReadAll(r.Body)
				if err != nil {
					writeHTTPError(w, r, http.StatusInternalServerError, "read_body_failed", "Error reading request body")
					return
				}

				c.hub.BroadcastAll(ws.OpcodeText, b)
				w.WriteHeader(http.StatusNoContent)
			})
		}
	})
}

func (c *webSocketComponent) RegisterMiddleware(_ *middleware.Registry) {}

func (c *webSocketComponent) Start(_ context.Context) error { return nil }

func (c *webSocketComponent) Stop(_ context.Context) error {
	if c.hub != nil {
		c.hub.Stop()
		c.hub = nil
	}
	return nil
}

func (c *webSocketComponent) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"broadcastEnabled": c.config.BroadcastEnabled}}

	if c.hub == nil {
		status.Status = health.StatusUnhealthy
		status.Message = "hub not initialized"
	}

	return "websocket", status
}

// ConfigureWebSocket configures WebSocket support for the app.
// It returns the Hub for advanced usage.
func (a *App) ConfigureWebSocket() (*ws.Hub, error) {
	if err := a.loadEnv(); err != nil {
		return nil, err
	}

	return a.ConfigureWebSocketWithOptions(DefaultWebSocketConfig())
}

// ConfigureWebSocketWithOptions configures WebSocket support with custom options.
func (a *App) ConfigureWebSocketWithOptions(config WebSocketConfig) (*ws.Hub, error) {
	if err := a.ensureMutable("configure_websocket", "configure websocket"); err != nil {
		return nil, err
	}

	cfg := a.configSnapshot()
	a.mu.RLock()
	logger := a.logger
	a.mu.RUnlock()

	comp, err := newWebSocketComponent(config, cfg.Debug, logger)
	if err != nil {
		return nil, err
	}

	comp.RegisterRoutes(a.ensureRouter())

	a.mu.Lock()
	a.components = append(a.components, comp)
	a.mu.Unlock()

	return comp.hub, nil
}

// NewWebSocketComponent builds a pluggable WebSocket component so examples can
// compose it via core.WithComponent.
func NewWebSocketComponent(config WebSocketConfig, logger log.StructuredLogger, debug bool) (Component, *ws.Hub, error) {
	if logger == nil {
		logger = log.NewGLogger()
	}

	comp, err := newWebSocketComponent(config, debug, logger)
	if err != nil {
		return nil, nil, err
	}

	return comp, comp.hub, nil
}

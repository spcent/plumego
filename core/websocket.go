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

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	ws "github.com/spcent/plumego/net/websocket"
	"github.com/spcent/plumego/router"
)

// WebSocketConfig defines the configuration for WebSocket.
type WebSocketConfig struct {
	WorkerCount      int             // Number of worker goroutines
	JobQueueSize     int             // Size of the job queue
	SendQueueSize    int             // Size of the send queue per connection
	SendTimeout      time.Duration   // Timeout for sending messages
	SendBehavior     ws.SendBehavior // Behavior when queue is full or timeout occurs
	Secret           []byte          // Secret key for JWT authentication
	WSRoutePath      string          // Path for WebSocket connection
	BroadcastPath    string          // Path for broadcasting messages
	BroadcastEnabled bool            // Enable broadcast endpoint when true
}

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	secret := []byte(os.Getenv("WS_SECRET"))

	return WebSocketConfig{
		WorkerCount:      16,
		JobQueueSize:     4096,
		SendQueueSize:    256,
		SendTimeout:      200 * time.Millisecond,
		SendBehavior:     ws.SendBlock,
		Secret:           secret,
		WSRoutePath:      "/ws",
		BroadcastPath:    "/_admin/broadcast",
		BroadcastEnabled: true,
	}
}

type webSocketComponent struct {
	config WebSocketConfig
	debug  bool
	logger log.StructuredLogger
	hub    *ws.Hub

	routesOnce sync.Once
}

func newWebSocketComponent(cfg WebSocketConfig, debug bool, logger log.StructuredLogger) (*webSocketComponent, error) {
	if len(cfg.Secret) == 0 {
		return nil, fmt.Errorf("websocket secret is required")
	}

	hub := ws.NewHub(cfg.WorkerCount, cfg.JobQueueSize)

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
					http.Error(w, "POST only", http.StatusMethodNotAllowed)
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
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
				}

				b, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Error reading request body", http.StatusInternalServerError)
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

func (c *webSocketComponent) Health() (string, any) {
	return "websocket", map[string]any{"enabled": true}
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
	comp, err := newWebSocketComponent(config, a.config.Debug, a.logger)
	if err != nil {
		return nil, err
	}

	comp.RegisterRoutes(a.router)
	a.wsHub = comp.hub
	a.components = append(a.components, comp)

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

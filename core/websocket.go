package core

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	ws "github.com/spcent/plumego/net/websocket"
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
	if len(config.Secret) == 0 {
		return nil, fmt.Errorf("websocket secret is required")
	}

	hub := ws.NewHub(config.WorkerCount, config.JobQueueSize)
	a.wsHub = hub
	wsAuth := ws.NewSimpleRoomAuth(config.Secret)

	a.Router().GetFunc(config.WSRoutePath, func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWSWithAuth(w, r, hub, wsAuth, config.SendQueueSize,
			config.SendTimeout, config.SendBehavior)
	})

	if config.BroadcastEnabled && config.BroadcastPath != "" {
		a.Router().PostFunc(config.BroadcastPath, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "POST only", http.StatusMethodNotAllowed)
				return
			}
			if !a.config.Debug {
				const bearerPrefix = "bearer "
				rawAuth := r.Header.Get("Authorization")
				var provided []byte
				if strings.HasPrefix(strings.ToLower(rawAuth), bearerPrefix) {
					provided = []byte(strings.TrimSpace(rawAuth[len("Bearer "):]))
				} else if q := r.URL.Query().Get("secret"); q != "" {
					provided = []byte(q)
				}

				if len(provided) == 0 || subtle.ConstantTimeCompare(provided, config.Secret) != 1 {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
			}

			b, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}

			hub.BroadcastAll(ws.OpcodeText, b)
			w.WriteHeader(http.StatusNoContent)
		})
	}

	return hub, nil
}

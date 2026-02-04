package core

import (
	"github.com/spcent/plumego/core/components/websocket"
	log "github.com/spcent/plumego/log"
	ws "github.com/spcent/plumego/net/websocket"
)

type WebSocketConfig = websocket.WebSocketConfig

const DefaultSendQueueSize = websocket.DefaultSendQueueSize

func DefaultWebSocketConfig() WebSocketConfig {
	return websocket.DefaultWebSocketConfig()
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

	comp, err := websocket.NewComponent(config, cfg.Debug, logger)
	if err != nil {
		return nil, err
	}

	comp.RegisterRoutes(a.ensureRouter())

	a.mu.Lock()
	a.components = append(a.components, comp)
	a.mu.Unlock()

	return comp.Hub(), nil
}

// NewWebSocketComponent builds a pluggable WebSocket component so examples can
// compose it via core.WithComponent.
func NewWebSocketComponent(config WebSocketConfig, logger log.StructuredLogger, debug bool) (Component, *ws.Hub, error) {
	if logger == nil {
		logger = log.NewGLogger()
	}

	comp, err := websocket.NewComponent(config, debug, logger)
	if err != nil {
		return nil, nil, err
	}

	return comp, comp.Hub(), nil
}

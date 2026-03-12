package websocket

import (
	"net"
	"net/http"
	"time"

	ws "github.com/spcent/plumego/net/websocket"
)

type (
	Hub               = ws.Hub
	HubConfig         = ws.HubConfig
	HubMetrics        = ws.HubMetrics
	SendBehavior      = ws.SendBehavior
	Conn              = ws.Conn
	Outbound          = ws.Outbound
	UserInfo          = ws.UserInfo
	SecurityConfig    = ws.SecurityConfig
	SecurityMetrics   = ws.SecurityMetrics
	SecurityEvent     = ws.SecurityEvent
	RoomAuthenticator = ws.RoomAuthenticator
	SimpleRoomAuth    = ws.SimpleRoomAuth
	SecureRoomAuth    = ws.SecureRoomAuth
	ServerConfig      = ws.ServerConfig
)

const (
	SendDrop  = ws.SendDrop
	SendBlock = ws.SendBlock
	SendClose = ws.SendClose

	OpcodeText   = ws.OpcodeText
	OpcodeBinary = ws.OpcodeBinary
)

func NewHub(workerCount int, jobQueueSize int) *Hub {
	return ws.NewHub(workerCount, jobQueueSize)
}

func NewHubWithConfig(cfg HubConfig) *Hub {
	return ws.NewHubWithConfig(cfg)
}

func NewConn(c net.Conn, queueSize int, sendTimeout time.Duration, behavior SendBehavior) *Conn {
	return ws.NewConn(c, queueSize, sendTimeout, behavior)
}

func NewSimpleRoomAuth(secret []byte) *SimpleRoomAuth {
	return ws.NewSimpleRoomAuth(secret)
}

func NewSecureRoomAuth(secret []byte, cfg SecurityConfig) (*SecureRoomAuth, error) {
	return ws.NewSecureRoomAuth(secret, cfg)
}

func ServeWSWithAuth(w http.ResponseWriter, r *http.Request, hub *Hub, auth RoomAuthenticator, queueSize int, sendTimeout time.Duration, behavior SendBehavior) {
	ws.ServeWSWithAuth(w, r, hub, auth, queueSize, sendTimeout, behavior)
}

func ServeWSWithConfig(w http.ResponseWriter, r *http.Request, cfg ServerConfig) {
	ws.ServeWSWithConfig(w, r, cfg)
}

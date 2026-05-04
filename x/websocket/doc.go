// Package websocket provides experimental WebSocket server helpers with
// room-based broadcasting.
//
// This package implements a WebSocket hub featuring:
//   - Room-based message broadcasting
//   - Explicit room authorization and token authentication hooks
//   - Connection lifecycle management
//   - Worker pool for concurrent message delivery
//   - Room-registration and per-room limits
//   - Metrics collection and monitoring
//   - Security event tracking
//
// The hub manages WebSocket connections organized into rooms, allowing efficient
// message broadcasting to specific groups of clients. Connections are
// authenticated or accepted anonymously according to ServerConfig.
//
// Example usage:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	// Create a WebSocket hub
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{
//		WorkerCount:              4,
//		JobQueueSize:             1024,
//		MaxRoomRegistrations:     10000,
//		MaxRoomConnections:       1000,
//	})
//	if err != nil {
//		// handle configuration error
//	}
//	defer hub.Stop()
//
//	auth := websocket.NewSimpleRoomAuth()
//
//	// Serve a room fanout WebSocket endpoint with auth and origin checks.
//	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
//		websocket.ServeRoomFanoutWS(w, r, websocket.ServerConfig{
//			Hub:                  hub,
//			RoomAuth:             auth,
//			AllowUnauthenticated: true,
//			QueueSize:            256,
//			SendTimeout:          200 * time.Millisecond,
//			SendBehavior:         websocket.SendBlock,
//			AllowedOrigins:       []string{"https://app.example.com"},
//		})
//	})
//
//	// Broadcast to a room
//	hub.BroadcastRoom("chat:general", websocket.OpcodeText, []byte("Hello"))
package websocket

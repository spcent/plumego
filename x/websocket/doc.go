// Package websocket provides a production-ready WebSocket server with room-based broadcasting.
//
// This package implements a high-performance WebSocket hub featuring:
//   - Room-based message broadcasting
//   - JWT authentication for secure connections
//   - Connection lifecycle management
//   - Worker pool for concurrent message delivery
//   - Connection limits (total and per-room)
//   - Metrics collection and monitoring
//   - Security event tracking
//
// The hub manages WebSocket connections organized into rooms, allowing efficient
// message broadcasting to specific groups of clients. All connections are authenticated
// using JWT tokens.
//
// Example usage:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	// Create a WebSocket hub
//	hub := websocket.NewHubWithConfig(websocket.HubConfig{
//		WorkerCount:        4,
//		JobQueueSize:       1024,
//		MaxConnections:     10000,
//		MaxRoomConnections: 1000,
//	})
//	defer hub.Stop()
//
//	auth := websocket.NewSimpleRoomAuth([]byte("your-32-byte-secret"))
//
//	// Serve a WebSocket endpoint with auth and origin checks.
//	// Inside the package handler, successful connections are registered to rooms.
//	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
//		websocket.ServeWSWithConfig(w, r, websocket.ServerConfig{
//			Hub:            hub,
//			Auth:           auth,
//			QueueSize:      256,
//			SendTimeout:    200 * time.Millisecond,
//			SendBehavior:   websocket.SendBlock,
//			AllowedOrigins: []string{"https://app.example.com"},
//		})
//	})
//
//	// Broadcast to a room
//	hub.BroadcastRoom("chat:general", websocket.OpcodeText, []byte("Hello"))
package websocket

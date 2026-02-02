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
//	import "github.com/spcent/plumego/net/websocket"
//
//	// Create a WebSocket hub
//	hub := websocket.NewHub(websocket.HubConfig{
//		Workers:            4,
//		JobQueueSize:       1024,
//		MaxConnections:     10000,
//		MaxRoomConnections: 1000,
//	})
//	hub.Start()
//	defer hub.Stop()
//
//	// Upgrade HTTP connection to WebSocket
//	conn, err := websocket.Upgrade(w, r, jwtSecret)
//	if err != nil {
//		// Upgrade failed
//	}
//
//	// Join a room and broadcast messages
//	conn.JoinRoom("chat:general")
//	hub.BroadcastToRoom("chat:general", []byte("Hello"))
package websocket

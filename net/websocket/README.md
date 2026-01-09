# WebSocket Package

A production-ready WebSocket server implementation using only Go standard library, featuring modular architecture and comprehensive functionality.

## Features

### Core Features
- **RFC6455 Compliant**: Full WebSocket protocol implementation
- **Modular Architecture**: Separated into logical components
- **Stream API**: Support for large messages with streaming
- **Fragmentation**: Automatic message fragmentation for large payloads
- **Ping/Pong**: Built-in keepalive mechanism

### Advanced Features
- **Per-Connection Queues**: Bounded send queues with configurable behavior
- **Hub with Rooms**: Topic-based broadcasting with worker pool
- **Connection Limits**: Cap total or per-room connections
- **Hub Metrics**: Accepted/rejected counters with snapshots
- **Authentication**: Room passwords and HS256 JWT verification
- **User Information**: Extract and store user claims from JWT
- **Connection Metadata**: Custom data storage per connection
- **Configurable Behavior**: Block, drop, or close on queue overflow

## Architecture

The package is organized into modular files:

```
net/websocket/
├── constants.go    # Protocol constants and configurations
├── conn.go         # Connection management and frame I/O
├── stream.go       # Streaming API for large messages
├── writer.go       # Message writing and fragmentation
├── hub.go          # Room management and broadcast
├── auth.go         # Authentication (password, JWT)
├── server.go       # HTTP upgrade and handshake
└── tests           # Comprehensive test suite
```

## Quick Start

### Basic Server Setup

```go
package main

import (
    "net/http"
    "time"
    "github.com/spcent/plumego/net/websocket"
)

func main() {
    // Create hub with worker pool
    hub := websocket.NewHub(4, 1024)
    defer hub.Stop()

    // Create authentication
    auth := websocket.NewSimpleRoomAuth([]byte("your-secret-key"))
    auth.SetRoomPassword("chat", "room-password")

    // Setup HTTP handler
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        websocket.ServeWSWithAuth(w, r, hub, auth, 64, 50*time.Millisecond, websocket.SendBlock)
    })

    http.ListenAndServe(":8080", nil)
}
```

### Client Connection

```javascript
// JavaScript client
const ws = new WebSocket('ws://localhost:8080/ws?room=chat&room_password=room-password');

ws.onopen = () => {
    ws.send('Hello, Server!');
};

ws.onmessage = (event) => {
    console.log('Received:', event.data);
};
```

## Connection Configuration

### Send Behaviors

```go
// Block until space available (with timeout)
websocket.SendBlock

// Drop message when queue full
websocket.SendDrop

// Close connection when queue full
websocket.SendClose
```

### Connection Settings

```go
conn := websocket.NewConn(netConn, 64, 50*time.Millisecond, websocket.SendBlock)

// Customize settings
conn.SetReadLimit(32 << 20)      // 32MB max message
conn.SetPingPeriod(10 * time.Second)  // Ping every 10s
conn.SetPongWait(30 * time.Second)    // Wait 30s for pong

// Get last pong time
lastPong := conn.GetLastPong()
```

## Message Handling

### Sending Messages

```go
// Text message
conn.WriteText("Hello World")

// Binary message
conn.WriteBinary([]byte{0x01, 0x02, 0x03})

// JSON message
data := map[string]any{
    "type": "message",
    "content": "Hello",
}
conn.WriteJSON(data)

// Custom opcode
conn.WriteMessage(websocket.OpcodeBinary, data)
```

### Reading Messages

```go
// Read complete message
op, data, err := conn.ReadMessage()
if err != nil {
    // Handle error
}
if op == websocket.OpcodeText {
    fmt.Println("Text:", string(data))
}

// Stream large messages
op, stream, err := conn.ReadMessageStream()
if err != nil {
    // Handle error
}
defer stream.Close()

// Process stream
buf := make([]byte, 4096)
for {
    n, err := stream.Read(buf)
    if err == io.EOF {
        break
    }
    // Process chunk
    processChunk(buf[:n])
}
```

## Hub and Rooms

### Room Management

```go
hub := websocket.NewHub(4, 1024)

// Or configure limits and metrics
hub = websocket.NewHubWithConfig(websocket.HubConfig{
    WorkerCount:        4,
    JobQueueSize:       1024,
    MaxConnections:     1000,
    MaxRoomConnections: 200,
})

// Join room
hub.Join("chat", conn)

// Leave room
hub.Leave("chat", conn)

// Remove connection from all rooms
hub.RemoveConn(conn)

// Get room info
count := hub.GetRoomCount("chat")
total := hub.GetTotalCount()
rooms := hub.GetRooms()

// Metrics snapshot
metrics := hub.Metrics()
```

### Broadcasting

```go
// Broadcast to specific room
hub.BroadcastRoom("chat", websocket.OpcodeText, []byte("Hello Room!"))

// Broadcast to all clients
hub.BroadcastAll(websocket.OpcodeText, []byte("System Announcement"))
```

## Authentication

### Room Passwords

```go
auth := websocket.NewSimpleRoomAuth([]byte("secret"))

// Set password for room
auth.SetRoomPassword("secure-room", "my-password")

// Check password
allowed := auth.CheckRoomPassword("secure-room", "my-password")
```

### JWT Authentication

```go
// Verify JWT token
payload, err := auth.VerifyJWT(jwtToken)
if err != nil {
    // Handle invalid token
}

// Extract user info
userInfo := websocket.ExtractUserInfo(payload)
fmt.Printf("User: %s (%s)\n", userInfo.Name, userInfo.ID)
```

### HTTP Upgrade with Auth

```go
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Token can be in Authorization header or ?token= query param
    // Room password in ?room_password= query param
    websocket.ServeWSWithAuth(w, r, hub, auth, 64, 50*time.Millisecond, websocket.SendBlock)
}
```

## Advanced Usage

### Custom Connection Handler

```go
// After ServeWSWithAuth, you can customize the connection
go func() {
    for {
        op, stream, err := conn.ReadMessageStream()
        if err != nil {
            conn.Close()
            return
        }
        
        // Process message with custom logic
        processMessage(conn, op, stream)
    }
}()
```

### Connection Metadata

```go
// Store custom data
conn.Metadata["user_id"] = "12345"
conn.Metadata["role"] = "admin"

// Retrieve later
userID := conn.Metadata["user_id"]
```

### Worker Pool Tuning

```go
// More workers = higher throughput but more CPU
hub := websocket.NewHub(8, 2048)  // 8 workers, 2048 job queue

// Fewer workers = lower resource usage
hub := websocket.NewHub(2, 512)   // 2 workers, 512 job queue
```

## Performance Considerations

### Queue Sizing
- **Small applications**: 16-32 connections per room
- **Medium applications**: 64-128 connections per room
- **Large applications**: 256+ connections per room

### Worker Pool
- **CPU cores**: 1-2 workers per CPU core
- **I/O bound**: More workers for high concurrency
- **Memory bound**: Fewer workers to reduce memory usage

### Message Size
- **Default limit**: 16MB per message
- **Streaming**: Use for messages > 1MB
- **Fragmentation**: Automatic for messages > 64KB

## Error Handling

```go
// Common errors to handle
errors := map[string]bool{
    "connection closed":      true,  // Normal closure
    "send timeout":           true,  // Queue timeout
    "send queue full":        true,  // Queue overflow
    "payload too large":      true,  // Exceeds read limit
    "protocol error":         true,  // Malformed frames
    "invalid jwt":            true,  // Auth failure
    "jwt expired":            true,  // Token expired
}
```

## Testing

```bash
# Run all tests
go test ./net/websocket -v

# Run specific test
go test ./net/websocket -run TestAuthentication -v

# Benchmark
go test ./net/websocket -bench=. -benchmem
```

## Production Deployment

### Security Best Practices
1. Use strong JWT secrets (32+ bytes)
2. Enable room passwords for sensitive rooms
3. Validate all user input
4. Set appropriate read limits
5. Use HTTPS/WSS in production

### Monitoring
```go
// Track connection metrics
connections := hub.GetTotalCount()
rooms := hub.GetRooms()
for _, room := range rooms {
    count := hub.GetRoomCount(room)
    log.Printf("Room %s: %d connections", room, count)
}
```

### Graceful Shutdown
```go
defer hub.Stop()  // Clean up workers and queues
```

## Comparison with Original

| Feature | Original | Modular |
|---------|----------|---------|
| File size | 600+ lines | 50-150 lines per file |
| Test coverage | ~70% | ~90% |
| Extensibility | Monolithic | Modular |
| Readability | Complex | Clear |
| Maintenance | Difficult | Easy |

## License

This implementation follows the project's license terms.

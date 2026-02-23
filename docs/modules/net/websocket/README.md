# WebSocket Hub

> **Package**: `github.com/spcent/plumego/net/websocket` | **Feature**: Real-time connections

## Overview

The `websocket` package provides a WebSocket hub with JWT authentication, channel-based broadcasting, connection management, and security features.

## Quick Start

```go
import "github.com/spcent/plumego/net/websocket"

// Create hub
hub := websocket.NewHub(
    websocket.WithJWTSecret(os.Getenv("WS_SECRET")),
    websocket.WithMaxConnections(10000),
    websocket.WithPingInterval(30*time.Second),
)
go hub.Run()

// Register WebSocket endpoint
app.Get("/ws", hub.Handler())

// Broadcast to all connections
hub.Broadcast("channel", []byte(`{"type":"update","data":{}}`))

// Send to specific user
hub.SendToUser("user-123", []byte(`{"type":"notification","message":"Hello!"}`))
```

## Authentication

WebSocket connections are authenticated via JWT tokens passed as query parameters:

```
ws://localhost:8080/ws?token=<jwt-token>
```

```go
hub := websocket.NewHub(
    websocket.WithJWTSecret(os.Getenv("WS_SECRET")),
    websocket.WithTokenParam("token"),  // Default: "token"
    websocket.WithTokenValidator(func(token string) (*Claims, error) {
        return jwtManager.Verify(token, jwt.TokenTypeAccess)
    }),
)
```

**Client-side**:
```javascript
const token = await getAuthToken();
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = () => console.log('Connected');
ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    handleMessage(msg);
};
ws.onclose = (event) => {
    console.log('Disconnected:', event.code, event.reason);
    // Reconnect logic
    setTimeout(() => connect(), 3000);
};
```

## Channels

Channels enable topic-based message routing:

```go
// Client subscribes to channels by sending a message
// {"type": "subscribe", "channels": ["room-123", "notifications"]}

// Server broadcasts to channel
hub.Broadcast("room-123", []byte(`{"type":"message","content":"Hello room!"}`))
hub.Broadcast("notifications", []byte(`{"type":"notification","count":5}`))
```

### Channel Management

```go
// Subscribe a connection to a channel
hub.Subscribe(connID, "room-123")

// Unsubscribe
hub.Unsubscribe(connID, "room-123")

// List connections in a channel
connections := hub.ChannelConnections("room-123")
fmt.Printf("Users in room: %d\n", len(connections))
```

## Connection Handlers

```go
hub := websocket.NewHub(
    websocket.WithJWTSecret(secret),

    // Called when client connects
    websocket.OnConnect(func(conn *websocket.Conn) {
        log.Infof("Client connected: userID=%s", conn.UserID)

        // Send welcome message
        conn.Send([]byte(`{"type":"connected","message":"Welcome!"}`))

        // Auto-subscribe to user's notification channel
        hub.Subscribe(conn.ID, "user:"+conn.UserID)

        // Track in presence system
        presence.SetOnline(conn.UserID)
    }),

    // Called when client disconnects
    websocket.OnDisconnect(func(conn *websocket.Conn, err error) {
        log.Infof("Client disconnected: userID=%s", conn.UserID)
        presence.SetOffline(conn.UserID)
    }),

    // Called for each message received from client
    websocket.OnMessage(func(conn *websocket.Conn, msg []byte) {
        var event struct {
            Type string `json:"type"`
        }
        json.Unmarshal(msg, &event)

        switch event.Type {
        case "subscribe":
            handleSubscribe(hub, conn, msg)
        case "ping":
            conn.Send([]byte(`{"type":"pong"}`))
        case "message":
            handleMessage(hub, conn, msg)
        }
    }),
)
```

## Targeted Messaging

```go
// Send to specific user (all their connections)
hub.SendToUser("user-123", payload)

// Send to specific connection
hub.SendToConn(connID, payload)

// Broadcast to channel (all subscribers)
hub.Broadcast("room-123", payload)

// Broadcast to all connections
hub.BroadcastAll(payload)

// Broadcast excluding sender
hub.BroadcastExclude("room-123", senderConnID, payload)
```

## Connection Info

```go
// Get connection details
conn, err := hub.GetConn(connID)
if err == nil {
    conn.ID          // Unique connection ID
    conn.UserID      // Authenticated user ID
    conn.TenantID    // Tenant (if multi-tenant)
    conn.Channels    // Subscribed channels
    conn.RemoteAddr  // Client IP address
    conn.ConnectedAt // Connection timestamp
    conn.Metadata    // Custom metadata
}

// List all active connections
connections := hub.Connections()
fmt.Printf("Active connections: %d\n", len(connections))

// Check if user is online
online := hub.IsUserOnline("user-123")
```

## Chat Room Example

```go
type ChatMessage struct {
    Type    string `json:"type"`
    Room    string `json:"room"`
    Content string `json:"content"`
    UserID  string `json:"user_id"`
}

hub := websocket.NewHub(
    websocket.WithJWTSecret(secret),
    websocket.OnMessage(func(conn *websocket.Conn, msg []byte) {
        var chatMsg ChatMessage
        json.Unmarshal(msg, &chatMsg)
        chatMsg.UserID = conn.UserID

        if chatMsg.Type == "join" {
            hub.Subscribe(conn.ID, "room:"+chatMsg.Room)
            // Notify room of new user
            notification, _ := json.Marshal(map[string]string{
                "type":   "user_joined",
                "user":   conn.UserID,
                "room":   chatMsg.Room,
            })
            hub.Broadcast("room:"+chatMsg.Room, notification)
        } else if chatMsg.Type == "message" {
            // Broadcast chat message to room
            response, _ := json.Marshal(chatMsg)
            hub.Broadcast("room:"+chatMsg.Room, response)

            // Persist to database
            chatStore.Save(chatMsg)
        }
    }),
)
```

## Security Configuration

```go
hub := websocket.NewHub(
    websocket.WithJWTSecret(secret),

    // Allowed origins (CORS for WebSocket)
    websocket.WithAllowedOrigins([]string{
        "https://app.example.com",
        "https://www.example.com",
    }),

    // Rate limiting per connection
    websocket.WithMessageRateLimit(10, time.Second), // 10 msgs/sec

    // Max message size
    websocket.WithMaxMessageSize(64*1024), // 64KB

    // Connection limits
    websocket.WithMaxConnections(10000),
    websocket.WithMaxConnectionsPerUser(5),

    // Timeouts
    websocket.WithReadTimeout(60*time.Second),
    websocket.WithWriteTimeout(10*time.Second),
    websocket.WithPingInterval(30*time.Second),
)
```

## Metrics

```go
// Active connections
websocket_connections_total
websocket_connections_active

// Messages
websocket_messages_received_total
websocket_messages_sent_total

// Errors
websocket_auth_failures_total
websocket_disconnections_total{reason="timeout|error|normal"}
```

## Related Documentation

- [Net Overview](../README.md) — Network module
- [Security: JWT](../../security/jwt.md) — Token verification
- [AI: SSE](../../ai/sse.md) — Alternative for one-way streaming

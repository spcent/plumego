package websocket

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
)

// computeAcceptKey computes the WebSocket accept key
func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + guid))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// headerContains checks if a header contains a value (case-insensitive)
func headerContains(h http.Header, key, val string) bool {
	v := h.Get(key)
	if v == "" {
		return false
	}
	parts := strings.Split(v, ",")
	for _, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), val) {
			return true
		}
	}
	return false
}

// isOriginAllowed checks if the request origin is in the allowed list
// Returns true if allowedOrigins is nil/empty (skip validation) or contains "*" or the specific origin
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	// Skip validation if no origins configured
	if len(allowedOrigins) == 0 {
		return true
	}

	// Check for wildcard
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
	}
	return false
}

// ServerConfig configures WebSocket server options
type ServerConfig struct {
	Hub            *Hub
	Auth           *simpleRoomAuth
	QueueSize      int
	SendTimeout    time.Duration
	SendBehavior   SendBehavior
	AllowedOrigins []string // Allowed origins for CORS, use ["*"] to allow all
}

// ServeWSWithAuth does the handshake and basic auth checks before accepting connection.
// Accepts token from Authorization header "Bearer <token>" or ?token= in query.
// room password from ?room_password= param.
// onConn will be called when Conn is ready.
func ServeWSWithAuth(w http.ResponseWriter, r *http.Request, hub *Hub, auth *simpleRoomAuth, queueSize int, sendTimeout time.Duration, behavior SendBehavior) {
	ServeWSWithConfig(w, r, ServerConfig{
		Hub:            hub,
		Auth:           auth,
		QueueSize:      queueSize,
		SendTimeout:    sendTimeout,
		SendBehavior:   behavior,
		AllowedOrigins: nil, // No origin validation for backward compatibility
	})
}

// ServeWSWithConfig does the handshake with full configuration options including origin validation.
// Accepts token from Authorization header "Bearer <token>" or ?token= in query.
// room password from ?room_password= param.
// onConn will be called when Conn is ready.
func ServeWSWithConfig(w http.ResponseWriter, r *http.Request, cfg ServerConfig) {
	// Origin validation (CSRF protection)
	origin := r.Header.Get("Origin")
	if origin != "" && !isOriginAllowed(origin, cfg.AllowedOrigins) {
		metricsMutex.Lock()
		securityMetrics.RejectedConnections++
		metricsMutex.Unlock()
		contract.WriteError(w, r, contract.NewForbiddenError("forbidden origin"))
		return
	}

	// Basic HTTP validation first
	if r.Method != http.MethodGet {
		contract.WriteError(w, r, contract.APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed", Category: contract.CategoryClient})
		return
	}
	if !headerContains(r.Header, "Connection", "Upgrade") || !headerContains(r.Header, "Upgrade", "websocket") {
		contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: "bad request", Category: contract.CategoryClient})
		return
	}

	// Validate WebSocket key
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: "bad request", Category: contract.CategoryClient})
		return
	}
	if err := ValidateWebSocketKey(key); err != nil {
		metricsMutex.Lock()
		securityMetrics.InvalidWebSocketKeys++
		metricsMutex.Unlock()
		contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: err.Error(), Category: contract.CategoryClient})
		return
	}
	// auth: room and JWT
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}
	// check room password
	roomPwd := r.URL.Query().Get("room_password")
	if !cfg.Auth.CheckRoomPassword(room, roomPwd) {
		metricsMutex.Lock()
		securityMetrics.RejectedConnections++
		metricsMutex.Unlock()
		contract.WriteError(w, r, contract.NewForbiddenError("forbidden: bad room password"))
		return
	}
	if err := cfg.Hub.CanJoin(room); err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, ErrRoomFull) {
			status = http.StatusTooManyRequests
		}
		contract.WriteError(w, r, contract.APIError{Status: status, Code: "JOIN_DENIED", Message: err.Error(), Category: contract.CategoryClient})
		return
	}
	// check token if present
	token := ""
	var userInfo *UserInfo
	if ah := r.Header.Get("Authorization"); ah != "" && strings.HasPrefix(strings.ToLower(ah), "bearer ") {
		token = strings.TrimSpace(ah[len("bearer "):])
	} else if t := r.URL.Query().Get("token"); t != "" {
		token = t
	}
	if token != "" {
		payload, err := cfg.Auth.VerifyJWT(token)
		if err != nil {
			metricsMutex.Lock()
			securityMetrics.RejectedConnections++
			metricsMutex.Unlock()
			contract.WriteError(w, r, contract.NewForbiddenError("forbidden: invalid token"))
			return
		}
		// Extract user information from JWT payload
		userInfo = ExtractUserInfo(payload)
		metricsMutex.Lock()
		securityMetrics.SuccessfulAuthentications++
		metricsMutex.Unlock()
	}
	accept := computeAcceptKey(key)
	hj, ok := w.(http.Hijacker)
	if !ok {
		contract.WriteError(w, r, contract.NewInternalError("server does not support hijacking"))
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		contract.WriteError(w, r, contract.NewInternalError("hijack failed"))
		return
	}
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n" +
		"\r\n"
	if _, err := buf.WriteString(resp); err != nil {
		conn.Close()
		return
	}
	if err := buf.Flush(); err != nil {
		conn.Close()
		return
	}

	c := NewConn(conn, cfg.QueueSize, cfg.SendTimeout, cfg.SendBehavior)
	// Set user information if authenticated
	c.UserInfo = userInfo
	// override br/bw with hijacked conn
	c.br = bufio.NewReaderSize(conn, 8192)
	c.bw = bufio.NewWriterSize(conn, 8192)

	// register in hub
	if err := cfg.Hub.TryJoin(room, c); err != nil {
		conn.Close()
		return
	}
	// cleanup on close
	go func() {
		<-c.closeC
		cfg.Hub.Leave(room, c)
		cfg.Hub.RemoveConn(c)
	}()
	// spawn a goroutine to read frames and push to room broadcast on completed stream
	go func() {
		for {
			op, rstream, err := c.ReadMessageStream()
			if err != nil {
				if err != io.EOF {
					log.Println("ReadMessageStream error:", err)
				}
				c.Close()
				return
			}
			// For demonstration: stream copy to a buffer but streaming-friendly.
			// If message is huge, we stream to temp file or process chunk-by-chunk.
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, rstream) // streaming
			_ = rstream.Close()
			// broadcast to room
			cfg.Hub.BroadcastRoom(room, op, buf.Bytes())
		}
	}()
}

// UpgradeClient performs client-side WebSocket upgrade
// This is useful for testing or when you need to create a client connection
func UpgradeClient(url string, header http.Header) (net.Conn, *bufio.ReadWriter, error) {
	// This is a simplified client upgrade - in production you'd want more features
	return nil, nil, errors.New("client upgrade not implemented")
}
